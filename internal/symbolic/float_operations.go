package symbolic

// HasFloats сообщает, содержит ли выражение значения типа float64.
func HasFloats(expr SymbolicExpression) bool {
	return Visit(NewFloatsVisitor(), expr).(bool)
}

type FloatsVisitor struct {
	visited map[SymbolicExpression]struct{}
}

func NewFloatsVisitor() *FloatsVisitor {
	return &FloatsVisitor{visited: make(map[SymbolicExpression]struct{})}
}

func (*FloatsVisitor) VisitVariable(expr *SymbolicVariable) interface{} {
	return expr.Type() == FloatType
}
func (*FloatsVisitor) VisitIntConstant(*IntConstant) interface{}     { return false }
func (*FloatsVisitor) VisitFloatConstant(*FloatConstant) interface{} { return true }
func (*FloatsVisitor) VisitBoolConstant(*BoolConstant) interface{}   { return false }
func (*FloatsVisitor) VisitRef(*Ref) interface{}                     { return false }

func (v *FloatsVisitor) VisitBinaryOperation(expr *BinaryOperation) interface{} {
	return v.anyOf(expr.Left, expr.Right)
}

func (v *FloatsVisitor) VisitNegateOperation(expr *NegateOperation) interface{} {
	return v.anyOf(expr.Operand)
}

func (v *FloatsVisitor) VisitLogicalOperation(expr *LogicalOperation) interface{} {
	return v.anyOf(expr.Operands...)
}

func (v *FloatsVisitor) VisitArrayAccess(expr *ArrayAccess) interface{} {
	return expr.Type() == FloatType || v.anyOf(expr.Array, expr.Index)
}

func (v *FloatsVisitor) VisitFunctionCall(expr *FunctionCall) interface{} {
	return expr.Type() == FloatType || v.anyOf(expr.Arguments...)
}

func (v *FloatsVisitor) anyOf(exprs ...SymbolicExpression) bool {
	for _, expr := range exprs {
		if _, seen := v.visited[expr]; seen {
			continue
		}
		v.visited[expr] = struct{}{}
		if Visit(v, expr).(bool) {
			return true
		}
	}
	return false
}
