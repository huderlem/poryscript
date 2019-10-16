package emitter

import (
	"testing"

	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/parser"
)

func TestEmit1(t *testing.T) {
	input := `
script Route29_EventScript_WaitingMan {
	lock
	faceplayer
	# Display message based on time of day.
	gettime
	if (var(VAR_0x8002) == TIME_NIGHT) {
		msgbox("I'm waiting for POKéMON that appear\n"
				"only in the morning.")
	} else {
		msgbox("I'm waiting for POKéMON that appear\n"
				"only at night.")
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
	if (flag(FLAG_LEARNED_TO_CATCH_POKEMON)) {
		msgbox(Route29_Text_PokemonInTheGrass)
	} elif (!flag(FLAG_GAVE_MYSTERY_EGG_TO_ELM)) {
		msgbox(Route29_Text_PokemonInTheGrass)
	} else {
		msgbox("Huh? You want me to show you how\nto catch POKéMON?$", MSGBOX_YESNO)
		if (!var(VAR_RESULT)) {
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

	expectedUnoptimized := `Route29_EventScript_WaitingMan::
	lock
	faceplayer
	gettime
	goto Route29_EventScript_WaitingMan_4

Route29_EventScript_WaitingMan_1:
	goto Route29_EventScript_WaitingMan_6

Route29_EventScript_WaitingMan_2:
	msgbox Route29_EventScript_WaitingMan_Text_0
	goto Route29_EventScript_WaitingMan_1

Route29_EventScript_WaitingMan_3:
	msgbox Route29_EventScript_WaitingMan_Text_1
	goto Route29_EventScript_WaitingMan_1

Route29_EventScript_WaitingMan_4:
	compare VAR_0x8002, TIME_NIGHT
	goto_if_eq Route29_EventScript_WaitingMan_2
	goto Route29_EventScript_WaitingMan_3

Route29_EventScript_WaitingMan_5:
	release
	return

Route29_EventScript_WaitingMan_6:
	goto Route29_EventScript_WaitingMan_8

Route29_EventScript_WaitingMan_7:
	advancetime 5
	gettime
	goto Route29_EventScript_WaitingMan_6

Route29_EventScript_WaitingMan_8:
	compare VAR_0x8002, TIME_NIGHT
	goto_if_eq Route29_EventScript_WaitingMan_7
	goto Route29_EventScript_WaitingMan_5


Route29_EventScript_Dude::
	lock
	faceplayer
	goto Route29_EventScript_Dude_6

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
	goto Route29_EventScript_Dude_9

Route29_EventScript_Dude_5:
	goto_if_unset FLAG_GAVE_MYSTERY_EGG_TO_ELM, Route29_EventScript_Dude_3
	goto Route29_EventScript_Dude_4

Route29_EventScript_Dude_6:
	goto_if_set FLAG_LEARNED_TO_CATCH_POKEMON, Route29_EventScript_Dude_2
	goto Route29_EventScript_Dude_5

Route29_EventScript_Dude_7:
	msgbox Route29_Text_Dude_CatchingTutRejected
	goto Route29_EventScript_Dude_1

Route29_EventScript_Dude_8:
	closemessage
	special StartDudeTutorialBattle
	waitstate
	lock
	msgbox Route29_EventScript_Dude_Text_1
	setflag FLAG_LEARNED_TO_CATCH_POKEMON
	goto Route29_EventScript_Dude_1

Route29_EventScript_Dude_9:
	compare VAR_RESULT, 0
	goto_if_eq Route29_EventScript_Dude_7
	goto Route29_EventScript_Dude_8


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

	expectedOptimized := `Route29_EventScript_WaitingMan::
	lock
	faceplayer
	gettime
	compare VAR_0x8002, TIME_NIGHT
	goto_if_eq Route29_EventScript_WaitingMan_2
	msgbox Route29_EventScript_WaitingMan_Text_1
Route29_EventScript_WaitingMan_1:
Route29_EventScript_WaitingMan_6:
	compare VAR_0x8002, TIME_NIGHT
	goto_if_eq Route29_EventScript_WaitingMan_7
	release
	return

Route29_EventScript_WaitingMan_2:
	msgbox Route29_EventScript_WaitingMan_Text_0
	goto Route29_EventScript_WaitingMan_1

Route29_EventScript_WaitingMan_7:
	advancetime 5
	gettime
	goto Route29_EventScript_WaitingMan_6


Route29_EventScript_Dude::
	lock
	faceplayer
	goto_if_set FLAG_LEARNED_TO_CATCH_POKEMON, Route29_EventScript_Dude_2
	goto_if_unset FLAG_GAVE_MYSTERY_EGG_TO_ELM, Route29_EventScript_Dude_3
	msgbox Route29_EventScript_Dude_Text_0, MSGBOX_YESNO
	compare VAR_RESULT, 0
	goto_if_eq Route29_EventScript_Dude_7
	closemessage
	special StartDudeTutorialBattle
	waitstate
	lock
	msgbox Route29_EventScript_Dude_Text_1
	setflag FLAG_LEARNED_TO_CATCH_POKEMON
Route29_EventScript_Dude_1:
	release
	return

Route29_EventScript_Dude_2:
	msgbox Route29_Text_PokemonInTheGrass
	goto Route29_EventScript_Dude_1

Route29_EventScript_Dude_3:
	msgbox Route29_Text_PokemonInTheGrass
	goto Route29_EventScript_Dude_1

Route29_EventScript_Dude_7:
	msgbox Route29_Text_Dude_CatchingTutRejected
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
	p := parser.New(l, "")
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := New(program, false)
	result, _ := e.Emit()
	if result != expectedUnoptimized {
		t.Errorf("Mismatching unoptimized emit -- Expected=%q, Got=%q", expectedUnoptimized, result)
	}

	e = New(program, true)
	result, _ = e.Emit()
	if result != expectedOptimized {
		t.Errorf("Mismatching optimized emit -- Expected=%q, Got=%q", expectedOptimized, result)
	}
}

func TestEmitDoWhile(t *testing.T) {
	input := `
script Route29_EventScript_WaitingMan {
	lock
	faceplayer
	# Force player to answer "Yes" to NPC question.
	msgbox("Do you agree to the quest?", MSGBOX_YESNO)
	do {
		if (!flag(FLAG_1)) {
			msgbox("...How about now?", MSGBOX_YESNO)
		} else {
			special(OtherThing)
		}
	} while (var(VAR_RESULT))
	release
}`

	expectedUnoptimized := `Route29_EventScript_WaitingMan::
	lock
	faceplayer
	msgbox Route29_EventScript_WaitingMan_Text_0, MSGBOX_YESNO
	goto Route29_EventScript_WaitingMan_3

Route29_EventScript_WaitingMan_1:
	release
	return

Route29_EventScript_WaitingMan_2:
	goto Route29_EventScript_WaitingMan_4

Route29_EventScript_WaitingMan_3:
	goto Route29_EventScript_WaitingMan_7

Route29_EventScript_WaitingMan_4:
	compare VAR_RESULT, 0
	goto_if_ne Route29_EventScript_WaitingMan_3
	goto Route29_EventScript_WaitingMan_1

Route29_EventScript_WaitingMan_5:
	msgbox Route29_EventScript_WaitingMan_Text_1, MSGBOX_YESNO
	goto Route29_EventScript_WaitingMan_2

Route29_EventScript_WaitingMan_6:
	special OtherThing
	goto Route29_EventScript_WaitingMan_2

Route29_EventScript_WaitingMan_7:
	goto_if_unset FLAG_1, Route29_EventScript_WaitingMan_5
	goto Route29_EventScript_WaitingMan_6


Route29_EventScript_WaitingMan_Text_0:
	.string "Do you agree to the quest?$"

Route29_EventScript_WaitingMan_Text_1:
	.string "...How about now?$"
`

	expectedOptimized := `Route29_EventScript_WaitingMan::
	lock
	faceplayer
	msgbox Route29_EventScript_WaitingMan_Text_0, MSGBOX_YESNO
Route29_EventScript_WaitingMan_3:
	goto_if_unset FLAG_1, Route29_EventScript_WaitingMan_5
	special OtherThing
Route29_EventScript_WaitingMan_2:
	compare VAR_RESULT, 0
	goto_if_ne Route29_EventScript_WaitingMan_3
	release
	return

Route29_EventScript_WaitingMan_5:
	msgbox Route29_EventScript_WaitingMan_Text_1, MSGBOX_YESNO
	goto Route29_EventScript_WaitingMan_2


Route29_EventScript_WaitingMan_Text_0:
	.string "Do you agree to the quest?$"

Route29_EventScript_WaitingMan_Text_1:
	.string "...How about now?$"
`
	l := lexer.New(input)
	p := parser.New(l, "")
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := New(program, false)
	result, _ := e.Emit()
	if result != expectedUnoptimized {
		t.Errorf("Mismatching unoptimized emit -- Expected=%q, Got=%q", expectedUnoptimized, result)
	}

	e = New(program, true)
	result, _ = e.Emit()
	if result != expectedOptimized {
		t.Errorf("Mismatching optimized emit -- Expected=%q, Got=%q", expectedOptimized, result)
	}
}

func TestEmitBreak(t *testing.T) {
	input := `
script MyScript {
	while (var(VAR_1) < 5) {
		first
		do {
			if (flag(FLAG_1) == true) {
				stuff
				before
				break
			}
			last
		} while (flag(FLAG_2) == false)
		if (flag(FLAG_3) == true) {
			continue
		}
		lastinwhile
	}
	release
}	
`

	expectedUnoptimized := `MyScript::
	goto MyScript_2

MyScript_1:
	release
	return

MyScript_2:
	goto MyScript_4

MyScript_3:
	first
	goto MyScript_7

MyScript_4:
	compare VAR_1, 5
	goto_if_lt MyScript_3
	goto MyScript_1

MyScript_5:
	goto MyScript_11

MyScript_6:
	goto MyScript_8

MyScript_7:
	goto MyScript_14

MyScript_8:
	goto_if_unset FLAG_2, MyScript_7
	goto MyScript_5

MyScript_9:
	lastinwhile
	goto MyScript_2

MyScript_10:
	goto MyScript_2

MyScript_11:
	goto_if_set FLAG_3, MyScript_10
	goto MyScript_9

MyScript_12:
	last
	goto MyScript_6

MyScript_13:
	stuff
	before
	goto MyScript_5

MyScript_14:
	goto_if_set FLAG_1, MyScript_13
	goto MyScript_12

`
	expectedOptimized := `MyScript::
MyScript_2:
	compare VAR_1, 5
	goto_if_lt MyScript_3
	release
	return

MyScript_3:
	first
MyScript_7:
	goto_if_set FLAG_1, MyScript_13
	last
	goto_if_unset FLAG_2, MyScript_7
MyScript_5:
	goto_if_set FLAG_3, MyScript_10
	lastinwhile
	goto MyScript_2

MyScript_10:
	goto MyScript_2

MyScript_13:
	stuff
	before
	goto MyScript_5

`

	l := lexer.New(input)
	p := parser.New(l, "")
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := New(program, false)
	result, _ := e.Emit()
	if result != expectedUnoptimized {
		t.Errorf("Mismatching unoptimized emit -- Expected=%q, Got=%q", expectedUnoptimized, result)
	}

	e = New(program, true)
	result, _ = e.Emit()
	if result != expectedOptimized {
		t.Errorf("Mismatching optimized emit -- Expected=%q, Got=%q", expectedOptimized, result)
	}
}

func TestEmitCompoundBooleanExpressions(t *testing.T) {
	input := `
script MyScript {
	do {
		message()
		if (!flag(FLAG_3) || (var(VAR_44) > 3 && var(VAR_55) <= 5)) {
			hey
		}
		if (defeated(TRAINER_BLUE) || !defeated(TRAINER_RED) && (defeated(TRAINER_FOO) == true)) {
			baz(-24, 17)
		}
	} while ((flag(FLAG_1) == true || flag(FLAG_2)) && (var(VAR_1) == 2 || var(VAR_2) == 3))
	blah
}
`

	expectedUnoptimized := `MyScript::
	goto MyScript_3

MyScript_1:
	blah
	return

MyScript_2:
	goto MyScript_6

MyScript_3:
	message
	goto MyScript_14

MyScript_4:
	goto MyScript_9

MyScript_5:
	goto MyScript_7

MyScript_6:
	goto_if_set FLAG_1, MyScript_4
	goto MyScript_5

MyScript_7:
	goto_if_set FLAG_2, MyScript_4
	goto MyScript_1

MyScript_8:
	goto MyScript_10

MyScript_9:
	compare VAR_1, 2
	goto_if_eq MyScript_3
	goto MyScript_8

MyScript_10:
	compare VAR_2, 3
	goto_if_eq MyScript_3
	goto MyScript_1

MyScript_11:
	goto MyScript_20

MyScript_12:
	hey
	goto MyScript_11

MyScript_13:
	goto MyScript_16

MyScript_14:
	goto_if_unset FLAG_3, MyScript_12
	goto MyScript_13

MyScript_15:
	goto MyScript_17

MyScript_16:
	compare VAR_44, 3
	goto_if_gt MyScript_15
	goto MyScript_11

MyScript_17:
	compare VAR_55, 5
	goto_if_le MyScript_12
	goto MyScript_11

MyScript_18:
	baz -24, 17
	goto MyScript_2

MyScript_19:
	goto MyScript_22

MyScript_20:
	checktrainerflag TRAINER_BLUE
	goto_if 1, MyScript_18
	goto MyScript_19

MyScript_21:
	goto MyScript_23

MyScript_22:
	checktrainerflag TRAINER_RED
	goto_if 0, MyScript_21
	goto MyScript_2

MyScript_23:
	checktrainerflag TRAINER_FOO
	goto_if 1, MyScript_18
	goto MyScript_2

`

	expectedOptimized := `MyScript::
MyScript_3:
	message
	goto_if_unset FLAG_3, MyScript_12
	compare VAR_44, 3
	goto_if_gt MyScript_15
MyScript_11:
	checktrainerflag TRAINER_BLUE
	goto_if 1, MyScript_18
	checktrainerflag TRAINER_RED
	goto_if 0, MyScript_21
MyScript_2:
	goto_if_set FLAG_1, MyScript_4
	goto_if_set FLAG_2, MyScript_4
MyScript_1:
	blah
	return

MyScript_4:
	compare VAR_1, 2
	goto_if_eq MyScript_3
	compare VAR_2, 3
	goto_if_eq MyScript_3
	goto MyScript_1

MyScript_12:
	hey
	goto MyScript_11

MyScript_15:
	compare VAR_55, 5
	goto_if_le MyScript_12
	goto MyScript_11

MyScript_18:
	baz -24, 17
	goto MyScript_2

MyScript_21:
	checktrainerflag TRAINER_FOO
	goto_if 1, MyScript_18
	goto MyScript_2

`
	l := lexer.New(input)
	p := parser.New(l, "")
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := New(program, false)
	result, _ := e.Emit()
	if result != expectedUnoptimized {
		t.Errorf("Mismatching unoptimized emit -- Expected=%q, Got=%q", expectedUnoptimized, result)
	}

	e = New(program, true)
	result, _ = e.Emit()
	if result != expectedOptimized {
		t.Errorf("Mismatching optimized emit -- Expected=%q, Got=%q", expectedOptimized, result)
	}
}

func TestSwitchStatements(t *testing.T) {
	input := `
script MyScript {
	while (var(VAR_2) == 2) {
		switch (var(VAR_1)) {
			default: messagedefault()
			case 0:
				if (flag(FLAG_1) == true) {
					delay(5)
				}
				message0()
			case 72:
				switch (var(VAR_5)) {
					case 434:
					case 2:
						secondfirst()
						if (((!flag(FLAG_TEMP_1)))) {
							break
						}
						foo()
					default:
						seconddefault()
						continue
				}
			case 1:
			case 2:
				message1()
				messagedefault()
		}
		afterswitch()
	}
	release
}`

	expectedUnoptimized := `MyScript::
	goto MyScript_2

MyScript_1:
	release
	return

MyScript_2:
	goto MyScript_4

MyScript_3:
	goto MyScript_6

MyScript_4:
	compare VAR_2, 2
	goto_if_eq MyScript_3
	goto MyScript_1

MyScript_5:
	afterswitch
	goto MyScript_2

MyScript_6:
	switch VAR_1
	case 0, MyScript_7
	case 72, MyScript_8
	case 1, MyScript_9
	case 2, MyScript_9
	goto MyScript_10

MyScript_7:
	goto MyScript_13

MyScript_8:
	goto MyScript_14

MyScript_9:
	message1
	messagedefault
	goto MyScript_5

MyScript_10:
	messagedefault
	goto MyScript_5

MyScript_11:
	message0
	goto MyScript_5

MyScript_12:
	delay 5
	goto MyScript_11

MyScript_13:
	goto_if_set FLAG_1, MyScript_12
	goto MyScript_11

MyScript_14:
	switch VAR_5
	case 434, MyScript_15
	case 2, MyScript_15
	goto MyScript_16

MyScript_15:
	secondfirst
	goto MyScript_19

MyScript_16:
	seconddefault
	goto MyScript_2

MyScript_17:
	foo
	goto MyScript_5

MyScript_18:
	goto MyScript_5

MyScript_19:
	goto_if_unset FLAG_TEMP_1, MyScript_18
	goto MyScript_17

`

	expectedOptimized := `MyScript::
MyScript_2:
	compare VAR_2, 2
	goto_if_eq MyScript_3
	release
	return

MyScript_3:
	switch VAR_1
	case 0, MyScript_7
	case 72, MyScript_8
	case 1, MyScript_9
	case 2, MyScript_9
	messagedefault
MyScript_5:
	afterswitch
	goto MyScript_2

MyScript_7:
	goto_if_set FLAG_1, MyScript_12
MyScript_11:
	message0
	goto MyScript_5

MyScript_8:
	switch VAR_5
	case 434, MyScript_15
	case 2, MyScript_15
	seconddefault
	goto MyScript_2

MyScript_9:
	message1
	messagedefault
	goto MyScript_5

MyScript_12:
	delay 5
	goto MyScript_11

MyScript_15:
	secondfirst
	goto_if_unset FLAG_TEMP_1, MyScript_18
	foo
	goto MyScript_5

MyScript_18:
	goto MyScript_5

`
	l := lexer.New(input)
	p := parser.New(l, "")
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := New(program, false)
	result, _ := e.Emit()
	if result != expectedUnoptimized {
		t.Errorf("Mismatching unoptimized emit -- Expected=%q, Got=%q", expectedUnoptimized, result)
	}

	e = New(program, true)
	result, _ = e.Emit()
	if result != expectedOptimized {
		t.Errorf("Mismatching optimized emit -- Expected=%q, Got=%q", expectedOptimized, result)
	}
}

func TestEmitTextStatements(t *testing.T) {
	input := `
script TextFormatLineBreaks {
    msgbox(format("Long cat is loooong once again\p"
                  "Very very loooong and we need to have"
                  "multiple lines to fit its loooongness"))
}

script MyScript {
	msgbox("Hello")
}

text MyText {
	"Hi, I'm first$"
}

text MyText2 { "Bye!" }
`

	expectedUnoptimized := `TextFormatLineBreaks::
	msgbox TextFormatLineBreaks_Text_0
	return


MyScript::
	msgbox MyScript_Text_0
	return


TextFormatLineBreaks_Text_0:
	.string "Long cat is loooong once again\p"
	.string "Very very loooong and we need to have\n"
	.string "multiple lines to fit its loooongness$"

MyScript_Text_0:
	.string "Hello$"

MyText::
	.string "Hi, I'm first$"

MyText2::
	.string "Bye!$"
`

	expectedOptimized := `TextFormatLineBreaks::
	msgbox TextFormatLineBreaks_Text_0
	return


MyScript::
	msgbox MyScript_Text_0
	return


TextFormatLineBreaks_Text_0:
	.string "Long cat is loooong once again\p"
	.string "Very very loooong and we need to have\n"
	.string "multiple lines to fit its loooongness$"

MyScript_Text_0:
	.string "Hello$"

MyText::
	.string "Hi, I'm first$"

MyText2::
	.string "Bye!$"
`
	l := lexer.New(input)
	p := parser.New(l, "../font_widths.json")
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := New(program, false)
	result, _ := e.Emit()
	if result != expectedUnoptimized {
		t.Errorf("Mismatching unoptimized emit -- Expected=%q, Got=%q", expectedUnoptimized, result)
	}

	e = New(program, true)
	result, _ = e.Emit()
	if result != expectedOptimized {
		t.Errorf("Mismatching optimized emit -- Expected=%q, Got=%q", expectedOptimized, result)
	}
}

func TestEmitEndTerminators(t *testing.T) {
	input := `
script ScripText {
    lock
    if (var(VAR_FOO)) {
    	end
    }
    release
    end
}`

	expectedUnoptimized := `ScripText::
	lock
	goto ScripText_3

ScripText_1:
	release
	end

ScripText_2:
	end

ScripText_3:
	compare VAR_FOO, 0
	goto_if_ne ScripText_2
	goto ScripText_1

`

	expectedOptimized := `ScripText::
	lock
	compare VAR_FOO, 0
	goto_if_ne ScripText_2
	release
	end

ScripText_2:
	end

`
	l := lexer.New(input)
	p := parser.New(l, "../font_widths.json")
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := New(program, false)
	result, _ := e.Emit()
	if result != expectedUnoptimized {
		t.Errorf("Mismatching unoptimized emit -- Expected=%q, Got=%q", expectedUnoptimized, result)
	}

	e = New(program, true)
	result, _ = e.Emit()
	if result != expectedOptimized {
		t.Errorf("Mismatching optimized emit -- Expected=%q, Got=%q", expectedOptimized, result)
	}
}

func TestEmitMovementStatements(t *testing.T) {
	input := `
script ScriptWithMovement {
	lock
	msgbox("Let's go for a walk.")
	applymovement(2, MovementWalk)
	waitmovement(0)
	applymovement(2, MovementWalk2)
	waitmovement(0)
	applymovement(2, MovementWalk3)
	waitmovement(0)
	release
}

movement MovementWalk {
	walk_left * 2
	walk_up * 3
	walk_down
	run_down
	face_left
	step_end
}

movement MovementWalk2 {
	run_left
	run_right * 2
}

movement MovementWalk3 {
	run_left * 2
	step_end
	run_right * 5
}
`

	expectedUnoptimized := `ScriptWithMovement::
	lock
	msgbox ScriptWithMovement_Text_0
	applymovement 2, MovementWalk
	waitmovement 0
	applymovement 2, MovementWalk2
	waitmovement 0
	applymovement 2, MovementWalk3
	waitmovement 0
	release
	return


MovementWalk:
	walk_left
	walk_left
	walk_up
	walk_up
	walk_up
	walk_down
	run_down
	face_left
	step_end

MovementWalk2:
	run_left
	run_right
	run_right
	step_end

MovementWalk3:
	run_left
	run_left
	step_end

ScriptWithMovement_Text_0:
	.string "Let's go for a walk.$"
`

	expectedOptimized := `ScriptWithMovement::
	lock
	msgbox ScriptWithMovement_Text_0
	applymovement 2, MovementWalk
	waitmovement 0
	applymovement 2, MovementWalk2
	waitmovement 0
	applymovement 2, MovementWalk3
	waitmovement 0
	release
	return


MovementWalk:
	walk_left
	walk_left
	walk_up
	walk_up
	walk_up
	walk_down
	run_down
	face_left
	step_end

MovementWalk2:
	run_left
	run_right
	run_right
	step_end

MovementWalk3:
	run_left
	run_left
	step_end

ScriptWithMovement_Text_0:
	.string "Let's go for a walk.$"
`
	l := lexer.New(input)
	p := parser.New(l, "")
	program, err := p.ParseProgram()
	if err != nil {
		t.Fatalf(err.Error())
	}

	e := New(program, false)
	result, _ := e.Emit()
	if result != expectedUnoptimized {
		t.Errorf("Mismatching unoptimized emit -- Expected=%q, Got=%q", expectedUnoptimized, result)
	}

	e = New(program, true)
	result, _ = e.Emit()
	if result != expectedOptimized {
		t.Errorf("Mismatching optimized emit -- Expected=%q, Got=%q", expectedOptimized, result)
	}
}

// Helper benchmark var to prevent compiler/runtime optimizations.
// https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go
var benchResult string

func BenchmarkEmit1(b *testing.B) {
	input := `
script Route29_EventScript_WaitingMan {
	lock
	faceplayer
	# Display message based on time of day.
	gettime
	if (var(VAR_0x8002) == TIME_NIGHT) {
		msgbox("I'm waiting for POKéMON that appear\n"
				"only in the morning.")
	} else {
		msgbox("I'm waiting for POKéMON that appear\n"
				"only at night.")
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

	// According to my benchmarks, Unoptimized and Optimized have seemingly-identical performance.
	// I would expect Optimized to be consistently a little bit slower,
	// but I guess the optimizations in the emitter are so computationally light, that they
	// doesn't incur a performance hit.
	var result string
	b.Run("unoptimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lexer.New(input)
			p := parser.New(l, "")
			program, _ := p.ParseProgram()
			e := New(program, false)
			result, _ = e.Emit()
		}
	})
	benchResult = result

	b.Run("optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lexer.New(input)
			p := parser.New(l, "")
			program, _ := p.ParseProgram()
			e := New(program, true)
			result, _ = e.Emit()
		}
	})
	benchResult = result
}
