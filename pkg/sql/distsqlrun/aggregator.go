// Copyright 2016 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package distsqlrun

import (
	"context"
	"strings"
	"sync"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/types"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stringarena"
	"github.com/pkg/errors"
)

// GetAggregateInfo returns the aggregate constructor and the return type for
// the given aggregate function when applied on the given type.
func GetAggregateInfo(
	fn AggregatorSpec_Func, inputTypes ...sqlbase.ColumnType,
) (
	aggregateConstructor func(*tree.EvalContext) tree.AggregateFunc,
	returnType sqlbase.ColumnType,
	err error,
) {
	if fn == AggregatorSpec_IDENT {
		if len(inputTypes) != 1 {
			return nil, sqlbase.ColumnType{}, errors.Errorf("ident aggregate needs 1 input")
		}
		return builtins.NewIdentAggregate, inputTypes[0], nil
	}

	datumTypes := make([]types.T, len(inputTypes))
	for i := range inputTypes {
		datumTypes[i] = inputTypes[i].ToDatumType()
	}

	builtins := builtins.Aggregates[strings.ToLower(fn.String())]
	for _, b := range builtins {
		types := b.Types.Types()
		if len(types) != len(inputTypes) {
			continue
		}
		match := true
		for i, t := range types {
			if !datumTypes[i].Equivalent(t) {
				match = false
				break
			}
		}
		if match {
			// Found!
			constructAgg := func(evalCtx *tree.EvalContext) tree.AggregateFunc {
				return b.AggregateFunc(datumTypes, evalCtx)
			}

			colTyp, err := sqlbase.DatumTypeToColumnType(b.FixedReturnType())
			if err != nil {
				return nil, sqlbase.ColumnType{}, err
			}
			return constructAgg, colTyp, nil
		}
	}
	return nil, sqlbase.ColumnType{}, errors.Errorf(
		"no builtin aggregate for %s on %v", fn, inputTypes,
	)
}

// aggregator is the processor core type that does "aggregation" in the SQL
// sense. It groups rows and computes an aggregate for each group. The group is
// configured using the group key and the aggregator can be configured with one
// or more aggregation functions, as defined in the AggregatorSpec_Func enum.
//
// aggregator's output schema is comprised of what is specified by the
// accompanying SELECT expressions.
type aggregator struct {
	processorBase
	rowSourceBase

	state       aggregatorState
	input       RowSource
	inputDone   bool
	inputTypes  []sqlbase.ColumnType
	funcs       []*aggregateFuncHolder
	outputTypes []sqlbase.ColumnType
	datumAlloc  sqlbase.DatumAlloc

	bucketsAcc mon.BoundAccount

	groupCols        columns
	orderedGroupCols columns
	aggregations     []AggregatorSpec_Aggregation

	// buckets is used during the accumulation phase to track the bucket keys
	// that have been seen. After accumulation, the keys are extracted into
	// bucketsIter for iteration.
	buckets          map[string]struct{}
	bucketsIter      []string
	lastOrdGroupCols sqlbase.EncDatumRow
	arena            stringarena.Arena
	row              sqlbase.EncDatumRow
	scratch          []byte

	cancelChecker *sqlbase.CancelChecker
}

var _ Processor = &aggregator{}
var _ RowSource = &aggregator{}

const aggregatorProcName = "aggregator"

// aggregatorState represents the state of the processor.
type aggregatorState int

const (
	// aggAccumulating means that rows are being read from the input and used to
	// compute intermediary aggregation results.
	aggAccumulating aggregatorState = iota
	// aggEmittingRows means that accumulation has finished and rows are being
	// sent to the output.
	aggEmittingRows
	// aggDraining means that the input is being drained of its metadata.
	aggDraining
	// aggEmittingMeta means that the processor is emitting trailing metadata.
	aggEmittingMeta
	// aggDone means that the processor has exhausted sent all its rows and
	// exhausted its input of metadata. This is the only state where an
	// aggregator can return nil, nil from Next().
	aggDone
)

