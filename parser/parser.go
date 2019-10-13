package parser

import (
	"fmt"
	"log"
	"strings"

	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/token"
)

// Parser is a Poryscript AST parser.
type Parser struct {
	l                  *lexer.Lexer
	curToken           token.Token
	peekToken          token.Token
	inlineTexts        []ast.Text
	inlineTextsSet     map[string]string
	inlineTextCounts   map[string]int
	textStatements     []*ast.TextStatement
	breakStack         []ast.Statement
	continueStack      []ast.Statement
	fontConfigFilepath string
	fonts              *FontWidthsConfig
}

// New creates a new Poryscript AST Parser.
func New(l *lexer.Lexer, fontConfigFilepath string) *Parser {
	p := &Parser{
		l:                  l,
		inlineTexts:        make([]ast.Text, 0),
		inlineTextsSet:     make(map[string]string),
		inlineTextCounts:   make(map[string]int),
		textStatements:     make([]*ast.TextStatement, 0),
		fontConfigFilepath: fontConfigFilepath,
	}
	// Read two tokens, so curToken and peekToken are both set.
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
	p.peekToken = p.l.NextToken()
}

func (p *Parser) peekTokenIs(expectedType token.Type) bool {
	return p.peekToken.Type == expectedType
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
			IsGlobal: true,
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
		statement, err := p.parseScriptStatement()
		if err != nil {
			return nil, err
		}
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
	}

	return nil, fmt.Errorf("line %d: could not parse top-level statement for '%s'", p.curToken.LineNumber, p.curToken.Literal)
}

func (p *Parser) parseScriptStatement() (*ast.ScriptStatement, error) {
	statement := &ast.ScriptStatement{Token: p.curToken}
	if err := p.expectPeek(token.IDENT); err != nil {
		return nil, fmt.Errorf("line %d: missing name for script", p.curToken.LineNumber)
	}

	statement.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, fmt.Errorf("line %d: missing opening curly brace for script '%s'", p.curToken.LineNumber, statement.Name.Value)
	}

	p.nextToken()

	blockStmt, err := p.parseBlockStatement(statement.Name.Value)
	if err != nil {
		return nil, err
	}
	statement.Body = blockStmt
	return statement, nil
}

func (p *Parser) parseBlockStatement(scriptName string) (*ast.BlockStatement, error) {
	block := &ast.BlockStatement{
		Token:      p.curToken,
		Statements: []ast.Statement{},
	}

	for p.curToken.Type != token.RBRACE {
		if p.curToken.Type == token.EOF {
			return nil, fmt.Errorf("line %d: missing closing curly brace for block statement", block.Token.LineNumber)
		}

		statement, err := p.parseStatement(scriptName)
		if err != nil {
			return nil, err
		}

		block.Statements = append(block.Statements, statement)
		p.nextToken()
	}

	return block, nil
}

func (p *Parser) parseSwitchBlockStatement(scriptName string) (*ast.BlockStatement, error) {
	block := &ast.BlockStatement{
		Token:      p.curToken,
		Statements: []ast.Statement{},
	}

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.CASE && p.curToken.Type != token.DEFAULT {
		if p.curToken.Type == token.EOF {
			return nil, fmt.Errorf("line %d: missing end for switch case body", block.Token.LineNumber)
		}

		statement, err := p.parseStatement(scriptName)
		if err != nil {
			return nil, err
		}

		block.Statements = append(block.Statements, statement)
		p.nextToken()
	}

	return block, nil
}

func (p *Parser) parseStatement(scriptName string) (ast.Statement, error) {
	var statement ast.Statement
	var err error
	switch p.curToken.Type {
	case token.IDENT:
		statement, err = p.parseCommandStatement(scriptName)
	case token.IF:
		statement, err = p.parseIfStatement(scriptName)
	case token.WHILE:
		statement, err = p.parseWhileStatement(scriptName)
	case token.DO:
		statement, err = p.parseDoWhileStatement(scriptName)
	case token.BREAK:
		statement, err = p.parseBreakStatement(scriptName)
	case token.CONTINUE:
		statement, err = p.parseContinueStatement(scriptName)
	case token.SWITCH:
		statement, err = p.parseSwitchStatement(scriptName)
	default:
		err = fmt.Errorf("line %d: could not parse statement for '%s'", p.curToken.LineNumber, p.curToken.Literal)
	}

	if err != nil {
		return nil, err
	}
	return statement, nil
}

