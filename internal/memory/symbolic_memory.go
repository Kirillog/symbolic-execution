package memory

import (
	"maps"
	"sync"

	"symbolic-execution-course/internal/symbolic"
)

type Memory interface {
	Allocate(tpe symbolic.ExpressionType) *symbolic.Ref
	AssignField(ref *symbolic.Ref, fieldIdx int, value symbolic.SymbolicExpression)
	GetFieldValue(ref *symbolic.Ref, fieldIdx int) symbolic.SymbolicExpression
	AssignToArray(ref *symbolic.Ref, index int, value symbolic.SymbolicExpression)
	GetFromArray(ref *symbolic.Ref, index int) symbolic.SymbolicExpression
}

type FieldKey struct {
	Type symbolic.ExpressionType
	Idx  int
}

type SymbolicMemory struct {
	regions map[FieldKey]symbolic.SymbolicExpression
	inputs   *sync.Map
	nextAddr int64
}

func NewSymbolicMemory() SymbolicMemory {
	return SymbolicMemory{
		regions: make(map[FieldKey]symbolic.SymbolicExpression),
		inputs:  &sync.Map{},
	}
}

func (mem *SymbolicMemory) Clone() *SymbolicMemory {
	clone := *mem
	clone.regions = maps.Clone(mem.regions)
	return &clone
}

func (mem *SymbolicMemory) Allocate(tpe symbolic.ExpressionType) *symbolic.Ref {
	mem.nextAddr++
	ref := &symbolic.Ref{Addr: symbolic.NewIntConstant(mem.nextAddr), ExprType: tpe}
	mem.Write(Root(ref), Zero(tpe))
	return ref
}

func (mem *SymbolicMemory) AssignField(ref *symbolic.Ref, fieldIdx int, value symbolic.SymbolicExpression) {
	mem.Write(Root(ref).Field(fieldIdx), value)
}

func (mem *SymbolicMemory) GetFieldValue(ref *symbolic.Ref, fieldIdx int) symbolic.SymbolicExpression {
	return mem.Read(Root(ref).Field(fieldIdx))
}

func (mem *SymbolicMemory) AssignToArray(ref *symbolic.Ref, index int, value symbolic.SymbolicExpression) {
	mem.Write(Root(ref).Element(symbolic.NewIntConstant(int64(index))), value)
}

func (mem *SymbolicMemory) GetFromArray(ref *symbolic.Ref, index int) symbolic.SymbolicExpression {
	return mem.Read(Root(ref).Element(symbolic.NewIntConstant(int64(index))))
}

func (mem *SymbolicMemory) Read(address Address) symbolic.SymbolicExpression {
	values := make([]symbolic.SymbolicExpression, len(columnTypes(address.Elem)))
	for i := range values {
		key := FieldKey{Type: address.Ref.ExprType, Idx: address.Slot + i}
		value := selectArray(mem.region(key), address.Ref.Addr)
		for _, index := range address.Indices {
			value = selectArray(value, index)
		}
		values[i] = value
	}
	return pack(address.Elem, values)
}

func (mem *SymbolicMemory) Write(address Address, value symbolic.SymbolicExpression) {
	indices := append([]symbolic.SymbolicExpression{address.Ref.Addr}, address.Indices...)
	for i, value := range columns(value) {
		key := FieldKey{Type: address.Ref.ExprType, Idx: address.Slot + i}
		mem.regions[key] = store(mem.region(key), indices, value)
	}
}

func (mem *SymbolicMemory) region(key FieldKey) symbolic.SymbolicExpression {
	if region, ok := mem.regions[key]; ok {
		return region
	}
	if region, ok := mem.inputs.Load(key); ok {
		return region.(symbolic.SymbolicExpression)
	}
	member := columnTypes(key.Type)[key.Idx]
	fresh := symbolic.NewFreshSymbolicVariable("heap", symbolic.ArrayT{Elem: member})
	region, _ := mem.inputs.LoadOrStore(key, fresh)
	return region.(symbolic.SymbolicExpression)
}

func store(array symbolic.SymbolicExpression, indices []symbolic.SymbolicExpression, value symbolic.SymbolicExpression) symbolic.SymbolicExpression {
	index := indices[0]
	if len(indices) > 1 {
		value = store(selectArray(array, index), indices[1:], value)
	}
	return &symbolic.ArrayStore{Array: array, Index: index, Value: value}
}

func selectArray(array, index symbolic.SymbolicExpression) symbolic.SymbolicExpression {
	for {
		switch node := array.(type) {
		case *symbolic.ArrayConstant:
			return node.Value
		case *symbolic.ArrayStore:
			if node.Index == index {
				return node.Value
			}
			writtenAt, writtenKnown := node.Index.(*symbolic.IntConstant)
			readAt, readKnown := index.(*symbolic.IntConstant)
			if writtenKnown && readKnown {
				if writtenAt.Value == readAt.Value {
					return node.Value
				}
				array = node.Array
				continue
			}
		}
		return &symbolic.ArrayAccess{Array: array, Index: index}
	}
}
