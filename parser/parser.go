package parser

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/token"
)

var topLevelTokens = map[token.Type]bool{
	token.SCRIPT:     true,
	token.RAW:        true,
	token.TEXT:       true,
	token.MOVEMENT:   true,
	token.MAPSCRIPTS: true,
	token.CONST:      true,
}

type impText struct {
	command    *ast.CommandStatement
	argPos     int
	text       string
	scriptName string
}

// Parser is a Poryscript AST parser.
type Parser struct {
	l                  *lexer.Lexer
	curToken           token.Token
	peekToken          token.Token
	peek2Token         token.Token
	inlineTexts        []ast.Text
	inlineTextsSet     map[string]string
	inlineTextCounts   map[string]int
	textStatements     []*ast.TextStatement
	breakStack         []ast.Statement
	continueStack      []ast.Statement
	fontConfigFilepath string
	fonts              *FontWidthsConfig
	compileSwitches    map[string]string
	constants          map[string]string
}

// New creates a new Poryscript AST Parser.
func New(l *lexer.Lexer, fontConfigFilepath string, compileSwitches map[string]string) *Parser {
	p := &Parser{
		l:                  l,
		inlineTexts:        make([]ast.Text, 0),
		inlineTextsSet:     make(map[string]string),
		inlineTextCounts:   make(map[string]int),
		textStatements:     make([]*ast.TextStatement, 0),
		fontConfigFilepath: fontConfigFilepath,
		compileSwitches:    compileSwitches,
		constants:          make(map[string]string),
	}
	// Read three tokens, so curToken, peekToken, and peek2Token are all set.
	p.nextToken()
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) pushBreakStack(statement ast.Statement) {
	p.breakStack = append(p.breakStack, statement)
}

func (p *Parser) popBreakStack() {
	p.breakStack = p.breakStack[:len(p.breakStack)-1]
}

func (p *Parser) peekBreakStack() ast.Statement {
	if len(p.breakStack) == 0 {
		return nil
	}
	return p.breakStack[len(p.breakStack)-1]
}

func (p *Parser) pushContinueStack(statement ast.Statement) {
	p.continueStack = append(p.continueStack, statement)
}

func (p *Parser) popContinueStack() {
	p.continueStack = p.continueStack[:len(p.continueStack)-1]
}

func (p *Parser) peekContinueStack() ast.Statement {
	if len(p.continueStack) == 0 {
		return nil
	}
	return p.continueStack[len(p.continueStack)-1]
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.peek2Token
	p.peek2Token = p.l.NextToken()
}

func (p *Parser) peekTokenIs(expectedType token.Type) bool {
	return p.peekToken.Type == expectedType
}

func (p *Parser) peek2TokenIs(expectedType token.Type) bool {
	return p.peek2Token.Type == expectedType
}

func (p *Parser) expectPeek(expectedType token.Type) error {
	if p.peekTokenIs(expectedType) {
		p.nextToken()
		return nil
	}

	return fmt.Errorf("line %d: expected next token to be '%s', got '%s' instead", p.peekToken.LineNumber, expectedType, p.peekToken.Literal)
}

func getImplicitTextLabel(scriptName string, i int) string {
	return fmt.Sprintf("%s_Text_%d", scriptName, i)
}

// ParseProgram parses a Poryscript file into an AST.
func (p *Parser) ParseProgram() (*ast.Program, error) {
	p.inlineTexts = make([]ast.Text, 0)
	p.inlineTextsSet = make(map[string]string)
	p.textStatements = make([]*ast.TextStatement, 0)
	program := &ast.Program{
		TopLevelStatements: []ast.Statement{},
		Texts:              []ast.Text{},
	}

	for p.curToken.Type != token.EOF {
		statement, err := p.parseTopLevelStatement()
		if err != nil {
			return nil, err
		}
		if statement != nil {
			program.TopLevelStatements = append(program.TopLevelStatements, statement)
		}
		p.nextToken()
	}

	// Build list of Texts from both inline and explicit texts.
	// Generate error if there are any name clashes.
	for _, text := range p.inlineTexts {
		program.Texts = append(program.Texts, text)
	}
	for _, textStmt := range p.textStatements {
		program.Texts = append(program.Texts, ast.Text{
			Value:    textStmt.Value,
			Name:     textStmt.Name.Value,
			IsGlobal: textStmt.Scope == token.GLOBAL,
		})
	}
	names := make(map[string]struct{}, 0)
	for _, text := range program.Texts {
		if _, ok := names[text.Name]; ok {
			return nil, fmt.Errorf("Duplicate text label '%s'. Choose a unique label that won't clash with the auto-generated text labels", text.Name)
		}
		names[text.Name] = struct{}{}
	}

	return program, nil
}

func (p *Parser) parseTopLevelStatement() (ast.Statement, error) {
	switch p.curToken.Type {
	case token.SCRIPT:
		statement, implicitTexts, err := p.parseScriptStatement()
		if err != nil {
			return nil, err
		}
		p.addImplicitTexts(implicitTexts)
		return statement, nil
	case token.RAW:
		statement, err := p.parseRawStatement()
		if err != nil {
			return nil, err
		}
		return statement, nil
	case token.TEXT:
		statement, err := p.parseTextStatement()
		if err != nil {
			return nil, err
		}
		return statement, nil
	case token.MOVEMENT:
		statement, err := p.parseMovementStatement()
		if err != nil {
			return nil, err
		}
		return statement, nil
	case token.MAPSCRIPTS:
		statement, implicitTexts, err := p.parseMapscriptsStatement()
		if err != nil {
			return nil, err
		}
		p.addImplicitTexts(implicitTexts)
		return statement, nil
	case token.CONST:
		err := p.parseConstant()
		return nil, err
	}

	return nil, fmt.Errorf("line %d: could not parse top-level statement for '%s'", p.curToken.LineNumber, p.curToken.Literal)
}

func (p *Parser) addImplicitTexts(implicitTexts []impText) {
	for _, t := range implicitTexts {
		if textLabel, ok := p.inlineTextsSet[t.text]; ok {
			t.command.Args[t.argPos] = textLabel
		} else {
			textLabel := getImplicitTextLabel(t.scriptName, p.inlineTextCounts[t.scriptName])
			t.command.Args[t.argPos] = textLabel
			p.inlineTextCounts[t.scriptName]++
			p.inlineTextsSet[t.text] = textLabel
			p.inlineTexts = append(p.inlineTexts, ast.Text{
				Name:     textLabel,
				Value:    t.text,
				IsGlobal: false,
			})
		}
	}
}

