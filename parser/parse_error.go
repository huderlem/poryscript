package parser

import (
	"fmt"

	"github.com/huderlem/poryscript/token"
)

type ParseError struct {
	LineNumberStart int
	LineNumberEnd   int
	CharStart       int
	CharEnd         int
	Message         string
}

func NewParseError(tok token.Token, message string) error {
	return ParseError{
		LineNumberStart: tok.LineNumber,
		LineNumberEnd:   tok.EndLineNumber,
		CharStart:       tok.StartCharIndex,
		CharEnd:         tok.EndCharIndex,
		Message:         message,
	}
}

func NewRangeParseError(tok1, tok2 token.Token, message string) error {
	return ParseError{
		LineNumberStart: tok1.LineNumber,
		LineNumberEnd:   tok2.EndLineNumber,
		CharStart:       tok1.StartCharIndex,
		CharEnd:         tok2.EndCharIndex,
		Message:         message,
	}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.LineNumberStart, e.Message)
}
