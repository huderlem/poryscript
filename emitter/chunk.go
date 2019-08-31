package emitter

import "github.com/huderlem/poryscript/ast"

// Represents a single chunk of script output. Each chunk has an associated label in
// the emitted bytecode output.
type chunk struct {
	id             int
	returnID       int
	statements     []ast.Statement
	branchBehavior brancher
}

func (c *chunk) splitChunkForBranch(statementIndex int, chunkCounter *int, remainingChunks []chunk) ([]chunk, int) {
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

func (c *chunk) createPostLogicChunk(id int, lastStatementIndex int) chunk {
	newChunk := chunk{
		id:         id,
		returnID:   c.returnID,
		statements: c.statements[lastStatementIndex+1:],
	}
	return newChunk
}
