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
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
}

// New creates a new Poryscript AST Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// Read two tokens, so curToken and peekToken are both set.
	p.nextToken()
	p.nextToken()
	return p
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

// ParseProgram parses a Poryscript file into an AST.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.TopLevelStatements = []ast.Statement{}
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

	statement.Body = p.parseBlockStatement()
	return statement
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
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

		statement := p.parseStatement()
		if statement == nil {
			return nil
		}

		block.Statements = append(block.Statements, statement)
		p.nextToken()
	}

	return block
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.IDENT:
		statement := p.parseCommandStatement()
		if statement == nil {
			return nil
		}
		return statement
	}

	msg := fmt.Sprintf("line %d: could not parse statement for '%s'\n", p.curToken.LineNumber, p.curToken.Literal)
	p.errors = append(p.errors, msg)
	return nil
}

func (p *Parser) parseCommandStatement() ast.Statement {
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
			} else {
				if p.curToken.Type == token.LPAREN {
					numOpenParens++
				} else if p.curToken.Type == token.RPAREN {
					numOpenParens--
				}
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
