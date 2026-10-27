// Package symbolic определяет базовые типы символьных выражений
package symbolic

// ExpressionType представляет тип символьного выражения
type ExpressionType interface {
    String() string
}

type IntT struct{}
type BoolT struct{}
type ArrayT struct{ Elem ExpressionType }
type PtrT   struct{ Elem ExpressionType }

func (IntT) String() string     { return "int" }
func (BoolT) String() string    { return "bool" }
func (a ArrayT) String() string { return "[]" + a.Elem.String() }
func (p PtrT) String() string   { return "*" + p.Elem.String() }

var (
    IntType  ExpressionType = IntT{}
    BoolType ExpressionType = BoolT{}
)
