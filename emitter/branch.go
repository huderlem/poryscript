package emitter

import (
	"fmt"
	"strings"

	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/token"
)

// Interface that manages chunk branching behavior.
type brancher interface {
	renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int)) bool
	getTailChunkID() int
}

// Helper types for keeping track of script chunk branching logic.
type conditionDestination struct {
	id                 int
	operatorExpression *ast.OperatorExpression
}

// Represents the initial jump to a loop or comparison chunk.
type jump struct {
	destChunkID int
}

// Satisfies brancher interface.
func (j *jump) renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int)) bool {
	if j.destChunkID != nextChunkID {
		registerJumpChunk(j.destChunkID)
		sb.WriteString(fmt.Sprintf("\tgoto %s_%d\n", scriptName, j.destChunkID))
		return false
	}
	return true
}

// Satisfies brancher interface.
func (j *jump) getTailChunkID() int {
	return j.destChunkID
}

// Represents a break statement, where it branches to after its loop scope.
type breakContext struct {
	destChunkID int
}

// Satisfies brancher interface.
func (bc *breakContext) renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int)) bool {
	if bc.destChunkID == -1 {
		sb.WriteString("\treturn\n")
		return false
	} else if bc.destChunkID != nextChunkID {
		registerJumpChunk(bc.destChunkID)
		sb.WriteString(fmt.Sprintf("\tgoto %s_%d\n", scriptName, bc.destChunkID))
		return false
	}
	return true
}

// Satisfies brancher interface.
func (bc *breakContext) getTailChunkID() int {
	if bc.destChunkID == -1 {
		return -1
	}
	return bc.destChunkID
}

// Represents a leaf expression of a compound boolean expression.
type leafExpressionBranch struct {
	truthyDest     *conditionDestination
	falseyReturnID int
}

// Satisfies brancher interface.
func (l *leafExpressionBranch) renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int)) bool {
	registerJumpChunk(l.truthyDest.id)
	renderBranchComparison(sb, l.truthyDest, scriptName)
	if l.falseyReturnID == -1 {
		sb.WriteString("\treturn\n")
		return false
	} else if l.falseyReturnID != nextChunkID {
		registerJumpChunk(l.falseyReturnID)
		sb.WriteString(fmt.Sprintf("\tgoto %s_%d\n", scriptName, l.falseyReturnID))
		return false
	}
	return true
}

// Satisfies brancher interface.
func (l *leafExpressionBranch) getTailChunkID() int {
	if l.falseyReturnID == -1 {
		return -1
	}
	return l.falseyReturnID
}

func renderBranchComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	if dest.operatorExpression.Type == token.FLAG {
		renderFlagComparison(sb, dest, scriptName)
	} else if dest.operatorExpression.Type == token.VAR {
		renderVarComparison(sb, dest, scriptName)
	}
}

func renderFlagComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	if dest.operatorExpression.ComparisonValue == token.TRUE {
		sb.WriteString(fmt.Sprintf("\tgoto_if_set %s, %s_%d\n", dest.operatorExpression.Operand, scriptName, dest.id))
	} else {
		sb.WriteString(fmt.Sprintf("\tgoto_if_unset %s, %s_%d\n", dest.operatorExpression.Operand, scriptName, dest.id))
	}
}

func renderVarComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	sb.WriteString(fmt.Sprintf("\tcompare %s, %s\n", dest.operatorExpression.Operand, dest.operatorExpression.ComparisonValue))
	switch dest.operatorExpression.Operator {
	case token.EQ:
		sb.WriteString(fmt.Sprintf("\tgoto_if_eq %s_%d\n", scriptName, dest.id))
	case token.NEQ:
		sb.WriteString(fmt.Sprintf("\tgoto_if_ne %s_%d\n", scriptName, dest.id))
	case token.LT:
		sb.WriteString(fmt.Sprintf("\tgoto_if_lt %s_%d\n", scriptName, dest.id))
	case token.LTE:
		sb.WriteString(fmt.Sprintf("\tgoto_if_le %s_%d\n", scriptName, dest.id))
	case token.GT:
		sb.WriteString(fmt.Sprintf("\tgoto_if_gt %s_%d\n", scriptName, dest.id))
	case token.GTE:
		sb.WriteString(fmt.Sprintf("\tgoto_if_ge %s_%d\n", scriptName, dest.id))
	}
}
