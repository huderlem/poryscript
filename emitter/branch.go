package emitter

import (
	"fmt"
	"strings"

	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/genconfig"
	"github.com/huderlem/poryscript/token"
	"github.com/huderlem/poryscript/types"
)

// Interface that manages chunk branching behavior.
type brancher interface {
	renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int), gen types.Gen) bool
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
func (j *jump) renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int), gen types.Gen) bool {
	if j.destChunkID != nextChunkID {
		registerJumpChunk(j.destChunkID)
		sb.WriteString(fmt.Sprintf("\t%s %s_%d\n", genconfig.GotoCommands[gen], scriptName, j.destChunkID))
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
func (bc *breakContext) renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int), gen types.Gen) bool {
	if bc.destChunkID == -1 {
		sb.WriteString(fmt.Sprintf("\t%s\n", genconfig.ReturnCommands[gen]))
		return false
	} else if bc.destChunkID != nextChunkID {
		registerJumpChunk(bc.destChunkID)
		sb.WriteString(fmt.Sprintf("\t%s %s_%d\n", genconfig.GotoCommands[gen], scriptName, bc.destChunkID))
		return false
	}
	return true
}

// Satisfies brancher interface.
func (bc *breakContext) getTailChunkID() int {
	return bc.destChunkID
}

// Represents a leaf expression of a compound boolean expression.
type leafExpressionBranch struct {
	truthyDest     *conditionDestination
	falseyReturnID int
}

// Satisfies brancher interface.
func (l *leafExpressionBranch) renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int), gen types.Gen) bool {
	registerJumpChunk(l.truthyDest.id)
	renderBranchComparison(sb, l.truthyDest, scriptName, gen)
	if l.falseyReturnID == -1 {
		sb.WriteString(fmt.Sprintf("\t%s\n", genconfig.ReturnCommands[gen]))
		return false
	} else if l.falseyReturnID != nextChunkID {
		registerJumpChunk(l.falseyReturnID)
		sb.WriteString(fmt.Sprintf("\t%s %s_%d\n", genconfig.GotoCommands[gen], scriptName, l.falseyReturnID))
		return false
	}
	return true
}

// Satisfies brancher interface.
func (l *leafExpressionBranch) getTailChunkID() int {
	return l.falseyReturnID
}

type switchCaseBranch struct {
	comparisonValue string
	destChunkID     int
}

// Represents the a switch statement branch behavior.
type switchBranch struct {
	operator    string
	operand     string
	cases       []*switchCaseBranch
	defaultCase *switchCaseBranch
	destChunkID int
}

// Satisfies brancher interface.
func (s *switchBranch) renderBranchConditions(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int), gen types.Gen) bool {
	switch gen {
	case types.GEN2:
		s.renderGen2SwitchCases(sb, scriptName, nextChunkID, registerJumpChunk)
	case types.GEN3:
		s.renderGen3SwitchCases(sb, scriptName, nextChunkID, registerJumpChunk)
	}

	if s.defaultCase != nil {
		if s.defaultCase.destChunkID != nextChunkID {
			registerJumpChunk(s.defaultCase.destChunkID)
			sb.WriteString(fmt.Sprintf("\t%s %s_%d\n", genconfig.GotoCommands[gen], scriptName, s.defaultCase.destChunkID))
			return false
		}
	} else if s.destChunkID != nextChunkID {
		if s.destChunkID == -1 {
			sb.WriteString(fmt.Sprintf("\t%s\n", genconfig.ReturnCommands[gen]))
		} else {
			registerJumpChunk(s.destChunkID)
			sb.WriteString(fmt.Sprintf("\t%s %s_%d\n", genconfig.GotoCommands[gen], scriptName, s.destChunkID))
		}
		return false
	}

	return true
}

func (s *switchBranch) renderGen2SwitchCases(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int)) {
	sb.WriteString(fmt.Sprintf("\t%s %s\n", s.operator, s.operand))
	for _, switchCase := range s.cases {
		registerJumpChunk(switchCase.destChunkID)
		sb.WriteString(fmt.Sprintf("\tifequal %s, %s_%d\n", switchCase.comparisonValue, scriptName, switchCase.destChunkID))
	}
}

