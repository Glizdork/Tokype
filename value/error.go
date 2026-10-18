package value

import "fmt"

type RuntimeError struct {
	Message string
	Line    int
	Column  int
}

func (e *RuntimeError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return e.Message
}

func NewRuntimeError(msg string, line, col int) *RuntimeError {
	return &RuntimeError{Message: msg, Line: line, Column: col}
}