func (p *Parser) parseScopeModifier(defaultScope token.Type) (token.Type, error) {
	var scope = defaultScope
	if !p.peekTokenIs(token.LPAREN) {
		return scope, nil
	}
	p.nextToken()
	if !p.peekTokenIs(token.GLOBAL) && !p.peekTokenIs(token.LOCAL) {
		return scope, fmt.Errorf("line %d: scope modifier must be 'global' or 'local', but got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	p.nextToken()
	if !p.peekTokenIs(token.RPAREN) {
		return scope, fmt.Errorf("line %d: missing ')' after scope modifier. Got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	scope = p.curToken.Type
	p.nextToken()
	return scope, nil
}

func (p *Parser) parseScriptStatement() (*ast.ScriptStatement, []impText, error) {
	statement := &ast.ScriptStatement{Token: p.curToken}
	scope, err := p.parseScopeModifier(token.GLOBAL)
	if err != nil {
		return nil, nil, err
	}
	statement.Scope = scope
	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing name for script", p.curToken.LineNumber)
	}

	statement.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing opening curly brace for script '%s'", p.curToken.LineNumber, statement.Name.Value)
	}

	p.nextToken()

	blockStmt, implicitTexts, err := p.parseBlockStatement(statement.Name.Value)
	if err != nil {
		return nil, nil, err
	}
	statement.Body = blockStmt
	return statement, implicitTexts, nil
}

func (p *Parser) parseBlockStatement(scriptName string) (*ast.BlockStatement, []impText, error) {
	block := &ast.BlockStatement{
		Token:      p.curToken,
		Statements: []ast.Statement{},
	}
	implicitTexts := make([]impText, 0)

	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.EOF {
			return nil, nil, fmt.Errorf("line %d: missing closing curly brace for block statement", block.Token.LineNumber)
		}

		statements, stmtTexts, err := p.parseStatement(scriptName)
		if err != nil {
			return nil, nil, err
		}
		implicitTexts = append(implicitTexts, stmtTexts...)

		block.Statements = append(block.Statements, statements...)
		p.nextToken()
	}

	return block, implicitTexts, nil
}

func (p *Parser) parseSwitchBlockStatement(scriptName string) (*ast.BlockStatement, []impText, error) {
	block := &ast.BlockStatement{
		Token:      p.curToken,
		Statements: []ast.Statement{},
	}
	implicitTexts := make([]impText, 0)

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.CASE && p.curToken.Type != token.DEFAULT {
		if p.curToken.Type == token.EOF {
			return nil, nil, fmt.Errorf("line %d: missing end for switch case body", block.Token.LineNumber)
		}

		statements, stmtTexts, err := p.parseStatement(scriptName)
		if err != nil {
			return nil, nil, err
		}

		implicitTexts = append(implicitTexts, stmtTexts...)
		block.Statements = append(block.Statements, statements...)
		p.nextToken()
	}

	return block, implicitTexts, nil
}

func (p *Parser) parseStatement(scriptName string) ([]ast.Statement, []impText, error) {
	statements := make([]ast.Statement, 0, 1)
	var implicitTexts []impText
	var err error
	var statement ast.Statement
	switch p.curToken.Type {
	case token.IDENT:
		statement, implicitTexts, err = p.parseCommandStatement(scriptName)
		statements = append(statements, statement)
	case token.IF:
		statement, implicitTexts, err = p.parseIfStatement(scriptName)
		statements = append(statements, statement)
	case token.WHILE:
		statement, implicitTexts, err = p.parseWhileStatement(scriptName)
		statements = append(statements, statement)
	case token.DO:
		statement, implicitTexts, err = p.parseDoWhileStatement(scriptName)
		statements = append(statements, statement)
	case token.BREAK:
		statement, err = p.parseBreakStatement(scriptName)
		statements = append(statements, statement)
	case token.CONTINUE:
		statement, err = p.parseContinueStatement(scriptName)
		statements = append(statements, statement)
	case token.SWITCH:
		statement, implicitTexts, err = p.parseSwitchStatement(scriptName)
		statements = append(statements, statement)
	case token.PORYSWITCH:
		var stmts []ast.Statement
		stmts, implicitTexts, err = p.parsePoryswitchStatement(scriptName)
		statements = append(statements, stmts...)
	default:
		err = fmt.Errorf("line %d: could not parse statement for '%s'", p.curToken.LineNumber, p.curToken.Literal)
	}

	if err != nil {
		return nil, nil, err
	}

	return statements, implicitTexts, nil
}

func (p *Parser) parseCommandStatement(scriptName string) (ast.Statement, []impText, error) {
	command := &ast.CommandStatement{
		Token: p.curToken,
		Name: &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
		Args: []string{},
	}

	implicitTexts := make([]impText, 0)

	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		p.nextToken()
		argParts := []string{}
		numOpenParens := 0
		for !(p.curToken.Type == token.RPAREN && numOpenParens == 0) {
			if p.curToken.Type == token.EOF {
				err := fmt.Errorf("line %d: missing closing parenthesis for command '%s'", command.Token.LineNumber, command.Name.TokenLiteral())
				return nil, nil, err
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
			} else if p.curToken.Type == token.FORMAT {
				strValue, err := p.parseFormatStringOperator()
				if err != nil {
					return nil, nil, err
				}
				implicitTexts = append(implicitTexts, impText{
					command:    command,
					argPos:     len(command.Args),
					text:       p.formatTextTerminator(strValue),
					scriptName: scriptName,
				})
				argParts = append(argParts, "")
			} else if p.curToken.Type == token.STRING {
				implicitTexts = append(implicitTexts, impText{
					command:    command,
					argPos:     len(command.Args),
					text:       p.formatTextTerminator(p.curToken.Literal),
					scriptName: scriptName,
				})
				argParts = append(argParts, "")
			} else {
				argParts = append(argParts, p.tryReplaceWithConstant(p.curToken.Literal))
			}

			p.nextToken()
		}

		if len(argParts) > 0 {
			arg := strings.Join(argParts, " ")
			command.Args = append(command.Args, arg)
		}
	}

	return command, implicitTexts, nil
}

func (p *Parser) parseRawStatement() (*ast.RawStatement, error) {
	statement := &ast.RawStatement{
		Token: p.curToken,
	}

	if err := p.expectPeek(token.RAWSTRING); err != nil {
		return nil, fmt.Errorf("line %d: raw statement must begin with a backtick character '`'", p.curToken.LineNumber)
	}

	statement.Value = p.curToken.Literal
	return statement, nil
}

