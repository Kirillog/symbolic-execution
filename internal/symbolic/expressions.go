// Package symbolic содержит конкретные реализации символьных выражений
package symbolic

import (
	"fmt"
	"strings"
	"symbolic-execution-course/internal/bslices"
)

// SymbolicExpression - базовый интерфейс для всех символьных выражений
type SymbolicExpression interface {
	// Type возвращает тип выражения
	Type() ExpressionType

	// String возвращает строковое представление выражения
	String() string

	// Accept принимает visitor для обхода дерева выражений
	Accept(visitor Visitor) (interface{}, error)
}

// SymbolicVariable представляет символьную переменную
type SymbolicVariable struct {
	Name     string
	ExprType ExpressionType
}

// NewSymbolicVariable создаёт новую символьную переменную
func NewSymbolicVariable(name string, exprType ExpressionType) *SymbolicVariable {
	return &SymbolicVariable{
		Name:     name,
		ExprType: exprType,
	}
}

// Type возвращает тип переменной
func (sv *SymbolicVariable) Type() ExpressionType {
	return sv.ExprType
}

// String возвращает строковое представление переменной
func (sv *SymbolicVariable) String() string {
	return sv.Name
}

// Accept реализует Visitor pattern
func (sv *SymbolicVariable) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitVariable(sv)
}

// IntConstant представляет целочисленную константу
type IntConstant struct {
	Value int64
}

// NewIntConstant создаёт новую целочисленную константу
func NewIntConstant(value int64) *IntConstant {
	return &IntConstant{Value: value}
}

// Type возвращает тип константы
func (ic *IntConstant) Type() ExpressionType {
	return IntType
}

// String возвращает строковое представление константы
func (ic *IntConstant) String() string {
	return fmt.Sprintf("%d", ic.Value)
}

// Accept реализует Visitor pattern
func (ic *IntConstant) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitIntConstant(ic)
}

// BoolConstant представляет булеву константу
type BoolConstant struct {
	Value bool
}

// NewBoolConstant создаёт новую булеву константу
func NewBoolConstant(value bool) *BoolConstant {
	return &BoolConstant{Value: value}
}

// Type возвращает тип константы
func (bc *BoolConstant) Type() ExpressionType {
	return BoolType
}

// String возвращает строковое представление константы
func (bc *BoolConstant) String() string {
	return fmt.Sprintf("%t", bc.Value)
}

// Accept реализует Visitor pattern
func (bc *BoolConstant) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitBoolConstant(bc)
}

// BinaryOperation представляет бинарную операцию
type BinaryOperation struct {
	Left     SymbolicExpression
	Right    SymbolicExpression
	Operator BinaryOperator
}

// TODO: Реализуйте следующие методы в рамках домашнего задания

// NewBinaryOperation создаёт новую бинарную операцию
func NewBinaryOperation(left, right SymbolicExpression, op BinaryOperator) *BinaryOperation {
	// TODO: Реализовать
	// Создать новую бинарную операцию и проверить совместимость типов
	lt, rt := left.Type(), right.Type()
	switch op {
	case ADD, SUB, MUL, DIV, MOD, LT, LE, GT, GE:
		if lt != IntType || rt != IntType {
			panic(fmt.Sprintf("type mismatch: %s %s %s", lt, op, rt))
		}
	case EQ, NE:
		if lt != rt {
			panic(fmt.Sprintf("type mismatch: %s %s %s", lt, op, rt))
		}
	}
	return &BinaryOperation{left, right, op}
}

// Type возвращает результирующий тип операции
func (bo *BinaryOperation) Type() ExpressionType {
	// TODO: Реализовать
	// Определить результирующий тип на основе операции и типов операндов
	// Например: int + int = int, int < int = bool
	switch bo.Operator {
	case ADD, SUB, MUL, DIV, MOD:
		return IntType
	case LT, LE, GT, GE:
		return BoolType
	case EQ, NE:
		return bo.Left.Type()
	}
	panic("unuspported operator " + bo.Operator.String())
}

// String возвращает строковое представление операции
func (bo *BinaryOperation) String() string {
	// TODO: Реализовать
	// Формат: "(left operator right)"
	return fmt.Sprintf("(%s %s %s)", bo.Left, bo.Operator, bo.Right)
}

// Accept реализует Visitor pattern
func (bo *BinaryOperation) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitBinaryOperation(bo)
}

type NegateOperation struct {
	Operand SymbolicExpression
}

func (no *NegateOperation) Type() ExpressionType {
	return no.Operand.Type()
}

func (no *NegateOperation) String() string {
	return fmt.Sprintf("-%s", no.Operand)
}

func (no *NegateOperation) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitNegateOperation(no)
}

