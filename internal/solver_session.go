package internal

import (
	"symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"

	"github.com/ebukreev/go-z3/z3"
)

// solverSession - долгоживущий транслятор одной теории. Контекст Z3 и кэш трансляции
// переиспользуются между запросами.
//
// В инкрементальной сессии солвер общий, а утверждения запроса снимаются через Pop.
// В неинкрементальной на каждый запрос создаётся новый солвер
type solverSession struct {
	translator *translator.Z3Translator
	solver     *z3.Solver
}

func newSolverSession(z3Translator *translator.Z3Translator, incremental bool) *solverSession {
	session := &solverSession{translator: z3Translator}
	if incremental {
		session.solver = session.newSolver()
	}
	return session
}

func (session *solverSession) newSolver() *z3.Solver {
	return z3.NewSolver(session.translator.GetContext().(*z3.Context))
}

// check сообщает, выполнима ли конъюнкция constraints.
func (session *solverSession) check(constraints []symbolic.SymbolicExpression) bool {
	solver := session.solver
	if solver != nil {
		solver.Push()
		defer solver.Pop()
	} else {
		solver = session.newSolver()
	}

	for _, constraint := range constraints {
		translated, err := session.translator.TranslateExpression(constraint)
		if err != nil {
			panic(err)
		}
		solver.Assert(translated.(z3.Bool))
	}

	satisfiable, err := solver.Check()
	if err != nil {
		panic(err)
	}
	return satisfiable
}
