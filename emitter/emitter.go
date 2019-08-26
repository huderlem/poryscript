package emitter

import (
	"fmt"
	"strings"

	"github.com/huderlem/poryscript/ast"
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
	for i, stmt := range e.program.TopLevelStatements {
		if i > 0 {
			sb.WriteString("\n")
		}

		scriptStmt, ok := stmt.(*ast.ScriptStatement)
		if !ok {
			fmt.Printf("Could not emit top-level statement because it is not a Script statement")
			return ""
		}

		emitted := emitScriptStatement(scriptStmt)
		if emitted == "" {
			return ""
		}

		sb.WriteString(emitted)
	}
	return sb.String()
}

func emitScriptStatement(scriptStmt *ast.ScriptStatement) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s::\n", scriptStmt.Name.Value))
	for _, stmt := range scriptStmt.Body.Statements {
		commandStmt, ok := stmt.(*ast.CommandStatement)
		if !ok {
			fmt.Printf("Could not emit statement because it is not a Command statement")
			return ""
		}

		sb.WriteString(fmt.Sprintf("\t%s", commandStmt.Name.Value))
		if len(commandStmt.Args) > 0 {
			sb.WriteString(fmt.Sprintf(" %s", strings.Join(commandStmt.Args, ", ")))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
