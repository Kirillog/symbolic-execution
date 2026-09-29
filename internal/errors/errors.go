// Package errors содержит общие ошибки проекта
package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrFunctionNotFound возвращается, если функция отсутствует в SSA представлении
	ErrFunctionNotFound = errors.New("function not found")
	// ErrParseSource возвращается, если не удалось разобрать исходный код
	ErrParseSource = errors.New("failed to parse source")
	// ErrBuildSSA возвращается, если не удалось построить SSA
	ErrBuildSSA = errors.New("failed to build ssa")
)

// FunctionNotFoundError описывает ненайденную функцию; совпадает с ErrFunctionNotFound через errors.Is
type FunctionNotFoundError struct {
	Name string
}

func (e *FunctionNotFoundError) Error() string {
	return fmt.Sprintf("function %s not found", e.Name)
}

func (e *FunctionNotFoundError) Unwrap() error { return ErrFunctionNotFound }

// NewFunctionNotFound создаёт ошибку ненайденной функции
func NewFunctionNotFound(name string) error {
	return &FunctionNotFoundError{Name: name}
}