func (p *Parser) parseTextStatement() (*ast.TextStatement, error) {
	statement := &ast.TextStatement{
		Token: p.curToken,
	}
	scope, err := p.parseScopeModifier(token.GLOBAL)
	if err != nil {
		return nil, err
	}
	statement.Scope = scope
	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, fmt.Errorf("line %d: missing name for text statement", p.curToken.LineNumber)
	}

	statement.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, fmt.Errorf("line %d: missing opening curly brace for text '%s'", p.peekToken.LineNumber, statement.Name.Value)
	}
	p.nextToken()

	var strValue string
	if p.curToken.Type == token.PORYSWITCH {
		strValue, err = p.parsePoryswitchTextStatement()
		if err != nil {
			return nil, err
		}
	} else {
		strValue, err = p.parseTextValue()
		if err != nil {
			return nil, err
		}
	}

	statement.Value = strValue
	p.textStatements = append(p.textStatements, statement)
	if err := p.expectPeek(token.RBRACE); err != nil {
		return nil, fmt.Errorf("line %d: expected closing curly brace for text. Got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	return statement, nil
}

func (p *Parser) parseTextValue() (string, error) {
	if p.curToken.Type == token.FORMAT {
		var err error
		strValue, err := p.parseFormatStringOperator()
		if err != nil {
			return "", err
		}
		return p.formatTextTerminator(strValue), nil
	} else if p.curToken.Type == token.STRING {
		return p.formatTextTerminator(p.curToken.Literal), nil
	} else {
		return "", fmt.Errorf("line %d: body of text statement must be a string or formatted string. Got '%s' instead", p.curToken.LineNumber, p.curToken.Literal)
	}
}

func (p *Parser) parsePoryswitchHeader() (string, string, error) {
	if len(p.compileSwitches) == 0 {
		return "", "", fmt.Errorf("line %d: poryswitch used, but no compile switches were specified with the '-s' option", p.curToken.LineNumber)
	}
	if err := p.expectPeek(token.LPAREN); err != nil {
		return "", "", fmt.Errorf("line %d: expected opening parenthesis for poryswitch value. Got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	if err := p.expectPeek(token.IDENT); err != nil {
		return "", "", fmt.Errorf("line %d: expected poryswitch identifier value. Got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	switchCase := p.curToken.Literal
	var switchValue string
	var ok bool
	if switchValue, ok = p.compileSwitches[switchCase]; !ok {
		return "", "", fmt.Errorf("line %d: no poryswitch for '%s' was specified with the '-s' option", p.curToken.LineNumber, switchCase)
	}

	if err := p.expectPeek(token.RPAREN); err != nil {
		return "", "", fmt.Errorf("line %d: expected closing parenthesis for poryswitch value. Got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	if err := p.expectPeek(token.LBRACE); err != nil {
		return "", "", fmt.Errorf("line %d: expected opening curly brace for poryswitch statement. Got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	p.nextToken()
	return switchCase, switchValue, nil
}

func (p *Parser) parsePoryswitchTextCases() (map[string]string, error) {
	textCases := make(map[string]string)
	startLineNumber := p.curToken.LineNumber
	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.EOF {
			return nil, fmt.Errorf("line %d: missing closing curly brace for poryswitch statement", startLineNumber)
		}
		if p.curToken.Type != token.IDENT && p.curToken.Type != token.INT {
			return nil, fmt.Errorf("line %d: invalid poryswitch case '%s'. Expected a simple identifier", p.curToken.LineNumber, p.curToken.Literal)
		}
		caseValue := p.curToken.Literal
		p.nextToken()
		if p.curToken.Type == token.COLON || p.curToken.Type == token.LBRACE {
			usedBrace := p.curToken.Type == token.LBRACE
			p.nextToken()
			strValue, err := p.parseTextValue()
			if err != nil {
				return nil, err
			}
			textCases[caseValue] = strValue
			p.nextToken()
			if usedBrace {
				if p.curToken.Type != token.RBRACE {
					return nil, fmt.Errorf("line %d: missing closing curly brace for poryswitch case '%s'", startLineNumber, caseValue)
				}
				p.nextToken()
			}
		} else {
			return nil, fmt.Errorf("line %d: invalid token '%s' after poryswitch case '%s'. Expected ':' or '{'", p.curToken.LineNumber, p.curToken.Literal, caseValue)
		}
	}
	return textCases, nil
}

func (p *Parser) parsePoryswitchTextStatement() (string, error) {
	startLineNumber := p.curToken.LineNumber
	switchCase, switchValue, err := p.parsePoryswitchHeader()
	if err != nil {
		return "", err
	}
	cases, err := p.parsePoryswitchTextCases()
	if err != nil {
		return "", err
	}
	strValue, ok := cases[switchValue]
	if !ok {
		strValue, ok = cases["_"]
		if !ok {
			return "", fmt.Errorf("line %d: no poryswitch case found for '%s=%s', which was specified with the '-s' option", startLineNumber, switchCase, switchValue)
		}
	}
	return strValue, nil
}

func (p *Parser) parseMovementStatement() (*ast.MovementStatement, error) {
	statement := &ast.MovementStatement{
		Token:            p.curToken,
		MovementCommands: []string{},
	}
	scope, err := p.parseScopeModifier(token.LOCAL)
	if err != nil {
		return nil, err
	}
	statement.Scope = scope
	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, fmt.Errorf("line %d: missing name for movement statement", p.curToken.LineNumber)
	}

	statement.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, fmt.Errorf("line %d: missing opening curly brace for movement '%s'", p.peekToken.LineNumber, statement.Name.Value)
	}
	p.nextToken()
	statement.MovementCommands, err = p.parseMovementValue(true)
	if err != nil {
		return nil, err
	}

	return statement, nil
}

func (p *Parser) parseMovementValue(allowMultiple bool) ([]string, error) {
	movementCommands := make([]string, 0)
	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.PORYSWITCH {
			poryswitchCommands, err := p.parsePoryswitchMovementStatement()
			if err != nil {
				return nil, err
			}
			movementCommands = append(movementCommands, poryswitchCommands...)
		} else if p.curToken.Type == token.IDENT {
			moveCommand := p.curToken.Literal
			p.nextToken()
			if p.curToken.Type == token.MUL {
				p.nextToken()
				if p.curToken.Type != token.INT {
					return nil, fmt.Errorf("line %d: expected mulplier number for movement command, but got '%s' instead", p.curToken.LineNumber, p.curToken.Literal)
				}
				num, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
				if err != nil {
					return nil, fmt.Errorf("line %d: invalid movement mulplier integer '%s': %s", p.curToken.LineNumber, p.curToken.Literal, err.Error())
				}
				if num <= 0 {
					return nil, fmt.Errorf("line %d: movement mulplier must be a positive integer, but got '%s' instead", p.curToken.LineNumber, p.curToken.Literal)
				}
				if num > 9999 {
					return nil, fmt.Errorf("line %d: movement mulplier '%s' is too large. Maximum is 9999", p.curToken.LineNumber, p.curToken.Literal)
				}
				var i int64
				for i = 0; i < num; i++ {
					movementCommands = append(movementCommands, moveCommand)
				}

				p.nextToken()
			} else {
				movementCommands = append(movementCommands, moveCommand)
			}
		} else {
			return nil, fmt.Errorf("line %d: expected movement command, but got '%s' instead", p.curToken.LineNumber, p.curToken.Literal)
		}
		if !allowMultiple {
			break
		}
	}
	return movementCommands, nil
}

func (p *Parser) parsePoryswitchMovementStatement() ([]string, error) {
	startLineNumber := p.curToken.LineNumber
	switchCase, switchValue, err := p.parsePoryswitchHeader()
	if err != nil {
		return nil, err
	}
	cases, err := p.parsePoryswitchMovementCases()
	if err != nil {
		return nil, err
	}
	movements, ok := cases[switchValue]
	if !ok {
		movements, ok = cases["_"]
		if !ok {
			return nil, fmt.Errorf("line %d: no poryswitch case found for '%s=%s', which was specified with the '-s' option", startLineNumber, switchCase, switchValue)
		}
	}
	p.nextToken()
	return movements, nil
}

func (p *Parser) parsePoryswitchMovementCases() (map[string][]string, error) {
	movementCases := make(map[string][]string)
	startLineNumber := p.curToken.LineNumber
	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.EOF {
			return nil, fmt.Errorf("line %d: missing closing curly braces for poryswitch statement", startLineNumber)
		}
		if p.curToken.Type != token.IDENT && p.curToken.Type != token.INT {
			return nil, fmt.Errorf("line %d: invalid poryswitch case '%s'. Expected a simple identifier", p.curToken.LineNumber, p.curToken.Literal)
		}
		caseValue := p.curToken.Literal
		p.nextToken()
		if p.curToken.Type == token.COLON || p.curToken.Type == token.LBRACE {
			usedBrace := p.curToken.Type == token.LBRACE
			p.nextToken()
			movements, err := p.parseMovementValue(usedBrace)
			if err != nil {
				return nil, err
			}
			movementCases[caseValue] = movements
			if usedBrace {
				if p.curToken.Type != token.RBRACE {
					return nil, fmt.Errorf("line %d: missing closing curly brace for poryswitch case '%s'", startLineNumber, caseValue)
				}
				p.nextToken()
			}
		} else {
			return nil, fmt.Errorf("line %d: invalid token '%s' after poryswitch case '%s'. Expected ':' or '{'", p.curToken.LineNumber, p.curToken.Literal, caseValue)
		}
	}
	return movementCases, nil
}

func (p *Parser) parseMapscriptsStatement() (*ast.MapScriptsStatement, []impText, error) {
	scope, err := p.parseScopeModifier(token.GLOBAL)
	if err != nil {
		return nil, nil, err
	}
	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing name for mapscripts statement", p.curToken.LineNumber)
	}

	statement := &ast.MapScriptsStatement{
		Token: p.curToken,
		Name: &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		},
		MapScripts:      []ast.MapScript{},
		TableMapScripts: []ast.TableMapScript{},
		Scope:           scope,
	}
	implicitTexts := make([]impText, 0)

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing opening curly brace for mapscripts '%s'", p.peekToken.LineNumber, statement.Name.Value)
	}
	p.nextToken()

	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type != token.IDENT {
			return nil, nil, fmt.Errorf("line %d: expected map script type, but got '%s' instead", p.curToken.LineNumber, p.curToken.Literal)
		}
		mapScriptType := p.curToken.Literal
		p.nextToken()
		if p.curToken.Type == token.COLON {
			if err := p.expectPeek(token.IDENT); err != nil {
				return nil, nil, fmt.Errorf("line %d: expected map script label after ':', but got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
			}
			statement.MapScripts = append(statement.MapScripts, ast.MapScript{
				Type:   mapScriptType,
				Name:   p.curToken.Literal,
				Script: nil,
			})
			p.nextToken()
		} else if p.curToken.Type == token.LBRACE {
			p.nextToken()
			scriptName := fmt.Sprintf("%s_%s", statement.Name.Value, mapScriptType)
			blockStmt, stmtTexts, err := p.parseBlockStatement(scriptName)
			if err != nil {
				return nil, nil, err
			}
			implicitTexts = append(implicitTexts, stmtTexts...)
			statement.MapScripts = append(statement.MapScripts, ast.MapScript{
				Type: mapScriptType,
				Name: scriptName,
				Script: &ast.ScriptStatement{
					Name: &ast.Identifier{
						Value: scriptName,
					},
					Body:  blockStmt,
					Scope: token.LOCAL,
				},
			})
			p.nextToken()
		} else if p.curToken.Type == token.LBRACKET {
			tableEntries := []ast.TableMapScriptEntry{}
			p.nextToken()
			i := 0
			for p.curToken.Type != token.RBRACKET {
				var sb strings.Builder
				startLineNumber := p.curToken.LineNumber
				for p.curToken.Type != token.COMMA {
					if sb.Len() != 0 {
						sb.WriteByte(' ')
					}
					sb.WriteString(p.tryReplaceWithConstant(p.curToken.Literal))
					p.nextToken()
					if p.curToken.Type == token.EOF {
						return nil, nil, fmt.Errorf("line %d: missing ',' to specify map script table entry comparison value", startLineNumber)
					}
				}
				conditionValue := sb.String()
				if len(conditionValue) == 0 {
					return nil, nil, fmt.Errorf("line %d: expected condition for map script table entry, but it was empty", p.curToken.LineNumber)
				}
				p.nextToken()
				sb.Reset()
				startLineNumber = p.curToken.LineNumber
				for p.curToken.Type != token.COLON && p.curToken.Type != token.LBRACE {
					if sb.Len() != 0 {
						sb.WriteByte(' ')
					}
					sb.WriteString(p.tryReplaceWithConstant(p.curToken.Literal))
					p.nextToken()
					if p.curToken.Type == token.EOF {
						return nil, nil, fmt.Errorf("line %d: missing ':' or '{' to specify map script table entry", startLineNumber)
					}
				}
				comparisonValue := sb.String()
				if len(comparisonValue) == 0 {
					return nil, nil, fmt.Errorf("line %d: expected comparison value for map script table entry, but it was empty", p.curToken.LineNumber)
				}

				if p.curToken.Type == token.COLON {
					if err := p.expectPeek(token.IDENT); err != nil {
						return nil, nil, fmt.Errorf("line %d: expected map script label after ':', but got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
					}
					tableEntries = append(tableEntries, ast.TableMapScriptEntry{
						Condition:  conditionValue,
						Comparison: comparisonValue,
						Name:       p.curToken.Literal,
						Script:     nil,
					})
					p.nextToken()
				} else if p.curToken.Type == token.LBRACE {
					p.nextToken()
					scriptName := fmt.Sprintf("%s_%s_%d", statement.Name.Value, mapScriptType, i)
					blockStmt, stmtTexts, err := p.parseBlockStatement(scriptName)
					if err != nil {
						return nil, nil, err
					}
					implicitTexts = append(implicitTexts, stmtTexts...)
					tableEntries = append(tableEntries, ast.TableMapScriptEntry{
						Condition:  conditionValue,
						Comparison: comparisonValue,
						Name:       scriptName,
						Script: &ast.ScriptStatement{
							Name: &ast.Identifier{
								Value: scriptName,
							},
							Body:  blockStmt,
							Scope: token.LOCAL,
						},
					})
					p.nextToken()
				}
				i++
			}
			statement.TableMapScripts = append(statement.TableMapScripts, ast.TableMapScript{
				Type:    mapScriptType,
				Name:    fmt.Sprintf("%s_%s", statement.Name.Value, mapScriptType),
				Entries: tableEntries,
			})
			p.nextToken()
		}
	}

	return statement, implicitTexts, nil
}

