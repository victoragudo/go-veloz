package compile

import "fmt"

type CompileError struct {
	Message string
}

func (e *CompileError) Error() string { return e.Message }

func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
