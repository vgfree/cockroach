// Code generated by optgen; DO NOT EDIT.

package opt

const (
	UnknownOp Operator = iota

	// ------------------------------------------------------------
	// Enforcer Operators
	// ------------------------------------------------------------

	// SortOp enforces the ordering of rows returned by its input expression. Rows can
	// be sorted by one or more of the input columns, each of which can be sorted in
	// either ascending or descending order. See the Ordering field in the
	// PhysicalProps struct.
	// TODO(andyk): Add the Ordering field.
	SortOp

	// ------------------------------------------------------------
	// Relational Operators
	// ------------------------------------------------------------

	// ScanOp returns a result set containing every row in the specified table, by
	// scanning one of the table's indexes according to its ordering. The private
	// Def field is an *opt.ScanOpDef that identifies the table and index to scan,
	// as well as the subset of columns to project from it.
	ScanOp

	// ValuesOp returns a manufactured result set containing a constant number of rows.
	// specified by the Rows list field. Each row must contain the same set of
	// columns in the same order.
	//
	// The Rows field contains a list of Tuples, one for each row. Each tuple has
	// the same length (same with that of Cols).
	//
	// The Cols field contains the set of column indices returned by each row
	// as an opt.ColList. It is legal for Cols to be empty.
	ValuesOp

	// SelectOp filters rows from its input result set, based on the boolean filter
	// predicate expression. Rows which do not match the filter are discarded. While
	// the Filter operand can be any boolean expression, normalization rules will
	// typically convert it to a Filters operator in order to make conjunction list
	// matching easier.
	SelectOp

	// ProjectOp modifies the set of columns returned by the input result set. Columns
	// can be removed, reordered, or renamed. In addition, new columns can be
	// synthesized. Projections is a scalar Projections list operator that contains
	// the list of expressions that describe the output columns. The Cols field of
	// the Projections operator provides the indexes of each of the output columns.
	ProjectOp

	// InnerJoinOp creates a result set that combines columns from its left and right
	// inputs, based upon its "on" join predicate. Rows which do not match the
	// predicate are filtered. While expressions in the predicate can refer to
	// columns projected by either the left or right inputs, the inputs are not
	// allowed to refer to the other's projected columns.
	InnerJoinOp

	LeftJoinOp

	RightJoinOp

	FullJoinOp

	SemiJoinOp

	AntiJoinOp

	// InnerJoinApplyOp has the same join semantics as InnerJoin. However, unlike
	// InnerJoin, it allows the right input to refer to columns projected by the
	// left input.
	InnerJoinApplyOp

	LeftJoinApplyOp

	RightJoinApplyOp

	FullJoinApplyOp

	SemiJoinApplyOp

	AntiJoinApplyOp

	// GroupByOp is an operator that is used for performing aggregations (for queries
	// with aggregate functions, HAVING clauses and/or group by expressions). It
	// groups results that are equal on the grouping columns and computes
	// aggregations as described by Aggregations (which is always an Aggregations
	// operator). The arguments of the aggregations are columns from the input.
	GroupByOp

	// UnionOp is an operator used to combine the Left and Right input relations into
	// a single set containing rows from both inputs. Duplicate rows are discarded.
	// The private field, ColMap, matches columns from the Left and Right inputs
	// of the Union with the output columns. See the comment above opt.SetOpColMap
	// for more details.
	UnionOp

	// IntersectOp is an operator used to perform an intersection between the Left
	// and Right input relations. The result consists only of rows in the Left
	// relation that are also present in the Right relation. Duplicate rows are
	// discarded.
	// The private field, ColMap, matches columns from the Left and Right inputs
	// of the Intersect with the output columns. See the comment above
	// opt.SetOpColMap for more details.
	IntersectOp

	// ExceptOp is an operator used to perform a set difference between the Left and
	// Right input relations. The result consists only of rows in the Left relation
	// that are not present in the Right relation. Duplicate rows are discarded.
	// The private field, ColMap, matches columns from the Left and Right inputs
	// of the Except with the output columns. See the comment above opt.SetOpColMap
	// for more details.
	ExceptOp

	// UnionAllOp is an operator used to combine the Left and Right input relations
	// into a single set containing rows from both inputs. Duplicate rows are
	// not discarded. For example:
	//   SELECT x FROM xx UNION ALL SELECT y FROM yy
	//     x       y         out
	//   -----   -----      -----
	//     1       1          1
	//     1       2    ->    1
	//     2       3          1
	//                        2
	//                        2
	//                        3
	//
	// The private field, ColMap, matches columns from the Left and Right inputs
	// of the UnionAll with the output columns. See the comment above
	// opt.SetOpColMap for more details.
	UnionAllOp

	// IntersectAllOp is an operator used to perform an intersection between the Left
	// and Right input relations. The result consists only of rows in the Left
	// relation that have a corresponding row in the Right relation. Duplicate rows
	// are not discarded. This effectively creates a one-to-one mapping between the
	// Left and Right rows. For example:
	//   SELECT x FROM xx INTERSECT ALL SELECT y FROM yy
	//     x       y         out
	//   -----   -----      -----
	//     1       1          1
	//     1       1    ->    1
	//     1       2          2
	//     2       2          2
	//     2       3
	//     4
	//
	// The private field, ColMap, matches columns from the Left and Right inputs
	// of the IntersectAll with the output columns. See the comment above
	// opt.SetOpColMap for more details.
	IntersectAllOp

	// ExceptAllOp is an operator used to perform a set difference between the Left
	// and Right input relations. The result consists only of rows in the Left
	// relation that do not have a corresponding row in the Right relation.
	// Duplicate rows are not discarded. This effectively creates a one-to-one
	// mapping between the Left and Right rows. For example:
	//   SELECT x FROM xx EXCEPT ALL SELECT y FROM yy
	//     x       y         out
	//   -----   -----      -----
	//     1       1    ->    1
	//     1       1          4
	//     1       2
	//     2       2
	//     2       3
	//     4
	//
	// The private field, ColMap, matches columns from the Left and Right inputs
	// of the ExceptAll with the output columns. See the comment above
	// opt.SetOpColMap for more details.
	ExceptAllOp

	// LimitOp returns a limited subset of the results in the input relation.
	// The limit expression is a scalar value; the operator returns at most this many
	// rows. The private field is an opt.Ordering which indicates the desired
	// row ordering (the first rows with respect to this ordering are returned).
	LimitOp

	// OffsetOp filters out the first Offset rows of the input relation; used in
	// conjunction with Limit.
	OffsetOp

	// Max1RowOp is an operator which enforces that its input must return at most one
	// row. It is used as input to the Subquery operator. See the comment above
	// Subquery for more details.
	Max1RowOp

	// ------------------------------------------------------------
	// Scalar Operators
	// ------------------------------------------------------------

	// SubqueryOp is a subquery in a single-row context such as
	// `SELECT 1 = (SELECT 1)` or `SELECT (1, 'a') = (SELECT 1, 'a')`.
	// In a single-row context, the outer query is only valid if the subquery
	// returns at most one row.
	//
	// Subqueries in a multi-row context such as
	// `SELECT 1 IN (SELECT c FROM t)` or `SELECT (1, 'a') IN (SELECT 1, 'a')`
	// can be transformed to a single row context using the Any operator. (Note that
	// this is different from the SQL ANY operator. See the comment above the Any
	// operator for more details.)
	//
	// We use the following transformations:
	// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	// `<var> IN (<subquery>)`
	//    ==> `Any(SELECT <var> = x FROM (<subquery>) AS q(x))`
	//
	// `<var> NOT IN (<subquery>)`
	//    ==> `NOT Any(SELECT <var> = x FROM (<subquery>) AS q(x))`
	//
	// `<var> <comp> {SOME|ANY}(<subquery>)`
	//    ==> `Any(SELECT <var> <comp> x FROM (<subquery>) AS q(x))`
	//
	// `<var> <comp> ALL(<subquery>)`
	//    ==> `NOT Any(SELECT NOT(<var> <comp> x) FROM (<subquery>) AS q(x))`
	// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
	//
	// The Input field contains the subquery itself, which should be wrapped in a
	// Max1Row operator to enforce that the subquery can return at most one row
	// (Max1Row may be removed by the optimizer later if it can determine statically
	// that the subquery will always return at most one row). The Projection field
	// contains a single column representing the output of the subquery. For
	// example, `(SELECT 1, 'a')` would be represented by the following structure:
	//
	// (Subquery
	//   (Max1Row
	//     (Project (Values (Tuple)) (Projections (Tuple (Const 1) (Const 'a'))))
	//   )
	//   (Variable 3)
	// )
	//
	// Here Variable 3 refers to the projection from the Input,
	// (Tuple (Const 1) (Const 'a')).
	SubqueryOp

	// AnyOp is a special operator that does not exist in SQL. However, it is very
	// similar to the SQL ANY, and can be converted to the SQL ANY operator using
	// the following transformation:
	//  `Any(<subquery>)` ==> `True = ANY(<subquery>)`
	//
	// Any expects the subquery to return a single boolean column. The semantics
	// are equivalent to the SQL ANY expression above on the right: Any returns true
	// if any of the values returned by the subquery are true, else returns NULL
	// if any of the values are NULL, else returns false.
	AnyOp

	// VariableOp is the typed scalar value of a column in the query. The private
	// field is a Metadata.ColumnID that references the column by index.
	VariableOp

	// ConstOp is a typed scalar constant value. The private field is a tree.Datum
	// value having any datum type that's legal in the expression's context.
	ConstOp

	// NullOp is the constant SQL null value that has "unknown value" semantics. If
	// the Typ field is not types.Unknown, then the value is known to be in the
	// domain of that type. This is important for preserving correct types in
	// replacement patterns. For example:
	//   (Plus (Function ...) (Const 1))
	//
	// If the function in that expression has a static type of Int, but then it gets
	// constant folded to (Null), then its type must remain as Int. Any other type
	// violates logical equivalence of the expression, breaking type inference and
	// possibly changing the results of execution. The solution is to tag the null
	// with the correct type:
	//   (Plus (Null (Int)) (Const 1))
	//
	// Null is its own operator rather than a Const datum in order to make matching
	// and replacement easier and more efficient, as patterns can contain (Null)
	// expressions.
	NullOp

	// TrueOp is the boolean true value that is equivalent to the tree.DBoolTrue datum
	// value. It is a separate operator to make matching and replacement simpler and
	// more efficient, as patterns can contain (True) expressions.
	TrueOp

	// FalseOp is the boolean false value that is equivalent to the tree.DBoolFalse
	// datum value. It is a separate operator to make matching and replacement
	// simpler and more efficient, as patterns can contain (False) expressions.
	FalseOp

	PlaceholderOp

	TupleOp

	// ProjectionsOp is a set of typed scalar expressions that will become output
	// columns for a containing Project operator. The private Cols field contains
	// the list of column indexes returned by the expression, as an opt.ColList. It
	// is not legal for Cols to be empty.
	ProjectionsOp

	// AggregationsOp is a set of aggregate expressions that will become output
	// columns for a containing GroupBy operator. The private Cols field contains
	// the list of column indexes returned by the expression, as an opt.ColList. It
	// is legal for Cols to be empty.
	AggregationsOp

	ExistsOp

	// FiltersOp is a boolean And operator that only appears as the Filters child of
	// a Select operator, or the On child of a Join operator. For example:
	//   (Select
	//     (Scan a)
	//     (Filters (Gt (Variable a) 1) (Lt (Variable a) 5))
	//   )
	//
	// Normalization rules ensure that a Filters expression is always created if
	// there is at least one condition, so that other rules can rely on its presence
	// when matching, even in the case where there is only one condition. The
	// semantics of the Filters operator are identical to those of the And operator.
	FiltersOp

	// AndOp is the boolean conjunction operator that evalutes to true if all of its
	// conditions evaluate to true. If the conditions list is empty, it evalutes to
	// true.
	AndOp

	// OrOp is the boolean disjunction operator that evalutes to true if any of its
	// conditions evaluate to true. If the conditions list is empty, it evaluates to
	// false.
	OrOp

	// NotOp is the boolean negation operator that evaluates to true if its input
	// evalutes to false.
	NotOp

	EqOp

	LtOp

	GtOp

	LeOp

	GeOp

	NeOp

	InOp

	NotInOp

	LikeOp

	NotLikeOp

	ILikeOp

	NotILikeOp

	SimilarToOp

	NotSimilarToOp

	RegMatchOp

	NotRegMatchOp

	RegIMatchOp

	NotRegIMatchOp

	IsOp

	IsNotOp

	ContainsOp

	JsonExistsOp

	JsonAllExistsOp

	JsonSomeExistsOp

	BitandOp

	BitorOp

	BitxorOp

	PlusOp

	MinusOp

	MultOp

	DivOp

	FloorDivOp

	ModOp

	PowOp

	ConcatOp

	LShiftOp

	RShiftOp

	FetchValOp

	FetchTextOp

	FetchValPathOp

	FetchTextPathOp

	UnaryMinusOp

	UnaryComplementOp

	CastOp

	// CaseOp is a CASE statement of the form:
	//   CASE [ <Input> ]
	//       WHEN <condval1> THEN <expr1>
	//     [ WHEN <condval2> THEN <expr2> ] ...
	//     [ ELSE <expr> ]
	//   END
	//
	// The Case operator evaluates <Input> (if not provided, Input is set to True),
	// then picks the WHEN branch where <condval> is equal to
	// <Input>, then evaluates and returns the corresponding THEN expression. If no
	// WHEN branch matches, the ELSE expression is evaluated and returned, if any.
	// Otherwise, NULL is returned.
	//
	// Note that the Whens list inside Case is used to represent all the WHEN
	// branches as well as the ELSE statement if it exists. It is of the form:
	// [(When <condval1> <expr1>),(When <condval2> <expr2>),...,<expr>]
	CaseOp

	// WhenOp represents a single WHEN ... THEN ... condition inside a CASE statement.
	// It is the type of each list item in Whens (except for the last item which is
	// a raw expression for the ELSE statement).
	WhenOp

	// ArrayOp is an ARRAY literal of the form ARRAY[<expr1>, <expr2>, ..., <exprN>].
	ArrayOp

	// FunctionOp invokes a builtin SQL function like CONCAT or NOW, passing the given
	// arguments. The private field is a *opt.FuncOpDef struct that provides the
	// name of the function as well as a pointer to the builtin overload definition.
	FunctionOp

	CoalesceOp

	// UnsupportedExprOp is used for interfacing with the old planner code. It can
	// encapsulate a TypedExpr that is otherwise not supported by the optimizer.
	UnsupportedExprOp

	// NumOperators tracks the total count of operators.
	NumOperators
)