func (p *Parser) parseFormatStringOperator() (string, error) {
	if err := p.expectPeek(token.LPAREN); err != nil {
		return "", fmt.Errorf("line %d: format operator must begin with an open parenthesis '('", p.peekToken.LineNumber)
	}
	if err := p.expectPeek(token.STRING); err != nil {
		return "", fmt.Errorf("line %d: invalid format() argument '%s'. Expected a string literal", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	lineNum := p.curToken.LineNumber
	rawText := p.curToken.Literal
	var fontID string
	setFontID := false
	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if err := p.expectPeek(token.STRING); err != nil {
			return "", fmt.Errorf("line %d: invalid format() fontId '%s'. Expected string", p.peekToken.LineNumber, p.peekToken.Literal)
		}
		fontID = p.curToken.Literal
		setFontID = true
	}
	maxTextLength := 208
	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if err := p.expectPeek(token.INT); err != nil {
			return "", fmt.Errorf("line %d: invalid format() maxLineLen '%s'. Expected integer", p.peekToken.LineNumber, p.peekToken.Literal)
		}
		num, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
		if err != nil {
			return "", fmt.Errorf("line %d: invalid format() maxLineLen '%s'. Expected integer", p.curToken.LineNumber, p.curToken.Literal)
		}
		maxTextLength = int(num)
	}
	if err := p.expectPeek(token.RPAREN); err != nil {
		return "", fmt.Errorf("line %d: missing closing parenthesis ')' for format()", p.peekToken.LineNumber)
	}
	if p.fonts == nil {
		fw, err := LoadFontWidths(p.fontConfigFilepath)
		if err != nil {
			log.Printf("PORYSCRIPT WARNING: Failed to load fonts JSON config file. Text auto-formatting will not work. Please specify a valid font config filepath with -fw option. '%s'\n", err.Error())
		}
		p.fonts = &fw
	}
	if !setFontID {
		fontID = p.fonts.DefaultFontID
	}
	formatted, err := p.fonts.FormatText(rawText, maxTextLength, fontID)
	if err != nil {
		return "", fmt.Errorf("line %d: %s", lineNum, err.Error())
	}
	return formatted, nil
}

