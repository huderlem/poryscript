package emitter

import (
	"fmt"
	"strings"

	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/genconfig"
	"github.com/huderlem/poryscript/types"
)

// Represents a single chunk of script output. Each chunk has an associated label in
// the emitted bytecode output.
type chunk struct {
	id               int
	returnID         int
	useEndTerminator bool
	statements       []ast.Statement
	branchBehavior   brancher
}

func (c *chunk) renderLabel(scriptName string, isGlobal bool, sb *strings.Builder) {
	if c.id == 0 {
		// Main script entrypoint label.
		if isGlobal {
			sb.WriteString(fmt.Sprintf("%s::\n", scriptName))
		} else {
			sb.WriteString(fmt.Sprintf("%s:\n", scriptName))
		}
	} else {
		sb.WriteString(fmt.Sprintf("%s_%d:\n", scriptName, c.id))
	}
}

func (c *chunk) renderStatements(sb *strings.Builder) error {
	// Render basic non-branching commands.
	for _, stmt := range c.statements {
		commandStmt, ok := stmt.(*ast.CommandStatement)
		if !ok {
			return fmt.Errorf("could not render chunk statement '%q' because it is not a command statement", stmt.TokenLiteral())
		}

		sb.WriteString(renderCommandStatement(commandStmt))
	}
	return nil
}

func (c *chunk) renderBranching(scriptName string, sb *strings.Builder, nextChunkID int, registerJumpChunk func(int), gen types.Gen) bool {
	if c.branchBehavior != nil {
		isFallThrough := c.branchBehavior.renderBranchConditions(sb, scriptName, nextChunkID, registerJumpChunk)
		return isFallThrough
	}

	// Handle natural return logic that wasn't covered by a branch behavior.
	if c.returnID == -1 {
		sb.WriteString(fmt.Sprintf("\t%s\n", c.getTerminatorCommand(gen)))
		return false
	} else if c.returnID != nextChunkID {
		registerJumpChunk(c.returnID)
		sb.WriteString(fmt.Sprintf("\tgoto %s_%d\n", scriptName, c.returnID))
		return false
	}

	// Fallthrough to next chunk.
	return true
}

func (c *chunk) getTerminatorCommand(gen types.Gen) string {
	if c.useEndTerminator {
		return genconfig.EndCommands[gen]
	}
	return genconfig.ReturnCommands[gen]
}

func (c *chunk) splitChunkForBranch(statementIndex int, chunkCounter *int, remainingChunks []*chunk) ([]*chunk, int) {
	var returnID int
	if c.isLastStatement(statementIndex) {
		// The statement is the last of the current chunk, so it
		// has the same return point as the current chunk.
		returnID = c.returnID
	} else {
		// The statement needs to return to a chunk of logic
		// that occurs directly after it. So, create a new Chunk for
		// that logic.
		*chunkCounter++
		newChunk := c.createPostLogicChunk(*chunkCounter, statementIndex)
		remainingChunks = append(remainingChunks, newChunk)
		returnID = newChunk.id
		c.returnID = newChunk.id
	}
	return remainingChunks, returnID
}

func (c *chunk) isLastStatement(statementIndex int) bool {
	return statementIndex == len(c.statements)-1
}

func (c *chunk) createPostLogicChunk(id int, lastStatementIndex int) *chunk {
	newChunk := &chunk{
		id:         id,
		returnID:   c.returnID,
		statements: c.statements[lastStatementIndex+1:],
	}
	return newChunk
}

func renderCommandStatement(commandStmt *ast.CommandStatement) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\t%s", commandStmt.Name.Value))
	if len(commandStmt.Args) > 0 {
		sb.WriteString(fmt.Sprintf(" %s", strings.Join(commandStmt.Args, ", ")))
	}
	sb.WriteString("\n")
	return sb.String()
}