const opNames = "unknownsortscanvaluesselectprojectinner-joinleft-joinright-joinfull-joinsemi-joinanti-joininner-join-applyleft-join-applyright-join-applyfull-join-applysemi-join-applyanti-join-applygroup-byunionintersectexceptunion-allintersect-allexcept-alllimitoffsetmax1-rowsubqueryanyvariableconstnulltruefalseplaceholdertupleprojectionsaggregationsexistsfiltersandornoteqltgtlegeneinnot-inlikenot-likei-likenot-i-likesimilar-tonot-similar-toreg-matchnot-reg-matchreg-i-matchnot-reg-i-matchisis-notcontainsjson-existsjson-all-existsjson-some-existsbitandbitorbitxorplusminusmultdivfloor-divmodpowconcatl-shiftr-shiftfetch-valfetch-textfetch-val-pathfetch-text-pathunary-minusunary-complementcastcasewhenarrayfunctioncoalesceunsupported-expr"

var opIndexes = [...]uint32{0, 7, 11, 15, 21, 27, 34, 44, 53, 63, 72, 81, 90, 106, 121, 137, 152, 167, 182, 190, 195, 204, 210, 219, 232, 242, 247, 253, 261, 269, 272, 280, 285, 289, 293, 298, 309, 314, 325, 337, 343, 350, 353, 355, 358, 360, 362, 364, 366, 368, 370, 372, 378, 382, 390, 396, 406, 416, 430, 439, 452, 463, 478, 480, 486, 494, 505, 520, 536, 542, 547, 553, 557, 562, 566, 569, 578, 581, 584, 590, 597, 604, 613, 623, 637, 652, 663, 679, 683, 687, 691, 696, 704, 712, 728}

