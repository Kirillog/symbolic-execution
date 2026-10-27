package symbolic

import "math"

// RangeConstraints возвращает ограничения MinInt64 <= e <= MaxInt64 для int-узлов выражения.
// Переменные ограничиваются всегда. Остальные int-узлы (арифметика, унарный минус,
// чтение из массива, вызов функции) - только при intermediates == true: тогда любая модель
// не содержит переполнений и совпадает с реальным исполнением. Каждый узел ограничивается один раз.
func RangeConstraints(expr SymbolicExpression, intermediates bool) []SymbolicExpression {
	v := &rangeConstraintsVisitor{
		intermediates: intermediates,
		seenVars:      make(map[string]struct{}),
		seenNodes:     make(map[SymbolicExpression]struct{}),
	}
	v.walk(expr)
	return v.result
}

// rangeConstraintsVisitor обходит выражение и накапливает ограничения диапазона в result.
type rangeConstraintsVisitor struct {
	intermediates bool
	seenVars      map[string]struct{}
	seenNodes     map[SymbolicExpression]struct{}
	result        []SymbolicExpression
}

func (v *rangeConstraintsVisitor) VisitVariable(expr *SymbolicVariable) interface{} {
	if _, seen := v.seenVars[expr.Name]; seen || expr.Type() != IntType {
		return nil
	}
	v.seenVars[expr.Name] = struct{}{}
	v.result = append(v.result, inRange(expr))
	return nil
}

func (v *rangeConstraintsVisitor) VisitIntConstant(*IntConstant) interface{}   { return nil }
func (v *rangeConstraintsVisitor) VisitBoolConstant(*BoolConstant) interface{} { return nil }
func (v *rangeConstraintsVisitor) VisitRef(*Ref) interface{}                   { return nil }

func (v *rangeConstraintsVisitor) VisitBinaryOperation(expr *BinaryOperation) interface{} {
	v.walk(expr.Left, expr.Right)
	v.bound(expr)
	return nil
}

func (v *rangeConstraintsVisitor) VisitNegateOperation(expr *NegateOperation) interface{} {
	v.walk(expr.Operand)
	v.bound(expr)
	return nil
}

func (v *rangeConstraintsVisitor) VisitLogicalOperation(expr *LogicalOperation) interface{} {
	v.walk(expr.Operands...)
	return nil
}

func (v *rangeConstraintsVisitor) VisitArrayAccess(expr *ArrayAccess) interface{} {
	v.walk(expr.Array, expr.Index)
	v.bound(expr)
	return nil
}

func (v *rangeConstraintsVisitor) VisitFunctionCall(expr *FunctionCall) interface{} {
	v.walk(expr.Arguments...)
	v.bound(expr)
	return nil
}

func (v *rangeConstraintsVisitor) walk(exprs ...SymbolicExpression) {
	for _, expr := range exprs {
		Visit(v, expr)
	}
}

// bound добавляет ограничение для вычисляемого int-узла, если включены промежуточные.
func (v *rangeConstraintsVisitor) bound(expr SymbolicExpression) {
	if !v.intermediates || expr.Type() != IntType {
		return
	}
	if _, seen := v.seenNodes[expr]; seen {
		return
	}
	v.seenNodes[expr] = struct{}{}
	v.result = append(v.result, inRange(expr))
}

func inRange(expr SymbolicExpression) SymbolicExpression {
	return NewLogicalOperation([]SymbolicExpression{
		NewBinaryOperation(expr, NewIntConstant(math.MinInt64), GE),
		NewBinaryOperation(expr, NewIntConstant(math.MaxInt64), LE),
	}, AND)
}
