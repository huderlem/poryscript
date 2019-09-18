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
	# this is a comment
	# another comment
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
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
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
raw ` + "`" + `
	step_up
	step_end
` + "`" + `

raw ` + "`" + `
	step_down
` + "`"
	l := lexer.New(input)
	p := New(l)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(program.TopLevelStatements) != 2 {
		t.Fatalf("program.TopLevelStatements does not contain 2 statements. got=%d", len(program.TopLevelStatements))
	}

	tests := []struct {
		expectedValue string
	}{
		{`	step_up
	step_end`},
		{`	step_down`},
	}

	for i, tt := range tests {
		stmt := program.TopLevelStatements[i]
		if !testRawStatement(t, stmt, tt.expectedValue) {
			return
		}
	}
}

func testRawStatement(t *testing.T, s ast.Statement, expectedValue string) bool {
	if s.TokenLiteral() != "raw" {
		t.Errorf("s.TokenLiteral not 'raw'. got=%q", s.TokenLiteral())
		return false
	}

	rawStmt, ok := s.(*ast.RawStatement)
	if !ok {
		t.Errorf("s not %T. got=%T", &ast.RawStatement{}, s)
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
	} elif (flag(FLAG_3)) {
		blah()
	} elif (!flag(FLAG_4)) {
		blah()
	} elif (var(VAR_1)) {
		blah()	
	} elif (!var(VAR_2)) {
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
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	scriptStmt := program.TopLevelStatements[0].(*ast.ScriptStatement)
	ifStmt := scriptStmt.Body.Statements[0].(*ast.IfStatement)
	testConditionExpression(t, ifStmt.Consequence.Expression.(*ast.OperatorExpression), token.VAR, "VAR_1", token.EQ, "1")
	testConditionExpression(t, ifStmt.ElifConsequences[0].Expression.(*ast.OperatorExpression), token.VAR, "VAR_2", token.NEQ, "2")
	testConditionExpression(t, ifStmt.ElifConsequences[1].Expression.(*ast.OperatorExpression), token.VAR, "VAR_3", token.LT, "3")
	testConditionExpression(t, ifStmt.ElifConsequences[2].Expression.(*ast.OperatorExpression), token.VAR, "VAR_4", token.LTE, "4")
	testConditionExpression(t, ifStmt.ElifConsequences[3].Expression.(*ast.OperatorExpression), token.VAR, "VAR_5", token.GT, "5")
	testConditionExpression(t, ifStmt.ElifConsequences[4].Expression.(*ast.OperatorExpression), token.VAR, "VAR_6", token.GTE, "6")
	testConditionExpression(t, ifStmt.ElifConsequences[5].Expression.(*ast.OperatorExpression), token.FLAG, "FLAG_1", token.EQ, token.TRUE)
	testConditionExpression(t, ifStmt.ElifConsequences[6].Expression.(*ast.OperatorExpression), token.FLAG, "FLAG_2 + BASE", token.EQ, token.FALSE)
	testConditionExpression(t, ifStmt.ElifConsequences[7].Expression.(*ast.OperatorExpression), token.FLAG, "FLAG_3", token.EQ, token.TRUE)
	testConditionExpression(t, ifStmt.ElifConsequences[8].Expression.(*ast.OperatorExpression), token.FLAG, "FLAG_4", token.EQ, token.FALSE)
	testConditionExpression(t, ifStmt.ElifConsequences[9].Expression.(*ast.OperatorExpression), token.VAR, "VAR_1", token.NEQ, "0")
	testConditionExpression(t, ifStmt.ElifConsequences[10].Expression.(*ast.OperatorExpression), token.VAR, "VAR_2", token.EQ, "0")
	nested := ifStmt.Consequence.Body.Statements[0].(*ast.IfStatement)
	testConditionExpression(t, nested.Consequence.Expression.(*ast.OperatorExpression), token.VAR, "VAR_7", token.NEQ, "1")

	if len(ifStmt.ElseConsequence.Statements) != 5 {
		t.Fatalf("len(ifStmt.ElseConsequences) should be '%d'. got=%d", 5, len(ifStmt.ElseConsequence.Statements))
	}
}

func testConditionExpression(t *testing.T, expression *ast.OperatorExpression, expectedType token.Type, expectedOperand string, expectedOperator token.Type, expectedComparisonValue string) {
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

func TestWhileStatements(t *testing.T) {
	input := `
script Test {
	while (var(VAR_1) < 1) {
		if (var(VAR_7) != 1) {
			continue
		}
		message()
	}
	while (flag(FLAG_1)) {
		message()
		break
	}
	do {
		message()
		break
	} while (var(VAR_1) > 2)
}
`
	l := lexer.New(input)
	p := New(l)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	scriptStmt := program.TopLevelStatements[0].(*ast.ScriptStatement)
	whileStmt := scriptStmt.Body.Statements[0].(*ast.WhileStatement)
	testConditionExpression(t, whileStmt.Consequence.Expression.(*ast.OperatorExpression), token.VAR, "VAR_1", token.LT, "1")
	ifStmt := whileStmt.Consequence.Body.Statements[0].(*ast.IfStatement)
	continueStmt := ifStmt.Consequence.Body.Statements[0].(*ast.ContinueStatement)
	if continueStmt.LoopStatment != whileStmt {
		t.Fatalf("continueStmt != whileStmt")
	}

	whileStmt = scriptStmt.Body.Statements[1].(*ast.WhileStatement)
	testConditionExpression(t, whileStmt.Consequence.Expression.(*ast.OperatorExpression), token.FLAG, "FLAG_1", token.EQ, "TRUE")
	breakStmt := whileStmt.Consequence.Body.Statements[1].(*ast.BreakStatement)
	if breakStmt.ScopeStatment != whileStmt {
		t.Fatalf("breakStmt != whileStmt")
	}

	doWhileStmt := scriptStmt.Body.Statements[2].(*ast.DoWhileStatement)
	testConditionExpression(t, doWhileStmt.Consequence.Expression.(*ast.OperatorExpression), token.VAR, "VAR_1", token.GT, "2")
	breakStmt = doWhileStmt.Consequence.Body.Statements[1].(*ast.BreakStatement)
	if breakStmt.ScopeStatment != doWhileStmt {
		t.Fatalf("breakStmt != doWhileStmt")
	}
}

func TestCompoundBooleanExpressions(t *testing.T) {
	input := `
script Test {
	if (var(VAR_1) < 1 || flag(FLAG_2) == true && var(VAR_3) > 4) {
		message()
	}
	if (var(VAR_1) < 1 && flag(FLAG_2) == true || var(VAR_3) > 4) {
		message()
	}
	if ((var(VAR_1) == 10 || ((var(VAR_1) == 12))) && !flag(FLAG_1)) {
		message()
	}
}
`
	l := lexer.New(input)
	p := New(l)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	scriptStmt := program.TopLevelStatements[0].(*ast.ScriptStatement)
	ifStmt := scriptStmt.Body.Statements[0].(*ast.IfStatement)
	ex := ifStmt.Consequence.Expression.(*ast.BinaryExpression)
	if ex.Operator != token.OR {
		t.Fatalf("ex.Operator != token.OR. Got '%s' instead.", ex.Operator)
	}
	op := ex.Left.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.VAR, "1", "VAR_1", token.LT)
	op = (ex.Right.(*ast.BinaryExpression)).Left.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.FLAG, "TRUE", "FLAG_2", token.EQ)
	op = (ex.Right.(*ast.BinaryExpression)).Right.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.VAR, "4", "VAR_3", token.GT)

	ifStmt = scriptStmt.Body.Statements[1].(*ast.IfStatement)
	ex = ifStmt.Consequence.Expression.(*ast.BinaryExpression)
	if ex.Operator != token.OR {
		t.Fatalf("ex.Operator != token.OR Got '%s' instead.", ex.Operator)
	}
	op = (ex.Left.(*ast.BinaryExpression)).Left.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.VAR, "1", "VAR_1", token.LT)
	op = (ex.Left.(*ast.BinaryExpression)).Right.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.FLAG, "TRUE", "FLAG_2", token.EQ)
	op = ex.Right.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.VAR, "4", "VAR_3", token.GT)

	ifStmt = scriptStmt.Body.Statements[2].(*ast.IfStatement)
	ex = ifStmt.Consequence.Expression.(*ast.BinaryExpression)
	if ex.Operator != token.AND {
		t.Fatalf("ex.Operator != token.AND Got '%s' instead.", ex.Operator)
	}
	op = (ex.Left.(*ast.BinaryExpression)).Left.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.VAR, "10", "VAR_1", token.EQ)
	op = (ex.Left.(*ast.BinaryExpression)).Right.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.VAR, "12", "VAR_1", token.EQ)
	op = ex.Right.(*ast.OperatorExpression)
	testOperatorExpression(t, op, token.FLAG, "FALSE", "FLAG_1", token.EQ)
}