var EnforcerOperators = [...]Operator{
	SortOp,
}

var RelationalOperators = [...]Operator{
	ScanOp,
	ValuesOp,
	SelectOp,
	ProjectOp,
	InnerJoinOp,
	LeftJoinOp,
	RightJoinOp,
	FullJoinOp,
	SemiJoinOp,
	AntiJoinOp,
	InnerJoinApplyOp,
	LeftJoinApplyOp,
	RightJoinApplyOp,
	FullJoinApplyOp,
	SemiJoinApplyOp,
	AntiJoinApplyOp,
	GroupByOp,
	UnionOp,
	IntersectOp,
	ExceptOp,
	UnionAllOp,
	IntersectAllOp,
	ExceptAllOp,
	LimitOp,
	OffsetOp,
	Max1RowOp,
}

var JoinOperators = [...]Operator{
	InnerJoinOp,
	LeftJoinOp,
	RightJoinOp,
	FullJoinOp,
	SemiJoinOp,
	AntiJoinOp,
	InnerJoinApplyOp,
	LeftJoinApplyOp,
	RightJoinApplyOp,
	FullJoinApplyOp,
	SemiJoinApplyOp,
	AntiJoinApplyOp,
}

var JoinApplyOperators = [...]Operator{
	InnerJoinApplyOp,
	LeftJoinApplyOp,
	RightJoinApplyOp,
	FullJoinApplyOp,
	SemiJoinApplyOp,
	AntiJoinApplyOp,
}