func (p *Parser) parseIfStatement(scriptName string) (*ast.IfStatement, []impText, error) {
	statement := &ast.IfStatement{
		Token: p.curToken,
	}
	implicitTexts := make([]impText, 0)

	// First if statement condition
	consequence, stmtTexts, err := p.parseConditionExpression(scriptName)
	if err != nil {
		return nil, nil, err
	}
	implicitTexts = append(implicitTexts, stmtTexts...)
	statement.Consequence = consequence

	// Possibly-many elif conditions
	for p.peekToken.Type == token.ELSEIF {
		p.nextToken()
		consequence, stmtTexts, err := p.parseConditionExpression(scriptName)
		if err != nil {
			return nil, nil, err
		}
		implicitTexts = append(implicitTexts, stmtTexts...)
		statement.ElifConsequences = append(statement.ElifConsequences, consequence)
	}

	// Trailing else block
	if p.peekToken.Type == token.ELSE {
		p.nextToken()
		if err := p.expectPeek(token.LBRACE); err != nil {
			return nil, nil, fmt.Errorf("line %d: missing opening curly brace of else statement", p.curToken.LineNumber)
		}
		p.nextToken()
		blockStmt, stmtTexts, err := p.parseBlockStatement(scriptName)
		if err != nil {
			return nil, nil, err
		}
		implicitTexts = append(implicitTexts, stmtTexts...)
		statement.ElseConsequence = blockStmt
	}

	return statement, implicitTexts, nil
}

func (p *Parser) parseWhileStatement(scriptName string) (*ast.WhileStatement, []impText, error) {
	statement := &ast.WhileStatement{
		Token: p.curToken,
	}
	implicitTexts := make([]impText, 0)
	p.pushBreakStack(statement)
	p.pushContinueStack(statement)

	// while statement condition
	consequence, stmtTexts, err := p.parseConditionExpression(scriptName)
	if err != nil {
		return nil, nil, err
	}
	implicitTexts = append(implicitTexts, stmtTexts...)
	p.popBreakStack()
	p.popContinueStack()
	statement.Consequence = consequence

	return statement, implicitTexts, nil
}

