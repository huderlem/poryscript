package emitter

import (
	"fmt"
	"strings"

	"github.com/huderlem/poryscript/token"
)

// Interface that manages chunk branching behavior.
type brancher interface {
	renderBranchConditions(sb *strings.Builder, scriptName string)
	requiresTailJump() bool
}

// Helper types for keeping track of script chunk branching logic.
type conditionDestination struct {
	id              int
	compareType     token.Type
	operand         string
	operator        token.Type
	comparisonValue string
}

// Represents the initial jump to a loop chunk.
type loopStart struct {
	destChunkID int
}

// Satisfies brancher interface.
func (wsb *loopStart) renderBranchConditions(sb *strings.Builder, scriptName string) {
	sb.WriteString(fmt.Sprintf("\tgoto %s_%d\n", scriptName, wsb.destChunkID))
}

// Satisfies brancher interface.
func (wsb *loopStart) requiresTailJump() bool {
	return false
}

// Represents a while loop header, where its branching conditions occur.
type whileHeader struct {
	dest *conditionDestination
}

// Satisfies brancher interface.
func (wh *whileHeader) renderBranchConditions(sb *strings.Builder, scriptName string) {
	renderBranchComparison(sb, wh.dest, scriptName)
}

// Satisfies brancher interface.
func (wh *whileHeader) requiresTailJump() bool {
	return true
}

// Represents a do-while loop header, where its branching conditions occur.
type doWhileHeader struct {
	dest *conditionDestination
}

// Satisfies brancher interface.
func (dwh *doWhileHeader) renderBranchConditions(sb *strings.Builder, scriptName string) {
	renderBranchComparison(sb, dwh.dest, scriptName)
}

// Satisfies brancher interface.
func (dwh *doWhileHeader) requiresTailJump() bool {
	return true
}

// Represents an if statement header, where its branching conditions occur.
type ifHeader struct {
	consequence      *conditionDestination
	elifConsequences []*conditionDestination
	elseConsequence  *conditionDestination
}

// Satisfies brancher interface.
func (ih *ifHeader) renderBranchConditions(sb *strings.Builder, scriptName string) {
	renderBranchComparison(sb, ih.consequence, scriptName)
	for _, dest := range ih.elifConsequences {
		renderBranchComparison(sb, dest, scriptName)
	}
	if ih.elseConsequence != nil {
		sb.WriteString(fmt.Sprintf("\tgoto %s_%d\n", scriptName, ih.elseConsequence.id))
	}
}

// Satisfies brancher interface.
func (ih *ifHeader) requiresTailJump() bool {
	if ih.elseConsequence != nil {
		return false
	}
	return true
}
