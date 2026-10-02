// Package translator содержит реализацию транслятора в Z3
package translator

import (
	"fmt"
	. "symbolic-execution-course/internal/bslices"
	"symbolic-execution-course/internal/symbolic"

	"github.com/ebukreev/go-z3/z3"
)

// Z3Translator транслирует символьные выражения в Z3 формулы
type Z3Translator struct {
	ctx    *z3.Context
	config *z3.Config
	vars   map[string]z3.Value // Кэш переменных
}

// NewZ3Translator создаёт новый экземпляр Z3 транслятора
func NewZ3Translator() *Z3Translator {
	config := &z3.Config{}
	ctx := z3.NewContext(config)

	return &Z3Translator{
		ctx:    ctx,
		config: config,
		vars:   make(map[string]z3.Value),
	}
}

// GetContext возвращает Z3 контекст
func (zt *Z3Translator) GetContext() interface{} {
	return zt.ctx
}

// Reset сбрасывает состояние транслятора
func (zt *Z3Translator) Reset() {
	zt.vars = make(map[string]z3.Value)
}

// Close освобождает ресурсы
func (zt *Z3Translator) Close() {
	// Z3 контекст закрывается автоматически
}

// TranslateExpression транслирует символьное выражение в Z3
func (zt *Z3Translator) TranslateExpression(expr symbolic.SymbolicExpression) (interface{}, error) {
	return expr.Accept(zt)
}

// TODO: Реализуйте следующие методы в рамках домашнего задания

// VisitVariable транслирует символьную переменную в Z3
func (zt *Z3Translator) VisitVariable(expr *symbolic.SymbolicVariable) (interface{}, error) {
	// TODO: Реализовать
	// Проверить, есть ли переменная в кэше
	// Если нет - создать новую Z3 переменную соответствующего типа
	// Добавить в кэш и вернуть

	// Подсказки:
	// - Используйте zt.ctx.IntConst(name) для int переменных
	// - Используйте zt.ctx.BoolConst(name) для bool переменных
	// - Храните переменные в zt.vars для повторного использования

	if z3Var, exists := zt.vars[expr.Name]; exists {
		return z3Var, nil
	}

	z3Var := zt.createZ3Variable(expr.Name, expr.ExprType)
	zt.vars[expr.Name] = z3Var
	return z3Var, nil
}

// VisitIntConstant транслирует целочисленную константу в Z3
func (zt *Z3Translator) VisitIntConstant(expr *symbolic.IntConstant) (interface{}, error) {
	// TODO: Реализовать
	// Создать Z3 константу с помощью zt.ctx.FromBigInt или аналогичного метода
	return zt.ctx.FromInt(expr.Value, zt.ctx.IntSort()), nil
}

// VisitBoolConstant транслирует булеву константу в Z3
func (zt *Z3Translator) VisitBoolConstant(expr *symbolic.BoolConstant) (interface{}, error) {
	// TODO: Реализовать
	// Использовать zt.ctx.FromBool для создания Z3 булевой константы

	return zt.ctx.FromBool(expr.Value), nil
}

// VisitBinaryOperation транслирует бинарную операцию в Z3
func (zt *Z3Translator) VisitBinaryOperation(expr *symbolic.BinaryOperation) (interface{}, error) {
	// TODO: Реализовать
	// 1. Транслировать левый и правый операнды
	// 2. В зависимости от оператора создать соответствующую Z3 операцию

	// Подсказки по операциям в Z3:
	// - Арифметические: left.Add(right), left.Sub(right), left.Mul(right), left.Div(right)
	// - Сравнения: left.Eq(right), left.LT(right), left.LE(right), etc.
	// - Приводите типы: left.(z3.Int), right.(z3.Int) для int операций

	if expr.Operator == symbolic.EQ || expr.Operator == symbolic.NE {
		left, err := translateAs[z3.Value](zt, expr.Left)
		if err != nil {
			return nil, err
		}
		right, err := translateAs[z3.Value](zt, expr.Right)
		if err != nil {
			return nil, err
		}
		distinct := zt.ctx.Distinct(left, right)
		if expr.Operator == symbolic.EQ {
			return distinct.Not(), nil
		}
		return distinct, nil
	}

	left, err := translateAs[z3.Int](zt, expr.Left)
	if err != nil {
		return nil, err
	}
	right, err := translateAs[z3.Int](zt, expr.Right)
	if err != nil {
		return nil, err
	}
	switch expr.Operator {
	case symbolic.ADD:
		return left.Add(right), nil
	case symbolic.SUB:
		return left.Sub(right), nil
	case symbolic.DIV:
		return left.Div(right), nil
	case symbolic.MUL:
		return left.Mul(right), nil
	case symbolic.MOD:
		return left.Mod(right), nil
	case symbolic.LT:
		return left.LT(right), nil
	case symbolic.LE:
		return left.LE(right), nil
	case symbolic.GT:
		return left.GT(right), nil
	case symbolic.GE:
		return left.GE(right), nil
	default:
		return nil, NewTranslationError(fmt.Sprintf("unsupported binary operator: %s", expr.Operator), expr)
	}
}

