package symbolic

// ExpressionVisitor - упрощённый посетитель: возвращает значение, но не возвращает ошибок.
// Рекурсивный спуск в подвыражения реализация выполняет сама через Visit.
type ExpressionVisitor interface {
	VisitVariable(expr *SymbolicVariable) interface{}
	VisitIntConstant(expr *IntConstant) interface{}
	VisitBoolConstant(expr *BoolConstant) interface{}
	VisitNegateOperation(expr *NegateOperation) interface{}
	VisitBinaryOperation(expr *BinaryOperation) interface{}
	VisitLogicalOperation(expr *LogicalOperation) interface{}
	VisitRef(expr *Ref) interface{}
	VisitArrayAccess(expr *ArrayAccess) interface{}
	VisitFunctionCall(expr *FunctionCall) interface{}
}

// Visit применяет посетителя без ошибок к выражению и возвращает результат его метода.
func Visit(visitor ExpressionVisitor, expr SymbolicExpression) interface{} {
	result, _ := expr.Accept(expressionVisitorAdapter{visitor})
	return result
}

// expressionVisitorAdapter адаптирует ExpressionVisitor к ExpressionVisitorError.
type expressionVisitorAdapter struct {
	visitor ExpressionVisitor
}

func (a expressionVisitorAdapter) VisitVariable(expr *SymbolicVariable) (interface{}, error) {
	return a.visitor.VisitVariable(expr), nil
}

func (a expressionVisitorAdapter) VisitIntConstant(expr *IntConstant) (interface{}, error) {
	return a.visitor.VisitIntConstant(expr), nil
}

func (a expressionVisitorAdapter) VisitBoolConstant(expr *BoolConstant) (interface{}, error) {
	return a.visitor.VisitBoolConstant(expr), nil
}

func (a expressionVisitorAdapter) VisitNegateOperation(expr *NegateOperation) (interface{}, error) {
	return a.visitor.VisitNegateOperation(expr), nil
}

func (a expressionVisitorAdapter) VisitBinaryOperation(expr *BinaryOperation) (interface{}, error) {
	return a.visitor.VisitBinaryOperation(expr), nil
}

func (a expressionVisitorAdapter) VisitLogicalOperation(expr *LogicalOperation) (interface{}, error) {
	return a.visitor.VisitLogicalOperation(expr), nil
}

func (a expressionVisitorAdapter) VisitRef(expr *Ref) (interface{}, error) {
	return a.visitor.VisitRef(expr), nil
}

func (a expressionVisitorAdapter) VisitArrayAccess(expr *ArrayAccess) (interface{}, error) {
	return a.visitor.VisitArrayAccess(expr), nil
}

func (a expressionVisitorAdapter) VisitFunctionCall(expr *FunctionCall) (interface{}, error) {
	return a.visitor.VisitFunctionCall(expr), nil
}