var ScalarOperators = [...]Operator{
	SubqueryOp,
	AnyOp,
	VariableOp,
	ConstOp,
	NullOp,
	TrueOp,
	FalseOp,
	PlaceholderOp,
	TupleOp,
	ProjectionsOp,
	AggregationsOp,
	ExistsOp,
	FiltersOp,
	AndOp,
	OrOp,
	NotOp,
	EqOp,
	LtOp,
	GtOp,
	LeOp,
	GeOp,
	NeOp,
	InOp,
	NotInOp,
	LikeOp,
	NotLikeOp,
	ILikeOp,
	NotILikeOp,
	SimilarToOp,
	NotSimilarToOp,
	RegMatchOp,
	NotRegMatchOp,
	RegIMatchOp,
	NotRegIMatchOp,
	IsOp,
	IsNotOp,
	ContainsOp,
	JsonExistsOp,
	JsonAllExistsOp,
	JsonSomeExistsOp,
	BitandOp,
	BitorOp,
	BitxorOp,
	PlusOp,
	MinusOp,
	MultOp,
	DivOp,
	FloorDivOp,
	ModOp,
	PowOp,
	ConcatOp,
	LShiftOp,
	RShiftOp,
	FetchValOp,
	FetchTextOp,
	FetchValPathOp,
	FetchTextPathOp,
	UnaryMinusOp,
	UnaryComplementOp,
	CastOp,
	CaseOp,
	WhenOp,
	ArrayOp,
	FunctionOp,
	CoalesceOp,
	UnsupportedExprOp,
}