func (s *switchBranch) renderGen3SwitchCases(sb *strings.Builder, scriptName string, nextChunkID int, registerJumpChunk func(int)) {
	sb.WriteString(fmt.Sprintf("\tswitch %s\n", s.operand))
	for _, switchCase := range s.cases {
		registerJumpChunk(switchCase.destChunkID)
		sb.WriteString(fmt.Sprintf("\tcase %s, %s_%d\n", switchCase.comparisonValue, scriptName, switchCase.destChunkID))
	}
}

// Satisfies brancher interface.
func (s *switchBranch) getTailChunkID() int {
	if s.defaultCase != nil {
		return s.defaultCase.destChunkID
	}
	return s.destChunkID
}

func renderBranchComparison(sb *strings.Builder, dest *conditionDestination, scriptName string, gen types.Gen) {
	switch gen {
	case types.GEN2:
		renderGen2BranchComparison(sb, dest, scriptName)
	case types.GEN3:
		renderGen3BranchComparison(sb, dest, scriptName)
	}
}

func renderGen2BranchComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	sb.WriteString(fmt.Sprintf("\t%s %s\n", dest.operatorExpression.Token.Literal, dest.operatorExpression.Operand))
	switch dest.operatorExpression.Operator {
	case token.EQ:
		if dest.operatorExpression.ComparisonValue == token.TRUE {
			sb.WriteString(fmt.Sprintf("\tiftrue %s_%d\n", scriptName, dest.id))
		} else if dest.operatorExpression.ComparisonValue == token.FALSE {
			sb.WriteString(fmt.Sprintf("\tiffalse %s_%d\n", scriptName, dest.id))
		} else {
			sb.WriteString(fmt.Sprintf("\tifequal %s, %s_%d\n", dest.operatorExpression.ComparisonValue, scriptName, dest.id))
		}
	case token.NEQ:
		if dest.operatorExpression.ComparisonValue == token.TRUE {
			sb.WriteString(fmt.Sprintf("\tiffalse %s_%d\n", scriptName, dest.id))
		} else if dest.operatorExpression.ComparisonValue == token.FALSE {
			sb.WriteString(fmt.Sprintf("\tiftrue %s_%d\n", scriptName, dest.id))
		} else {
			sb.WriteString(fmt.Sprintf("\tifnotequal %s, %s_%d\n", dest.operatorExpression.ComparisonValue, scriptName, dest.id))
		}
	case token.LT:
		sb.WriteString(fmt.Sprintf("\tifless %s, %s_%d\n", dest.operatorExpression.ComparisonValue, scriptName, dest.id))
	case token.GT:
		sb.WriteString(fmt.Sprintf("\tifgreater %s, %s_%d\n", dest.operatorExpression.ComparisonValue, scriptName, dest.id))
	case token.LTE:
		sb.WriteString(fmt.Sprintf("\tifless (%s) + 1, %s_%d\n", dest.operatorExpression.ComparisonValue, scriptName, dest.id))
	case token.GTE:
		sb.WriteString(fmt.Sprintf("\tifgreater (%s) - 1, %s_%d\n", dest.operatorExpression.ComparisonValue, scriptName, dest.id))
	}
}

func renderGen3BranchComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	switch dest.operatorExpression.Token.Type {
	case token.FLAG:
		renderFlagComparison(sb, dest, scriptName)
	case token.VAR:
		renderVarComparison(sb, dest, scriptName)
	case token.DEFEATED:
		renderDefeatedComparison(sb, dest, scriptName)
	}
}

func renderFlagComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	if (dest.operatorExpression.Operator == token.EQ && dest.operatorExpression.ComparisonValue == token.TRUE) ||
		(dest.operatorExpression.Operator == token.NEQ && dest.operatorExpression.ComparisonValue == token.FALSE) {
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

func renderDefeatedComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	sb.WriteString(fmt.Sprintf("\tchecktrainerflag %s\n", dest.operatorExpression.Operand))
	if (dest.operatorExpression.Operator == token.EQ && dest.operatorExpression.ComparisonValue == token.TRUE) ||
		(dest.operatorExpression.Operator == token.NEQ && dest.operatorExpression.ComparisonValue == token.FALSE) {
		sb.WriteString(fmt.Sprintf("\tgoto_if 1, %s_%d\n", scriptName, dest.id))
	} else {
		sb.WriteString(fmt.Sprintf("\tgoto_if 0, %s_%d\n", scriptName, dest.id))
	}
}
