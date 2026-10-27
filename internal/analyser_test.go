package internal

import (
	"fmt"
	"os"
	"slices"
	"testing"
)

const examplesPath = "../homework3/examples/test_functions.go"

type analyseCase struct {
	name   string
	source string
	want   []string
}

func TestAnalyse(t *testing.T) {
	examples, err := os.ReadFile(examplesPath)
	if err != nil {
		t.Fatal(err)
	}

	cases := []analyseCase{
		{
			name: "test1",
			want: []string{
				"true && (x > 10) => (x + 1)",
				"true && !(x > 10) => (x - 1)",
			},
		},
		{
			name: "test2",
			want: []string{
				"true && a && b => true",
				"true && a && !b => false",
				"true && !a => false",
			},
		},
		{
			name: "testArithmetic",
			want: []string{
				"true && (((x + y) * (x - y)) < 0) => (-((x + y) * (x - y)) % 10)",
				"true && !(((x + y) * (x - y)) < 0) => (((x + y) * (x - y)) % 10)",
			},
		},
		{
			name: "testComparisons",
			want: []string{
				"true && (a == b) => true",
				"true && !(a == b) && (a != b) && (a > 0) => (b < 0)",
				"true && !(a == b) && (a != b) && !(a > 0) && (a >= b) => true",
				"true && !(a == b) && (a != b) && !(a > 0) && !(a >= b) => (a <= -b)",
			},
		},
		{
			name: "testLogicalOps",
			want: []string{
				"true && (x > 0) && !(x < 0) => ((y > 0) != (y < 0))",
				"true && !(x > 0) && (x < 0) => (false != true)",
				"true && !(x > 0) && !(x < 0) => (false != (y < 0))",
			},
		},
		{
			name: "testComplexConditions",
			want: []string{
				"true && (a > b) && (b < c) => 1",
				"true && (a > b) && !(b < c) && !(a < b) && !(a == b) => 3",
				"true && !(a > b) && (a < b) && (b > c) => 1",
				"true && !(a > b) && (a < b) && !(b > c) && !(a == b) => 3",
				"true && !(a > b) && !(a < b) && (a == b) && (b == c) => 2",
				"true && !(a > b) && !(a < b) && (a == b) && !(b == c) => 3",
			},
		},
		{
			name: "testNestedIf",
			want: []string{
				"true && (x > 0) && (y > 0) => 1",
				"true && (x > 0) && !(y > 0) => 2",
				"true && !(x > 0) && (y > 0) => 3",
				"true && !(x > 0) && !(y > 0) => 4",
			},
		},
		{
			name: "testEdgeCases",
			want: []string{
				"true && (a == 0) => 0",
				"true && !(a == 0) && (a < 0) => -1",
				"true && !(a == 0) && !(a < 0) => 1",
			},
		},
		{
			name: "testMultipleReturns",
			want: []string{
				"true && (x < 0) => -x",
				"true && !(x < 0) && (x == 0) => 0",
				"true && !(x < 0) && !(x == 0) => (x * 2)",
			},
		},
		{
			name: "testBitwise",
			want: []string{
				"true && (((a & b) | (a | b)) > 0) => ((a ^ b) << 1)",
				"true && !(((a & b) | (a | b)) > 0) => ((a ^ b) >> 1)",
			},
		},
		{
			name: "testUnary",
			want: []string{
				"true && ((-x ^ -1) > 0) => (-x ^ -1)",
				"true && !((-x ^ -1) > 0) => -x",
			},
		},
		{
			name: "testSimpleSum",
			want: []string{
				"true && ((b + (a + 1.1)) > 10.1) && ((b + (a + 1.1)) < 11.125) => 1.1",
				"true && ((b + (a + 1.1)) > 10.1) && !((b + (a + 1.1)) < 11.125) => 1.2",
				"true && !((b + (a + 1.1)) > 10.1) => 1.2",
			},
		},
		{
			name: "pruned",
			source: `package main
func pruned(x int) int {
	if x > 0 {
		if x < 0 {
			return 1
		}
	}
	return 0
}`,
			want: []string{
				"true && (x > 0) && !(x < 0) => 0",
				"true && !(x > 0) => 0",
			},
		},
		{
			name: "swap",
			source: `package main
func swap(n int) int {
	if n > 2 {
		return 0
	}
	a, b := 1, 2
	for i := 0; i < n; i++ {
		a, b = b, a
	}
	return a
}`,
			want: []string{
				"true && (n > 2) => 0",
				"true && !(n > 2) && !(0 < n) => 1",
				"true && !(n > 2) && (0 < n) && !((0 + 1) < n) => 2",
				"true && !(n > 2) && (0 < n) && ((0 + 1) < n) && !(((0 + 1) + 1) < n) => 1",
			},
		},
		{
			name: "loop",
			source: `package main
func loop(n int) int {
	if n > 2 {
		return -1
	}
	result := 0
	for i := 0; i < n; i++ {
		result += i
	}
	return result
}`,
			want: []string{
				"true && (n > 2) => -1",
				"true && !(n > 2) && !(0 < n) => 0",
				"true && !(n > 2) && (0 < n) && !((0 + 1) < n) => (0 + 0)",
				"true && !(n > 2) && (0 < n) && ((0 + 1) < n) && !(((0 + 1) + 1) < n) => ((0 + 0) + (0 + 1))",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Analyse(%s) panicked: %v", tc.name, r)
				}
			}()

			source := tc.source
			if source == "" {
				source = string(examples)
			}

			var got []string
			for _, result := range Analyse(source, tc.name) {
				frame := result.CallStack[0]
				got = append(got, fmt.Sprintf("%s => %s", result.PathCondition, frame.ReturnValue))
			}

			want := slices.Clone(tc.want)
			slices.Sort(got)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("paths of %s:\n got:  %q\n want: %q", tc.name, got, want)
			}
		})
	}
}

// TestAnalyseLoops проверяет только число путей: каждая итерация цикла даёт свой путь,
// а число итераций ограничено maxBlockVisits.
func TestAnalyseLoops(t *testing.T) {
	examples, err := os.ReadFile(examplesPath)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		want int
	}{
		{"testWhileLoop", maxBlockVisits},
		{"testForLoop", maxBlockVisits},
		{"testInfiniteLoopBreak", maxBlockVisits},
		{"testLoopWithConcreteBoundAndSymbolicBranching", 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(Analyse(string(examples), tc.name)); got != tc.want {
				t.Fatalf("paths of %s: got %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}