func newAggregator(
	flowCtx *FlowCtx,
	spec *AggregatorSpec,
	input RowSource,
	post *PostProcessSpec,
	output RowReceiver,
) (*aggregator, error) {
	ag := &aggregator{
		input:            input,
		groupCols:        spec.GroupCols,
		orderedGroupCols: spec.OrderedGroupCols,
		aggregations:     spec.Aggregations,
		buckets:          make(map[string]struct{}),
		funcs:            make([]*aggregateFuncHolder, len(spec.Aggregations)),
		outputTypes:      make([]sqlbase.ColumnType, len(spec.Aggregations)),
		row:              make(sqlbase.EncDatumRow, len(spec.Aggregations)),
		bucketsAcc:       flowCtx.EvalCtx.Mon.MakeBoundAccount(),
	}
	ag.arena = stringarena.Make(&ag.bucketsAcc)

	// Loop over the select expressions and extract any aggregate functions --
	// non-aggregation functions are replaced with parser.NewIdentAggregate,
	// (which just returns the last value added to them for a bucket) to provide
	// grouped-by values for each bucket.  ag.funcs is updated to contain all
	// the functions which need to be fed values.
	ag.inputTypes = input.OutputTypes()
	for i, aggInfo := range spec.Aggregations {
		if aggInfo.FilterColIdx != nil {
			col := *aggInfo.FilterColIdx
			if col >= uint32(len(ag.inputTypes)) {
				return nil, errors.Errorf("FilterColIdx out of range (%d)", col)
			}
			t := ag.inputTypes[col].SemanticType
			if t != sqlbase.ColumnType_BOOL && t != sqlbase.ColumnType_NULL {
				return nil, errors.Errorf(
					"filter column %d must be of boolean type, not %s", *aggInfo.FilterColIdx, t,
				)
			}
		}
		argTypes := make([]sqlbase.ColumnType, len(aggInfo.ColIdx))
		for i, c := range aggInfo.ColIdx {
			if c >= uint32(len(ag.inputTypes)) {
				return nil, errors.Errorf("ColIdx out of range (%d)", aggInfo.ColIdx)
			}
			argTypes[i] = ag.inputTypes[c]
		}
		aggConstructor, retType, err := GetAggregateInfo(aggInfo.Func, argTypes...)
		if err != nil {
			return nil, err
		}

		ag.funcs[i] = ag.newAggregateFuncHolder(aggConstructor)
		if aggInfo.Distinct {
			ag.funcs[i].seen = make(map[string]struct{})
		}

		ag.outputTypes[i] = retType
	}
	if err := ag.init(post, ag.outputTypes, flowCtx, output); err != nil {
		return nil, err
	}

	return ag, nil
}

// Start is part of the RowSource interface.
func (ag *aggregator) Start(ctx context.Context) context.Context {
	ag.input.Start(ctx)
	ctx = ag.startInternal(ctx, aggregatorProcName)
	ag.cancelChecker = sqlbase.NewCancelChecker(ctx)
	ag.state = aggAccumulating
	return ctx
}

// Run is part of the Processor interface.
func (ag *aggregator) Run(ctx context.Context, wg *sync.WaitGroup) {
	if ag.out.output == nil {
		panic("aggregator output not initialized for emitting rows")
	}

	ctx = ag.Start(ctx)
	Run(ctx, ag, ag.out.output)
	if wg != nil {
		wg.Done()
	}
}

func (ag *aggregator) close() {
	if !ag.closed {
		log.VEventf(ag.ctx, 2, "exiting aggregator")
		ag.bucketsAcc.Close(ag.ctx)
		for _, f := range ag.funcs {
			for _, aggFunc := range f.buckets {
				aggFunc.Close(ag.ctx)
			}
		}
	}
	ag.internalClose()
}

