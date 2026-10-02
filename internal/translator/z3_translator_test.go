package translator

import (
	"fmt"
	"testing"

	. "symbolic-execution-course/internal/symbolic"
)

type expr = SymbolicExpression

func iv(name string) expr { return NewSymbolicVariable(name, IntType) }
func bv(name string) expr { return NewSymbolicVariable(name, BoolType) }
func c(n int64) expr      { return NewIntConstant(n) }

func bin(l expr, op BinaryOperator, r expr) expr { return NewBinaryOperation(l, r, op) }
func and(xs ...expr) expr                        { return NewLogicalOperation(xs, AND) }
func or(xs ...expr) expr                         { return NewLogicalOperation(xs, OR) }
func not(x expr) expr                            { return NewLogicalOperation([]expr{x}, NOT) }
func neg(x expr) expr                            { return &NegateOperation{Operand: x} }

type testCase struct {
	name string
	expr expr
	want string
}

func runCases(t *testing.T, cases []testCase) {
	t.Helper()
	for _, tc := range cases {
		ok := t.Run(tc.name, func(t *testing.T) {
			tr := NewZ3Translator()
			defer tr.Close()

			got, err := tr.TranslateExpression(tc.expr)
			if err != nil {
				t.Fatalf("translate %s: %v", tc.expr, err)
			}
			if fmt.Sprint(got) != tc.want {
				t.Fatalf("translate %s:\n got:  %s\n want: %s", tc.expr, got, tc.want)
			}
		})
		if !ok {
			t.FailNow()
		}
	}
}

func TestAdd(t *testing.T) {
	runCases(t, []testCase{
		{"return a+b", bin(iv("a"), ADD, iv("b")), "(+ a b)"},
	})
}

func TestMax(t *testing.T) {
	runCases(t, []testCase{
		{"then: a>b", bin(iv("a"), GT, iv("b")), "(> a b)"},
		{"else: !(a>b)", not(bin(iv("a"), GT, iv("b"))), "(not (> a b))"},
	})
}

func TestCalculate(t *testing.T) {
	sum := bin(iv("x"), ADD, iv("y"))
	diff := bin(iv("x"), SUB, iv("y"))
	runCases(t, []testCase{
		{"(x+y)*(x-y)", bin(sum, MUL, diff), "(* (+ x y) (- x y))"},
	})
}

func TestIsValid(t *testing.T) {
	runCases(t, []testCase{
		{"x>0 && y>0 && x<100",
			and(bin(iv("x"), GT, c(0)), bin(iv("y"), GT, c(0)), bin(iv("x"), LT, c(100))),
			"(and (> x 0) (> y 0) (< x 100))"},
	})
}

func TestClassify(t *testing.T) {
	ge := func(n int64) expr { return bin(iv("score"), GE, c(n)) }
	runCases(t, []testCase{
		{"A", ge(90), "(>= score 90)"},
		{"B", and(not(ge(90)), ge(80)), "(and (not (>= score 90)) (>= score 80))"},
		{"C", and(not(ge(90)), not(ge(80)), ge(70)),
			"(and (not (>= score 90)) (not (>= score 80)) (>= score 70))"},
		{"F", and(not(ge(90)), not(ge(80)), not(ge(70))),
			"(and (not (>= score 90)) (not (>= score 80)) (not (>= score 70)))"},
	})
}

func TestSumLoopUnrolled(t *testing.T) {
	le := func(i int64) expr { return bin(c(i), LE, iv("n")) }
	runCases(t, []testCase{
		{"0 iterations: !(1<=n)", not(le(1)), "(not (<= 1 n))"},
		{"2 iterations: path", and(le(1), le(2), not(le(3))), "(and (<= 1 n) (<= 2 n) (not (<= 3 n)))"},
		{"2 iterations: result", bin(bin(c(0), ADD, c(1)), ADD, c(2)), "(+ 0 1 2)"},
	})
}

func TestComplexCondition(t *testing.T) {
	runCases(t, []testCase{
		{"(x>0 || y>0) && z>x+y",
			and(
				or(bin(iv("x"), GT, c(0)), bin(iv("y"), GT, c(0))),
				bin(iv("z"), GT, bin(iv("x"), ADD, iv("y"))),
			),
			"(and (or (> x 0) (> y 0)) (> z (+ x y)))"},
	})
}

