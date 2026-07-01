package soargs

import "fmt"

// Fehler beim Parsen des vom Client empfangenen Kommandos ("parse").
type ParseError string

func newParseError(format string, args ...any) error {
	return ParseError(fmt.Sprintf(format, args...))
}

// error-Interface
func (p ParseError) Error() string {
	return string(p)
}

// terror.NamedError interface.
func (p ParseError) ExitName() string {
	return "parse"
}