// VisitLogicalOperation транслирует логическую операцию в Z3
func (zt *Z3Translator) VisitLogicalOperation(expr *symbolic.LogicalOperation) (interface{}, error) {
	// TODO: Реализовать
	// 1. Транслировать все операнды
	// 2. Применить соответствующую логическую операцию

	// Подсказки:
	// - AND: zt.ctx.And(operands...)
	// - OR: zt.ctx.Or(operands...)
	// - NOT: operand.Not() (для единственного операнда)
	// - IMPLIES: antecedent.Implies(consequent)

	operands, err := MapWithError(expr.Operands,
		func(operand symbolic.SymbolicExpression) (z3.Bool, error) {
			return translateAs[z3.Bool](zt, operand)
		})
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case symbolic.AND:
		return operands[0].And(operands[1:]...), nil
	case symbolic.OR:
		return operands[0].Or(operands[1:]...), nil
	case symbolic.NOT:
		return operands[0].Not(), nil
	case symbolic.IMPLIES:
		return operands[0].Implies(operands[1]), nil
	default:
		return nil, NewTranslationError(fmt.Sprintf("unsupported logical operator: %s", expr.Operator), expr)
	}
}

// VisitArrayAccess implements [symbolic.Visitor].
func (zt *Z3Translator) VisitArrayAccess(expr *symbolic.ArrayAccess) (interface{}, error) {
	array, err := translateAs[z3.Array](zt, expr.Array)
	if err != nil {
		return nil, err
	}
	index, err := translateAs[z3.Int](zt, expr.Index)
	if err != nil {
		return nil, err
	}
	return array.Select(index), nil
}

// VisitFunctionCall implements [symbolic.Visitor].
func (zt *Z3Translator) VisitFunctionCall(expr *symbolic.FunctionCall) (interface{}, error) {
	arguments, err := MapWithError(expr.Arguments, func(arg symbolic.SymbolicExpression) (z3.Value, error) {
		return translateAs[z3.Value](zt, arg)
	})
	if err != nil {
		return nil, err
	}

	domain := Map(expr.Arguments, func(arg symbolic.SymbolicExpression) z3.Sort {
		return zt.sortOf(arg.Type())
	})
	decl := zt.ctx.FuncDecl(expr.FunctionName, domain, zt.sortOf(expr.RetType))
	return decl.Apply(arguments...), nil
}

// VisitNegateOperation implements [symbolic.Visitor].
func (zt *Z3Translator) VisitNegateOperation(expr *symbolic.NegateOperation) (interface{}, error) {
	val, err := translateAs[z3.Int](zt, expr.Operand)
	if err != nil {
		return nil, err
	}
	return val.Neg(), nil
}

// VisitRef implements [symbolic.Visitor].
func (zt *Z3Translator) VisitRef(expr *symbolic.Ref) (interface{}, error) {
	return zt.ctx.FromInt(int64(expr.ID), zt.ctx.IntSort()), nil
}

// Вспомогательные методы

func (zt *Z3Translator) sortOf(exprType symbolic.ExpressionType) z3.Sort {
	switch exprType := exprType.(type) {
	case symbolic.IntT:
		return zt.ctx.IntSort()
	case symbolic.BoolT:
		return zt.ctx.BoolSort()
	case symbolic.ArrayT:
		return zt.ctx.ArraySort(zt.ctx.IntSort(), zt.sortOf(exprType.Elem))
	case symbolic.PtrT:
		return zt.ctx.IntSort()
	default:
		panic("unsupported expression type for sort: " + exprType.String())
	}
}

// createZ3Variable создаёт Z3 переменную соответствующего типа
func (zt *Z3Translator) createZ3Variable(name string, exprType symbolic.ExpressionType) z3.Value {
	// TODO: Реализовать (вспомогательный метод)
	// Создать Z3 переменную на основе типа
	return zt.ctx.Const(name, zt.sortOf(exprType))
}

// castToZ3Type приводит значение к нужному Z3 типу
func castToZ3Type[T z3.Value](value interface{}) (T, error) {
	v, ok := value.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("expected %T, got %T", zero, value)
	}
	return v, nil
}

// translateAs транслирует выражение и приводит результат к Z3 типу T
func translateAs[T z3.Value](zt *Z3Translator, expr symbolic.SymbolicExpression) (T, error) {
	v, err := expr.Accept(zt)
	if err != nil {
		var zero T
		return zero, err
	}
	res, err := castToZ3Type[T](v)
	if err != nil {
		return res, NewTranslationError(err.Error(), expr)
	}
	return res, nil
}
