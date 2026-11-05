package internal

import (
	"container/heap"
	"symbolic-execution-course/internal/ssa"
	"symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"

	xssa "golang.org/x/tools/go/ssa"
)

type Analyser struct {
	Function     *xssa.Function
	StatesQueue  PriorityQueue
	PathSelector PathSelector
	Results      []Interpreter
	satisfiable  map[symbolic.SymbolicExpression]bool
	sessions     map[sessionOpt]*solverSession
}

type sessionOpt struct {
	bitvectors bool
	floats     bool
}

const maxSteps = 10000
const maxBlockVisits = 50

func Analyse(source string, functionName string) []Interpreter {
	builder := ssa.NewBuilder()

	function, err := builder.ParseAndBuildSSA(source, functionName)
	if err != nil {
		panic(err)
	}

	analyser := &Analyser{
		Function:     function,
		PathSelector: &RandomPathSelector{},
		satisfiable:  make(map[symbolic.SymbolicExpression]bool),
		sessions:     make(map[sessionOpt]*solverSession),
	}
	analyser.push(newInterpreter(analyser))

	for step := 0; step < maxSteps && analyser.StatesQueue.Len() > 0; step++ {
		state := heap.Pop(&analyser.StatesQueue).(*Item).value
		frame := state.CallStack[len(state.CallStack)-1]
		if frame.BlockVisits[frame.Block] > maxBlockVisits || !analyser.isSatisfiable(state.PathCondition) {
			continue
		}

		if frame.Block == nil {
			analyser.Results = append(analyser.Results, state)
			continue
		}

		for _, next := range state.interpretDynamically(frame.Block.Instrs[frame.InstrIdx]) {
			analyser.push(next)
		}
	}
	return analyser.Results
}

func (analyser *Analyser) push(state Interpreter) {
	heap.Push(&analyser.StatesQueue, &Item{
		value:    state,
		priority: analyser.PathSelector.CalculatePriority(state),
	})
}

func newInterpreter(analyser *Analyser) Interpreter {
	locals := make(map[string]symbolic.SymbolicExpression)
	for _, param := range analyser.Function.Params {
		locals[param.Name()] = symbolic.NewSymbolicVariable(param.Name(), symbolicType(param.Type()))
	}

	return Interpreter{
		CallStack: []CallStackFrame{{
			Function:    analyser.Function,
			LocalMemory: locals,
			Block:       analyser.Function.Blocks[0],
			BlockVisits: make(map[*xssa.BasicBlock]int),
		}},
		Analyser:      analyser,
		PathCondition: symbolic.NewBoolConstant(true),
	}
}

func (analyser *Analyser) isSatisfiable(pathCondition symbolic.SymbolicExpression) bool {
	if satisfiable, cached := analyser.satisfiable[pathCondition]; cached {
		return satisfiable
	}
	satisfiable := analyser.checkSatisfiability(pathCondition)
	analyser.satisfiable[pathCondition] = satisfiable
	return satisfiable
}

func (analyser *Analyser) checkSatisfiability(pathCondition symbolic.SymbolicExpression) bool {
	key := sessionOpt{
		bitvectors: symbolic.HasBitwiseOperations(pathCondition),
		floats:     symbolic.HasFloats(pathCondition),
	}
	session, ok := analyser.sessions[key]
	if !ok {
		session = newSolverSession(newTranslator(key), !key.floats)
		analyser.sessions[key] = session
	}

	constraints := []symbolic.SymbolicExpression{pathCondition}
	if !key.bitvectors {
		constraints = append(constraints, symbolic.RangeConstraints(pathCondition, true)...)
	}
	return session.check(constraints)
}

func newTranslator(key sessionOpt) *translator.Z3Translator {
	if key.bitvectors {
		return translator.NewBV64Translator()
	}
	return translator.NewIntTranslator()
}
