package internal

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"maps"
	"slices"
	"symbolic-execution-course/internal/memory"
	"symbolic-execution-course/internal/symbolic"

	"golang.org/x/tools/go/ssa"
)

type Interpreter struct {
	CallStack     []CallStackFrame
	Analyser      *Analyser
	PathCondition symbolic.SymbolicExpression
	Heap          memory.Memory
	Panic         *PanicInfo
}

type PanicInfo struct {
	Message string
}

type CallStackFrame struct {
	Function    *ssa.Function
	LocalMemory map[string]symbolic.SymbolicExpression
	ReturnValue symbolic.SymbolicExpression

	Block     *ssa.BasicBlock
	PrevBlock *ssa.BasicBlock
	InstrIdx  int

	BlockVisits map[*ssa.BasicBlock]int
}

func (frame *CallStackFrame) enter(block *ssa.BasicBlock) {
	frame.PrevBlock = frame.Block
	frame.Block = block
	frame.InstrIdx = 0
	frame.BlockVisits[block]++
}

func (interpreter Interpreter) Clone() Interpreter {
	clone := interpreter
	clone.CallStack = make([]CallStackFrame, len(interpreter.CallStack))
	for i, frame := range interpreter.CallStack {
		frame.LocalMemory = maps.Clone(frame.LocalMemory)
		frame.BlockVisits = maps.Clone(frame.BlockVisits)
		clone.CallStack[i] = frame
	}
	return clone
}

func symbolicType(tpe types.Type) symbolic.ExpressionType {
	basic, ok := tpe.Underlying().(*types.Basic)
	if !ok {
		panic(fmt.Sprintf("unsupported type %s", tpe))
	}
	switch {
	case basic.Info() & types.IsInteger != 0:
		return symbolic.IntType
	case basic.Info() & types.IsBoolean != 0:
		return symbolic.BoolType
	case basic.Info() & types.IsFloat != 0:
		return symbolic.FloatType
	}
	panic(fmt.Sprintf("unsupported type %s", tpe))
}

var binaryOperators = map[token.Token]symbolic.BinaryOperator{
	token.ADD: symbolic.ADD,
	token.SUB: symbolic.SUB,
	token.MUL: symbolic.MUL,
	token.QUO: symbolic.DIV,
	token.REM: symbolic.MOD,
	token.EQL: symbolic.EQ,
	token.NEQ: symbolic.NE,
	token.LSS: symbolic.LT,
	token.LEQ: symbolic.LE,
	token.GTR: symbolic.GT,
	token.GEQ: symbolic.GE,
	token.AND: symbolic.BAND,
	token.OR:  symbolic.BOR,
	token.XOR: symbolic.XOR,
	token.SHL: symbolic.SHL,
	token.SHR: symbolic.SHR,
}

