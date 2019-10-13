package parser

import "testing"

func TestFormatText(t *testing.T) {
	tests := []struct {
		maxWidth  int
		inputText string
		expected  string
	}{
		{40, "", ""},
		{40, "Hello", "Hello"},
		{100, "Foo bar", "Foo bar"},
		{140, "Foo {MUS} baz", "Foo {MUS}\\n\nbaz"},
		{139, "Foo {MUS} baz", "Foo\\n\n{MUS}\\l\nbaz"},
		{40, "ßŒœ ♂Üあ   ", "ßŒœ\\n\n♂Üあ"},
		{40, "   Foo    bar          baz  baz2", "Foo\\n\nbar\\l\nbaz\\l\nbaz2"},
		{100, `Hello.\pI am writing a test.`, "Hello.\\p\nI am\\n\nwriting a\\l\ntest."},
		{100, `Hello.\nI am writing a longer \l“test.”`, "Hello.\\n\nI am\\l\nwriting a\\l\nlonger\\l\n“test.”"},
	}

	fw := FontWidthsConfig{}

	for i, tt := range tests {
		result, _ := fw.FormatText(tt.inputText, tt.maxWidth, testFontID)
		if result != tt.expected {
			t.Errorf("FormatText Test %d: Expected '%s', but Got '%s'", i, tt.expected, result)
		}
	}
}

func TestGetNextWord(t *testing.T) {
	tests := []struct {
		inputText     string
		expectedPos   int
		expectedValue string
	}{
		{"", 0, ""},
		{"    ", 4, ""},
		{"A", 1, "A"},
		{"  B", 3, "B"},
		{"  C  ", 3, "C"},
		{"Hello", 5, "Hello"},
		{"{PLAYER} is cool", 8, "{PLAYER}"},
		{"{PLAYER}is cool", 10, "{PLAYER}is"},
		{"Foo Bar", 3, "Foo"},
		{`Foo\nBar`, 3, "Foo"},
		{`Foo\lBar`, 3, "Foo"},
		{`Foo\pBar`, 3, "Foo"},
		{`Foo\tBar`, 8, `Foo\tBar`},
		{` \nBar`, 3, `\n`},
		{`   \l  Bar`, 5, `\l`},
		{`\p  Bar`, 2, `\p`},
		{` \ \p  Bar`, 2, `\`},
	}

	fw := FontWidthsConfig{}

	for i, tt := range tests {
		resultPos, resultValue := fw.getNextWord(tt.inputText)
		if resultPos != tt.expectedPos {
			t.Errorf("TestGetNextWord Test %d: Expected Pos '%d', but Got '%d'", i, tt.expectedPos, resultPos)
		}
		if resultValue != tt.expectedValue {
			t.Errorf("TestGetNextWord Test %d: Expected Value '%s', but Got '%s'", i, tt.expectedValue, resultValue)
		}
	}
}

func TestProcessControlCodes(t *testing.T) {
	tests := []struct {
		inputText     string
		expectedValue string
		expectedWidth int
	}{
		{"", "", 0},
		{"NoCodes", "NoCodes", 0},
		{"}NoCodes{", "}NoCodes{", 0},
		{"Has{}Codes", "HasCodes", 100},
		{"Has{FOO}{PLAYE{R}}Codes", "Has}Codes", 200},
	}

	fw := FontWidthsConfig{}

	for i, tt := range tests {
		resultValue, resultWidth := fw.processControlCodes(tt.inputText, testFontID)
		if resultValue != tt.expectedValue {
			t.Errorf("TestProcessControlCodes Test %d: Expected Value '%s', but Got '%s'", i, tt.expectedValue, resultValue)
		}
		if resultWidth != tt.expectedWidth {
			t.Errorf("TestProcessControlCodes Test %d: Expected Width '%d', but Got '%d'", i, tt.expectedWidth, resultWidth)
		}
	}
}