func testOperatorExpression(t *testing.T, ex *ast.OperatorExpression, expectType token.Type, comparisonValue string, operand string, operator token.Type) {
	if ex.Type != expectType {
		t.Fatalf("ex.Type != %s. Got '%s' instead.", expectType, ex.Type)
	}
	if ex.ComparisonValue != comparisonValue {
		t.Fatalf("ex.ComparisonValue != %s. Got '%s' instead.", comparisonValue, ex.ComparisonValue)
	}
	if ex.Operand != operand {
		t.Fatalf("ex.Operand != %s. Got '%s' instead.", operand, ex.Operand)
	}
	if ex.Operator != operator {
		t.Fatalf("ex.Operator != %s. Got '%s' instead.", operator, ex.Operator)
	}
}

func TestSwitchStatements(t *testing.T) {
	input := `
script Test {
	switch (var(VAR_1)) {
		case 0: message1()
		case 1:
		case 2: case 3:
			message2()
			switch (var(VAR_2)) {
				case 67: message67()
			}
			blah
		default:
			message3()
	}
}
`
	l := lexer.New(input)
	p := New(l)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	scriptStmt := program.TopLevelStatements[0].(*ast.ScriptStatement)
	switchStmt, ok := scriptStmt.Body.Statements[0].(*ast.SwitchStatement)
	if !ok {
		t.Fatalf("not a switch statement\n")
	}
	if switchStmt.Operand != "VAR_1" {
		t.Fatalf("switchStmt.Operand != VAR_1. Got '%s' instead.", switchStmt.Operand)
	}
	if len(switchStmt.Cases) != 4 {
		t.Fatalf("len(switchStmt.Cases) != 4. Got '%d' instead.", len(switchStmt.Cases))
	}
	if switchStmt.DefaultCase == nil {
		t.Fatalf("switchStmt.DefaultCase == nil")
	}
	testSwitchCase(t, switchStmt.Cases[0], "0", 1)
	testSwitchCase(t, switchStmt.Cases[1], "1", 0)
	testSwitchCase(t, switchStmt.Cases[2], "2", 0)
	testSwitchCase(t, switchStmt.Cases[3], "3", 3)
	testSwitchCase(t, switchStmt.DefaultCase, "", 1)
	testSwitchCase(t, (switchStmt.Cases[3].Body.Statements[1].(*ast.SwitchStatement)).Cases[0], "67", 1)
}

