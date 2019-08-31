package emitter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/huderlem/poryscript/ast"
	"github.com/huderlem/poryscript/token"
)

// Emitter is responsible for transforming a parsed Poryscript program into
// the target assembler bytecode script.
type Emitter struct {
	program *ast.Program
}

// New creates a new Poryscript program emitter.
func New(program *ast.Program) *Emitter {
	return &Emitter{
		program: program,
	}
}

// Emit the target assembler bytecode script.
func (e *Emitter) Emit() string {
	var sb strings.Builder
	i := 0
	for _, stmt := range e.program.TopLevelStatements {
		if i > 0 {
			sb.WriteString("\n")
		}

		scriptStmt, ok := stmt.(*ast.ScriptStatement)
		if ok {
			sb.WriteString(emitScriptStatement(scriptStmt))
			i++
			continue
		}

		rawStmt, ok := stmt.(*ast.RawStatement)
		if ok {
			sb.WriteString(emitRawStatement(rawStmt))
			i++
			continue
		}

		fmt.Printf("Could not emit top-level statement because it is not recognized: %q", stmt.TokenLiteral())
		return ""
	}

	for j, text := range e.program.Texts {
		if i+j > 0 {
			sb.WriteString("\n")
		}

		emitted := emitText(text)
		sb.WriteString(emitted)
	}
	return sb.String()
}

func emitScriptStatement(scriptStmt *ast.ScriptStatement) string {
	// The algorithm for emitting script statements is to split the scripts into
	// self-contained chunks that logically branch to one another. When branching logic
	// occurs, create a new chunk for any shared logic that follows the branching, as well
	// as new chunks for the destination of the branching logic. When creating and processing
	// new chunks, it's important to remember where the chunks should return to.
	chunkCounter := 0
	finalChunks := make(map[int]chunk)
	remainingChunks := []chunk{
		{id: chunkCounter, returnID: -1, statements: scriptStmt.Body.Statements[:]},
	}
	loopStatementReturnChunks := make(map[ast.Statement]int)
	loopStatementOriginChunks := make(map[ast.Statement]int)
	for len(remainingChunks) > 0 {
		ids := []int{}
		for _, c := range remainingChunks {
			ids = append(ids, c.id)
		}

		// Grab an unprocessed script chunk.
		curChunk := remainingChunks[0]
		remainingChunks = remainingChunks[1:]

		// Skip over basic command statements.
		i := 0
		shouldContinue := false
		for _, stmt := range curChunk.statements {
			commandStmt, ok := stmt.(*ast.CommandStatement)
			if !ok {
				break
			}
			// "end" and "return" are special control-flow commands that end execution of
			// the current logic scope. Therefore, we should not process any further into the
			// current chunk, and mark it as finalized.
			if commandStmt.Name.Value == "end" || commandStmt.Name.Value == "return" {
				completeChunk := chunk{id: curChunk.id, returnID: -1, statements: curChunk.statements[:i]}
				finalChunks[completeChunk.id] = completeChunk
				shouldContinue = true
				break
			}
			i++
		}
		if shouldContinue {
			continue
		}

		if i == len(curChunk.statements) {
			// Finalize a new chunk, if we reached the end of the statements.
			finalChunks[curChunk.id] = curChunk
			continue
		}

		// Create new chunks from if statement blocks.
		if stmt, ok := curChunk.statements[i].(*ast.IfStatement); ok {
			newRemainingChunks, ifBranch := createIfStatementChunks(stmt, i, &curChunk, remainingChunks, &chunkCounter)
			remainingChunks = newRemainingChunks
			completeChunk := chunk{
				id:             curChunk.id,
				returnID:       curChunk.returnID,
				statements:     curChunk.statements[:i],
				branchBehavior: ifBranch,
			}
			finalChunks[completeChunk.id] = completeChunk
		} else if stmt, ok := curChunk.statements[i].(*ast.WhileStatement); ok {
			newRemainingChunks, loopStart, returnID := createWhileStatementChunks(stmt, i, &curChunk, remainingChunks, &chunkCounter)
			remainingChunks = newRemainingChunks
			completeChunk := chunk{
				id:             curChunk.id,
				returnID:       curChunk.returnID,
				statements:     curChunk.statements[:i],
				branchBehavior: loopStart,
			}
			finalChunks[completeChunk.id] = completeChunk
			loopStatementReturnChunks[stmt] = returnID
			loopStatementOriginChunks[stmt] = loopStart.destChunkID
		} else if stmt, ok := curChunk.statements[i].(*ast.DoWhileStatement); ok {
			newRemainingChunks, loopStart, returnID := createDoWhileStatementChunks(stmt, i, &curChunk, remainingChunks, &chunkCounter)
			remainingChunks = newRemainingChunks
			completeChunk := chunk{
				id:             curChunk.id,
				returnID:       curChunk.returnID,
				statements:     curChunk.statements[:i],
				branchBehavior: loopStart,
			}
			finalChunks[completeChunk.id] = completeChunk
			loopStatementReturnChunks[stmt] = returnID
			loopStatementOriginChunks[stmt] = loopStart.destChunkID
		} else if stmt, ok := curChunk.statements[i].(*ast.BreakStatement); ok {
			destChunkID, ok := loopStatementReturnChunks[stmt.LoopStatment]
			if !ok {
				panic("Could not emit 'break' statement because its return point is unknown.")
			}
			completeChunk := chunk{
				id:             curChunk.id,
				returnID:       curChunk.returnID,
				statements:     curChunk.statements[:i],
				branchBehavior: &breakContext{destChunkID: destChunkID},
			}
			finalChunks[completeChunk.id] = completeChunk
		} else if stmt, ok := curChunk.statements[i].(*ast.ContinueStatement); ok {
			destChunkID, ok := loopStatementOriginChunks[stmt.LoopStatment]
			if !ok {
				panic("Could not emit 'continue' statement because its return point is unknown.")
			}
			completeChunk := chunk{
				id:             curChunk.id,
				returnID:       curChunk.returnID,
				statements:     curChunk.statements[:i],
				branchBehavior: &breakContext{destChunkID: destChunkID},
			}
			finalChunks[completeChunk.id] = completeChunk
		} else {
			completeChunk := chunk{id: curChunk.id, returnID: curChunk.returnID, statements: curChunk.statements[:i]}
			finalChunks[completeChunk.id] = completeChunk
		}
	}

	return renderChunks(finalChunks, scriptStmt.Name.Value)
}