func (p *Parser) parseCommandStatement(scriptName string) (ast.Statement, error) {
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
				err := fmt.Errorf("line %d: missing closing parenthesis for command '%s'", command.Token.LineNumber, command.Name.TokenLiteral())
				return nil, err
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
					return nil, err
				}
				textLabel := p.addText(scriptName, strValue)
				argParts = append(argParts, textLabel)
			} else if p.curToken.Type == token.STRING {
				textLabel := p.addText(scriptName, p.formatTextTerminator(p.curToken.Literal))
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

	return command, nil
}

func (p *Parser) addText(scriptName string, text string) string {
	if textLabel, ok := p.inlineTextsSet[text]; ok {
		return textLabel
	}
	textLabel := getImplicitTextLabel(scriptName, p.inlineTextCounts[scriptName])
	p.inlineTextCounts[scriptName]++
	p.inlineTextsSet[text] = textLabel
	p.inlineTexts = append(p.inlineTexts, ast.Text{
		Name:     textLabel,
		Value:    text,
		IsGlobal: false,
	})
	return textLabel
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
	if p.curToken.Type == token.FORMAT {
		var err error
		strValue, err = p.parseFormatStringOperator()
		if err != nil {
			return nil, err
		}
		strValue = p.formatTextTerminator(strValue)
	} else if p.curToken.Type == token.STRING {
		strValue = p.formatTextTerminator(p.curToken.Literal)
	} else {
		return nil, fmt.Errorf("line %d: body of text statement must be a string or formatted string. Got '%s' instead", p.curToken.LineNumber, p.curToken.Literal)
	}

	statement.Value = strValue
	p.textStatements = append(p.textStatements, statement)
	if err := p.expectPeek(token.RBRACE); err != nil {
		return nil, fmt.Errorf("line %d: expected closing curly brace for text. Got '%s' instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	return statement, nil
}

func (p *Parser) parseFormatStringOperator() (string, error) {
	if err := p.expectPeek(token.LPAREN); err != nil {
		return "", fmt.Errorf("line %d: format operator must begin with an open parenthesis '('", p.peekToken.LineNumber)
	}
	if err := p.expectPeek(token.STRING); err != nil {
		return "", fmt.Errorf("line %d: invalid format() argument '%s'. Expected a string literal", p.peekToken.LineNumber, p.peekToken.Literal)
	}
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
	return p.fonts.FormatText(rawText, 208, fontID)
}

func (p *Parser) parseIfStatement(scriptName string) (*ast.IfStatement, error) {
	statement := &ast.IfStatement{
		Token: p.curToken,
	}

	// First if statement condition
	consequence, err := p.parseConditionExpression(scriptName)
	if err != nil {
		return nil, err
	}
	statement.Consequence = consequence

	// Possibly-many elif conditions
	for p.peekToken.Type == token.ELSEIF {
		p.nextToken()
		consequence, err := p.parseConditionExpression(scriptName)
		if err != nil {
			return nil, err
		}
		statement.ElifConsequences = append(statement.ElifConsequences, consequence)
	}

	// Trailing else block
	if p.peekToken.Type == token.ELSE {
		p.nextToken()
		if err := p.expectPeek(token.LBRACE); err != nil {
			return nil, fmt.Errorf("line %d: missing opening curly brace of else statement", p.curToken.LineNumber)
		}
		p.nextToken()
		blockStmt, err := p.parseBlockStatement(scriptName)
		if err != nil {
			return nil, err
		}
		statement.ElseConsequence = blockStmt
	}

	return statement, nil
}

func (p *Parser) parseWhileStatement(scriptName string) (*ast.WhileStatement, error) {
	statement := &ast.WhileStatement{
		Token: p.curToken,
	}
	p.pushBreakStack(statement)
	p.pushContinueStack(statement)

	// while statement condition
	consequence, err := p.parseConditionExpression(scriptName)
	if err != nil {
		return nil, err
	}
	p.popBreakStack()
	p.popContinueStack()
	statement.Consequence = consequence

	return statement, nil
}

func (p *Parser) parseDoWhileStatement(scriptName string) (*ast.DoWhileStatement, error) {
	statement := &ast.DoWhileStatement{
		Token: p.curToken,
	}
	p.pushBreakStack(statement)
	p.pushContinueStack(statement)
	expression := &ast.ConditionExpression{}

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, fmt.Errorf("line %d: missing opening curly brace of do...while statement", p.curToken.LineNumber)
	}
	p.nextToken()
	blockStmt, err := p.parseBlockStatement(scriptName)
	if err != nil {
		return nil, err
	}
	expression.Body = blockStmt
	p.popBreakStack()
	p.popContinueStack()

	if err := p.expectPeek(token.WHILE); err != nil {
		return nil, fmt.Errorf("line %d: missing 'while' after body of do...while statement", p.curToken.LineNumber)
	}
	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, fmt.Errorf("line %d: missing '(' to start condition for do...while statement", p.curToken.LineNumber)
	}

	boolExpression, err := p.parseBooleanExpression(false)
	if err != nil {
		return nil, err
	}
	expression.Expression = boolExpression
	statement.Consequence = expression
	return statement, nil
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

func (p *Parser) parseSwitchStatement(scriptName string) (*ast.SwitchStatement, error) {
	statement := &ast.SwitchStatement{
		Token: p.curToken,
		Cases: []*ast.SwitchCase{},
	}
	p.pushBreakStack(statement)
	originalLineNumber := p.curToken.LineNumber

	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, fmt.Errorf("line %d: missing opening parenthesis of switch statement operand", p.curToken.LineNumber)
	}
	if err := p.expectPeek(token.VAR); err != nil {
		return nil, fmt.Errorf("line %d: invalid switch statement operand '%s'. Must be 'var`", p.peekToken.LineNumber, p.peekToken.Literal)
	}
	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, fmt.Errorf("line %d: missing '(' after var operator. Got '%s` instead", p.peekToken.LineNumber, p.peekToken.Literal)
	}

	p.nextToken()
	parts := []string{}
	for p.curToken.Type != token.RPAREN {
		parts = append(parts, p.curToken.Literal)
		p.nextToken()
	}
	p.nextToken()
	statement.Operand = strings.Join(parts, " ")

	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, fmt.Errorf("line %d: missing opening curly brace of switch statement", p.curToken.LineNumber)
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
				parts = append(parts, p.curToken.Literal)
				p.nextToken()
				if p.curToken.Type == token.EOF {
					return nil, fmt.Errorf("line %d: missing `:` after 'case'", caseLineNum)
				}
			}
			caseValue := strings.Join(parts, " ")
			if caseValues[caseValue] {
				return nil, fmt.Errorf("line %d: duplicate switch cases detected for case '%s'", p.curToken.LineNumber, caseValue)
			}
			caseValues[caseValue] = true
			p.nextToken()

			body, err := p.parseSwitchBlockStatement(scriptName)
			if err != nil {
				return nil, err
			}
			statement.Cases = append(statement.Cases, &ast.SwitchCase{
				Value: caseValue,
				Body:  body,
			})
		} else if p.curToken.Type == token.DEFAULT {
			if statement.DefaultCase != nil {
				return nil, fmt.Errorf("line %d: multiple `default` cases found in switch statement. Only one `default` case is allowed", p.peekToken.LineNumber)
			}
			if err := p.expectPeek(token.COLON); err != nil {
				return nil, fmt.Errorf("line %d: missing `:` after default", p.curToken.LineNumber)
			}
			p.nextToken()
			body, err := p.parseSwitchBlockStatement(scriptName)
			if err != nil {
				return nil, err
			}
			statement.DefaultCase = &ast.SwitchCase{
				Body: body,
			}
		} else {
			return nil, fmt.Errorf("line %d: invalid start of switch case '%s'. Expected 'case' or 'default'", p.curToken.LineNumber, p.curToken.Literal)
		}
	}

	p.popBreakStack()

	if len(statement.Cases) == 0 && statement.DefaultCase == nil {
		return nil, fmt.Errorf("line %d: switch statement has no cases or default case", originalLineNumber)
	}

	return statement, nil
}

