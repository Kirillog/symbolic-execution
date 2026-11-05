package translator

import (
	"fmt"

	"symbolic-execution-course/internal/symbolic"

	"github.com/ebukreev/go-z3/z3"
)

// FloatTheory инкапсулирует всё, что зависит от выбранной SMT-теории для float64.
type FloatTheory interface {
	// Sort возвращает сорт, в который отображается symbolic.FloatType.
	Sort() z3.Sort

	// Const создаёт константу float64.
	Const(value float64) z3.Value

	// Binary транслирует арифметическую операцию или сравнение над float64.
	// Операторы EQ и NE выражаются через Equal и сюда не попадают.
	Binary(op symbolic.BinaryOperator, left, right z3.Value) (z3.Value, error)

	// Neg транслирует унарный минус.
	Neg(value z3.Value) (z3.Value, error)

	// Equal транслирует равенство == языка Go; операнды должны иметь сорт Sort.
	Equal(left, right z3.Value) z3.Bool
}

// Параметры IEEE 754 binary64: 11 бит экспоненты, 53 бита мантиссы (с неявной единицей).
const (
	float64ExponentBits    = 11
	float64SignificandBits = 53
)

// FP64Theory реализует float64 как IEEE-число binary64. Режим округления - режим контекста
// по умолчанию (к ближайшему чётному), как в Go. Равенство - IEEE: NaN != NaN, 0.0 == -0.0.
type FP64Theory struct {
	ctx *z3.Context
}

var _ FloatTheory = (*FP64Theory)(nil)

// NewFP64Theory создаёт теорию IEEE binary64 в заданном Z3 контексте.
func NewFP64Theory(ctx *z3.Context) *FP64Theory {
	return &FP64Theory{ctx: ctx}
}

func (t *FP64Theory) Sort() z3.Sort {
	return t.ctx.FloatSort(float64ExponentBits, float64SignificandBits)
}

func (t *FP64Theory) Const(value float64) z3.Value {
	return t.ctx.FromFloat64(value, t.Sort())
}

func (t *FP64Theory) Binary(op symbolic.BinaryOperator, left, right z3.Value) (z3.Value, error) {
	l, err := castToZ3Type[z3.Float](left)
	if err != nil {
		return nil, err
	}
	r, err := castToZ3Type[z3.Float](right)
	if err != nil {
		return nil, err
	}
	switch op {
	case symbolic.ADD:
		return l.Add(r), nil
	case symbolic.SUB:
		return l.Sub(r), nil
	case symbolic.MUL:
		return l.Mul(r), nil
	case symbolic.DIV:
		return l.Div(r), nil
	case symbolic.LT:
		return l.LT(r), nil
	case symbolic.LE:
		return l.LE(r), nil
	case symbolic.GT:
		return l.GT(r), nil
	case symbolic.GE:
		return l.GE(r), nil
	default:
		return nil, fmt.Errorf("unsupported binary operator for float64 theory: %s", op)
	}
}

func (t *FP64Theory) Neg(value z3.Value) (z3.Value, error) {
	v, err := castToZ3Type[z3.Float](value)
	if err != nil {
		return nil, err
	}
	return v.Neg(), nil
}

func (t *FP64Theory) Equal(left, right z3.Value) z3.Bool {
	return left.(z3.Float).IEEEEq(right.(z3.Float))
}
