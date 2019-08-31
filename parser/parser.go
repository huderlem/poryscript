package parser

import (
	"fmt"
	"strings"

	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/token"
)

// Parser is a Poryscript AST parser.
type Parser struct {
	l                *lexer.Lexer
	curToken         token.Token
	peekToken        token.Token
	errors           []string
	inlineTexts      []ast.Text
	inlineTextCounts map[string]int
	loopStack        []ast.Statement
}

// New creates a new Poryscript AST Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:                l,
		errors:           []string{},
		inlineTexts:      []ast.Text{},
		inlineTextCounts: make(map[string]int),
	}
	// Read two tokens, so curToken and peekToken are both set.
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) pushLoopStack(statement ast.Statement) {
	p.loopStack = append(p.loopStack, statement)
}

func (p *Parser) popLoopStack() {
	p.loopStack = p.loopStack[:len(p.loopStack)-1]
}

func (p *Parser) peekLoopStack() ast.Statement {
	if len(p.loopStack) == 0 {
		return nil
	}
	return p.loopStack[len(p.loopStack)-1]
}

// Errors returns the list of parser error messages.
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) peekTokenIs(expectedType token.Type) bool {
	return p.peekToken.Type == expectedType
}

func (p *Parser) expectPeek(expectedType token.Type) bool {
	if p.peekTokenIs(expectedType) {
		p.nextToken()
		return true
	}

	p.peekError(expectedType)
	return false
}

func (p *Parser) peekError(expectedType token.Type) {
	msg := fmt.Sprintf("expected next token to be type %s, got %s instead", expectedType, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func getImplicitTextLabel(scriptName string, i int) string {
	return fmt.Sprintf("%s_Text_%d", scriptName, i)
}

// ParseProgram parses a Poryscript file into an AST.
func (p *Parser) ParseProgram() *ast.Program {
	p.inlineTexts = nil
	program := &ast.Program{
		TopLevelStatements: []ast.Statement{},
		Texts:              []ast.Text{},
	}

	for p.curToken.Type != token.EOF {
		statement := p.parseTopLevelStatement()
		if len(p.errors) > 0 {
			for _, err := range p.errors {
				fmt.Printf("ERROR: %s\n", err)
			}
			return nil
		}
		if statement != nil {
			program.TopLevelStatements = append(program.TopLevelStatements, statement)
		}
		p.nextToken()
	}

	for _, text := range p.inlineTexts {
		program.Texts = append(program.Texts, text)
	}

	return program
}

func (p *Parser) parseTopLevelStatement() ast.Statement {
	switch p.curToken.Type {
	case token.SCRIPT:
		statement := p.parseScriptStatement()
		if statement == nil {
			return nil
		}
		return statement
	case token.RAW:
		statement := p.parseRawStatement()
		if statement == nil {
			return nil
		}
		return statement
	}

	msg := fmt.Sprintf("line %d: could not parse top-level statement for '%s'", p.curToken.LineNumber, p.curToken.Literal)
	p.errors = append(p.errors, msg)
	return nil
}

func (p *Parser) parseScriptStatement() *ast.ScriptStatement {
	statement := &ast.ScriptStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	p.nextToken()

	statement.Body = p.parseBlockStatement(statement.Name.Value)
	return statement
}

func (p *Parser) parseBlockStatement(scriptName string) *ast.BlockStatement {
	block := &ast.BlockStatement{
		Token:      p.curToken,
		Statements: []ast.Statement{},
	}

	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.EOF {
			msg := fmt.Sprintf("line %d: missing closing curly brace for block statement", block.Token.LineNumber)
			p.errors = append(p.errors, msg)
			return nil
		}

		statement := p.parseStatement(scriptName)
		if statement == nil {
			return nil
		}

		block.Statements = append(block.Statements, statement)
		p.nextToken()
	}

	return block
}

func (p *Parser) parseStatement(scriptName string) ast.Statement {
	switch p.curToken.Type {
	case token.IDENT:
		statement := p.parseCommandStatement(scriptName)
		if statement == nil {
			return nil
		}
		return statement
	case token.IF:
		statement := p.parseIfStatement(scriptName)
		if statement == nil {
			return nil
		}
		return statement
	case token.WHILE:
		statement := p.parseWhileStatement(scriptName)
		if statement == nil {
			return nil
		}
		return statement
	case token.DO:
		statement := p.parseDoWhileStatement(scriptName)
		if statement == nil {
			return nil
		}
		return statement
	case token.BREAK:
		statement := p.parseBreakStatement(scriptName)
		if statement == nil {
			return nil
		}
		return statement
	}

	msg := fmt.Sprintf("line %d: could not parse statement for '%s'\n", p.curToken.LineNumber, p.curToken.Literal)
	p.errors = append(p.errors, msg)
	return nil
}

func (p *Parser) parseCommandStatement(scriptName string) ast.Statement {
	command := &ast.CommandStatement{
		Token: p.curToken,
		Name: &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
		Args: []string{},
	}

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		p.nextToken()
		argParts := []string{}
		numOpenParens := 0
		for !(p.curToken.Type == token.RPAREN && numOpenParens == 0) {
			if p.curToken.Type == token.EOF {
				msg := fmt.Sprintf("line %d: missing closing parenthesis for command '%s'", command.Token.LineNumber, command.Name.TokenLiteral())
				p.errors = append(p.errors, msg)
				return nil
			}

			if p.curToken.Type == token.COMMA {
				arg := strings.Join(argParts, " ")
				command.Args = append(command.Args, arg)
				argParts = []string{}
			} else if p.curToken.Type == token.LPAREN {
				numOpenParens++
				argParts = append(argParts, p.curToken.Literal)
			} else if p.curToken.Type == token.RPAREN {
				numOpenParens--
				argParts = append(argParts, p.curToken.Literal)
			} else if p.curToken.Type == token.STRING {
				textLabel := getImplicitTextLabel(scriptName, p.inlineTextCounts[scriptName])
				p.inlineTextCounts[scriptName]++
				p.inlineTexts = append(p.inlineTexts, ast.Text{Name: textLabel, Value: p.curToken.Literal})
				argParts = append(argParts, textLabel)
			} else {
				argParts = append(argParts, p.curToken.Literal)
			}

			p.nextToken()
		}

		if len(argParts) > 0 {
			arg := strings.Join(argParts, " ")
			command.Args = append(command.Args, arg)
		}
	}

	return command
}