func (p *Parser) parseConditionExpression(scriptName string) (*ast.ConditionExpression, error) {
	if err := p.expectPeek(token.LPAREN); err != nil {
		return nil, fmt.Errorf("line %d: missing '(' to start boolean expression", p.peekToken.LineNumber)
	}

	expression := &ast.ConditionExpression{}
	boolExpression, err := p.parseBooleanExpression(false)
	if err != nil {
		return nil, err
	}
	expression.Expression = boolExpression
	if err := p.expectPeek(token.LBRACE); err != nil {
		return nil, err
	}
	p.nextToken()

	blockStmt, err := p.parseBlockStatement(scriptName)
	if err != nil {
		return nil, err
	}
	expression.Body = blockStmt
	return expression, nil
}

func (p *Parser) parseBooleanExpression(single bool) (ast.BooleanExpression, error) {
	if p.peekTokenIs(token.LPAREN) {
		// Open parenthesis indicates a nested expression.
		p.nextToken()
		nestedExpression, err := p.parseBooleanExpression(false)
		if err != nil {
			return nil, err
		}
		if p.curToken.Type != token.RPAREN {
			return nil, fmt.Errorf("line %d: missing closing ')' for nested boolean expression", p.curToken.LineNumber)
		}
		if p.peekTokenIs(token.AND) || p.peekTokenIs(token.OR) {
			p.nextToken()
			return p.parseRightSideExpression(nestedExpression, single)
		}
		p.nextToken()
		return nestedExpression, nil
	}

	leaf, err := p.parseLeafBooleanExpression()
	if err != nil {
		return nil, err
	}
	if single {
		return leaf, nil
	}
	return p.parseRightSideExpression(leaf, single)
}

