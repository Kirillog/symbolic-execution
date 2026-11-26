package translator

import (
	"fmt"
	"math"

	"symbolic-execution-course/internal/symbolic"

	"github.com/ebukreev/go-z3/z3"
)

// Evaluate вычисляет выражение в модели и возвращает его значение как символьную константу.
func (zt *Z3Translator) Evaluate(model *z3.Model, expr symbolic.SymbolicExpression) (symbolic.SymbolicExpression, error) {
	translated, err := zt.translate(expr)
	if err != nil {
		return nil, err
	}
	value := model.Eval(translated.(z3.Value), true)
	if value == nil {
		return nil, NewTranslationError("cannot evaluate expression in model", expr)
	}

	switch literal := value.(type) {
	case z3.Bool:
		if v, ok := literal.AsBool(); ok {
			return symbolic.NewBoolConstant(v), nil
		}
	case z3.Int:
		if expr.Type() == symbolic.StringType {
			if v, ok := zt.strings.resolve(literal); ok {
				return &symbolic.StringConstant{Value: v}, nil
			}
		} else if v, isLiteral, ok := literal.AsInt64(); isLiteral && ok {
			return integerConstant(v, expr.Type()), nil
		}
	case z3.BV:
		if v, isLiteral, ok := literal.AsInt64(); isLiteral && ok {
			return integerConstant(v, expr.Type()), nil
		}
	case z3.Float:
		if v, isLiteral := literal.AsBigFloat(); isLiteral {
			if v == nil {
				return symbolic.NewFloatConstant(math.NaN()), nil
			}
			f, _ := v.Float64()
			return symbolic.NewFloatConstant(f), nil
		}
	}
	return nil, NewTranslationError(fmt.Sprintf("model value %s is not a literal of type %s", value, expr.Type()), expr)
}

func integerConstant(value int64, exprType symbolic.ExpressionType) symbolic.SymbolicExpression {
	constant := symbolic.NewIntConstant(value)
	if typ, ok := exprType.(symbolic.PtrT); ok {
		return &symbolic.Ref{Addr: constant, ExprType: typ.Elem}
	}
	return constant
}
