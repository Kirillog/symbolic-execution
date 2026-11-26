package memory

import (
	"fmt"
	"slices"

	"symbolic-execution-course/internal/symbolic"
)

type Address struct {
	Ref     *symbolic.Ref
	Elem    symbolic.ExpressionType
	Slot    int
	Indices []symbolic.SymbolicExpression
}

func Root(ref *symbolic.Ref) Address { return Address{Ref: ref, Elem: ref.ExprType} }

func (a Address) Field(index int) Address {
	fields := a.Elem.(*symbolic.StructT).Fields
	a.Slot += fieldSlot(fields, index)
	a.Elem = fields[index]
	return a
}

func (a Address) Element(index symbolic.SymbolicExpression) Address {
	a.Elem = a.Elem.(symbolic.ArrayT).Elem
	a.Indices = append(slices.Clip(a.Indices), index)
	return a
}

type StructValue struct {
	tpe    *symbolic.StructT
	fields []symbolic.SymbolicExpression
}

func (v *StructValue) Type() symbolic.ExpressionType { return v.tpe }
func (v *StructValue) String() string                { return fmt.Sprint(v.fields) }
func (v *StructValue) Accept(symbolic.ExpressionVisitorError) (interface{}, error) {
	return nil, fmt.Errorf("struct %s requires a field access", v.tpe)
}

func (v *StructValue) Field(index int) symbolic.SymbolicExpression {
	fields := v.tpe.Fields
	field := fields[index]
	start := fieldSlot(fields, index)
	return pack(field, v.fields[start:start+len(columnTypes(field))])
}

type ArrayValue struct {
	tpe      symbolic.ArrayT
	elements []symbolic.SymbolicExpression
}

func (v *ArrayValue) Type() symbolic.ExpressionType { return v.tpe }
func (v *ArrayValue) String() string                { return fmt.Sprint(v.elements) }
func (v *ArrayValue) Accept(symbolic.ExpressionVisitorError) (interface{}, error) {
	return nil, fmt.Errorf("array %s requires an element access", v.tpe)
}

func (v *ArrayValue) Element(index symbolic.SymbolicExpression) symbolic.SymbolicExpression {
	values := make([]symbolic.SymbolicExpression, len(v.elements))
	for i, column := range v.elements {
		values[i] = selectArray(column, index)
	}
	return pack(v.tpe.Elem, values)
}

func fieldSlot(fields []symbolic.ExpressionType, index int) int {
	slot := 0
	for _, field := range fields[:index] {
		slot += len(columnTypes(field))
	}
	return slot
}

func columnTypes(tpe symbolic.ExpressionType) []symbolic.ExpressionType {
	switch tpe := tpe.(type) {
	case *symbolic.StructT:
		var result []symbolic.ExpressionType
		for _, field := range tpe.Fields {
			result = append(result, columnTypes(field)...)
		}
		return result
	case symbolic.ArrayT:
		result := columnTypes(tpe.Elem)
		for i, elem := range result {
			result[i] = symbolic.ArrayT{Elem: elem, Index: tpe.Index}
		}
		return result
	default:
		return []symbolic.ExpressionType{tpe}
	}
}

func columns(value symbolic.SymbolicExpression) []symbolic.SymbolicExpression {
	switch value := value.(type) {
	case *StructValue:
		return value.fields
	case *ArrayValue:
		return value.elements
	default:
		return []symbolic.SymbolicExpression{value}
	}
}

func pack(tpe symbolic.ExpressionType, values []symbolic.SymbolicExpression) symbolic.SymbolicExpression {
	switch tpe := tpe.(type) {
	case *symbolic.StructT:
		return &StructValue{tpe: tpe, fields: values}
	case symbolic.ArrayT:
		return &ArrayValue{tpe: tpe, elements: values}
	default:
		return values[0]
	}
}

func Zero(tpe symbolic.ExpressionType) symbolic.SymbolicExpression {
	types := columnTypes(tpe)
	values := make([]symbolic.SymbolicExpression, len(types))
	for i, column := range types {
		values[i] = zeroColumn(column)
	}
	return pack(tpe, values)
}

func zeroColumn(tpe symbolic.ExpressionType) symbolic.SymbolicExpression {
	switch tpe := tpe.(type) {
	case symbolic.IntT:
		return symbolic.NewIntConstant(0)
	case symbolic.BoolT:
		return symbolic.NewBoolConstant(false)
	case symbolic.FloatT:
		return symbolic.NewFloatConstant(0)
	case symbolic.StringT:
		return &symbolic.StringConstant{}
	case symbolic.PtrT:
		return &symbolic.Ref{Addr: symbolic.NullAddr, ExprType: tpe.Elem}
	case symbolic.ArrayT:
		return &symbolic.ArrayConstant{Value: zeroColumn(tpe.Elem), Index: tpe.Index}
	default:
		panic(fmt.Sprintf("unsupported zero value for %s", tpe))
	}
}