// producerMeta constructs the ProducerMetadata after consumption of rows has
// terminated and the input has been drained. It also closes the processor, if
// it has not been closed already.
func (ag *aggregator) producerMeta() (aggregatorState, sqlbase.EncDatumRow, *ProducerMetadata) {
	var meta *ProducerMetadata
	if trace := getTraceData(ag.ctx); trace != nil {
		meta = &ProducerMetadata{TraceData: trace}
	}
	// We need to close now that we've collected tracing data. The consumer is
	// allowed to not call ConsumerDone().
	ag.close()
	return aggDone, nil, meta
}

// matchLastOrdGroupCols takes a row and matches it with the row stored by
// lastOrdGroupCols. It returns true if the two rows are equal on the grouping
// columns, and false otherwise.
func (ag *aggregator) matchLastOrdGroupCols(row sqlbase.EncDatumRow) (bool, error) {
	for _, colIdx := range ag.orderedGroupCols {
		res, err := ag.lastOrdGroupCols[colIdx].Compare(
			&ag.inputTypes[colIdx], &ag.datumAlloc, &ag.flowCtx.EvalCtx, &row[colIdx],
		)
		if res != 0 || err != nil {
			return false, err
		}
	}
	return true, nil
}

// accumulateRows continually reads rows from the input and accumulates them
// into intermediary aggregate results. If it encounters metadata, the metadata
// is immediately returned. Subsequent calls of this function will resume row
// accumulation.
func (ag *aggregator) accumulateRows() (aggregatorState, sqlbase.EncDatumRow, *ProducerMetadata) {
	for {
		row, meta := ag.input.Next()
		if meta != nil {
			if meta.Err != nil {
				return ag.startDrain(meta.Err)
			}
			return aggAccumulating, nil, meta
		}
		if row == nil {
			log.VEvent(ag.ctx, 1, "accumulation complete")
			ag.inputDone = true
			break
		}

		if ag.lastOrdGroupCols == nil {
			ag.lastOrdGroupCols = row
		} else {
			matched, err := ag.matchLastOrdGroupCols(row)
			if err != nil {
				return ag.startDrain(err)
			}
			if !matched {
				ag.lastOrdGroupCols = row
				break
			}
		}
		if err := ag.accumulateRow(row); err != nil {
			return ag.startDrain(err)
		}
	}

	// Queries like `SELECT MAX(n) FROM t` expect a row of NULLs if nothing was
	// aggregated.
	if len(ag.buckets) < 1 && len(ag.groupCols) == 0 {
		ag.buckets[""] = struct{}{}
	}

	ag.bucketsIter = make([]string, 0, len(ag.buckets))
	for bucket := range ag.buckets {
		ag.bucketsIter = append(ag.bucketsIter, bucket)
	}
	ag.buckets = nil

	// Transition to aggEmittingRows, and let it generate the next row/meta.
	return aggEmittingRows, nil, nil
}

// drainInput reads the input until it finds metadata, then returns it.
func (ag *aggregator) drainInput() (aggregatorState, sqlbase.EncDatumRow, *ProducerMetadata) {
	for {
		row, meta := ag.input.Next()
		if row != nil {
			continue
		}
		if meta != nil {
			return aggDraining, nil, meta
		}

		// The input has no more metadata, so send our own trailing metadata.
		return aggEmittingMeta, nil, nil
	}
}

// startDrain informs the consumer that we're going to drain it, and transitions
// the processor to the draining state. If given an error, it will produce error
// metadata as well.
func (ag *aggregator) startDrain(
	err error,
) (aggregatorState, sqlbase.EncDatumRow, *ProducerMetadata) {
	var meta *ProducerMetadata
	ag.input.ConsumerDone()
	if err != nil {
		meta = &ProducerMetadata{Err: err}
	}
	return aggDraining, nil, meta
}