func (p *Parser) parseRawStatement() *ast.RawStatement {
	statement := &ast.RawStatement{
		Token: p.curToken,
	}

	if !p.expectPeek(token.RAWSTRING) {
		return nil
	}

	statement.Value = p.curToken.Literal
	return statement
}

func (p *Parser) parseIfStatement(scriptName string) *ast.IfStatement {
	statement := &ast.IfStatement{
		Token: p.curToken,
	}

	// First if statement condition
	consequence := p.parseConditionExpression(scriptName, statement.Token.LineNumber)
	if consequence == nil {
		return nil
	}
	statement.Consequence = consequence

	// Possibly-many elif conditions
	for p.peekToken.Type == token.ELSEIF {
		p.nextToken()
		consequence = p.parseConditionExpression(scriptName, p.peekToken.LineNumber)
		if consequence == nil {
			return nil
		}
		statement.ElifConsequences = append(statement.ElifConsequences, consequence)
	}

	// Trailing else block
	if p.peekToken.Type == token.ELSE {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			msg := fmt.Sprintf("line %d: missing opening curly brace of else statement '%s'", p.peekToken.LineNumber, p.peekToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}
		p.nextToken()
		statement.ElseConsequence = p.parseBlockStatement(scriptName)
	}

	return statement
}

func (p *Parser) parseWhileStatement(scriptName string) *ast.WhileStatement {
	statement := &ast.WhileStatement{
		Token: p.curToken,
	}
	p.pushLoopStack(statement)

	// while statement condition
	consequence := p.parseConditionExpression(scriptName, statement.Token.LineNumber)
	p.popLoopStack()
	if consequence == nil {
		return nil
	}
	statement.Consequence = consequence

	return statement
}

