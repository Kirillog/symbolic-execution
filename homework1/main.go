// Демонстрационная программа для тестирования SSA построения
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"symbolic-execution-course/internal/ssa"

	xssa "golang.org/x/tools/go/ssa"
)

func main() {
	fmt.Println("=== SSA Builder Demo ===")

	file := flag.String("file", "homework1/examples/test_functions.go", "путь к исходному файлу Go")
	funcName := flag.String("func", "simpleIf", "имя анализируемой функции")
	flag.Parse()

	source, err := os.ReadFile(*file)
	if err != nil {
		log.Fatalf("Ошибка чтения файла %s: %v", *file, err)
	}

	// Создаём builder для SSA
	builder := ssa.NewBuilder()

	// Строим SSA из исходного кода
	graph, err := builder.ParseAndBuildSSA(string(source), *funcName)
	if err != nil {
		log.Fatalf("Ошибка построения SSA: %v", err)
	}
	fmt.Printf("CFG построен для функции с %d блоками\n", len(graph.Blocks))
	for _, block := range graph.Blocks {
		fmt.Printf("Блок %d:\n", block.Index)
		fmt.Printf("\tПредшественники: %s\n", block.Preds)
		fmt.Printf("\tПотомки: %s\n", block.Succs)
		fmt.Println("\tИнструкции:")
		for _, instr := range block.Instrs {
			fmt.Printf("\t\t%s\n", formatInstr(instr))
		}
	}
}

func formatInstr(instr xssa.Instruction) string {
	if v, ok := instr.(xssa.Value); ok {
		return fmt.Sprintf("%s = %s", v.Name(), v.String())
	}
	return instr.String()
}