func (p *Parser) parseRightSideExpression(left ast.BooleanExpression, single bool) (ast.BooleanExpression, error) {
	if p.curToken.Type == token.AND {
		operator := p.curToken.Type
		right, err := p.parseBooleanExpression(true)
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
		binaryExpression := &ast.BinaryExpression{Left: grouped, Operator: p.curToken.Type}
		boolExpression, err := p.parseBooleanExpression(false)
		if err != nil {
			return nil, err
		}
		binaryExpression.Right = boolExpression
		return binaryExpression, nil
	} else if p.curToken.Type == token.OR {
		operator := p.curToken.Type
		right, err := p.parseBooleanExpression(false)
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
		parts = append(parts, p.curToken.Literal)
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
		parts = append(parts, p.curToken.Literal)
		p.nextToken()
		if p.curToken.Type == token.EOF {
			return fmt.Errorf("line %d: missing ')', '&&' or '||' when evaluating 'var' operator", lineNum)
		}
	}

	expression.ComparisonValue = strings.Join(parts, " ")
	return nil
}

func (p *Parser) parseConditionFlagLikeOperator(expression *ast.OperatorExpression, operatorName string) error {
	if p.curToken.Type != token.EQ {
		// Missing '==' means test for implicit truthiness.
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
		return fmt.Errorf("line %d: invalid %s comparison value '%s'. Only 'TRUE' and 'FALSE' are allowed", p.curToken.LineNumber, operatorName, p.curToken.Literal)
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