func createConditionDestination(compareType token.Type, operand string, operator token.Type, comparisonValue string, destinationChunk int) *conditionDestination {
	return &conditionDestination{
		id:              destinationChunk,
		compareType:     compareType,
		operand:         operand,
		operator:        operator,
		comparisonValue: comparisonValue,
	}
}

func renderChunks(finalChunks map[int]chunk, scriptName string) string {
	// Get sorted list of final chunk ids.
	chunkIDs := make([]int, 0)
	for k := range finalChunks {
		chunkIDs = append(chunkIDs, k)
	}
	sort.Ints(chunkIDs)

	var sb strings.Builder
	for _, chunkID := range chunkIDs {
		chunk := finalChunks[chunkID]
		if chunk.id == 0 {
			// Main script entrypoint, so it gets a global label.
			sb.WriteString(fmt.Sprintf("%s::\n", scriptName))
		} else {
			sb.WriteString(fmt.Sprintf("%s_%d:\n", scriptName, chunk.id))
		}

		// Render basic non-branching commands.
		for _, stmt := range chunk.statements {
			commandStmt, ok := stmt.(*ast.CommandStatement)
			if !ok {
				fmt.Printf("Could not render chunk statement because it is not a command statement %q", stmt.TokenLiteral())
				return ""
			}

			sb.WriteString(renderCommandStatement(commandStmt))
		}

		// Render branching conditions.
		requiresTailJump := true
		if chunk.branchBehavior != nil {
			chunk.branchBehavior.renderBranchConditions(&sb, scriptName)
			requiresTailJump = chunk.branchBehavior.requiresTailJump()
		}

		// Sometimes, a tail jump/return isn't needed.  For example, a chunk that ends in an "else"
		// branch will always naturally end with a "goto" bytecode command.
		if requiresTailJump {
			if chunk.returnID == -1 {
				sb.WriteString("\treturn\n")
			} else {
				sb.WriteString(fmt.Sprintf("\tgoto %s_%d\n", scriptName, chunk.returnID))
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func renderBranchComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	if dest.compareType == token.FLAG {
		renderFlagComparison(sb, dest, scriptName)
	} else if dest.compareType == token.VAR {
		renderVarComparison(sb, dest, scriptName)
	}
}

func renderFlagComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	if dest.comparisonValue == token.TRUE {
		sb.WriteString(fmt.Sprintf("\tgoto_if_set %s, %s_%d\n", dest.operand, scriptName, dest.id))
	} else {
		sb.WriteString(fmt.Sprintf("\tgoto_if_unset %s, %s_%d\n", dest.operand, scriptName, dest.id))
	}
}

func renderVarComparison(sb *strings.Builder, dest *conditionDestination, scriptName string) {
	sb.WriteString(fmt.Sprintf("\tcompare %s, %s\n", dest.operand, dest.comparisonValue))
	switch dest.operator {
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

func createIfStatementChunks(stmt *ast.IfStatement, i int, curChunk *chunk, remainingChunks []chunk, chunkCounter *int) ([]chunk, *ifHeader) {
	remainingChunks, returnID := curChunk.splitChunkForBranch(i, chunkCounter, remainingChunks)

	*chunkCounter++
	consequenceChunk := chunk{
		id:         *chunkCounter,
		returnID:   returnID,
		statements: stmt.Consequence.Body.Statements,
	}
	remainingChunks = append(remainingChunks, consequenceChunk)
	branch := &ifHeader{}
	branch.consequence = createConditionDestination(stmt.Consequence.Type, stmt.Consequence.Operand, stmt.Consequence.Operator, stmt.Consequence.ComparisonValue, consequenceChunk.id)

	branch.elifConsequences = []*conditionDestination{}
	for _, elifStmt := range stmt.ElifConsequences {
		*chunkCounter++
		elifChunk := chunk{
			id:         *chunkCounter,
			returnID:   returnID,
			statements: elifStmt.Body.Statements,
		}
		remainingChunks = append(remainingChunks, elifChunk)
		branch.elifConsequences = append(branch.elifConsequences,
			createConditionDestination(elifStmt.Type, elifStmt.Operand, elifStmt.Operator, elifStmt.ComparisonValue, elifChunk.id))
	}

	if stmt.ElseConsequence != nil {
		*chunkCounter++
		elseChunk := chunk{
			id:         *chunkCounter,
			returnID:   returnID,
			statements: stmt.ElseConsequence.Statements,
		}
		remainingChunks = append(remainingChunks, elseChunk)
		branch.elseConsequence = &conditionDestination{id: elseChunk.id}
	}

	return remainingChunks, branch
}

func createWhileStatementChunks(stmt *ast.WhileStatement, i int, curChunk *chunk, remainingChunks []chunk, chunkCounter *int) ([]chunk, *loopStart, int) {
	remainingChunks, returnID := curChunk.splitChunkForBranch(i, chunkCounter, remainingChunks)

	*chunkCounter++
	headerChunk := chunk{
		id:         *chunkCounter,
		returnID:   returnID,
		statements: []ast.Statement{},
	}

	*chunkCounter++
	consequenceChunk := chunk{
		id:         *chunkCounter,
		returnID:   headerChunk.id,
		statements: stmt.Consequence.Body.Statements,
	}
	dest := createConditionDestination(stmt.Consequence.Type, stmt.Consequence.Operand, stmt.Consequence.Operator, stmt.Consequence.ComparisonValue, consequenceChunk.id)
	headerChunk.branchBehavior = &whileHeader{dest: dest}
	remainingChunks = append(remainingChunks, consequenceChunk)
	remainingChunks = append(remainingChunks, headerChunk)

	return remainingChunks, &loopStart{destChunkID: headerChunk.id}, returnID
}

func createDoWhileStatementChunks(stmt *ast.DoWhileStatement, i int, curChunk *chunk, remainingChunks []chunk, chunkCounter *int) ([]chunk, *loopStart, int) {
	remainingChunks, returnID := curChunk.splitChunkForBranch(i, chunkCounter, remainingChunks)

	*chunkCounter++
	headerChunk := chunk{
		id:         *chunkCounter,
		returnID:   returnID,
		statements: []ast.Statement{},
	}

	*chunkCounter++
	consequenceChunk := chunk{
		id:         *chunkCounter,
		returnID:   headerChunk.id,
		statements: stmt.Consequence.Body.Statements,
	}
	dest := createConditionDestination(stmt.Consequence.Type, stmt.Consequence.Operand, stmt.Consequence.Operator, stmt.Consequence.ComparisonValue, consequenceChunk.id)
	headerChunk.branchBehavior = &whileHeader{dest: dest}
	remainingChunks = append(remainingChunks, consequenceChunk)
	remainingChunks = append(remainingChunks, headerChunk)

	return remainingChunks, &loopStart{destChunkID: consequenceChunk.id}, returnID
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

func emitText(text ast.Text) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s:\n", text.Name))
	lines := strings.Split(text.Value, "\n")
	for _, line := range lines {
		sb.WriteString(fmt.Sprintf("\t.string \"%s\"\n", line))
	}
	return sb.String()
}

func emitRawStatement(rawStmt *ast.RawStatement) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s\n", rawStmt.Value))
	return sb.String()
}
