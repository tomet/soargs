package soargs

import "fmt"

type namedError struct {
	name string
	err  error
}

func (e namedError) Error() string {
	return e.err.Error()
}

func (e namedError) Unwrap() error {
	return e.err
}

func (e namedError) ExitName() string {
	return e.name
}

var (
	parseError      = newErrorFunc("parse")
	connectionError = newErrorFunc("connection")
	osError         = newErrorFunc("os")
	deniedError     = newErrorFunc("denied")
	existsError     = newErrorFunc("exists")
)

func newErrorFunc(name string) func(string, ...any) error {
	return func(format string, args ...any) error {
		return namedError{
			name: name,
			err:  fmt.Errorf(format, args...),
		}
	}
}