func testSwitchCase(t *testing.T, sc *ast.SwitchCase, expectValue string, expectBodyLength int) {
	if sc.Value != expectValue {
		t.Fatalf("sc.Value != %s. Got '%s' instead.", expectValue, sc.Value)
	}
	if len(sc.Body.Statements) != expectBodyLength {
		t.Fatalf("len(sc.Body.Statements) != %d. Got '%d' instead.", expectBodyLength, len(sc.Body.Statements))
	}
}

func TestDuplicateTexts(t *testing.T) {
	input := `
script Script1 {
	msgbox("Hello$")
	msgbox("Goodbye$")
	msgbox("Hello\n"
		   "Multiline$")
}

script Script2 {
	msgbox("Test$")
	msgbox("Goodbye$")
	msgbox("Hello$")
	msgbox("Hello\n"
		"Multiline$", MSGBOX_DEFAULT)
}
`
	l := lexer.New(input)
	p := New(l)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(program.Texts) != 4 {
		t.Fatalf("len(program.Texts) != 4. Got '%d' instead.", len(program.Texts))
	}
}

func TestBreakContinueErrors(t *testing.T) {
	tests := []struct {
		input         string
		expectedError string
	}{
		{
			input: `
script Script1 {
	break
}
`,
			expectedError: "line 3: 'break' statement outside of any break-able scope",
		},
		{
			input: `
script Script1 {
	while (flag(FLAG_1)) {
		somestuff
		break
	}
	break
}
`,
			expectedError: "line 7: 'break' statement outside of any break-able scope",
		},
		{
			input: `
script Script1 {
	while (flag(FLAG_1)) {
		somestuff
		continue
	}
	continue
}
`,
			expectedError: "line 7: 'continue' statement outside of any continue-able scope",
		},
		{
			input: `
script Script1 {
	switch (var(VAR_1)) {
	case 1:
		msgbox
		break
	case 2:
		continue
	}
}
`,
			expectedError: "line 8: 'continue' statement outside of any continue-able scope",
		},
	}

	for _, test := range tests {
		testForParseError(t, test.input, test.expectedError)
	}
}

func testForParseError(t *testing.T, input string, expectedErrorText string) {
	l := lexer.New(input)
	p := New(l)
	_, err := p.ParseProgram()
	if err == nil {
		t.Fatalf("Expected error '%s', but no error occurred", expectedErrorText)
	}
	if err.Error() != expectedErrorText {
		t.Fatalf("Expected error '%s', but got '%s'", expectedErrorText, err.Error())
	}
}
