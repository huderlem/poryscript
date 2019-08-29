package emitter

import (
	"testing"

	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/parser"
)

func TestEmit(t *testing.T) {
	input := `
script Route29_EventScript_WaitingMan {
	lock
	faceplayer
	# Display message based on time of day.
	gettime
	if (var(VAR_0x8002) == TIME_NIGHT) {
		msgbox("I'm waiting for POKéMON that appear\n"
				"only in the morning.$")
	} else {
		msgbox("I'm waiting for POKéMON that appear\n"
				"only at night.$")
	}
	# Wait for morning.
	while (var(VAR_0x8002) == TIME_NIGHT) {
		advancetime(5)
		gettime
	}
	release
}

script Route29_EventScript_Dude {
	lock
	faceplayer
	if (flag(FLAG_LEARNED_TO_CATCH_POKEMON) == true) {
		msgbox(Route29_Text_PokemonInTheGrass)
	} elif (flag(FLAG_GAVE_MYSTERY_EGG_TO_ELM) == false) {
		msgbox(Route29_Text_PokemonInTheGrass)
	} else {
		msgbox("Huh? You want me to show you how\nto catch POKéMON?$", MSGBOX_YESNO)
		if (var(VAR_RESULT) == 0) {
			msgbox(Route29_Text_Dude_CatchingTutRejected)
		} else {
			# Teach the player how to catch.
			closemessage
			special(StartDudeTutorialBattle)
			waitstate
			lock
			msgbox("That's how you do it.\p"
					"If you weaken them first, POKéMON\n"
					"are easier to catch.$")
			setflag(FLAG_LEARNED_TO_CATCH_POKEMON)
		}
	}
	release
}

raw ` + "`" + `
Route29_Text_PokemonInTheGrass:
	.string "POKéMON hide in the grass.\n"
	.string "Who knows when they'll pop out…$"
` + "`" + `

raw ` + "`" + `
Route29_Text_Dude_CatchingTutRejected:
	.string "Oh.\n"
	.string "Fine, then.\p"
	.string "Anyway, if you want to catch\n"
	.string "POKéMON, you have to walk a lot.$"
` + "`"

	expected := `Route29_EventScript_WaitingMan::
	lock
	faceplayer
	gettime
	compare VAR_0x8002, TIME_NIGHT
	goto_if_eq Route29_EventScript_WaitingMan_2
	goto Route29_EventScript_WaitingMan_3

Route29_EventScript_WaitingMan_1:
	goto Route29_EventScript_WaitingMan_5

Route29_EventScript_WaitingMan_2:
	msgbox Route29_EventScript_WaitingMan_Text_0
	goto Route29_EventScript_WaitingMan_1

Route29_EventScript_WaitingMan_3:
	msgbox Route29_EventScript_WaitingMan_Text_1
	goto Route29_EventScript_WaitingMan_1

Route29_EventScript_WaitingMan_4:
	release
	return

Route29_EventScript_WaitingMan_5:
	compare VAR_0x8002, TIME_NIGHT
	goto_if_eq Route29_EventScript_WaitingMan_6
	goto Route29_EventScript_WaitingMan_4

Route29_EventScript_WaitingMan_6:
	advancetime 5
	gettime
	goto Route29_EventScript_WaitingMan_5


Route29_EventScript_Dude::
	lock
	faceplayer
	goto_if_set FLAG_LEARNED_TO_CATCH_POKEMON, Route29_EventScript_Dude_2
	goto_if_unset FLAG_GAVE_MYSTERY_EGG_TO_ELM, Route29_EventScript_Dude_3
	goto Route29_EventScript_Dude_4

Route29_EventScript_Dude_1:
	release
	return

Route29_EventScript_Dude_2:
	msgbox Route29_Text_PokemonInTheGrass
	goto Route29_EventScript_Dude_1

Route29_EventScript_Dude_3:
	msgbox Route29_Text_PokemonInTheGrass
	goto Route29_EventScript_Dude_1

Route29_EventScript_Dude_4:
	msgbox Route29_EventScript_Dude_Text_0, MSGBOX_YESNO
	compare VAR_RESULT, 0
	goto_if_eq Route29_EventScript_Dude_5
	goto Route29_EventScript_Dude_6

Route29_EventScript_Dude_5:
	msgbox Route29_Text_Dude_CatchingTutRejected
	goto Route29_EventScript_Dude_1

Route29_EventScript_Dude_6:
	closemessage
	special StartDudeTutorialBattle
	waitstate
	lock
	msgbox Route29_EventScript_Dude_Text_1
	setflag FLAG_LEARNED_TO_CATCH_POKEMON
	goto Route29_EventScript_Dude_1


Route29_Text_PokemonInTheGrass:
	.string "POKéMON hide in the grass.\n"
	.string "Who knows when they'll pop out…$"

Route29_Text_Dude_CatchingTutRejected:
	.string "Oh.\n"
	.string "Fine, then.\p"
	.string "Anyway, if you want to catch\n"
	.string "POKéMON, you have to walk a lot.$"

Route29_EventScript_WaitingMan_Text_0:
	.string "I'm waiting for POKéMON that appear\n"
	.string "only in the morning.$"

Route29_EventScript_WaitingMan_Text_1:
	.string "I'm waiting for POKéMON that appear\n"
	.string "only at night.$"

Route29_EventScript_Dude_Text_0:
	.string "Huh? You want me to show you how\nto catch POKéMON?$"

Route29_EventScript_Dude_Text_1:
	.string "That's how you do it.\p"
	.string "If you weaken them first, POKéMON\n"
	.string "are easier to catch.$"
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
