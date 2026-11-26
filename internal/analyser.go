package internal

import (
	"container/heap"
	"symbolic-execution-course/internal/ssa"
	"symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"

	"github.com/ebukreev/go-z3/z3"
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

func sessionKey(pathCondition symbolic.SymbolicExpression, terms []symbolic.SymbolicExpression) sessionOpt {
	var key sessionOpt
	for _, expr := range append([]symbolic.SymbolicExpression{pathCondition}, terms...) {
		key.bitvectors = key.bitvectors || symbolic.HasBitwiseOperations(expr)
		key.floats = key.floats || symbolic.HasFloats(expr)
	}
	return key
}

func (analyser *Analyser) checkSatisfiability(pathCondition symbolic.SymbolicExpression) bool {
	_, model := analyser.solve(sessionKey(pathCondition, nil), pathCondition, nil)
	return model != nil
}

func (analyser *Analyser) Concretize(state Interpreter) (args []symbolic.SymbolicExpression, result symbolic.SymbolicExpression, ok bool) {
	frame := state.CallStack[0]
	params := analyser.Function.Params
	terms := make([]symbolic.SymbolicExpression, 0, len(params)+1)
	for _, param := range params {
		terms = append(terms, frame.LocalMemory[param.Name()])
	}
	if state.Panic == nil && frame.ReturnValue != nil {
		terms = append(terms, frame.ReturnValue)
	}

	key := sessionKey(state.PathCondition, terms)
	session, model := analyser.solve(key, state.PathCondition, terms)
	if model == nil && !key.bitvectors {
		key.bitvectors = true
		session, model = analyser.solve(key, state.PathCondition, terms)
	}
	if model == nil {
		return nil, nil, false
	}
	values := make([]symbolic.SymbolicExpression, len(terms))
	for i, term := range terms {
		value, err := session.translator.Evaluate(model, term)
		if err != nil {
			panic(err)
		}
		values[i] = value
	}
	if len(values) > len(params) {
		result = values[len(params)]
	}
	return values[:len(params)], result, true
}

func (analyser *Analyser) solve(key sessionOpt, pathCondition symbolic.SymbolicExpression, terms []symbolic.SymbolicExpression) (*solverSession, *z3.Model) {
	exprs := append([]symbolic.SymbolicExpression{pathCondition}, terms...)
	session, ok := analyser.sessions[key]
	if !ok {
		session = newSolverSession(newTranslator(key), !key.floats)
		analyser.sessions[key] = session
	}

	constraints := []symbolic.SymbolicExpression{pathCondition}
	if !key.bitvectors {
		for _, expr := range exprs {
			constraints = append(constraints, symbolic.RangeConstraints(expr, true)...)
		}
	}
	return session, session.solve(constraints)
}

func newTranslator(key sessionOpt) *translator.Z3Translator {
	if key.bitvectors {
		return translator.NewBV64Translator()
	}
	return translator.NewIntTranslator()
}