func (p *Parser) parseDoWhileStatement(scriptName string) (*ast.DoWhileStatement, []impText, error) {
	statement := &ast.DoWhileStatement{
		Token: p.curToken,
	}
	implicitTexts := make([]impText, 0)

	p.pushBreakStack(statement)
	p.pushContinueStack(statement)
	expression := &ast.ConditionExpression{}

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing opening curly brace of do...while statement", p.curToken.LineNumber)
	}
	p.nextToken()
	blockStmt, stmtTexts, err := p.parseBlockStatement(scriptName)
	if err != nil {
		return nil, nil, err
	}
	implicitTexts = append(implicitTexts, stmtTexts...)
	expression.Body = blockStmt
	p.popBreakStack()
	p.popContinueStack()

	if err := p.expectPeek(token.WHILE); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing 'while' after body of do...while statement", p.curToken.LineNumber)
	}
	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing '(' to start condition for do...while statement", p.curToken.LineNumber)
	}

	boolExpression, err := p.parseBooleanExpression(false, false)
	if err != nil {
		return nil, nil, err
	}
	expression.Expression = boolExpression
	statement.Consequence = expression
	return statement, implicitTexts, nil
}

func (p *Parser) parseBreakStatement(scriptName string) (*ast.BreakStatement, error) {
	statement := &ast.BreakStatement{
		Token: p.curToken,
	}

	if p.peekBreakStack() == nil {
		return nil, fmt.Errorf("line %d: 'break' statement outside of any break-able scope", p.curToken.LineNumber)
	}
	statement.ScopeStatment = p.peekBreakStack()

	return statement, nil
}

func (p *Parser) parseContinueStatement(scriptName string) (*ast.ContinueStatement, error) {
	statement := &ast.ContinueStatement{
		Token: p.curToken,
	}

	if p.peekContinueStack() == nil {
		return nil, fmt.Errorf("line %d: 'continue' statement outside of any continue-able scope", p.curToken.LineNumber)
	}
	statement.LoopStatment = p.peekContinueStack()

	if p.peekToken.Type != token.RBRACE {
		return nil, fmt.Errorf("line %d: 'continue' must be the last statement in block scope", p.peekToken.LineNumber)
	}

	return statement, nil
}

func (p *Parser) parseSwitchStatement(scriptName string) (*ast.SwitchStatement, []impText, error) {
	statement := &ast.SwitchStatement{
		Token: p.curToken,
		Cases: []*ast.SwitchCase{},
	}
	implicitTexts := make([]impText, 0)
	p.pushBreakStack(statement)
	originalLineNumber := p.curToken.LineNumber

	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing opening parenthesis of switch statement operand", p.curToken.LineNumber)
	}
	if err := p.expectPeek(token.VAR); err != nil {
		return nil, nil, fmt.Errorf("line %d: invalid switch statement operand '%s'. Must be 'var`", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing '(' after var operator. Got '%s` instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}

	p.nextToken()
	parts := []string{}
	for p.curToken.Type != token.RPAREN {
		if p.curToken.Type == token.EOF {
			return nil, nil, fmt.Errorf("line %d: missing closing parenthesis of switch statement value", originalLineNumber)
		}
		parts = append(parts, p.tryReplaceWithConstant(p.curToken.Literal))
		p.nextToken()
	}
	p.nextToken()
	statement.Operand = strings.Join(parts, " ")

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing opening curly brace of switch statement", p.curToken.LineNumber)
	}
	p.nextToken()

	// Parse each of the switch cases, including "default".
	caseValues := make(map[string]bool)
	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.CASE {
			caseLineNum := p.curToken.LineNumber
			p.nextToken()
			parts := []string{}
			for p.curToken.Type != token.COLON {
				parts = append(parts, p.tryReplaceWithConstant(p.curToken.Literal))
				p.nextToken()
				if p.curToken.Type == token.EOF {
					return nil, nil, fmt.Errorf("line %d: missing `:` after 'case'", caseLineNum)
				}
			}
			caseValue := strings.Join(parts, " ")
			if caseValues[caseValue] {
				return nil, nil, fmt.Errorf("line %d: duplicate switch cases detected for case '%s'", p.curToken.LineNumber, caseValue)
			}
			caseValues[caseValue] = true
			p.nextToken()

			body, stmtTexts, err := p.parseSwitchBlockStatement(scriptName)
			if err != nil {
				return nil, nil, err
			}
			implicitTexts = append(implicitTexts, stmtTexts...)
			statement.Cases = append(statement.Cases, &ast.SwitchCase{
				Value: caseValue,
				Body:  body,
			})
		} else if p.curToken.Type == token.DEFAULT {
			if statement.DefaultCase != nil {
				return nil, nil, fmt.Errorf("line %d: multiple `default` cases found in switch statement. Only one `default` case is allowed", p.peekToken.LineNumber)
			}
			if err := p.expectPeek(token.COLON); err != nil {
				return nil, nil, fmt.Errorf("line %d: missing `:` after default", p.curToken.LineNumber)
			}
			p.nextToken()
			body, stmtTexts, err := p.parseSwitchBlockStatement(scriptName)
			if err != nil {
				return nil, nil, err
			}
			implicitTexts = append(implicitTexts, stmtTexts...)
			statement.Cases = append(statement.Cases, &ast.SwitchCase{
				IsDefault: true,
				Body:      body,
			})
			statement.DefaultCase = &ast.SwitchCase{
				Body: body,
			}
		} else {
			return nil, nil, fmt.Errorf("line %d: invalid start of switch case '%s'. Expected 'case' or 'default'", p.curToken.LineNumber, p.curToken.Literal)
		}
	}

	p.popBreakStack()

	if len(statement.Cases) == 0 && statement.DefaultCase == nil {
		return nil, nil, fmt.Errorf("line %d: switch statement has no cases or default case", originalLineNumber)
	}

	return statement, implicitTexts, nil
}

func (p *Parser) parseConditionExpression(scriptName string) (*ast.ConditionExpression, []impText, error) {
	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, nil, fmt.Errorf("line %d: missing '(' to start boolean expression", p.peekToken.LineNumber)
	}

	expression := &ast.ConditionExpression{}
	implicitTexts := make([]impText, 0)
	boolExpression, err := p.parseBooleanExpression(false, false)
	if err != nil {
		return nil, nil, err
	}
	expression.Expression = boolExpression
	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, nil, err
	}
	p.nextToken()

	blockStmt, stmtTexts, err := p.parseBlockStatement(scriptName)
	if err != nil {
		return nil, nil, err
	}
	implicitTexts = append(implicitTexts, stmtTexts...)
	expression.Body = blockStmt
	return expression, implicitTexts, nil
}

