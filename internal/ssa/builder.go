// Package ssa предоставляет функции для построения SSA представления
package ssa

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	ierrors "symbolic-execution-course/internal/errors"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// Builder отвечает за построение SSA из исходного кода Go
type Builder struct {
	fset *token.FileSet
}

// NewBuilder создаёт новый экземпляр Builder
func NewBuilder() *Builder {
	return &Builder{
		fset: token.NewFileSet(),
	}
}

// TODO: Реализуйте следующие методы в рамках домашнего задания

// ParseAndBuildSSA парсит исходный код Go и создаёт SSA представление
// Возвращает SSA программу и функцию по имени
func (b *Builder) ParseAndBuildSSA(source string, funcName string) (*ssa.Function, error) {
	// TODO: Реализовать
	// Шаги:
	// 1. Парсинг исходного кода с помощью go/parser
	// 2. Создание SSA программы
	// 3. Поиск нужной функции по имени

	f, err := parser.ParseFile(b.fset, "source.go", source, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ierrors.ErrParseSource, err)
	}
	files := []*ast.File{f}
	pkg := types.NewPackage(f.Name.Name, f.Name.Name)
	config := &types.Config{Importer: importer.Default()}
	hello, _, err := ssautil.BuildPackage(
		config, b.fset, pkg, files, ssa.SanityCheckFunctions)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ierrors.ErrBuildSSA, err)
	}
	var fun = hello.Func(funcName)
	if fun == nil {
		return nil, ierrors.NewFunctionNotFound(funcName)
	}
	return fun, nil

	// Подсказки:
	// - Используйте parser.ParseFile для парсинга
	// - Создайте packages.Config и загрузите пакет
	// - Используйте ssautil.CreateProgram для создания SSA
	// - Найдите функцию в SSA программе
}
