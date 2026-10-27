package symbolic

// HasBitwiseOperations сообщает, содержит ли выражение побитовые операции или сдвиги.
// Такие выражения нужно транслировать в теорию битовых векторов, остальные - в целые числа.
func HasBitwiseOperations(expr SymbolicExpression) bool {
	return Visit(NewBitwiseOperationsVisitor(), expr).(bool)
}

// BitwiseOperationsVisitor возвращает для каждого узла bool: есть ли побитовая операция в его поддереве.
// Обход прекращается на первой найденной операции, поэтому уже посещённый узел заведомо не содержит
// побитовых операций, и общие поддеревья выражения обходятся один раз.
type BitwiseOperationsVisitor struct {
	visited map[SymbolicExpression]struct{}
}

func NewBitwiseOperationsVisitor() *BitwiseOperationsVisitor {
	return &BitwiseOperationsVisitor{visited: make(map[SymbolicExpression]struct{})}
}

func (*BitwiseOperationsVisitor) VisitVariable(*SymbolicVariable) interface{} { return false }
func (*BitwiseOperationsVisitor) VisitIntConstant(*IntConstant) interface{}   { return false }
func (*BitwiseOperationsVisitor) VisitBoolConstant(*BoolConstant) interface{} { return false }
func (*BitwiseOperationsVisitor) VisitRef(*Ref) interface{}                   { return false }

func (v *BitwiseOperationsVisitor) VisitBinaryOperation(expr *BinaryOperation) interface{} {
	switch expr.Operator {
	case BAND, BOR, XOR, SHL, SHR:
		return true
	}
	return v.anyOf(expr.Left, expr.Right)
}

func (v *BitwiseOperationsVisitor) VisitNegateOperation(expr *NegateOperation) interface{} {
	return v.anyOf(expr.Operand)
}

func (v *BitwiseOperationsVisitor) VisitLogicalOperation(expr *LogicalOperation) interface{} {
	return v.anyOf(expr.Operands...)
}

func (v *BitwiseOperationsVisitor) VisitArrayAccess(expr *ArrayAccess) interface{} {
	return v.anyOf(expr.Array, expr.Index)
}

func (v *BitwiseOperationsVisitor) VisitFunctionCall(expr *FunctionCall) interface{} {
	return v.anyOf(expr.Arguments...)
}

// anyOf сообщает, содержит ли хотя бы одно из выражений побитовую операцию.
func (v *BitwiseOperationsVisitor) anyOf(exprs ...SymbolicExpression) bool {
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
