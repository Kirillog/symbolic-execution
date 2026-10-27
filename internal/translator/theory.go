package translator

import (
	"fmt"

	"symbolic-execution-course/internal/symbolic"

	"github.com/ebukreev/go-z3/z3"
)

// Theory инкапсулирует всё, что зависит от выбранной SMT-теории для целых чисел.
// Остальная часть транслятора (булевы операции, массивы, вызовы функций) от теории не зависит.
type Theory interface {
	// Sort возвращает сорт, в который отображается symbolic.IntType (и указатели).
	Sort() z3.Sort

	// Const создаёт целочисленную константу.
	Const(value int64) z3.Value

	// Binary транслирует арифметическую операцию или сравнение над целыми.
	// Операторы EQ и NE обрабатываются транслятором и сюда не попадают.
	Binary(op symbolic.BinaryOperator, left, right z3.Value) (z3.Value, error)

	// Neg транслирует унарный минус.
	Neg(value z3.Value) (z3.Value, error)
}

// IntTheory реализует теорию линейной целочисленной арифметики (z3.Int).
type IntTheory struct {
	ctx *z3.Context
}

// NewIntTheory создаёт теорию целых чисел в заданном Z3 контексте.
func NewIntTheory(ctx *z3.Context) *IntTheory {
	return &IntTheory{ctx: ctx}
}

func (t *IntTheory) Sort() z3.Sort {
	return t.ctx.IntSort()
}

func (t *IntTheory) Const(value int64) z3.Value {
	return t.ctx.FromInt(value, t.ctx.IntSort())
}

func (t *IntTheory) Binary(op symbolic.BinaryOperator, left, right z3.Value) (z3.Value, error) {
	l, err := castToZ3Type[z3.Int](left)
	if err != nil {
		return nil, err
	}
	r, err := castToZ3Type[z3.Int](right)
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
	case symbolic.MOD:
		return l.Mod(r), nil
	case symbolic.LT:
		return l.LT(r), nil
	case symbolic.LE:
		return l.LE(r), nil
	case symbolic.GT:
		return l.GT(r), nil
	case symbolic.GE:
		return l.GE(r), nil
	default:
		return nil, fmt.Errorf("unsupported binary operator for int theory: %s", op)
	}
}

func (t *IntTheory) Neg(value z3.Value) (z3.Value, error) {
	v, err := castToZ3Type[z3.Int](value)
	if err != nil {
		return nil, err
	}
	return v.Neg(), nil
}
