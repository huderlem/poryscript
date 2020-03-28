package token

import "github.com/huderlem/poryscript/types"

// Type distinguishes between different types of tokens in the Poryscript lexer.
type Type string

// Token represents a single token in the Poryscript lexer.
type Token struct {
	Type       Type
	Literal    string
	LineNumber int
}

// Token types
const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers and literals
	IDENT     = "IDENT"
	INT       = "INT"
	STRING    = "STRING"
	RAWSTRING = "RAWSTRING"

	// Operators
	ASSIGN = "="
	EQ     = "=="
	NEQ    = "!="
	LT     = "<"
	GT     = ">"
	LTE    = "<="
	GTE    = ">="
	AND    = "&&"
	OR     = "||"
	NOT    = "!"
	MUL    = "*"

	// Delimeters
	COMMA = ","
	COLON = ":"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	SCRIPT     = "SCRIPT"
	RAW        = "RAW"
	TEXT       = "TEXT"
	MOVEMENT   = "MOVEMENT"
	MAPSCRIPTS = "MAPSCRIPTS"
	FORMAT     = "FORMAT"
	VAR        = "VAR"
	FLAG       = "FLAG"
	DEFEATED   = "DEFEATED"
	TRUE       = "TRUE"
	FALSE      = "FALSE"
	IF         = "IF"
	ELSE       = "ELSE"
	ELSEIF     = "ELSEIF"
	DO         = "DO"
	WHILE      = "WHILE"
	BREAK      = "BREAK"
	CONTINUE   = "CONTINUE"
	SWITCH     = "SWITCH"
	CASE       = "CASE"
	DEFAULT    = "DEFAULT"
	GLOBAL     = "GLOBAL"
	LOCAL      = "LOCAL"
	PORYSWITCH = "PORYSWITCH"
	CONST      = "CONST"
)

var keywords = map[string]Type{
	"script":     SCRIPT,
	"raw":        RAW,
	"text":       TEXT,
	"movement":   MOVEMENT,
	"mapscripts": MAPSCRIPTS,
	"format":     FORMAT,
	"TRUE":       TRUE,
	"FALSE":      FALSE,
	"true":       TRUE,
	"false":      FALSE,
	"if":         IF,
	"else":       ELSE,
	"elif":       ELSEIF,
	"do":         DO,
	"while":      WHILE,
	"break":      BREAK,
	"continue":   CONTINUE,
	"switch":     SWITCH,
	"case":       CASE,
	"default":    DEFAULT,
	"global":     GLOBAL,
	"local":      LOCAL,
	"poryswitch": PORYSWITCH,
	"const":      CONST,
}

var gen3Keywords = map[string]Type{
	"var":      VAR,
	"flag":     FLAG,
	"defeated": DEFEATED,
}

// GetIdentType looks up the token type for the given identifier in the given Gen context.
func GetIdentType(ident string, gen types.Gen) Type {
	if tokType, ok := keywords[ident]; ok {
		return tokType
	}
	if gen == types.GEN3 {
		if tokType, ok := gen3Keywords[ident]; ok {
			return tokType
		}
	}
	return IDENT
}