func TestPolynomial(t *testing.T) {
	x := iv("x")
	runCases(t, []testCase{
		{"3*x*x + 2*x + 1",
			bin(bin(bin(bin(c(3), MUL, x), MUL, x), ADD, bin(c(2), MUL, x)), ADD, c(1)),
			"(+ (* 3 x x) (* 2 x) 1)"},
	})
}

func TestModOperations(t *testing.T) {
	runCases(t, []testCase{
		{"x%2==0 && y%3==1",
			and(
				bin(bin(iv("x"), MOD, c(2)), EQ, c(0)),
				bin(bin(iv("y"), MOD, c(3)), EQ, c(1)),
			),
			"(and (not (distinct (mod x 2) 0)) (not (distinct (mod y 3) 1)))"},
	})
}

func TestCompareAll(t *testing.T) {
	a, b, cc := iv("a"), iv("b"), iv("c")
	runCases(t, []testCase{
		{"a==b alone", bin(a, EQ, b), "(not (distinct a b))"},
		{"b!=c alone", bin(b, NE, cc), "(distinct b c)"},
		{"a==b || b!=c || a<c || c>=b",
			or(bin(a, EQ, b), bin(b, NE, cc), bin(a, LT, cc), bin(cc, GE, b)),
			"(or (not (distinct a b)) (distinct b c) (< a c) (>= c b))"},
	})
}

func TestDivide(t *testing.T) {
	nz := bin(iv("y"), NE, c(0))
	runCases(t, []testCase{
		{"guard: y!=0", nz, "(distinct y 0)"},
		{"guard else: !(y!=0)", not(nz), "(not (distinct y 0))"},
		{"return x/y", bin(iv("x"), DIV, iv("y")), "(div x y)"},
	})
}

// В SymbolicExpression нет битовых операторов (&, |, ^), поэтому bitwiseOps не выразима.
func TestBitwiseOps(t *testing.T) {
	t.Skip("bitwise operators (&, |, ^) are not supported by symbolic.BinaryOperator")
}

func TestUnaryOps(t *testing.T) {
	runCases(t, []testCase{
		{"result = -x", neg(iv("x")), "(- x)"},
		{"result = -(-x)", neg(neg(iv("x"))), "(- (- x))"},
		{"!flag", not(bv("flag")), "(not flag)"},
	})
}

func TestTernary(t *testing.T) {
	runCases(t, []testCase{
		{"bool variable", bv("condition"), "condition"},
		{"negated", not(bv("condition")), "(not condition)"},
		{"return a", iv("a"), "a"},
	})
}

func TestAbs(t *testing.T) {
	runCases(t, []testCase{
		{"x<0", bin(iv("x"), LT, c(0)), "(< x 0)"},
		{"return -x", neg(iv("x")), "(- x)"},
		{"!(x<0)", not(bin(iv("x"), LT, c(0))), "(not (< x 0))"},
	})
}

func TestSignFunction(t *testing.T) {
	pos := bin(iv("x"), GT, c(0))
	negative := bin(iv("x"), LT, c(0))
	runCases(t, []testCase{
		{"return 1", pos, "(> x 0)"},
		{"return -1", and(not(pos), negative), "(and (not (> x 0)) (< x 0))"},
		{"return 0", and(not(pos), not(negative)), "(and (not (> x 0)) (not (< x 0)))"},
		{"constant -1", c(-1), "(- 1)"},
	})
}

func TestInRange(t *testing.T) {
	runCases(t, []testCase{
		{"x>=min && x<=max",
			and(bin(iv("x"), GE, iv("min")), bin(iv("x"), LE, iv("max"))),
			"(and (>= x min) (<= x max))"},
	})
}

// Одна и та же переменная в разных частях выражения должна транслироваться в один Z3 символ.
func TestVariableCache(t *testing.T) {
	runCases(t, []testCase{
		{"x+x", bin(iv("x"), ADD, iv("x")), "(+ x x)"},
		{"bool and int variables side by side", and(bv("b"), bin(iv("x"), GT, c(0))), "(and b (> x 0))"},
	})
}