// emitRow constructs an output row from an accumulated bucket and returns it.
// Once there are no more buckets, the aggregator will drain.
func (ag *aggregator) emitRow() (aggregatorState, sqlbase.EncDatumRow, *ProducerMetadata) {
	for {
		if len(ag.bucketsIter) == 0 {
			// We've exhausted all of the aggregation buckets.
			if ag.inputDone {
				// The input has been fully consumed. Transition to draining so that we
				// emit any metadata that we've produced.
				return ag.startDrain(nil /* err */)
			}

			// We've only consumed part of the input where the rows are equal over
			// the columns specified by ag.orderedGroupCols, so we need to continue
			// accumulating the remaining rows.

			if err := ag.arena.UnsafeReset(ag.ctx); err != nil {
				return ag.startDrain(err)
			}
			ag.buckets = make(map[string]struct{})
			ag.bucketsIter = nil
			for _, f := range ag.funcs {
				for _, aggFunc := range f.buckets {
					aggFunc.Close(ag.ctx)
				}
				f.buckets = make(map[string]tree.AggregateFunc)
				if f.seen != nil {
					f.seen = make(map[string]struct{})
				}
			}

			if err := ag.accumulateRow(ag.lastOrdGroupCols); err != nil {
				return ag.startDrain(err)
			}

			return aggAccumulating, nil, nil
		}

		bucket := ag.bucketsIter[0]
		ag.bucketsIter = ag.bucketsIter[1:]

		for i, f := range ag.funcs {
			result, err := f.get(bucket)
			if err != nil {
				return ag.startDrain(err)
			}
			if result == nil {
				// Special case useful when this is a local stage of a distributed
				// aggregation.
				result = tree.DNull
			}
			ag.row[i] = sqlbase.DatumToEncDatum(ag.outputTypes[i], result)
		}

		outRow, status, err := ag.out.ProcessRow(ag.ctx, ag.row)
		if outRow != nil {
			return aggEmittingRows, outRow, nil
		}
		if outRow == nil && err == nil && status == NeedMoreRows {
			continue
		}
		return ag.startDrain(err)
	}
}

// Next is part of the RowSource interface.
func (ag *aggregator) Next() (sqlbase.EncDatumRow, *ProducerMetadata) {
	for ag.state != aggDone {
		var row sqlbase.EncDatumRow
		var meta *ProducerMetadata
		switch ag.state {
		case aggAccumulating:
			ag.state, row, meta = ag.accumulateRows()
		case aggEmittingRows:
			ag.state, row, meta = ag.emitRow()
		case aggDraining:
			ag.state, row, meta = ag.drainInput()
		case aggEmittingMeta:
			ag.state, row, meta = ag.producerMeta()
		}

		// If row == nil and meta == nil, but we aren't in the aggDone state, we
		// have a state transition without any output. As a result, we need to
		// execute the next state to produce output.
		if !(row == nil && meta == nil) {
			return row, meta
		}
	}

	return nil, nil
}

// ConsumerDone is part of the RowSource interface.
func (ag *aggregator) ConsumerDone() {
	ag.consumerDone()
	ag.input.ConsumerDone()
}

// ConsumerClosed is part of the RowSource interface.
func (ag *aggregator) ConsumerClosed() {
	ag.consumerClosed("aggregator")
	ag.input.ConsumerClosed()
	// The consumer is done, Next() will not be called again.
	ag.close()
}

// accumulateRow accumulates a single row, returning an error if accumulation
// failed for any reason.
func (ag *aggregator) accumulateRow(row sqlbase.EncDatumRow) error {
	if err := ag.cancelChecker.Check(); err != nil {
		return err
	}
	// The encoding computed here determines which bucket the non-grouping
	// datums are accumulated to.
	encoded, err := ag.encode(ag.scratch, row)
	if err != nil {
		return err
	}
	ag.scratch = encoded[:0]

	if _, ok := ag.buckets[string(encoded)]; !ok {
		s, err := ag.arena.AllocBytes(ag.ctx, encoded)
		if err != nil {
			return err
		}
		ag.buckets[s] = struct{}{}
	}

	// Feed the func holders for this bucket the non-grouping datums.
	for i, a := range ag.aggregations {
		if a.FilterColIdx != nil {
			col := *a.FilterColIdx
			if err := row[col].EnsureDecoded(&ag.inputTypes[col], &ag.datumAlloc); err != nil {
				return err
			}
			if row[*a.FilterColIdx].Datum != tree.DBoolTrue {
				// This row doesn't contribute to this aggregation.
				continue
			}
		}
		// Extract the corresponding arguments from the row to feed into the
		// aggregate function.
		// Most functions require at most one argument thus we separate
		// the first argument and allocation of (if applicable) a variadic
		// collection of arguments thereafter.
		var firstArg tree.Datum
		var otherArgs tree.Datums
		if len(a.ColIdx) > 1 {
			otherArgs = make(tree.Datums, len(a.ColIdx)-1)
		}
		isFirstArg := true
		for j, c := range a.ColIdx {
			if err := row[c].EnsureDecoded(&ag.inputTypes[c], &ag.datumAlloc); err != nil {
				return err
			}
			if isFirstArg {
				firstArg = row[c].Datum
				isFirstArg = false
				continue
			}
			otherArgs[j-1] = row[c].Datum
		}

		if err := ag.funcs[i].add(ag.ctx, encoded, firstArg, otherArgs); err != nil {
			return err
		}
	}
	return nil
}

