package lexer

import (
	"strings"
	"unicode"

	"github.com/huderlem/poryscript/token"
)

// Lexer produces tokens from a Poryscript file
type Lexer struct {
	input        string
	position     int           // current position in input (points to current char)
	readPosition int           // current reading position in input (after current char)
	ch           byte          // current char under examination
	lineNumber   int           // current line number
	queuedTokens []token.Token // extra tokens that were read ahead of time
}

// New initializes a new lexer for the given Poryscript file
func New(input string) *Lexer {
	l := &Lexer{input: input, lineNumber: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	prevCh := l.ch
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if prevCh == '\n' {
		l.lineNumber++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken builds the next token of the Poryscript file
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	// Return the next queued token, if there is one.
	// Tokens can be queued if there are tokens that rely
	// ok look-ahead functionality to determine their type.
	if len(l.queuedTokens) > 0 {
		tok = l.queuedTokens[0]
		l.queuedTokens = l.queuedTokens[1:]
		return tok
	}

	l.skipWhitespace()

	// Check for single-line comment.
	// Both '#' and '//' are valid comment styles.
	for l.ch == '#' || (l.ch == '/' && l.peekChar() == '/') {
		l.skipToNextLine()
		l.skipWhitespace()
	}

	switch l.ch {
	case '*':
		tok = newToken(token.MUL, l.ch, l.lineNumber)
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch), LineNumber: l.lineNumber}
		} else {
			tok = newToken(token.ASSIGN, l.ch, l.lineNumber)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NEQ, Literal: string(ch) + string(l.ch), LineNumber: l.lineNumber}
		} else {
			tok = newToken(token.NOT, l.ch, l.lineNumber)
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: string(ch) + string(l.ch), LineNumber: l.lineNumber}
		} else {
			tok = newToken(token.LT, l.ch, l.lineNumber)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: string(ch) + string(l.ch), LineNumber: l.lineNumber}
		} else {
			tok = newToken(token.GT, l.ch, l.lineNumber)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: string(ch) + string(l.ch), LineNumber: l.lineNumber}
		} else {
			tok = newToken(token.ILLEGAL, l.ch, l.lineNumber)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: string(ch) + string(l.ch), LineNumber: l.lineNumber}
		} else {
			tok = newToken(token.ILLEGAL, l.ch, l.lineNumber)
		}
	case '(':
		tok = newToken(token.LPAREN, l.ch, l.lineNumber)
	case ')':
		tok = newToken(token.RPAREN, l.ch, l.lineNumber)
	case '[':
		tok = newToken(token.LBRACKET, l.ch, l.lineNumber)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, l.lineNumber)
	case ',':
		tok = newToken(token.COMMA, l.ch, l.lineNumber)
	case ':':
		tok = newToken(token.COLON, l.ch, l.lineNumber)
	case '"':
		return l.readStringToken()
	case '`':
		tok.LineNumber = l.lineNumber
		tok.Literal = l.readRaw()
		tok.Type = token.RAWSTRING
		return tok
	case '{':
		tok = newToken(token.LBRACE, l.ch, l.lineNumber)
	case '}':
		tok = newToken(token.RBRACE, l.ch, l.lineNumber)
	case '0':
		if l.peekChar() == 'x' {
			l.readChar()
			l.readChar()
			tok.Type = token.INT
			tok.LineNumber = l.lineNumber
			tok.Literal = "0x" + l.readHexNumber()
			return tok
		}

		tok.Type = token.INT
		tok.LineNumber = l.lineNumber
		tok.Literal = l.readNumber()
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.LineNumber = l.lineNumber
	default:
		if isLetter(l.ch) {
			tok.LineNumber = l.lineNumber
			tok.Literal = l.readIdentifier()
			tok.Type = token.GetIdentType(tok.Literal)
			// If the immediately-next character is the start of a
			// STRING token, then this is a STRINGTYPE token, instead
			// of an IDENT.
			if l.ch == '"' {
				nextToken := l.readStringToken()
				l.queuedTokens = append(l.queuedTokens, nextToken)
				tok.Type = token.STRINGTYPE
			}
			return tok
		} else if isDigit(l.ch) || (l.ch == '-' && isDigit(l.peekChar())) {
			tok.Type = token.INT
			tok.LineNumber = l.lineNumber
			if l.ch == '-' {
				l.readChar()
				tok.Literal = "-" + l.readNumber()
			} else {
				tok.Literal = l.readNumber()
			}
			return tok
		}
		tok = newToken(token.ILLEGAL, l.ch, l.lineNumber)
	}

	l.readChar()
	return tok
}

func (l *Lexer) readStringToken() token.Token {
	var t token.Token
	t.LineNumber = l.lineNumber
	t.Literal = l.readString()
	t.Type = token.STRING
	return t
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipToNextLine() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.readChar()
}

func (l *Lexer) skipNewlineWhitespace() {
	for l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func newToken(tokenType token.Type, ch byte, lineNumber int) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), LineNumber: lineNumber}
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || (start != l.position && isDigit(l.ch)) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readString() string {
	var sb strings.Builder
	for l.ch == '"' {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		l.readChar()
		for l.ch != '"' && l.ch != 0 {
			sb.WriteByte(l.ch)
			l.readChar()
		}
		l.readChar()
		l.skipWhitespace()
	}
	return sb.String()
}

func (l *Lexer) readRaw() string {
	var sb strings.Builder
	l.readChar()
	l.skipNewlineWhitespace()
	for l.ch != '`' && l.ch != 0 {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	l.readChar()
	return strings.TrimRightFunc(sb.String(), unicode.IsSpace)
}

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func (l *Lexer) readNumber() string {
	start := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readHexNumber() string {
	start := l.position
	for isHexDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}
