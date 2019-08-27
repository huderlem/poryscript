package parser

import (
	"testing"

	"github.com/huderlem/poryscript/token"

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

func TestRawStatements(t *testing.T) {
	input := `
raw RawTest ` + "`" + `
	step_up
	step_end
` + "`" + `

raw_global RawTestGlobal ` + "`" + `
	step_down
` + "`"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(program.TopLevelStatements) != 2 {
		t.Fatalf("program.TopLevelStatements does not contain 2 statements. got=%d", len(program.TopLevelStatements))
	}

	tests := []struct {
		expectedName   string
		expectedGlobal bool
		expectedValue  string
	}{
		{"RawTest", false, `	step_up
	step_end`},
		{"RawTestGlobal", true, `	step_down`},
	}

	for i, tt := range tests {
		stmt := program.TopLevelStatements[i]
		if !testRawStatement(t, stmt, tt.expectedName, tt.expectedGlobal, tt.expectedValue) {
			return
		}
	}
}

func testRawStatement(t *testing.T, s ast.Statement, expectedName string, expectedGlobal bool, expectedValue string) bool {
	if s.TokenLiteral() != "raw" && s.TokenLiteral() != "raw_global" {
		t.Errorf("s.TokenLiteral not 'raw' or 'raw_global'. got=%q", s.TokenLiteral())
		return false
	}

	rawStmt, ok := s.(*ast.RawStatement)
	if !ok {
		t.Errorf("s not %T. got=%T", &ast.RawStatement{}, s)
		return false
	}

	if rawStmt.Name.Value != expectedName {
		t.Errorf("rawStmt.Name.Value not '%s'. got=%s", expectedName, rawStmt.Name.Value)
		return false
	}

	if rawStmt.IsGlobal != expectedGlobal {
		t.Errorf("rawStmt.IsGlobal not '%t'. got=%t", expectedGlobal, rawStmt.IsGlobal)
		return false
	}

	if rawStmt.Value != expectedValue {
		t.Errorf("rawStmt.Value not '%s'. got=%s", expectedValue, rawStmt.Value)
		return false
	}

	return true
}

func TestIfStatements(t *testing.T) {
	input := `
script Test {
	if (var(VAR_1) == 1) {
		if (var(VAR_7) != 1) {
			message()
		}
		message()
	} elif (var(VAR_2) != 2) {
		blah()
	} elif (var(VAR_3) < 3) {
		blah()
	} elif (var(VAR_4) <= 4) {
		blah()
	} elif (var(VAR_5) > 5) {
		blah()
	} elif (var(VAR_6) >= 6) {
		blah()
	} elif (flag(FLAG_1) == TRUE) {
		blah()
	} elif (flag(FLAG_2 +  BASE) == false) {
		blah()
	} else {
		message()
		lock
		faceplayer
		facepalm
		blah(1, 3, 4)
	}
}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	scriptStmt := program.TopLevelStatements[0].(*ast.ScriptStatement)
	ifStmt := scriptStmt.Body.Statements[0].(*ast.IfStatement)
	testConditionExpression(t, ifStmt.Consequence, token.VAR, "VAR_1", token.EQ, "1")
	testConditionExpression(t, ifStmt.ElifConsequences[0], token.VAR, "VAR_2", token.NEQ, "2")
	testConditionExpression(t, ifStmt.ElifConsequences[1], token.VAR, "VAR_3", token.LT, "3")
	testConditionExpression(t, ifStmt.ElifConsequences[2], token.VAR, "VAR_4", token.LTE, "4")
	testConditionExpression(t, ifStmt.ElifConsequences[3], token.VAR, "VAR_5", token.GT, "5")
	testConditionExpression(t, ifStmt.ElifConsequences[4], token.VAR, "VAR_6", token.GTE, "6")
	testConditionExpression(t, ifStmt.ElifConsequences[5], token.FLAG, "FLAG_1", token.EQ, "TRUE")
	testConditionExpression(t, ifStmt.ElifConsequences[6], token.FLAG, "FLAG_2 + BASE", token.EQ, "false")
	nested := ifStmt.Consequence.Body.Statements[0].(*ast.IfStatement)
	testConditionExpression(t, nested.Consequence, token.VAR, "VAR_7", token.NEQ, "1")

	if len(ifStmt.ElseConsequence.Statements) != 5 {
		t.Fatalf("len(ifStmt.ElseConsequences) should be '%d'. got=%d", 5, len(ifStmt.ElseConsequence.Statements))
	}
}

func testConditionExpression(t *testing.T, expression *ast.ConditionExpression, expectedType token.Type, expectedOperand string, expectedOperator token.Type, expectedComparisonValue string) {
	if expression.Type != expectedType {
		t.Fatalf("expression.Type not '%s'. got=%s", expectedType, expression.Type)
	}
	if expression.Operand != expectedOperand {
		t.Fatalf("expression.Operand not '%s'. got=%s", expectedOperand, expression.Operand)
	}
	if expression.Operator != expectedOperator {
		t.Fatalf("expression.Operator not '%s'. got=%s", expectedOperator, expression.Operator)
	}
	if expression.ComparisonValue != expectedComparisonValue {
		t.Fatalf("expression.ComparisonValue not '%s'. got=%s", expectedComparisonValue, expression.ComparisonValue)
	}
}
