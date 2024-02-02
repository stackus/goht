package compiler

import (
	"fmt"
)

type PositionalError struct {
	Line   int
	Column int
	Err    error
}

func (e PositionalError) Error() string {
	return fmt.Sprintf("%s: %s", e.Position(), e.Err)
}

func (e PositionalError) Position() string {
	return fmt.Sprintf("[%d:%d]", e.Line, e.Column)
}

func (e PositionalError) Unwrap() error {
	return e.Err
}