func (interpreter *Interpreter) interpretDynamically(element ssa.Instruction) []Interpreter {
	frame := &interpreter.CallStack[len(interpreter.CallStack)-1]
	frame.InstrIdx++

	switch instr := element.(type) {
	case *ssa.BinOp:
		op, ok := binaryOperators[instr.Op]
		if !ok {
			panic(fmt.Sprintf("unsupported binary operator %s", instr.Op))
		}
		left, right := interpreter.resolveExpression(instr.X), interpreter.resolveExpression(instr.Y)
		frame.LocalMemory[instr.Name()] = symbolic.NewBinaryOperation(left, right, op)

		if (op == symbolic.DIV || op == symbolic.MOD) && right.Type() == symbolic.IntType && !isNonZeroConstant(right) {
			isZero := symbolic.NewBinaryOperation(right, symbolic.NewIntConstant(0), symbolic.EQ)
			notZero := symbolic.NewLogicalOperation([]symbolic.SymbolicExpression{isZero}, symbolic.NOT)

			panicked := interpreter.assume(isZero)
			panicked.Panic = &PanicInfo{Message: "integer divide by zero"}
			panicked.CallStack[len(panicked.CallStack)-1].Block = nil
			return []Interpreter{interpreter.assume(notZero), panicked}
		}
	case *ssa.UnOp:
		operand := interpreter.resolveExpression(instr.X)
		switch instr.Op {
		case token.SUB:
			frame.LocalMemory[instr.Name()] = &symbolic.NegateOperation{Operand: operand}
		case token.NOT:
			frame.LocalMemory[instr.Name()] = symbolic.NewLogicalOperation([]symbolic.SymbolicExpression{operand}, symbolic.NOT)
		case token.XOR:
			frame.LocalMemory[instr.Name()] = symbolic.NewBinaryOperation(operand, symbolic.NewIntConstant(-1), symbolic.XOR)
		default:
			panic(fmt.Sprintf("unsupported unary operator %s", instr.Op))
		}
	case *ssa.Jump:
		frame.enter(frame.Block.Succs[0])
	case *ssa.Phi:
		// все phi блока выполняются одновременно: сначала читаем значения, потом пишем,
		// иначе phi, читающий другой phi того же блока, увидит уже новое значение
		edge := slices.Index(frame.Block.Preds, frame.PrevBlock)
		if edge < 0 {
			panic(fmt.Sprintf("block %s was not entered from %s", frame.Block, frame.PrevBlock))
		}
		var phis []*ssa.Phi
		idx := frame.InstrIdx - 1
		for ; idx < len(frame.Block.Instrs); idx++ {
			phi, ok := frame.Block.Instrs[idx].(*ssa.Phi)
			if !ok {
				break
			}
			phis = append(phis, phi)
		}
		values := make([]symbolic.SymbolicExpression, len(phis))
		for i, phi := range phis {
			values[i] = interpreter.resolveExpression(phi.Edges[edge])
		}
		for i, phi := range phis {
			frame.LocalMemory[phi.Name()] = values[i]
		}
		frame.InstrIdx = idx
	case *ssa.Return:
		switch len(instr.Results) {
		case 0:
		case 1:
			frame.ReturnValue = interpreter.resolveExpression(instr.Results[0])
		default:
			panic("multiple return values are not supported")
		}
		frame.Block = nil
	case *ssa.MakeInterface:
	case *ssa.Panic:
		interpreter.Panic = &PanicInfo{Message: panicMessage(instr.X)}
		frame.Block = nil
	case *ssa.If:
		condition := interpreter.resolveExpression(instr.Cond)
		negation := symbolic.NewLogicalOperation([]symbolic.SymbolicExpression{condition}, symbolic.NOT)
		return []Interpreter{
			interpreter.branch(condition, frame.Block.Succs[0]),
			interpreter.branch(negation, frame.Block.Succs[1]),
		}
	default:
		panic(fmt.Sprintf("unsupported instruction %s", element))
	}
	return []Interpreter{*interpreter}
}

func (interpreter *Interpreter) assume(condition symbolic.SymbolicExpression) Interpreter {
	clone := interpreter.Clone()
	clone.PathCondition = symbolic.NewLogicalOperation(
		[]symbolic.SymbolicExpression{interpreter.PathCondition, condition}, symbolic.AND)
	return clone
}

func (interpreter *Interpreter) branch(condition symbolic.SymbolicExpression, target *ssa.BasicBlock) Interpreter {
	clone := interpreter.assume(condition)
	frame := &clone.CallStack[len(clone.CallStack)-1]
	frame.enter(target)
	return clone
}

func isNonZeroConstant(expr symbolic.SymbolicExpression) bool {
	constant, ok := expr.(*symbolic.IntConstant)
	return ok && constant.Value != 0
}

func panicMessage(value ssa.Value) string {
	makeInterface, ok := value.(*ssa.MakeInterface)
	if !ok {
		return ""
	}
	if text, ok := makeInterface.X.(*ssa.Const); ok && text.Value != nil && text.Value.Kind() == constant.String {
		return constant.StringVal(text.Value)
	}
	return ""
}

func (interpreter *Interpreter) resolveExpression(value ssa.Value) symbolic.SymbolicExpression {
	switch value := value.(type) {
	case *ssa.Const:
		switch symbolicType(value.Type()) {
		case symbolic.IntType:
			return symbolic.NewIntConstant(value.Int64())
		case symbolic.BoolType:
			return symbolic.NewBoolConstant(constant.BoolVal(value.Value))
		case symbolic.FloatType:
			float, _ := constant.Float64Val(value.Value)
			return symbolic.NewFloatConstant(float)
		}
	default:
		frame := interpreter.CallStack[len(interpreter.CallStack)-1]
		if expr, ok := frame.LocalMemory[value.Name()]; ok {
			return expr
		}
	}
	panic(fmt.Sprintf("cannot resolve value %s", value))
}
