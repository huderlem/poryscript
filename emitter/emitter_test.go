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
}

script MyScript2 {
	lock waitstate
	bufferitemname(0, VAR_BUG_CONTEST_PRIZE)
}
`
	expected := `MyScript::
	lock
	waitstate
	special DoThingZhuLi

MyScript2::
	lock
	waitstate
	bufferitemname 0, VAR_BUG_CONTEST_PRIZE
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
