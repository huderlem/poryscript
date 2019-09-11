package token

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
	EQ  = "=="
	NEQ = "!="
	LT  = "<"
	GT  = ">"
	LTE = "<="
	GTE = ">="
	AND = "&&"
	OR  = "||"
	NOT = "!"

	// Delimeters
	COMMA = ","
	COLON = ":"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"

	// Keywords
	SCRIPT   = "SCRIPT"
	RAW      = "RAW"
	VAR      = "VAR"
	FLAG     = "FLAG"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	ELSEIF   = "ELSEIF"
	DO       = "DO"
	WHILE    = "WHILE"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	SWITCH   = "SWITCH"
	CASE     = "CASE"
	DEFAULT  = "DEFAULT"
)

// If statement comparison types
const (
	CMPVAR  = "CMPVAR"
	CMPFLAG = "CMPFLAG"
)

var keywords = map[string]Type{
	"script":   SCRIPT,
	"raw":      RAW,
	"var":      VAR,
	"flag":     FLAG,
	"TRUE":     TRUE,
	"FALSE":    FALSE,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"elif":     ELSEIF,
	"do":       DO,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
	"switch":   SWITCH,
	"case":     CASE,
	"default":  DEFAULT,
}

// GetIdentType looks up the token type for the given identifier
func GetIdentType(ident string) Type {
	if tokType, ok := keywords[ident]; ok {
		return tokType
	}
	return IDENT
}
