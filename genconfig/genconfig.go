package genconfig

import (
	"github.com/huderlem/poryscript/token"
	"github.com/huderlem/poryscript/types"
)

// SupportsDollarSignHexNotation tells whether or not the given Gen supports
// dollar-sign-prefixed hexadecimal notation. Example: $4AD0
func SupportsDollarSignHexNotation(gen types.Gen) bool {
	switch gen {
	case types.GEN2:
		return true
	case types.GEN3:
		return false
	default:
		return true
	}
}

// Supports0xHexNotation tells whether or not the given Gen supports
// 0x-prefixed hexadecimal notation. Example: 0x4AD0
func Supports0xHexNotation(gen types.Gen) bool {
	switch gen {
	case types.GEN2:
		return false
	case types.GEN3:
		return true
	default:
		return true
	}
}

// SupportedBooleanOperators gets the list of supported operators that can be used in a boolean expression.
// An empty list indicates that arbitrary operators can be used.
func SupportedBooleanOperators(gen types.Gen) []token.Type {
	switch gen {
	case types.GEN2:
		return []token.Type{}
	case types.GEN3:
		return []token.Type{token.VAR, token.FLAG, token.DEFEATED}
	default:
		return []token.Type{}
	}
}

// SupportedSwitchOperators gets the list of supported operators that can be used in a switch statement.
// An empty list indicates that arbitrary operators can be used.
func SupportedSwitchOperators(gen types.Gen) []token.Type {
	switch gen {
	case types.GEN2:
		return []token.Type{}
	case types.GEN3:
		return []token.Type{token.VAR}
	default:
		return []token.Type{}
	}
}

// LocalScriptNamePrefixes is a mapping of the prefix for
// local script label names for each Gen.
var LocalScriptNamePrefixes = map[types.Gen]string{
	types.GEN2: ".",
	types.GEN3: "",
}

// ReturnCommands is a mapping of the return command used for
// each Gen. These return from the called script execution, or
// halt script execution if there is no current call context.
var ReturnCommands = map[types.Gen]string{
	types.GEN2: "end",
	types.GEN3: "return",
}

// EndCommands is a mapping of the halting end command used for
// each Gen. These halt script execution.
var EndCommands = map[types.Gen]string{
	types.GEN2: "endall",
	types.GEN3: "end",
}

// GotoCommands is a mapping of the goto command used for
// each Gen. These jump directly to another related script.
var GotoCommands = map[types.Gen]string{
	types.GEN2: "sjump",
	types.GEN3: "goto",
}

// AltReturnCommands is a mapping of the alternative return command
// used for each Gen. For example, in pokecrystal, "return" must be
// used to exit from a map callback script.
var AltReturnCommands = map[types.Gen]string{
	types.GEN2: "return",
	types.GEN3: "return",
}

// DefaultFontIDs is a mapping of the default font ids for
// formatted text. These IDs are found within font_widths.json.
var DefaultFontIDs = map[types.Gen]string{
	types.GEN2: "return",
	types.GEN3: "return",
}

// TextBoxCharWidths is a mapping of the maximum width for
// formatted text.
var TextBoxCharWidths = map[types.Gen]int{
	types.GEN2: 144,
	types.GEN3: 208,
}

// TextControlCodeStarts is a mapping of the characters that start
// control codes in text.
var TextControlCodeStarts = map[types.Gen]map[rune]struct{}{
	types.GEN2: map[rune]struct{}{
		'<': struct{}{},
		'{': struct{}{},
	},
	types.GEN3: map[rune]struct{}{
		'{': struct{}{},
	},
}

// TextControlCodeEnds is a mapping of the characters that end
// control codes in text.
var TextControlCodeEnds = map[types.Gen]map[rune]struct{}{
	types.GEN2: map[rune]struct{}{
		'>': struct{}{},
		'}': struct{}{},
	},
	types.GEN3: map[rune]struct{}{
		'}': struct{}{},
	},
}
