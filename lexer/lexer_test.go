package lexer

import (
	"testing"

	"github.com/huderlem/poryscript/config"
	"github.com/huderlem/poryscript/token"
)

func TestNextToken(t *testing.T) {
	input := `script(local) BugContestOfficer_EnterContest_23
{
	if (var(VAR_BUG_CONTEST_PRIZE) != ITEM_NONE) {
		giveitem_std(VAR_BUG_CONTEST_PRIZE)
		if (flag(FLAG_TEST) == TRUE) {
			setvar(VAR_BUG_CONTEST_PRIZE, ITEM_NONE)
		} elif (var(VAR_TEST) <= 5) {
		} else { ##
		#}
// >
		< // >=
		> //#
		>= #
		=
		!
		/
		&&&
		|||
		do
		break
		global
		continue
		[]
		switch
		0x5ABCDEF
		0435
		-23
		case: default
		$4acb
		while
		defeated
		text
		poryswitch
		const
		movement
		mapscripts
		*
		format
		("Hello\n"
		"I'm glad to see$")
		raw RawTest ` + "`" + `
	step` + "`" + `
	>`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.SCRIPT, "script"},
		{token.LPAREN, "("},
		{token.LOCAL, "local"},
		{token.RPAREN, ")"},
		{token.IDENT, "BugContestOfficer_EnterContest_23"},
		{token.LBRACE, "{"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.VAR, "var"},
		{token.LPAREN, "("},
		{token.IDENT, "VAR_BUG_CONTEST_PRIZE"},
		{token.RPAREN, ")"},
		{token.NEQ, "!="},
		{token.IDENT, "ITEM_NONE"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "giveitem_std"},
		{token.LPAREN, "("},
		{token.IDENT, "VAR_BUG_CONTEST_PRIZE"},
		{token.RPAREN, ")"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.FLAG, "flag"},
		{token.LPAREN, "("},
		{token.IDENT, "FLAG_TEST"},
		{token.RPAREN, ")"},
		{token.EQ, "=="},
		{token.TRUE, "TRUE"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "setvar"},
		{token.LPAREN, "("},
		{token.IDENT, "VAR_BUG_CONTEST_PRIZE"},
		{token.COMMA, ","},
		{token.IDENT, "ITEM_NONE"},
		{token.RPAREN, ")"},
		{token.RBRACE, "}"},
		{token.ELSEIF, "elif"},
		{token.LPAREN, "("},
		{token.VAR, "var"},
		{token.LPAREN, "("},
		{token.IDENT, "VAR_TEST"},
		{token.RPAREN, ")"},
		{token.LTE, "<="},
		{token.INT, "5"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.LT, "<"},
		{token.GT, ">"},
		{token.GTE, ">="},
		{token.ASSIGN, "="},
		{token.NOT, "!"},
		{token.ILLEGAL, "/"},
		{token.AND, "&&"},
		{token.ILLEGAL, "&"},
		{token.OR, "||"},
		{token.ILLEGAL, "|"},
		{token.DO, "do"},
		{token.BREAK, "break"},
		{token.GLOBAL, "global"},
		{token.CONTINUE, "continue"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.SWITCH, "switch"},
		{token.INT, "0x5ABCDEF"},
		{token.INT, "0435"},
		{token.INT, "-23"},
		{token.CASE, "case"},
		{token.COLON, ":"},
		{token.DEFAULT, "default"},
		{token.INT, "$4acb"},
		{token.WHILE, "while"},
		{token.DEFEATED, "defeated"},
		{token.TEXT, "text"},
		{token.PORYSWITCH, "poryswitch"},
		{token.CONST, "const"},
		{token.MOVEMENT, "movement"},
		{token.MAPSCRIPTS, "mapscripts"},
		{token.MUL, "*"},
		{token.FORMAT, "format"},
		{token.LPAREN, "("},
		{token.STRING, "Hello\\n\nI'm glad to see$"},
		{token.RPAREN, ")"},
		{token.RAW, "raw"},
		{token.IDENT, "RawTest"},
		{token.RAWSTRING, "\tstep"},
		{token.GT, ">"},
		{token.EOF, ""},
	}

	l := New(input, config.GEN3)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("tests[%d] - tokenType wrong. Expected=%q, Got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. Expected=%q, Got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}
