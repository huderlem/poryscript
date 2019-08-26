package parser

import (
	"testing"

	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/lexer"
)

type commandArgs struct {
	name string
	args []string
}

func TestScriptStatements(t *testing.T) {
	input := `
script MyScript {
	lock
	bufferitemname(0, VAR_BUG_CONTEST_PRIZE)
	message() waitstate
	somecommand(foo,
		4+   6,,(CONST_FOO) +  1)
}

script MyScript2 {}
script MyScript3 {
		}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.TopLevelStatements) != 3 {
		t.Fatalf("program.TopLevelStatements does not contain 3 statements. got=%d", len(program.TopLevelStatements))
	}

	tests := []struct {
		expectedName     string
		expectedCommands []commandArgs
	}{
		{"MyScript", []commandArgs{
			{"lock", []string{}},
			{"bufferitemname", []string{"0", "VAR_BUG_CONTEST_PRIZE"}},
			{"message", []string{}},
			{"waitstate", []string{}},
			{"somecommand", []string{"foo", "4 + 6", "", "( CONST_FOO ) + 1"}},
		}},
		{"MyScript2", []commandArgs{}},
		{"MyScript3", []commandArgs{}},
	}

	for i, tt := range tests {
		stmt := program.TopLevelStatements[i]
		if !testScriptStatement(t, stmt, tt.expectedName, tt.expectedCommands) {
			return
		}
	}
}

func testScriptStatement(t *testing.T, s ast.Statement, expectedName string, expectedCommandArgs []commandArgs) bool {
	if s.TokenLiteral() != "script" {
		t.Errorf("s.TokenLiteral not 'script'. got=%q", s.TokenLiteral())
		return false
	}

	scriptStmt, ok := s.(*ast.ScriptStatement)
	if !ok {
		t.Errorf("s not %T. got=%T", &ast.ScriptStatement{}, s)
		return false
	}

	if scriptStmt.Name.Value != expectedName {
		t.Errorf("scriptStmt.Name.Value not '%s'. got=%s", expectedName, scriptStmt.Name.Value)
		return false
	}

	if scriptStmt.Name.TokenLiteral() != expectedName {
		t.Errorf("scriptStmt.Name not '%s'. got=%s", expectedName, scriptStmt.Name.TokenLiteral())
		return false
	}

	if scriptStmt.Body == nil {
		t.Errorf("scriptStmt.Body was nil")
		return false
	}

	if len(scriptStmt.Body.Statements) != len(expectedCommandArgs) {
		t.Errorf("scriptStmt.Body statements size not %d. got=%d", len(expectedCommandArgs), len(scriptStmt.Body.Statements))
		return false
	}

	for i, stmt := range scriptStmt.Body.Statements {
		expectedCommand := expectedCommandArgs[i]
		if stmt.TokenLiteral() != expectedCommand.name {
			t.Errorf("scriptStmt.Body statement %d not '%s'. got=%s", i, expectedCommand, stmt.TokenLiteral())
			return false
		}

		commandStmt, ok := stmt.(*ast.CommandStatement)
		if !ok {
			t.Errorf("s not %T. got=%T", &ast.CommandStatement{}, s)
			return false
		}

		if len(commandStmt.Args) != len(expectedCommand.args) {
			t.Errorf("commandStmt.Args size not %d. got=%d", len(expectedCommand.args), len(commandStmt.Args))
			return false
		}

		for j, arg := range commandStmt.Args {
			if arg != expectedCommand.args[j] {
				t.Errorf("Command statement %d of body statement %d not '%s'. got='%s'", j, i, expectedCommand.args[j], arg)
				return false
			}
		}
	}

	return true
}