func (p *Parser) parseBooleanExpression(single bool, negated bool) (ast.BooleanExpression, error) {
	nested := p.peekTokenIs(token.LPAREN)
	negatedNested := p.peekTokenIs(token.NOT) && p.peek2TokenIs(token.LPAREN)
	if nested || negatedNested {
		// Open parenthesis indicates a nested expression.
		// If a NOT operator is used before a nested expression, distribute
		// it to the nested expression (i.e. De Morgan's law).
		p.nextToken()
		if nested {
			negatedNested = negated
		} else if negatedNested {
			p.nextToken()
			negatedNested = !negated
		}
		nestedExpression, err := p.parseBooleanExpression(false, negatedNested)
		if err != nil {
			return nil, err
		}
		if p.curToken.Type != token.RPAREN {
			return nil, fmt.Errorf("line %d: missing closing ')' for nested boolean expression", p.curToken.LineNumber)
		}
		if p.peekTokenIs(token.AND) || p.peekTokenIs(token.OR) {
			p.nextToken()
			return p.parseRightSideExpression(nestedExpression, single, negated)
		}
		p.nextToken()
		return nestedExpression, nil
	}

	leaf, err := p.parseLeafBooleanExpression()
	if err != nil {
		return nil, err
	}

	if negated {
		leaf.Operator = getNegatedBooleanOperator(leaf.Operator)
	}
	if single {
		return leaf, nil
	}
	return p.parseRightSideExpression(leaf, single, negated)
}

func getNegatedBooleanOperator(operator token.Type) token.Type {
	switch operator {
	case token.EQ:
		return token.NEQ
	case token.NEQ:
		return token.EQ
	case token.LT:
		return token.GTE
	case token.GT:
		return token.LTE
	case token.LTE:
		return token.GT
	case token.GTE:
		return token.LT
	case token.AND:
		return token.OR
	case token.OR:
		return token.AND
	default:
		return operator
	}
}

func (p *Parser) parseRightSideExpression(left ast.BooleanExpression, single bool, negated bool) (ast.BooleanExpression, error) {
	curTokenType := p.curToken.Type
	if negated {
		curTokenType = getNegatedBooleanOperator(curTokenType)
	}

	if p.curToken.Type == token.AND {
		operator := curTokenType
		right, err := p.parseBooleanExpression(true, negated)
		if err != nil {
			return nil, err
		}
		grouped := &ast.BinaryExpression{
			Left:     left,
			Operator: operator,
			Right:    right,
		}
		if p.curToken.Literal == token.RPAREN {
			return grouped, nil
		}
		operator = p.curToken.Type
		if negated {
			operator = getNegatedBooleanOperator(p.curToken.Type)
		}
		binaryExpression := &ast.BinaryExpression{Left: grouped, Operator: operator}
		boolExpression, err := p.parseBooleanExpression(false, negated)
		if err != nil {
			return nil, err
		}
		binaryExpression.Right = boolExpression
		return binaryExpression, nil
	} else if p.curToken.Type == token.OR {
		operator := curTokenType
		right, err := p.parseBooleanExpression(false, negated)
		if err != nil {
			return nil, err
		}
		binaryExpression := &ast.BinaryExpression{Left: left, Operator: operator, Right: right}
		return binaryExpression, nil
	} else {
		return left, nil
	}
}

func (p *Parser) parseLeafBooleanExpression() (*ast.OperatorExpression, error) {
	// Left-side of binary expression must be a special condition statement.
	usedNotOperator := false
	operatorExpression := &ast.OperatorExpression{}
	if p.peekTokenIs(token.NOT) {
		operatorExpression.Operator = token.EQ
		p.nextToken()
		usedNotOperator = true
	}

	if !p.peekTokenIs(token.VAR) && !p.peekTokenIs(token.FLAG) && !p.peekTokenIs(token.DEFEATED) {
		return nil, fmt.Errorf("line %d: left side of binary expression must be var(), flag(), or defeated() operator. Instead, found '%s'", p.curToken.LineNumber, p.peekToken.Literal)
	}
	p.nextToken()
	operatorExpression.Type = p.curToken.Type

	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, fmt.Errorf("line %d: missing opening parenthesis for condition operator '%s'", p.curToken.LineNumber, operatorExpression.Type)
	}
	if p.peekToken.Type == token.RPAREN {
		return nil, fmt.Errorf("line %d: missing value for condition operator '%s'", p.curToken.LineNumber, operatorExpression.Type)
	}
	p.nextToken()
	parts := []string{}
	lineNum := p.curToken.LineNumber
	for p.curToken.Type != token.RPAREN {
		parts = append(parts, p.tryReplaceWithConstant(p.curToken.Literal))
		p.nextToken()
		if p.curToken.Type == token.EOF {
			return nil, fmt.Errorf("line %d: missing closing ')' for condition operator value", lineNum)
		}
	}
	operatorExpression.Operand = strings.Join(parts, " ")
	p.nextToken()

	if usedNotOperator {
		if operatorExpression.Type == token.VAR {
			operatorExpression.ComparisonValue = "0"
		} else if operatorExpression.Type == token.FLAG || operatorExpression.Type == token.DEFEATED {
			operatorExpression.ComparisonValue = token.FALSE
		}
	} else {
		if operatorExpression.Type == token.VAR {
			err := p.parseConditionVarOperator(operatorExpression)
			if err != nil {
				return nil, err
			}
		} else if operatorExpression.Type == token.FLAG {
			err := p.parseConditionFlagLikeOperator(operatorExpression, "flag")
			if err != nil {
				return nil, err
			}
		} else if operatorExpression.Type == token.DEFEATED {
			err := p.parseConditionFlagLikeOperator(operatorExpression, "defeated")
			if err != nil {
				return nil, err
			}
		}
	}

	return operatorExpression, nil
}

func (p *Parser) parseConditionVarOperator(expression *ast.OperatorExpression) error {
	if p.curToken.Type != token.GT && p.curToken.Type != token.GTE && p.curToken.Type != token.LT &&
		p.curToken.Type != token.LTE && p.curToken.Type != token.EQ && p.curToken.Type != token.NEQ {
		// Missing condition operator means test for implicit truthiness.
		expression.Operator = token.NEQ
		expression.ComparisonValue = "0"
		return nil
	}
	expression.Operator = p.curToken.Type
	p.nextToken()

	if p.curToken.Type == token.RPAREN {
		return fmt.Errorf("line %d: missing comparison value for var operator", p.curToken.LineNumber)
	}
	parts := []string{}
	lineNum := p.curToken.LineNumber
	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.AND && p.curToken.Type != token.OR {
		parts = append(parts, p.tryReplaceWithConstant(p.curToken.Literal))
		p.nextToken()
		if p.curToken.Type == token.EOF {
			return fmt.Errorf("line %d: missing ')', '&&' or '||' when evaluating 'var' operator", lineNum)
		}
	}

	expression.ComparisonValue = strings.Join(parts, " ")
	return nil
}