var ConstValueOperators = [...]Operator{
	ConstOp,
	NullOp,
	TrueOp,
	FalseOp,
}

var BooleanOperators = [...]Operator{
	TrueOp,
	FalseOp,
	FiltersOp,
	AndOp,
	OrOp,
	NotOp,
}

var ComparisonOperators = [...]Operator{
	EqOp,
	LtOp,
	GtOp,
	LeOp,
	GeOp,
	NeOp,
	InOp,
	NotInOp,
	LikeOp,
	NotLikeOp,
	ILikeOp,
	NotILikeOp,
	SimilarToOp,
	NotSimilarToOp,
	RegMatchOp,
	NotRegMatchOp,
	RegIMatchOp,
	NotRegIMatchOp,
	IsOp,
	IsNotOp,
	ContainsOp,
	JsonExistsOp,
	JsonAllExistsOp,
	JsonSomeExistsOp,
}

var BinaryOperators = [...]Operator{
	BitandOp,
	BitorOp,
	BitxorOp,
	PlusOp,
	MinusOp,
	MultOp,
	DivOp,
	FloorDivOp,
	ModOp,
	PowOp,
	ConcatOp,
	LShiftOp,
	RShiftOp,
	FetchValOp,
	FetchTextOp,
	FetchValPathOp,
	FetchTextPathOp,
}

var UnaryOperators = [...]Operator{
	UnaryMinusOp,
	UnaryComplementOp,
}
