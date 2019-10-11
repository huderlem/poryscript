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
		{40, "ßŒœ ♂Üあ   ", "ßŒœ\\n\n♂Üあ"},
		{40, "   Foo    bar          baz  baz2", "Foo\\n\nbar\\l\nbaz\\l\nbaz2"},
		{100, `Hello.\pI am writing a test.`, "Hello.\\p\nI am\\n\nwriting a\\l\ntest."},
		{100, `Hello.\nI am writing a longer \l“test.”`, "Hello.\\n\nI am\\l\nwriting a\\l\nlonger\\l\n“test.”"},
	}

	for i, tt := range tests {
		result := FormatText(tt.inputText, tt.maxWidth, "TEST")
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

	for i, tt := range tests {
		resultPos, resultValue := getNextWord(tt.inputText)
		if resultPos != tt.expectedPos {
			t.Errorf("TestGetNextWord Test %d: Expected Pos '%d', but Got '%d'", i, tt.expectedPos, resultPos)
		}
		if resultValue != tt.expectedValue {
			t.Errorf("TestGetNextWord Test %d: Expected Value '%s', but Got '%s'", i, tt.expectedValue, resultValue)
		}
	}
}
