// Package symbolic определяет базовые типы символьных выражений
package symbolic

import (
	"strings"
	"symbolic-execution-course/internal/bslices"
)

// ExpressionType представляет тип символьного выражения
type ExpressionType interface {
	String() string
}

type IntT struct{}
type AddrT struct{}
type BoolT struct{}
type FloatT struct{}
type ArrayT struct {
	Elem ExpressionType
	Index ExpressionType
}

func (a ArrayT) IndexType() ExpressionType {
	if a.Index == nil {
		return IntType
	}
	return a.Index
}

type PtrT struct{ Elem ExpressionType }
type StructT struct {
	Name   string
	Fields []ExpressionType
}

func (IntT) String() string     { return "int" }
func (AddrT) String() string    { return "addr" }
func (BoolT) String() string    { return "bool" }
func (FloatT) String() string   { return "float64" }
func (a ArrayT) String() string { return "[]" + a.Elem.String() }
func (p PtrT) String() string   { return "*" + p.Elem.String() }
func (s *StructT) String() string {
	fields := bslices.Map(s.Fields, func(t ExpressionType) string { return t.String() })
	return s.Name + "{" + strings.Join(fields, ", ") + "}"
}

var (
	IntType   ExpressionType = IntT{}
	AddrType  ExpressionType = AddrT{}
	BoolType  ExpressionType = BoolT{}
	FloatType ExpressionType = FloatT{}
)