func (p *Parser) parseDoWhileStatement(scriptName string) *ast.DoWhileStatement {
	statement := &ast.DoWhileStatement{
		Token: p.curToken,
	}
	p.pushLoopStack(statement)
	expression := &ast.ConditionExpression{}

	if !p.expectPeek(token.LBRACE) {
		msg := fmt.Sprintf("line %d: missing opening curly brace of do-while statement '%s'", p.peekToken.LineNumber, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()
	expression.Body = p.parseBlockStatement(scriptName)
	if expression.Body == nil {
		return nil
	}
	p.popLoopStack()

	if !p.expectPeek(token.WHILE) {
		msg := fmt.Sprintf("line %d: missing 'while' after body of do-while statement '%s'", p.peekToken.LineNumber, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	if !p.parseCondition(expression, statement.Token.LineNumber) {
		return nil
	}

	statement.Consequence = expression

	return statement
}

func (p *Parser) parseBreakStatement(scriptName string) *ast.BreakStatement {
	statement := &ast.BreakStatement{
		Token: p.curToken,
	}

	if p.peekLoopStack() == nil {
		msg := fmt.Sprintf("line %d: 'break' statement outside of loop scope.", p.peekToken.LineNumber)
		p.errors = append(p.errors, msg)
		return nil
	}
	statement.LoopStatment = p.peekLoopStack()

	if p.peekToken.Type != token.RBRACE {
		msg := fmt.Sprintf("line %d: missing '{' after 'break'. 'break' must be the last statement in block scope.", p.peekToken.LineNumber)
		p.errors = append(p.errors, msg)
		return nil
	}

	return statement
}

func (p *Parser) parseConditionExpression(scriptName string, lineNumber int) *ast.ConditionExpression {
	expression := &ast.ConditionExpression{}
	if !p.parseCondition(expression, lineNumber) {
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()

	expression.Body = p.parseBlockStatement(scriptName)
	return expression
}

func (p *Parser) parseCondition(expression *ast.ConditionExpression, lineNumber int) bool {
	if !p.expectPeek(token.LPAREN) {
		msg := fmt.Sprintf("line %d: missing opening parenthesis of condition '%s'", lineNumber, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return false
	}

	if !p.peekTokenIs(token.VAR) && !p.peekTokenIs(token.FLAG) {
		msg := fmt.Sprintf("line %d: invalid condition statement command '%s'", lineNumber, p.peekToken.Literal)
		p.errors = append(p.errors, msg)
		return false
	}

	p.nextToken()
	expression.Type = p.curToken.Type
	if !p.expectPeek(token.LPAREN) {
		msg := fmt.Sprintf("line %d: missing opening parenthesis for condition operator '%s'", lineNumber, expression.Type)
		p.errors = append(p.errors, msg)
		return false
	}
	if p.peekToken.Type == token.RPAREN {
		msg := fmt.Sprintf("line %d: missing value for condition operator '%s'", lineNumber, expression.Type)
		p.errors = append(p.errors, msg)
		return false
	}
	p.nextToken()

	parts := []string{}
	for p.curToken.Type != token.RPAREN {
		parts = append(parts, p.curToken.Literal)
		p.nextToken()
	}
	expression.Operand = strings.Join(parts, " ")
	p.nextToken()

	if expression.Type == token.VAR {
		ok := p.parseConditionVarOperator(expression)
		if !ok {
			return false
		}
	} else if expression.Type == token.FLAG {
		ok := p.parseConditionFlagOperator(expression)
		if !ok {
			return false
		}
	}
	return true
}

func (p *Parser) parseConditionVarOperator(expression *ast.ConditionExpression) bool {
	if p.curToken.Type != token.GT && p.curToken.Type != token.GTE && p.curToken.Type != token.LT &&
		p.curToken.Type != token.LTE && p.curToken.Type != token.EQ && p.curToken.Type != token.NEQ {
		msg := fmt.Sprintf("line %d: invalid condition operator '%s'", p.curToken.LineNumber, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return false
	}
	expression.Operator = p.curToken.Type
	p.nextToken()

	if p.curToken.Type == token.RPAREN {
		msg := fmt.Sprintf("line %d: missing comparison value for if statement", p.curToken.LineNumber)
		p.errors = append(p.errors, msg)
		return false
	}
	parts := []string{}
	for p.curToken.Type != token.RPAREN {
		parts = append(parts, p.curToken.Literal)
		p.nextToken()
	}

	expression.ComparisonValue = strings.Join(parts, " ")
	return true
}

func (p *Parser) parseConditionFlagOperator(expression *ast.ConditionExpression) bool {
	if p.curToken.Type != token.EQ {
		msg := fmt.Sprintf("line %d: invalid condition operator '%s'. Only '==' is allowed.", p.curToken.LineNumber, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return false
	}
	expression.Operator = p.curToken.Type
	p.nextToken()

	if p.curToken.Type == token.RPAREN {
		msg := fmt.Sprintf("line %d: missing comparison value for if statement", p.curToken.LineNumber)
		p.errors = append(p.errors, msg)
		return false
	}

	if p.curToken.Type != token.TRUE && p.curToken.Type != token.FALSE {
		msg := fmt.Sprintf("line %d: invalid flag comparison value '%s'. Only 'TRUE' and 'FALSE' are allowed.", p.curToken.LineNumber, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return false
	}

	expression.ComparisonValue = string(p.curToken.Type)
	if !p.expectPeek(token.RPAREN) {
		return false
	}

	return true
}