func (p *Parser) parseConditionFlagLikeOperator(expression *ast.OperatorExpression, operatorName string) error {
	if p.curToken.Type != token.EQ && p.curToken.Type != token.NEQ {
		// Missing '==' or '!=' means test for implicit truthiness.
		expression.Operator = token.EQ
		expression.ComparisonValue = token.TRUE
		return nil
	}

	expression.Operator = p.curToken.Type
	p.nextToken()

	if p.curToken.Type == token.RPAREN {
		return fmt.Errorf("line %d: missing comparison value for %s operator", p.curToken.LineNumber, operatorName)
	}

	if p.curToken.Type != token.TRUE && p.curToken.Type != token.FALSE {
		return fmt.Errorf("line %d: invalid %s comparison value '%s'. Only TRUE and FALSE are allowed", p.curToken.LineNumber, operatorName, p.curToken.Literal)
	}
	expression.ComparisonValue = string(p.curToken.Type)
	p.nextToken()
	return nil
}

// Automatically adds a terminator character to the text, if it doesn't already have one.
func (p *Parser) formatTextTerminator(text string) string {
	if !strings.HasSuffix(text, "$") {
		text += "$"
	}
	return text
}

func (p *Parser) parsePoryswitchStatement(scriptName string) ([]ast.Statement, []impText, error) {
	startLineNumber := p.curToken.LineNumber
	switchCase, switchValue, err := p.parsePoryswitchHeader()
	if err != nil {
		return nil, nil, err
	}
	cases, caseTexts, err := p.parsePoryswitchStatementCases(scriptName)
	if err != nil {
		return nil, nil, err
	}
	statements, ok := cases[switchValue]
	if !ok {
		statements, ok = cases["_"]
		if !ok {
			return nil, nil, fmt.Errorf("line %d: no poryswitch case found for '%s=%s', which was specified with the '-s' option", startLineNumber, switchCase, switchValue)
		}
	}
	implicitTexts, ok := caseTexts[switchValue]
	if !ok {
		implicitTexts, ok = caseTexts["_"]
		if !ok {
			return nil, nil, fmt.Errorf("line %d: no poryswitch case found for '%s=%s', which was specified with the '-s' option", startLineNumber, switchCase, switchValue)
		}
	}
	return statements, implicitTexts, nil
}

func (p *Parser) parsePoryswitchStatementCases(scriptName string) (map[string][]ast.Statement, map[string][]impText, error) {
	statementCases := make(map[string][]ast.Statement)
	implicitTexts := make(map[string][]impText)
	startLineNumber := p.curToken.LineNumber
	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.EOF {
			return nil, nil, fmt.Errorf("line %d: missing closing curly braces for poryswitch statement", startLineNumber)
		}
		if p.curToken.Type != token.IDENT && p.curToken.Type != token.INT {
			return nil, nil, fmt.Errorf("line %d: invalid poryswitch case '%s'. Expected a simple identifier", p.curToken.LineNumber, p.curToken.Literal)
		}
		caseValue := p.curToken.Literal
		p.nextToken()
		if p.curToken.Type == token.COLON || p.curToken.Type == token.LBRACE {
			usedBrace := p.curToken.Type == token.LBRACE
			p.nextToken()
			statements, stmtTexts, err := p.parsePoryswitchStatements(scriptName, usedBrace)
			if err != nil {
				return nil, nil, err
			}
			statementCases[caseValue] = statements
			implicitTexts[caseValue] = stmtTexts
			if usedBrace {
				if p.curToken.Type != token.RBRACE {
					return nil, nil, fmt.Errorf("line %d: missing closing curly brace for poryswitch case '%s'", startLineNumber, caseValue)
				}
				p.nextToken()
			}
		} else {
			return nil, nil, fmt.Errorf("line %d: invalid token '%s' after poryswitch case '%s'. Expected ':' or '{'", p.curToken.LineNumber, p.curToken.Literal, caseValue)
		}
	}
	return statementCases, implicitTexts, nil
}

func (p *Parser) parsePoryswitchStatements(scriptName string, allowMultiple bool) ([]ast.Statement, []impText, error) {
	statements := make([]ast.Statement, 0)
	implicitTexts := make([]impText, 0)
	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.PORYSWITCH {
			poryswitchStatements, stmtTexts, err := p.parsePoryswitchStatement(scriptName)
			if err != nil {
				return nil, nil, err
			}
			statements = append(statements, poryswitchStatements...)
			implicitTexts = append(implicitTexts, stmtTexts...)
			p.nextToken()
		} else {
			stmts, stmtTexts, err := p.parseStatement(scriptName)
			if err != nil {
				return nil, nil, err
			}
			statements = append(statements, stmts...)
			implicitTexts = append(implicitTexts, stmtTexts...)
			p.nextToken()
		}
		if !allowMultiple {
			break
		}
	}
	return statements, implicitTexts, nil
}

func (p *Parser) parseConstant() error {
	initialLineNumber := p.curToken.LineNumber
	if err := p.expectPeek(token.IDENT); err != nil {
		return fmt.Errorf("line %d: expected identifier after const, but got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	constName := p.curToken.Literal
	if _, ok := p.constants[constName]; ok {
		return fmt.Errorf("line %d: duplicate const '%s'. Must use unique const names", p.curToken.LineNumber, constName)
	}
	if err := p.expectPeek(token.ASSIGN); err != nil {
		return fmt.Errorf("line %d: missing equals sign after const name '%s'", p.peekToken.LineNumber, constName)
	}

	var sb strings.Builder
	for {
		_, ok := topLevelTokens[p.peekToken.Type]
		if ok || p.curToken.Type == token.EOF {
			break
		}
		p.nextToken()
		if sb.Len() > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteString(p.tryReplaceWithConstant(p.curToken.Literal))
	}

	if sb.Len() == 0 {
		return fmt.Errorf("line %d: missing value for const '%s'", initialLineNumber, constName)
	}
	p.constants[constName] = sb.String()
	return nil
}

func (p *Parser) tryReplaceWithConstant(value string) string {
	if constValue, ok := p.constants[p.curToken.Literal]; ok {
		return constValue
	}
	return value
}
