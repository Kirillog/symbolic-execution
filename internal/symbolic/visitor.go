package symbolic

// Visitor интерфейс для обхода символьных выражений (Visitor Pattern)
type Visitor interface {
	VisitVariable(expr *SymbolicVariable) (interface{}, error)
	VisitIntConstant(expr *IntConstant) (interface{}, error)
	VisitBoolConstant(expr *BoolConstant) (interface{}, error)
	VisitNegateOperation(expr *NegateOperation) (interface{}, error)
	VisitBinaryOperation(expr *BinaryOperation) (interface{}, error)
	VisitLogicalOperation(expr *LogicalOperation) (interface{}, error)
	VisitRef(expr *Ref) (interface{}, error)
	VisitArrayAccess(expr *ArrayAccess) (interface{}, error)
	VisitFunctionCall(expr *FunctionCall) (interface{}, error)
	// TODO: Добавьте методы для других типов выражений по мере необходимости
}
