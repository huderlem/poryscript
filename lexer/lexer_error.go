package lexer

import (
  "fmt"
)

type LexerError struct {
	LineNumberStart int
	CharStart       int
	Message         string
}

func NewLexerError(l *Lexer, message string) error {
	return LexerError{
		LineNumberStart: l.lineNumber,
		CharStart:       l.charNumber,
		Message:         message,
	}
}

func (e LexerError) Error() string {
  return fmt.Sprintf("line %d: %s", e.LineNumberStart, e.Message)
}
