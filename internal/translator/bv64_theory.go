package translator

import (
	"fmt"

	"symbolic-execution-course/internal/symbolic"

	"github.com/ebukreev/go-z3/z3"
)

const bv64Width = 64

// BV64Theory реализует int как знаковый 64-битный вектор: арифметика оборачивается
// при переполнении, / и % усекают к нулю, а сдвиги ведут себя как в Go.
type BV64Theory struct {
	ctx *z3.Context
}

var _ Theory = (*BV64Theory)(nil)

// NewBV64Theory создаёт теорию 64-битных векторов в заданном Z3 контексте.
func NewBV64Theory(ctx *z3.Context) *BV64Theory {
	return &BV64Theory{ctx: ctx}
}

func (t *BV64Theory) Sort() z3.Sort {
	return t.ctx.BVSort(bv64Width)
}

func (t *BV64Theory) Const(value int64) z3.Value {
	return t.ctx.FromInt(value, t.ctx.BVSort(bv64Width))
}

func (t *BV64Theory) Binary(op symbolic.BinaryOperator, left, right z3.Value) (z3.Value, error) {
	l, err := castToZ3Type[z3.BV](left)
	if err != nil {
		return nil, err
	}
	r, err := castToZ3Type[z3.BV](right)
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
		return l.SDiv(r), nil
	case symbolic.MOD:
		return l.SRem(r), nil
	case symbolic.LT:
		return l.SLT(r), nil
	case symbolic.LE:
		return l.SLE(r), nil
	case symbolic.GT:
		return l.SGT(r), nil
	case symbolic.GE:
		return l.SGE(r), nil
	case symbolic.BAND:
		return l.And(r), nil
	case symbolic.BOR:
		return l.Or(r), nil
	case symbolic.XOR:
		return l.Xor(r), nil
	case symbolic.SHL:
		return l.Lsh(r), nil
	case symbolic.SHR:
		return l.SRsh(r), nil
	default:
		return nil, fmt.Errorf("unsupported binary operator for bv64 theory: %s", op)
	}
}

func (t *BV64Theory) Neg(value z3.Value) (z3.Value, error) {
	v, err := castToZ3Type[z3.BV](value)
	if err != nil {
		return nil, err
	}
	return v.Neg(), nil
}
