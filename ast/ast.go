package ast

import "github.com/huderlem/poryscript/token"

// Node is an interface that represents a node in a Poryscript AST.
type Node interface {
	TokenLiteral() string
}

// Statement is an interface that represents a statement node in a Poryscript AST.
type Statement interface {
	Node
	statementNode()
}

// Text holds a label and value for some script text.
type Text struct {
	Name  string
	Value string
}

// Program represents the root-level Node in any Poryscript AST.
type Program struct {
	TopLevelStatements []Statement
	Texts              []Text
}

// TokenLiteral returns a string representation of the Program node.
func (p *Program) TokenLiteral() string {
	if len(p.TopLevelStatements) > 0 {
		return p.TopLevelStatements[0].TokenLiteral()
	}
	return ""
}

// ScriptStatement is a Poryscript script statement. Script statements define
// the block of a script's execution.
type ScriptStatement struct {
	Token token.Token
	Name  *Identifier
	Body  *BlockStatement
}

func (ss *ScriptStatement) statementNode() {}

// TokenLiteral returns a string representation of the script statement.
func (ss *ScriptStatement) TokenLiteral() string { return ss.Token.Literal }

// BlockStatement is a Poryscript block, which can hold many statements and blocks inside.
// It is defined by curly braces.
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns a string representation of the block statement.
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// CommandStatement is a Poryscript command statement. Command statements map directly to
// original engine script commands.
type CommandStatement struct {
	Token token.Token
	Name  *Identifier
	Args  []string
}

func (cs *CommandStatement) statementNode() {}

// TokenLiteral returns a string representation of the command statement.
func (cs *CommandStatement) TokenLiteral() string { return cs.Token.Literal }

// Identifier represents a Poryscript identifier.
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}

// TokenLiteral returns a string representation of the identifier.
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// RawStatement is a Poryscript raw statement. Raw statements are directly
// included into the target bytecode script.
type RawStatement struct {
	Token token.Token
	Value string
}

func (rs *RawStatement) statementNode() {}

// TokenLiteral returns a string representation of the raw statement.
func (rs *RawStatement) TokenLiteral() string { return rs.Token.Literal }

// ConditionExpression is the condition for an if statement.
type ConditionExpression struct {
	Type            token.Type
	Operand         string
	Operator        token.Type
	ComparisonValue string
	Body            *BlockStatement
}

// IfStatement is an if statement in Poryscript.
type IfStatement struct {
	Token            token.Token
	Consequence      *ConditionExpression
	ElifConsequences []*ConditionExpression
	ElseConsequence  *BlockStatement
}

func (is *IfStatement) statementNode() {}

// TokenLiteral returns a string representation of the if statement.
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }

// WhileStatement is a while statement in Poryscript.
type WhileStatement struct {
	Token       token.Token
	Consequence *ConditionExpression
}

func (ws *WhileStatement) statementNode() {}

// TokenLiteral returns a string representation of the if statement.
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }

// DoWhileStatement is a do-while statement in Poryscript.
type DoWhileStatement struct {
	Token       token.Token
	Consequence *ConditionExpression
}

func (dws *DoWhileStatement) statementNode() {}

// TokenLiteral returns a string representation of the if statement.
func (dws *DoWhileStatement) TokenLiteral() string { return dws.Token.Literal }

// BreakStatement is a break statement in Poryscript.
type BreakStatement struct {
	Token        token.Token
	LoopStatment Statement
}

func (dws *BreakStatement) statementNode() {}

// TokenLiteral returns a string representation of the if statement.
func (dws *BreakStatement) TokenLiteral() string { return dws.Token.Literal }
