package symbolic

// ExpressionVisitorError - посетитель, возвращающий значение или ошибку; интерфейс для обхода символьных выражений (Visitor Pattern)
type ExpressionVisitorError interface {
	VisitVariable(expr *SymbolicVariable) (interface{}, error)
	VisitIntConstant(expr *IntConstant) (interface{}, error)
	VisitFloatConstant(expr *FloatConstant) (interface{}, error)
	VisitBoolConstant(expr *BoolConstant) (interface{}, error)
	VisitNegateOperation(expr *NegateOperation) (interface{}, error)
	VisitBinaryOperation(expr *BinaryOperation) (interface{}, error)
	VisitLogicalOperation(expr *LogicalOperation) (interface{}, error)
	VisitRef(expr *Ref) (interface{}, error)
	VisitArrayConstant(expr *ArrayConstant) (interface{}, error)
	VisitArrayAccess(expr *ArrayAccess) (interface{}, error)
	VisitArrayStore(expr *ArrayStore) (interface{}, error)
	VisitFunctionCall(expr *FunctionCall) (interface{}, error)
	// TODO: Добавьте методы для других типов выражений по мере необходимости
}