type aggregateFuncHolder struct {
	create        func(*tree.EvalContext) tree.AggregateFunc
	group         *aggregator
	buckets       map[string]tree.AggregateFunc
	seen          map[string]struct{}
	arena         *stringarena.Arena
	bucketsMemAcc *mon.BoundAccount
}

const sizeOfAggregateFunc = int64(unsafe.Sizeof(tree.AggregateFunc(nil)))

func (ag *aggregator) newAggregateFuncHolder(
	create func(*tree.EvalContext) tree.AggregateFunc,
) *aggregateFuncHolder {
	return &aggregateFuncHolder{
		create:        create,
		group:         ag,
		buckets:       make(map[string]tree.AggregateFunc),
		bucketsMemAcc: &ag.bucketsAcc,
		arena:         &ag.arena,
	}
}

func (a *aggregateFuncHolder) add(
	ctx context.Context, bucket []byte, firstArg tree.Datum, otherArgs tree.Datums,
) error {
	if a.seen != nil {
		encoded, err := sqlbase.EncodeDatum(bucket, firstArg)
		if err != nil {
			return err
		}
		// Encode additional arguments if necessary.
		if otherArgs != nil {
			encoded, err = sqlbase.EncodeDatums(bucket, otherArgs)
			if err != nil {
				return err
			}
		}

		if _, ok := a.seen[string(encoded)]; ok {
			// skip
			return nil
		}
		s, err := a.arena.AllocBytes(ctx, encoded)
		if err != nil {
			return err
		}
		a.seen[s] = struct{}{}
	}

	impl, ok := a.buckets[string(bucket)]
	if !ok {
		// TODO(radu): we should account for the size of impl (this needs to be done
		// in each aggregate constructor).
		impl = a.create(&a.group.flowCtx.EvalCtx)
		usage := int64(len(bucket))
		usage += sizeOfAggregateFunc
		// TODO(radu): this model of each func having a map of buckets (one per
		// group) for each func plus a global map is very wasteful. We should have a
		// single map that stores all the AggregateFuncs.
		if err := a.bucketsMemAcc.Grow(ctx, usage); err != nil {
			return err
		}
		a.buckets[string(bucket)] = impl
	}

	return impl.Add(ctx, firstArg, otherArgs...)
}

func (a *aggregateFuncHolder) get(bucket string) (tree.Datum, error) {
	found, ok := a.buckets[bucket]
	if !ok {
		found = a.create(&a.group.flowCtx.EvalCtx)
	}

	return found.Result()
}

// encode returns the encoding for the grouping columns, this is then used as
// our group key to determine which bucket to add to.
func (ag *aggregator) encode(
	appendTo []byte, row sqlbase.EncDatumRow,
) (encoding []byte, err error) {
	for _, colIdx := range ag.groupCols {
		appendTo, err = row[colIdx].Encode(&ag.inputTypes[colIdx], &ag.datumAlloc, sqlbase.DatumEncoding_ASCENDING_KEY, appendTo)
		if err != nil {
			return appendTo, err
		}
	}
	return appendTo, nil
}