// LogicalOperation представляет логическую операцию
type LogicalOperation struct {
	Operands []SymbolicExpression
	Operator LogicalOperator
}

// TODO: Реализуйте следующие методы в рамках домашнего задания

// NewLogicalOperation создаёт новую логическую операцию
func NewLogicalOperation(operands []SymbolicExpression, op LogicalOperator) *LogicalOperation {
	// TODO: Реализовать
	// Создать логическую операцию и проверить типы операндов
	if !bslices.All(operands, func(op SymbolicExpression) bool {
		return op.Type() == BoolType
	}) {
		panic("all operands must be of boolean type")
	}
	return &LogicalOperation{operands, op}
}

// Type возвращает тип логической операции (всегда bool)
func (lo *LogicalOperation) Type() ExpressionType {
	return BoolType
}

// String возвращает строковое представление логической операции
func (lo *LogicalOperation) String() string {
	// TODO: Реализовать
	// Для NOT: "!operand"
	// Для AND/OR: "(operand1 && operand2 && ...)"
	// Для IMPLIES: "(operand1 => operand2)"
	switch lo.Operator {
	case NOT:
		return fmt.Sprintf("!%s", lo.Operands[0])
	case AND:
		return strings.Join(bslices.Map(lo.Operands, func(op SymbolicExpression) string { return op.String() }), " && ")
	case OR:
		return strings.Join(bslices.Map(lo.Operands, func(op SymbolicExpression) string { return op.String() }), " || ")
	case IMPLIES:
		return fmt.Sprintf("(%s => %s)", lo.Operands[0], lo.Operands[1])
	default:
		panic("unsupported logical operator " + lo.Operator.String())
	}
}

// Accept реализует Visitor pattern
func (lo *LogicalOperation) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitLogicalOperation(lo)
}

// Операторы для бинарных выражений
type BinaryOperator int

const (
	// Арифметические операторы
	ADD BinaryOperator = iota
	SUB
	MUL
	DIV
	MOD

	// Операторы сравнения
	EQ // равно
	NE // не равно
	LT // меньше
	LE // меньше или равно
	GT // больше
	GE // больше или равно
)

// String возвращает строковое представление оператора
func (op BinaryOperator) String() string {
	switch op {
	case ADD:
		return "+"
	case SUB:
		return "-"
	case MUL:
		return "*"
	case DIV:
		return "/"
	case MOD:
		return "%"
	case EQ:
		return "=="
	case NE:
		return "!="
	case LT:
		return "<"
	case LE:
		return "<="
	case GT:
		return ">"
	case GE:
		return ">="
	default:
		return "unknown"
	}
}

// Логические операторы
type LogicalOperator int

const (
	AND LogicalOperator = iota
	OR
	NOT
	IMPLIES
)

// String возвращает строковое представление логического оператора
func (op LogicalOperator) String() string {
	switch op {
	case AND:
		return "&&"
	case OR:
		return "||"
	case NOT:
		return "!"
	case IMPLIES:
		return "=>"
	default:
		return "unknown"
	}
}

type Ref struct {
	// TODO: Выбрать и написать внутреннее представление символьной ссылки
	ID       int64
	ExprType ExpressionType
}

func (ref *Ref) Type() ExpressionType {
	return PtrT{Elem: ref.ExprType}
}

func (ref *Ref) String() string {
	return fmt.Sprintf("ref(%d)", ref.ID)
}

func (ref *Ref) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitRef(ref)
}

type ArrayAccess struct {
	Array SymbolicExpression
	Index SymbolicExpression
}

func (aa *ArrayAccess) Type() ExpressionType {
	return aa.Array.Type().(ArrayT).Elem
}

func (aa *ArrayAccess) String() string {
	return fmt.Sprintf("%s[%s]", aa.Array, aa.Index)
}

func (aa *ArrayAccess) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitArrayAccess(aa)
}

type FunctionCall struct {
	FunctionName string
	Arguments    []SymbolicExpression
	RetType      ExpressionType
}

func (fc *FunctionCall) Type() ExpressionType {
	return fc.RetType
}

func (fc *FunctionCall) String() string {
	args := bslices.Map(fc.Arguments, func(arg SymbolicExpression) string { return arg.String() })
	return fmt.Sprintf("%s(%s)", fc.FunctionName, strings.Join(args, ", "))
}

func (fc *FunctionCall) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitFunctionCall(fc)
}

// TODO: Добавьте дополнительные типы выражений по необходимости:
// - UnaryOperation (унарные операции: -x, !x)
// - ArrayAccess (доступ к элементам массива: arr[index])
// - FunctionCall (вызовы функций: f(x, y))
// - ConditionalExpression (тернарный оператор: condition ? true_expr : false_expr)
