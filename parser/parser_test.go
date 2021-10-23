package parser

import (
	"regexp"
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
	bufferitemname("Bar", 0, VAR_BUG_CONTEST_PRIZE, ascii"Foo", format(braille"Baz"))
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
	p := New(l, "../font_widths.json", "", 0, nil)
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
			{"bufferitemname", []string{"MyScript_Text_0", "0", "VAR_BUG_CONTEST_PRIZE", "MyScript_Text_1", "MyScript_Text_2"}},
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

func TestPoryswitchStatements(t *testing.T) {
	input := `
script MyScript {
	lock
	poryswitch(GAME_VERSION) {
		RUBY: foo
		SAPPHIRE {
			bar
			poryswitch(LANG) {
				DE {
					de_command
					de_command2
				}
				EN: en_command
				_: fallback
			}
			baz
		}
	}
	release
}

text MyText {
	poryswitch(LANG) {
		DE: ascii"Das ist MAY's Haus."
		_: format("formatted")
		EN {
			"Two\n"
			"Lines"
		}
	}
}

movement MyMovement {
	walk_up
	poryswitch(GAME_VERSION) {
		_: walk_default
		RUBY: walk_ruby * 2
		SAPPHIRE {
			face_up
			poryswitch(LANG) {
				DE: face_de
				_ { face_default * 2  }
			}
			face_down
		}
	}
	walk_down
}
`
	scriptTests := []struct {
		switches map[string]string
		commands []string
	}{
		{map[string]string{"GAME_VERSION": "RUBY", "LANG": "FR"}, []string{"lock", "foo", "release"}},
		{map[string]string{"GAME_VERSION": "SAPPHIRE", "LANG": "DE"}, []string{"lock", "bar", "de_command", "de_command2", "baz", "release"}},
		{map[string]string{"GAME_VERSION": "SAPPHIRE", "LANG": "BLAH"}, []string{"lock", "bar", "fallback", "baz", "release"}},
	}

	for _, tt := range scriptTests {
		l := lexer.New(input)
		p := New(l, "../font_widths.json", "", 208, tt.switches)
		program, err := p.ParseProgram()
		if err != nil {
			t.Fatalf(err.Error())
		}
		stmt := program.TopLevelStatements[0].(*ast.ScriptStatement)
		if len(stmt.Body.Statements) != len(tt.commands) {
			t.Fatalf("Incorrect number of statements. Expected %d, got %d", len(tt.commands), len(stmt.Body.Statements))
		}
		for i, expectedCommand := range tt.commands {
			command := stmt.Body.Statements[i].TokenLiteral()
			if expectedCommand != command {
				t.Fatalf("Incorrect statement %d. Expected %s, got %s", i, expectedCommand, command)
			}
		}
	}

	textTests := []struct {
		switches map[string]string
		text     string
	}{
		{map[string]string{"GAME_VERSION": "RUBY", "LANG": "DE"}, "Das ist MAY's Haus.\\0"},
		{map[string]string{"GAME_VERSION": "SAPPHIRE", "LANG": "EN"}, "Two\\n\nLines$"},
		{map[string]string{"GAME_VERSION": "SAPPHIRE", "LANG": "BLAH"}, "formatted$"},
	}

	for i, tt := range textTests {
		l := lexer.New(input)
		p := New(l, "../font_widths.json", "", 208, tt.switches)
		program, err := p.ParseProgram()
		if err != nil {
			t.Fatalf(err.Error())
		}
		stmt := program.TopLevelStatements[1].(*ast.TextStatement)
		text := stmt.Value
		if tt.text != text {
			t.Fatalf("Incorrect text %d. Expected %s, got %s", i, tt.text, text)
		}
	}

	movementTests := []struct {
		switches map[string]string
		commands []string
	}{
		{map[string]string{"GAME_VERSION": "RUBY", "LANG": "DE"}, []string{"walk_up", "walk_ruby", "walk_ruby", "walk_down"}},
		{map[string]string{"GAME_VERSION": "SAPPHIRE", "LANG": "DE"}, []string{"walk_up", "face_up", "face_de", "face_down", "walk_down"}},
		{map[string]string{"GAME_VERSION": "SAPPHIRE", "LANG": "EN"}, []string{"walk_up", "face_up", "face_default", "face_default", "face_down", "walk_down"}},
	}

	for _, tt := range movementTests {
		l := lexer.New(input)
		p := New(l, "../font_widths.json", "", 208, tt.switches)
		program, err := p.ParseProgram()
		if err != nil {
			t.Fatalf(err.Error())
		}
		stmt := program.TopLevelStatements[2].(*ast.MovementStatement)
		if len(stmt.MovementCommands) != len(tt.commands) {
			t.Fatalf("Incorrect number of movement commands. Expected %d, got %d", len(tt.commands), len(stmt.MovementCommands))
		}
		for i, expectedCommand := range tt.commands {
			command := stmt.MovementCommands[i]
			if expectedCommand != command {
				t.Fatalf("Incorrect movement command %d. Expected %s, got %s", i, expectedCommand, command)
			}
		}
	}
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
	p := New(l, "", "", 0, nil)
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
	} elif (defeated(TRAINER_GARY)) {
		blah()
	} elif (!defeated(TRAINER_BLUE)) {
		blah()
	} elif (defeated(TRAINER_GERALD) == TRUE) {
		blah()
	} elif (defeated(TRAINER_AXLE) == false) {
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
	p := New(l, "", "", 0, nil)
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
	testConditionExpression(t, ifStmt.ElifConsequences[11].Expression.(*ast.OperatorExpression), token.DEFEATED, "TRAINER_GARY", token.EQ, token.TRUE)
	testConditionExpression(t, ifStmt.ElifConsequences[12].Expression.(*ast.OperatorExpression), token.DEFEATED, "TRAINER_BLUE", token.EQ, token.FALSE)
	testConditionExpression(t, ifStmt.ElifConsequences[13].Expression.(*ast.OperatorExpression), token.DEFEATED, "TRAINER_GERALD", token.EQ, token.TRUE)
	testConditionExpression(t, ifStmt.ElifConsequences[14].Expression.(*ast.OperatorExpression), token.DEFEATED, "TRAINER_AXLE", token.EQ, token.FALSE)
	nested := ifStmt.Consequence.Body.Statements[0].(*ast.IfStatement)
	testConditionExpression(t, nested.Consequence.Expression.(*ast.OperatorExpression), token.VAR, "VAR_7", token.NEQ, "1")

	if len(ifStmt.ElseConsequence.Statements) != 5 {
		t.Errorf("len(ifStmt.ElseConsequences) should be '%d'. got=%d", 5, len(ifStmt.ElseConsequence.Statements))
	}
}

func testConditionExpression(t *testing.T, expression *ast.OperatorExpression, expectedType token.Type, expectedOperand string, expectedOperator token.Type, expectedComparisonValue string) {
	if expression.Type != expectedType {
		t.Errorf("expression.Type not '%s'. got=%s", expectedType, expression.Type)
	}
	if expression.Operand != expectedOperand {
		t.Errorf("expression.Operand not '%s'. got=%s", expectedOperand, expression.Operand)
	}
	if expression.Operator != expectedOperator {
		t.Errorf("expression.Operator not '%s'. got=%s", expectedOperator, expression.Operator)
	}
	if expression.ComparisonValue != expectedComparisonValue {
		t.Errorf("expression.ComparisonValue not '%s'. got=%s", expectedComparisonValue, expression.ComparisonValue)
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
	p := New(l, "", "", 0, nil)
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
	p := New(l, "", "", 0, nil)
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
	p := New(l, "", "", 0, nil)
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
	if len(switchStmt.Cases) != 5 {
		t.Fatalf("len(switchStmt.Cases) != 5. Got '%d' instead.", len(switchStmt.Cases))
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
	msgbox(braille"Goodbye$")
	msgbox(ascii"StringType$")
	msgbox("Hello\n"
		   "Multiline$")
}

script Script2 {
	msgbox("Test$")
	msgbox(braille"Goodbye$")
	msgbox("Hello$")
	msgbox("StringType$")
	msgbox("Hello\n"
		"Multiline$", MSGBOX_DEFAULT)
}
`
	l := lexer.New(input)
	p := New(l, "", "", 0, nil)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(program.Texts) != 6 {
		t.Fatalf("len(program.Texts) != 6. Got '%d' instead.", len(program.Texts))
	}
}

func TestTextStatements(t *testing.T) {
	input := `
script MyScript1 {
	msgbox(ascii"Test$")
	msgbox("Hello$")
}

text MyText1 {
	"Hello$"
}

text MyText2 {
	"Foo"
	"Bar$"
}

script MyScript1 {
	foo()
}
`
	l := lexer.New(input)
	p := New(l, "", "", 0, nil)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(program.Texts) != 4 {
		t.Fatalf("len(program.Texts) != 4. Got '%d' instead.", len(program.Texts))
	}
}

func TestFormatOperator(t *testing.T) {
	input := `
script MyScript1 {
	msgbox(format("Test»{BLAH}$"))
}

text MyText {
	format("FooBar", "TEST")
}

text MyText1 {
	format("FooBar", "TEST", 100)
}

text MyText2 {
	format("FooBar", 100, "TEST")
}
`
	l := lexer.New(input)
	p := New(l, "../font_widths.json", "1_latin_frlg", 150, nil)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(program.Texts) != 4 {
		t.Fatalf("len(program.Texts) != 3. Got '%d' instead.", len(program.Texts))
	}
	if program.Texts[0].Value != "Test»{BLAH}$" {
		t.Fatalf("Incorrect format() evaluation. Got '%s' instead of '%s'", program.Texts[0].Value, "Test»{BLAH}$")
	}
	if program.Texts[1].Value != "FooBar$" {
		t.Fatalf("Incorrect format() evaluation. Got '%s' instead of '%s'", program.Texts[1].Value, "FooBar$")
	}
	if program.Texts[2].Value != "FooBar$" {
		t.Fatalf("Incorrect format() evaluation. Got '%s' instead of '%s'", program.Texts[2].Value, "FooBar$")
	}
	if program.Texts[3].Value != "FooBar$" {
		t.Fatalf("Incorrect format() evaluation. Got '%s' instead of '%s'", program.Texts[3].Value, "FooBar$")
	}
}

func TestMovementStatements(t *testing.T) {
	input := `
movement MyMovement {
	walk_up
	walk_down walk_left
	step_end
}

movement MyMovement2 {
}

movement MyMovement3 {
	run_up * 3
	face_down
	delay_16*2
}
`
	l := lexer.New(input)
	p := New(l, "", "", 0, nil)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(program.TopLevelStatements) != 3 {
		t.Fatalf("len(program.TopLevelStatements) != 3. Got '%d' instead.", len(program.TopLevelStatements))
	}
	testMovement(t, program.TopLevelStatements[0], "MyMovement", []string{"walk_up", "walk_down", "walk_left", "step_end"})
	testMovement(t, program.TopLevelStatements[1], "MyMovement2", []string{})
	testMovement(t, program.TopLevelStatements[2], "MyMovement3", []string{"run_up", "run_up", "run_up", "face_down", "delay_16", "delay_16"})
}

func testMovement(t *testing.T, stmt ast.Statement, expectedName string, expectedCommands []string) {
	movementStmt := stmt.(*ast.MovementStatement)
	if movementStmt.Name.Value != expectedName {
		t.Errorf("Incorrect movement name. Got '%s' instead of '%s'", movementStmt.Name.Value, expectedName)
	}
	if len(movementStmt.MovementCommands) != len(expectedCommands) {
		t.Fatalf("Incorrect number of movement commands. Got %d commands instead of %d", len(movementStmt.MovementCommands), len(expectedCommands))
	}
	for i, cmd := range expectedCommands {
		if movementStmt.MovementCommands[i] != cmd {
			t.Errorf("Incorrect movement command at index %d. Got '%s' instead of '%s'", i, movementStmt.MovementCommands[i], cmd)
		}
	}
}

func TestMapScriptStatements(t *testing.T) {
	input := `
mapscripts MyMap_MapScripts {
	MAP_SCRIPT_ON_TRANSITION: MyMap_OnTransition
	MAP_SCRIPT_ON_RESUME: MyMap_OnResume
	MAP_SCRIPT_ON_TRANSITION {
		lock
		release
	}
	MAP_SCRIPT_ON_FRAME_TABLE [
		VAR_LITTLEROOT_INTRO_STATE, 1: MyMap_OnFrame_FirstThing
		VAR_LITTLEROOT_INTRO_STATE    + 2, 1+ BASE_THING: MyMap_OnFrame_2
		VAR_OTHER, 3 {
			lock
			foo(1, 2)
			release
		}
	]
	MAP_SCRIPT_ON_WARP_INTO_MAP_TABLE [
		VAR_TEMP_1, 0 { foo }
		VAR_TEMP_1, 1 { some more commands }
	]
}
`
	l := lexer.New(input)
	p := New(l, "", "", 0, nil)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(program.TopLevelStatements) != 1 {
		t.Fatalf("len(program.TopLevelStatements) != 1. Got '%d' instead.", len(program.TopLevelStatements))
	}
	stmt := program.TopLevelStatements[0].(*ast.MapScriptsStatement)
	if stmt.Name.Value != "MyMap_MapScripts" {
		t.Errorf("Incorrect mapscripts name. Got '%s' instead of '%s'", stmt.Name.Value, "MyMap_MapScripts")
	}
	if len(stmt.MapScripts) != 3 {
		t.Fatalf("Incorrect length of MapScripts. Got '%d' instead of '%d'", len(stmt.MapScripts), 3)
	}
	testNamedMapScript(t, stmt.MapScripts[0], "MyMap_OnTransition", "MAP_SCRIPT_ON_TRANSITION")
	testNamedMapScript(t, stmt.MapScripts[1], "MyMap_OnResume", "MAP_SCRIPT_ON_RESUME")
	testScriptedMapScript(t, stmt.MapScripts[2], "MyMap_MapScripts_MAP_SCRIPT_ON_TRANSITION", "MAP_SCRIPT_ON_TRANSITION", 2)

	if len(stmt.TableMapScripts) != 2 {
		t.Fatalf("Incorrect length of TableMapScripts. Got '%d' instead of '%d'", len(stmt.TableMapScripts), 1)
	}
	testTableMapScript(t, stmt.TableMapScripts[0], "MyMap_MapScripts_MAP_SCRIPT_ON_FRAME_TABLE", "MAP_SCRIPT_ON_FRAME_TABLE", 3)
	testTableMapScriptEntry(t, stmt.TableMapScripts[0].Entries[0], "VAR_LITTLEROOT_INTRO_STATE", "1", "MyMap_OnFrame_FirstThing")
	testTableMapScriptEntry(t, stmt.TableMapScripts[0].Entries[1], "VAR_LITTLEROOT_INTRO_STATE + 2", "1 + BASE_THING", "MyMap_OnFrame_2")
	testScriptedTableMapScriptEntry(t, stmt.TableMapScripts[0].Entries[2], "VAR_OTHER", "3", "MyMap_MapScripts_MAP_SCRIPT_ON_FRAME_TABLE_2", 3)
	testTableMapScript(t, stmt.TableMapScripts[1], "MyMap_MapScripts_MAP_SCRIPT_ON_WARP_INTO_MAP_TABLE", "MAP_SCRIPT_ON_WARP_INTO_MAP_TABLE", 2)
	testScriptedTableMapScriptEntry(t, stmt.TableMapScripts[1].Entries[0], "VAR_TEMP_1", "0", "MyMap_MapScripts_MAP_SCRIPT_ON_WARP_INTO_MAP_TABLE_0", 1)
	testScriptedTableMapScriptEntry(t, stmt.TableMapScripts[1].Entries[1], "VAR_TEMP_1", "1", "MyMap_MapScripts_MAP_SCRIPT_ON_WARP_INTO_MAP_TABLE_1", 3)
}

func testNamedMapScript(t *testing.T, mapScript ast.MapScript, expectedName string, expectedType string) {
	if mapScript.Name != expectedName {
		t.Errorf("Incorrect mapScript name. Got '%s' instead of '%s'", mapScript.Name, expectedName)
	}
	if mapScript.Type != expectedType {
		t.Errorf("Incorrect mapScript type. Got '%s' instead of '%s'", mapScript.Type, expectedType)
	}
	if mapScript.Script != nil {
		t.Errorf("mapScript is supposed to be nil")
	}
}

func testScriptedMapScript(t *testing.T, mapScript ast.MapScript, expectedName string, expectedType string, expectedNumStatements int) {
	if mapScript.Name != expectedName {
		t.Errorf("Incorrect scripted mapScript name. Got '%s' instead of '%s'", mapScript.Name, expectedName)
	}
	if mapScript.Type != expectedType {
		t.Errorf("Incorrect scripted mapScript type. Got '%s' instead of '%s'", mapScript.Type, expectedType)
	}
	if mapScript.Script == nil {
		t.Errorf("mapScript.Script is not supposed to be nil")
	}
	if len(mapScript.Script.Body.Statements) != expectedNumStatements {
		t.Errorf("Incorrect scripted mapScript number of statements. Got '%d' instead of '%d'", len(mapScript.Script.Body.Statements), expectedNumStatements)
	}
}

func testTableMapScript(t *testing.T, mapScript ast.TableMapScript, expectedName string, expectedType string, expectedNumEntries int) {
	if mapScript.Name != expectedName {
		t.Errorf("Incorrect table mapScript name. Got '%s' instead of '%s'", mapScript.Name, expectedName)
	}
	if mapScript.Type != expectedType {
		t.Errorf("Incorrect table mapScript type. Got '%s' instead of '%s'", mapScript.Type, expectedType)
	}
	if len(mapScript.Entries) != expectedNumEntries {
		t.Errorf("Incorrect table mapScript number of entries. Got '%d' instead of '%d'", len(mapScript.Entries), expectedNumEntries)
	}
}

func testTableMapScriptEntry(t *testing.T, mapScriptEntry ast.TableMapScriptEntry, expectedCondition, expectedComparison, expectedName string) {
	if mapScriptEntry.Condition != expectedCondition {
		t.Errorf("Incorrect table mapScript entry condition. Got '%s' instead of '%s'", mapScriptEntry.Condition, expectedCondition)
	}
	if mapScriptEntry.Comparison != expectedComparison {
		t.Errorf("Incorrect table mapScript entry comparison. Got '%s' instead of '%s'", mapScriptEntry.Comparison, expectedComparison)
	}
	if mapScriptEntry.Name != expectedName {
		t.Errorf("Incorrect table mapScript entry name. Got '%s' instead of '%s'", mapScriptEntry.Name, expectedName)
	}
	if mapScriptEntry.Script != nil {
		t.Errorf("mapScriptEntry.Script is supposed to be nil")
	}
}

func testScriptedTableMapScriptEntry(t *testing.T, mapScriptEntry ast.TableMapScriptEntry, expectedCondition, expectedComparison, expectedName string, expectedNumStatements int) {
	if mapScriptEntry.Condition != expectedCondition {
		t.Errorf("Incorrect table mapScript entry condition. Got '%s' instead of '%s'", mapScriptEntry.Condition, expectedCondition)
	}
	if mapScriptEntry.Comparison != expectedComparison {
		t.Errorf("Incorrect table mapScript entry comparison. Got '%s' instead of '%s'", mapScriptEntry.Comparison, expectedComparison)
	}
	if mapScriptEntry.Name != expectedName {
		t.Errorf("Incorrect table mapScript entry name. Got '%s' instead of '%s'", mapScriptEntry.Name, expectedName)
	}
	if mapScriptEntry.Script == nil {
		t.Errorf("mapScriptEntry.Script is supposed to be nil")
	}
	if len(mapScriptEntry.Script.Body.Statements) != expectedNumStatements {
		t.Errorf("Incorrect scripted mapScriptEntry number of statements. Got '%d' instead of '%d'", len(mapScriptEntry.Script.Body.Statements), expectedNumStatements)
	}
}

func TestScopeModifiers(t *testing.T) {
	input := `
script Script1 {}
script(local) Script2 {}
script(global) Script3 {}
text Text1 {"test"}
text(local) Text2 {"test"}
text(global) Text3 {"test"}
movement Movement1 {}
movement(local) Movement2 {}
movement(global) Movement3 {}
mapscripts MapScripts1 {}
mapscripts(local) MapScripts2 {}
mapscripts(global) MapScripts3 {}
`
	l := lexer.New(input)
	p := New(l, "", "", 0, nil)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(program.TopLevelStatements) != 12 {
		t.Fatalf("len(program.TopLevelStatements) != 12. Got '%d' instead.", len(program.TopLevelStatements))
	}
	expectedScopes := []token.Type{
		token.GLOBAL,
		token.LOCAL,
		token.GLOBAL,
		token.GLOBAL,
		token.LOCAL,
		token.GLOBAL,
		token.LOCAL,
		token.LOCAL,
		token.GLOBAL,
		token.GLOBAL,
		token.LOCAL,
		token.GLOBAL,
	}
	for i, expectedScope := range expectedScopes {
		testScope(t, i, program.TopLevelStatements[i], expectedScope)
	}
}

func testScope(t *testing.T, i int, statement ast.Statement, expectedScope token.Type) {
	if script, ok := statement.(*ast.ScriptStatement); ok {
		if script.Scope != expectedScope {
			t.Errorf("%d: Expected script scope %s, but got %s", i, expectedScope, script.Scope)
		}
	} else if text, ok := statement.(*ast.TextStatement); ok {
		if text.Scope != expectedScope {
			t.Errorf("%d: Expected text scope %s, but got %s", i, expectedScope, text.Scope)
		}
	} else if movement, ok := statement.(*ast.MovementStatement); ok {
		if movement.Scope != expectedScope {
			t.Errorf("%d: Expected movement scope %s, but got %s", i, expectedScope, movement.Scope)
		}
	} else if mapscripts, ok := statement.(*ast.MapScriptsStatement); ok {
		if mapscripts.Scope != expectedScope {
			t.Errorf("%d: Expected mapscripts scope %s, but got %s", i, expectedScope, mapscripts.Scope)
		}
	}
}

func TestConstants(t *testing.T) {
	input := `
const FOO = 2
const BAR = FLAG_TEMP_1 +   3
	- FLAG_BASE   const PROF_BIRCH = FOO+1
const PROF_ELM = PROF_BIRCH - 1

script Script1 {
	command1(FOO)
	command2(1, 2 + BAR)
	command3(2, (PROF_ELM) - ((PROF_BIRCH + FOO)))
	if (flag(PROF_ELM)) {}
	if (var(PROF_BIRCH) == PROF_ELM +1) {}
	switch (var(PROF_ELM)) {
		case FOO: commandfood()
		default: command()
	}
}

mapscripts MyMapScript {
	MAP_SCRIPT_ON_FRAME_TABLE [
		PROF_ELM, FOO: MyOnFrameScript
	]
}
`
	l := lexer.New(input)
	p := New(l, "", "", 0, nil)
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	script := program.TopLevelStatements[0].(*ast.ScriptStatement)
	command1 := script.Body.Statements[0].(*ast.CommandStatement)
	command2 := script.Body.Statements[1].(*ast.CommandStatement)
	command3 := script.Body.Statements[2].(*ast.CommandStatement)
	testConstant(t, "2", command1.Args[0])
	testConstant(t, "2 + FLAG_TEMP_1 + 3 - FLAG_BASE", command2.Args[1])
	testConstant(t, "( 2 + 1 - 1 ) - ( ( 2 + 1 + 2 ) )", command3.Args[1])

	if1 := script.Body.Statements[3].(*ast.IfStatement)
	op1 := if1.Consequence.Expression.(*ast.OperatorExpression)
	testConstant(t, "2 + 1 - 1", op1.Operand)

	if2 := script.Body.Statements[4].(*ast.IfStatement)
	op2 := if2.Consequence.Expression.(*ast.OperatorExpression)
	testConstant(t, "2 + 1", op2.Operand)
	testConstant(t, "2 + 1 - 1 + 1", op2.ComparisonValue)

	sw := script.Body.Statements[5].(*ast.SwitchStatement)
	testConstant(t, "2 + 1 - 1", sw.Operand)
	testConstant(t, "2", sw.Cases[0].Value)

	ms := program.TopLevelStatements[1].(*ast.MapScriptsStatement)
	frame := ms.TableMapScripts[0].Entries[0]
	testConstant(t, "2 + 1 - 1", frame.Condition)
	testConstant(t, "2", frame.Comparison)
}

func testConstant(t *testing.T, expected, actual string) {
	if actual != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, actual)
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		input              string
		expectedError      string
		expectedErrorRegex string
	}{
		{
			input: `
script Script1 {
	break
}`,
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
}`,
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
}`,
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
}`,
			expectedError: "line 8: 'continue' statement outside of any continue-able scope",
		},
		{
			input: `
script Script1 {}
raw ` + "``" + `
invalid
`,
			expectedError: "line 4: could not parse top-level statement for 'invalid'",
		},
		{
			input: `
raw "stuff"
`,
			expectedError: "line 2: raw statement must begin with a backtick character '`'",
		},
		{
			input: `
script {
	foo
}`,
			expectedError: "line 2: missing name for script",
		},
		{
			input: `
script MyScript
	foo
}`,
			expectedError: "line 2: missing opening curly brace for script 'MyScript'",
		},
		{
			input: `
script MyScript {
	if (var(VAR_1)) {
	foo
}`,
			expectedError: "line 3: missing closing curly brace for block statement",
		},
		{
			input: `
script MyScript {
	switch (var(VAR_1)) {
	case 1: foo
`,
			expectedError: "line 4: missing end for switch case body",
		},
		{
			input: `
script MyScript {
	foo
	<
}`,
			expectedError: "line 4: could not parse statement for '<'",
		},
		{
			input: `
script MyScript {
	if (flag(FLAG_1)) {
		foo
	} else 
		bar
	}
}`,
			expectedError: "line 5: missing opening curly brace of else statement",
		},
		{
			input: `
script MyScript {
	do 
		foo
	while (flag(FLAG_1))
}`,
			expectedError: "line 3: missing opening curly brace of do...while statement",
		},
		{
			input: `
script MyScript {
	do {
		foo
	} (flag(FLAG_1))
}`,
			expectedError: "line 5: missing 'while' after body of do...while statement",
		},
		{
			input: `
script MyScript {
	do {
		foo
	} while flag(FLAG_1)
}`,
			expectedError: "line 5: missing '(' to start condition for do...while statement",
		},
		{
			input: `
script MyScript {
	while (flag(FLAG_1)) {
		continue
		foo
	}
}`,
			expectedError: "line 5: 'continue' must be the last statement in block scope",
		},
		{
			input: `
script MyScript {
	switch flag(FLAG_1) {
	case 1:
		foo
	}
}`,
			expectedError: "line 3: missing opening parenthesis of switch statement operand",
		},
		{
			input: `
script MyScript {
	switch (flag(FLAG_1)) {
	case 1:
		foo
	}
}`,
			expectedError: "line 3: invalid switch statement operand 'flag'. Must be 'var`",
		},
		{
			input: `
script MyScript {
	switch (var FLAG_1) {
	case 1:
		foo
	}
}`,
			expectedError: "line 3: missing '(' after var operator. Got 'FLAG_1` instead",
		},
		{
			input: `
script MyScript {
	switch (var(FLAG_1))
	case 1:
		foo
	}
}`,
			expectedError: "line 3: missing opening curly brace of switch statement",
		},
		{
			input: `
script MyScript {
	switch (var(FLAG_1)) {
	case 1:
	case 2:
		foo
	case 1:
		bar
	}
}`,
			expectedError: "line 7: duplicate switch cases detected for case '1'",
		},
		{
			input: `
script MyScript {
	switch (var(FLAG_1)) {
	case 2:
		foo
	default
		baz
	}
}`,
			expectedError: "line 6: missing `:` after default",
		},
		{
			input: `
script MyScript {
	switch (var(FLAG_1)) {
	case 2:
	case 7
		foo
	}
}`,
			expectedError: "line 5: missing `:` after 'case'",
		},
		{
			input: `
script MyScript {
	switch (var(FLAG_1)) {
		foo
	}`,
			expectedError: "line 4: invalid start of switch case 'foo'. Expected 'case' or 'default'",
		},
		{
			input: `
script MyScript {
	switch (var(FLAG_1)) {
	}`,
			expectedError: "line 3: switch statement has no cases or default case",
		},
		{
			input: `
script MyScript {
	if var(FLAG_1)) {
	}`,
			expectedError: "line 3: missing '(' to start boolean expression",
		},
		{
			input: `
script MyScript {
	if (var(FLAG_1) ||) {
	}`,
			expectedError: "line 3: left side of binary expression must be var(), flag(), or defeated() operator. Instead, found ')'",
		},
		{
			input: `
script MyScript {
	if (var(FLAG_1)) 
		foo
	}`,
			expectedError: "line 4: expected next token to be '{', got 'foo' instead",
		},
		{
			input: `
script MyScript {
	if (var(FLAG_1) && (var(VAR_1) == 1 {
		foo
	}`,
			expectedError: "line 3: missing ')', '&&' or '||' when evaluating 'var' operator",
		},
		{
			input: `
script MyScript {
	if (var{FLAG_1) {
		foo
	}`,
			expectedError: "line 3: missing opening parenthesis for condition operator 'VAR'",
		},
		{
			input: `
script MyScript {
	if (flag{FLAG_1) {
		foo
	}`,
			expectedError: "line 3: missing opening parenthesis for condition operator 'FLAG'",
		},
		{
			input: `
script MyScript {
	if (flag()) {
		foo
	}`,
			expectedError: "line 3: missing value for condition operator 'FLAG'",
		},
		{
			input: `
script MyScript {
	foo(sdfa
	bar()
}`,
			expectedError: "line 3: missing closing parenthesis for command 'foo'",
		},
		{
			input: `
script MyScript {
	if (flag(FLAG_1)) {
		foo
	} elif (fla(FLAG_2)) {
		bar
	}
}`,
			expectedError: "line 5: left side of binary expression must be var(), flag(), or defeated() operator. Instead, found 'fla'",
		},
		{
			input: `
script MyScript {
	if (flag(FLAG_1)) {
		foo
	} else {
		else
	}
}`,
			expectedError: "line 6: could not parse statement for 'else'",
		},
		{
			input: `
script MyScript {
	do {
		continue
		break
	} while (flag(FLAG_1))
}`,
			expectedError: "line 5: 'continue' must be the last statement in block scope",
		},
		{
			input: `
script MyScript {
	do {
		continue
	} while (flag(FLAG_1) == 45)
}`,
			expectedError: "line 5: invalid flag comparison value '45'. Only TRUE and FALSE are allowed",
		},
		{
			input: `
script MyScript {
	switch (var(VAR_1)) {
	default:
		foo
	case 1:
		bar
	default:
		baz
	}
}`,
			expectedError: "line 8: multiple `default` cases found in switch statement. Only one `default` case is allowed",
		},
		{
			input: `
script MyScript {
	switch (var(VAR_1)) {
	case 1:
		bar
	default:
		baz
		continue
	}
}`,
			expectedError: "line 8: 'continue' statement outside of any continue-able scope",
		},
		{
			input: `
script MyScript {
	while ((flag(FLAG_1) {
		foo
	}
}`,
			expectedError: "line 3: missing closing ')' for nested boolean expression",
		},
		{
			input: `
script MyScript {
	while (flag(FLAG_1 {
		foo
	}
}`,
			expectedError: "line 3: missing closing ')' for condition operator value",
		},
		{
			input: `
script MyScript {
	while (flag(FLAG_1) == ) {
		foo
	}
}`,
			expectedError: "line 3: missing comparison value for flag operator",
		},
		{
			input: `
script MyScript {
	while (var(VAR_1) == ) {
		foo
	}
}`,
			expectedError: "line 3: missing comparison value for var operator",
		},
		{
			input: `
script MyScript {
	if (defeated(TRAINER_FOO) == ) {
		foo
	}
}`,
			expectedError: "line 3: missing comparison value for defeated operator",
		},
		{
			input: `
script MyScript {
	while (var(VAR_1) == 1 && flag(FLAG_1) == true && flag()) {
		foo
	}
}`,
			expectedError: "line 3: missing value for condition operator 'FLAG'",
		},
		{
			input: `
text {
	"MyText$"
}`,
			expectedError: "line 2: missing name for text statement",
		},
		{
			input: `
text Text1
	"MyText$"
}`,
			expectedError: "line 3: missing opening curly brace for text 'Text1'",
		},
		{
			input: `
text Text1 {
	nottext
	"MyText$"
}`,
			expectedError: "line 3: body of text statement must be a string or formatted string. Got 'nottext' instead",
		},
		{
			input: `
text Text1 {
	"MyText$"
	notcurlybrace
}`,
			expectedError: "line 4: expected closing curly brace for text. Got 'notcurlybrace' instead",
		},
		{
			input: `
script Script1 {
	msgbox("Hello")
}
text Script1_Text_0 {
	"MyText$"
}`,
			expectedError: "Duplicate text label 'Script1_Text_0'. Choose a unique label that won't clash with the auto-generated text labels",
		},
		{
			input: `
movement {
	
}`,
			expectedError: "line 2: missing name for movement statement",
		},
		{
			input: `
movement Foo
	walk_up
}`,
			expectedError: "line 3: missing opening curly brace for movement 'Foo'",
		},
		{
			input: `
movement Foo {
	+
}`,
			expectedError: "line 3: expected movement command, but got '+' instead",
		},
		{
			input: `
movement Foo {
	walk_up * walk_down
}`,
			expectedError: "line 3: expected mulplier number for movement command, but got 'walk_down' instead",
		},
		{
			input: `
movement Foo {
	walk_up * 999999999999999999999999999999999
}`,
			expectedError: "line 3: invalid movement mulplier integer '999999999999999999999999999999999': strconv.ParseInt: parsing \"999999999999999999999999999999999\": value out of range",
		},
		{
			input: `
movement Foo {
	walk_up * 10000
}`,
			expectedError: "line 3: movement mulplier '10000' is too large. Maximum is 9999",
		},
		{
			input: `
movement Foo {
	walk_up * 0
}`,
			expectedError: "line 3: movement mulplier must be a positive integer, but got '0' instead",
		},
		{
			input: `
movement Foo {
	walk_up * -2
}`,
			expectedError: "line 3: movement mulplier must be a positive integer, but got '-2' instead",
		},
		{
			input: `
text Foo {
	format asdf
}`,
			expectedError: "line 3: format operator must begin with an open parenthesis '('",
		},
		{
			input: `
text Foo {
	format()
}`,
			expectedError: "line 3: invalid format() argument ')'. Expected a string literal",
		},
		{
			input: `
text Foo {
	format("Hi", )
}`,
			expectedError: "line 3: invalid format() parameter ')'. Expected either fontId (string) or maxLineLength (integer)",
		},
		{
			input: `
text Foo {
	format("Hi"
}`,
			expectedError: "line 4: missing closing parenthesis ')' for format()",
		},
		{
			input: `
text Foo {
	format("Hi", "TEST", "NOT_AN_INT")
}`,
			expectedError: "line 3: invalid format() maxLineLength 'NOT_AN_INT'. Expected integer",
		},
		{
			input: `
text Foo {
	format("Hi", 100, 42)
}`,
			expectedError: "line 3: invalid format() fontId '42'. Expected string",
		},
		{
			input: `
script Foo {
	msgbox(format("Hi", ))
}`,
			expectedError: "line 3: invalid format() parameter ')'. Expected either fontId (string) or maxLineLength (integer)",
		},
		{
			input: `
text Foo {
	format("Hi", "invalidFontID")
}`,
			expectedErrorRegex: `line 3: Unknown fontID 'invalidFontID' used in format\(\)\. List of valid fontIDs are '\[((1_latin)|(1_latin_frlg)| )+\]'`,
		},
		{
			input: `
mapscripts {
}`,
			expectedError: "line 2: missing name for mapscripts statement",
		},
		{
			input: `
mapscripts MyMapScripts
}`,
			expectedError: "line 3: missing opening curly brace for mapscripts 'MyMapScripts'",
		},
		{
			input: `
mapscripts MyMapScripts {
	+
}`,
			expectedError: "line 3: expected map script type, but got '+' instead",
		},
		{
			input: `
mapscripts MyMapScripts {
	SOME_TYPE: 5
}`,
			expectedError: "line 3: expected map script label after ':', but got '5' instead",
		},
		{
			input: `
mapscripts MyMapScripts {
	SOME_TYPE {
		if (sdf)
	}
}`,
			expectedError: "line 4: left side of binary expression must be var(), flag(), or defeated() operator. Instead, found 'sdf'",
		},
		{
			input: `
mapscripts MyMapScripts {
	SOME_TYPE [
		VAR_TEMP
	]
}`,
			expectedError: "line 4: missing ',' to specify map script table entry comparison value",
		},
		{
			input: `
mapscripts MyMapScripts {
	SOME_TYPE [
		VAR_TEMP, : Foo
	]
}`,
			expectedError: "line 4: expected comparison value for map script table entry, but it was empty",
		},
		{
			input: `
mapscripts MyMapScripts {
	SOME_TYPE [
		VAR_TEMP, FOO
	]
}`,
			expectedError: "line 4: missing ':' or '{' to specify map script table entry",
		},
		{
			input: `
mapscripts MyMapScripts {
	SOME_TYPE [
		, FOO: Foo Script
	]
}`,
			expectedError: "line 4: expected condition for map script table entry, but it was empty",
		},
		{
			input: `
mapscripts MyMapScripts {
	SOME_TYPE [
		VAR_TEMP, 1: 5
	]
}`,
			expectedError: "line 4: expected map script label after ':', but got '5' instead",
		},
		{
			input: `
mapscripts MyMapScripts {
	SOME_TYPE [
		VAR_TEMP, 1 {
			msgbox(
		}
	]
}`,
			expectedError: "line 5: missing closing parenthesis for command 'msgbox'",
		},
		{
			input: `
script(asdf) MyScript {}`,
			expectedError: "line 2: scope modifier must be 'global' or 'local', but got 'asdf' instead",
		},
		{
			input: `
script(local MyScript {}`,
			expectedError: "line 2: missing ')' after scope modifier. Got 'MyScript' instead",
		},
		{
			input: `
text(local MyText {"test"}`,
			expectedError: "line 2: missing ')' after scope modifier. Got 'MyText' instead",
		},
		{
			input: `
movement() MyMovement {walk_left}`,
			expectedError: "line 2: scope modifier must be 'global' or 'local', but got ')' instead",
		},
		{
			input: `
mapscripts() MyMapScripts {}`,
			expectedError: "line 2: scope modifier must be 'global' or 'local', but got ')' instead",
		},
		{
			input:         `const 45`,
			expectedError: "line 1: expected identifier after const, but got '45' instead",
		},
		{
			input:         `const FOO`,
			expectedError: "line 1: missing equals sign after const name 'FOO'",
		},
		{
			input:         `const FOO = 4 const FOO = 5`,
			expectedError: "line 1: duplicate const 'FOO'. Must use unique const names",
		},
		{
			input:         `const FOO = `,
			expectedError: "line 1: missing value for const 'FOO'",
		},
		{
			input: `const FOO = 
			script MyScript {}`,
			expectedError: "line 1: missing value for const 'FOO'",
		},
	}

	for _, test := range tests {
		testForParseError(t, test.input, test.expectedError, test.expectedErrorRegex)
	}
}

func testForParseError(t *testing.T, input string, expectedError, expectedErrorRegex string) {
	l := lexer.New(input)
	p := New(l, "../font_widths.json", "", 208, nil)
	_, err := p.ParseProgram()
	if err == nil {
		t.Fatalf("Expected error '%s', but no error occurred", expectedError)
	}
	if expectedError != "" {
		if err.Error() != expectedError {
			t.Fatalf("Expected error '%s', but got '%s'", expectedError, err.Error())
		}
	}
	if expectedErrorRegex != "" {
		isMatch, _ := regexp.MatchString(expectedErrorRegex, err.Error())
		if !isMatch {
			t.Fatalf("Expected error to match regex '%s', but got '%s'", expectedErrorRegex, err.Error())
		}
	}
}
