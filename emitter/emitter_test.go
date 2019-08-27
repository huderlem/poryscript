package emitter

import (
	"testing"

	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/parser"
)

func TestEmit(t *testing.T) {
	input := `
script MyScript {
	lock waitstate
	special(DoThingZhuLi)
	message("Hi\n"
			"I'm Marcus$")
}

script MyScript2 {
	lock waitstate
	bufferitemname(0, VAR_BUG_CONTEST_PRIZE)
}

raw RawTest1 ` + "`" + `
	step_end
` + "`" + `

raw_global RawTest2 ` + "`" + `
	stuff
	morestuff
` + "`" + `
`
	expected := `MyScript::
	lock
	waitstate
	special DoThingZhuLi
	message Text_0

MyScript2::
	lock
	waitstate
	bufferitemname 0, VAR_BUG_CONTEST_PRIZE

RawTest1:
	step_end

RawTest2::
	stuff
	morestuff

Text_0:
	.string "Hi\n"
	.string "I'm Marcus$"
`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	e := New(program)
	result := e.Emit()
	if result != expected {
		t.Errorf("Mismatching emit -- Expected=%q, Got=%q", expected, result)
	}
}
