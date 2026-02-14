package parser

import "testing"

func TestFormatText(t *testing.T) {
	tests := []struct {
		maxWidth           int
		cursorOverlapWidth int
		inputText          string
		expected           string
		numLines           int
	}{
		{40, 0, "", "", 2},
		{40, 0, "Hello", "Hello", 2},
		{100, 0, "Foo bar", "Foo bar", 2},
		{140, 0, "Foo {MUS} baz", "Foo {MUS}\\n\nbaz", 2},
		{139, 0, "Foo {MUS} baz", "Foo\\n\n{MUS}\\l\nbaz", 2},
		{40, 0, "ßŒœ ♂Üあ   ", "ßŒœ\\n\n♂Üあ", 2},
		{40, 0, "   Foo    bar          baz  baz2", "Foo\\n\nbar\\l\nbaz\\l\nbaz2", 2},
		{100, 0, `Hello.\pI am writing a test.`, "Hello.\\p\nI am\\n\nwriting a\\l\ntest.", 2},
		{100, 0, `Hello.\nI am writing a longer \l“test.”`, "Hello.\\n\nI am\\l\nwriting a\\l\nlonger\\l\n“test.”", 2},
		{190, 0, `First\nSecond Third Fourth`, "First\\n\nSecond Third Fourth", 2},
		{190, 1, `First\nSecond Third Fourth`, "First\\n\nSecond Third Fourth", 2},
		{190, 0, `First\nSecond Third Fourth Second Third Fourth`, "First\\n\nSecond Third Fourth\\l\nSecond Third Fourth", 2},
		{190, 1, `First\nSecond Third Fourth Second Third Fourth`, "First\\n\nSecond Third\\l\nFourth Second\\l\nThird Fourth", 2},
		{130, 0, `Apple Banana\pOrange`, "Apple Banana\\p\nOrange", 2},
		{130, 10, `Apple Banana\pOrange`, "Apple Banana\\p\nOrange", 2},
		{130, 11, `Apple Banana\pOrange`, "Apple\\n\nBanana\\p\nOrange", 2},
		{100, 0, `Hello.\NI am writing a longer \N“test.”`, "Hello.\\n\nI am\\l\nwriting a\\l\nlonger\\l\n“test.”", 2},
		{100, 0, `Hello.\NI am\Nwriting\la longer \N“test.”`, "Hello.\\n\nI am\\l\nwriting\\l\na longer\\l\n“test.”", 2},
		{100, 0, `Hello.\NI am\Nwriting\pa longer \N“test.”`, "Hello.\\n\nI am\\l\nwriting\\p\na longer\\n\n“test.”", 2},
		{100, 0, `Hello.\NI am\Nwriting\la longer \N“test.”`, "Hello.\\n\nI am\\n\nwriting\\l\na longer\\l\n“test.”", 3},
		{100, 0, `Hello.\NI am\Nwriting\pa longer \N“test.”`, "Hello.\\n\nI am\\n\nwriting\\p\na longer\\n\n“test.”", 3},
		{100, 0, `Hello.\NI am\Nwriting\pa longer \N“test.”`, "Hello.\\l\nI am\\l\nwriting\\p\na longer\\l\n“test.”", 1},
	}

	fc := FontConfig{}

	for i, tt := range tests {
		result, _ := fc.FormatText(tt.inputText, tt.maxWidth, tt.cursorOverlapWidth, testFontID, tt.numLines)
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
		{"{COLOR BLUE}Player is cool", 18, "{COLOR BLUE}Player"},
		{"{  COLOR BLUE RED }Player is cool", 25, "{  COLOR BLUE RED }Player"},
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

	fc := FontConfig{}

	for i, tt := range tests {
		resultPos, resultValue := fc.getNextWord(tt.inputText)
		if resultPos != tt.expectedPos {
			t.Errorf("TestGetNextWord Test %d: Expected Pos '%d', but Got '%d'", i, tt.expectedPos, resultPos)
		}
		if resultValue != tt.expectedValue {
			t.Errorf("TestGetNextWord Test %d: Expected Value '%s', but Got '%s'", i, tt.expectedValue, resultValue)
		}
	}
}

func TestValidateLineWidths(t *testing.T) {
	fc := FontConfig{}

	// Lines within limit — no errors
	errors := fc.ValidateLineWidths("Hello", testFontID, 100)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for short text, got %d", len(errors))
	}

	// Empty text — no errors
	errors = fc.ValidateLineWidths("", testFontID, 100)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for empty text, got %d", len(errors))
	}

	// maxWidth <= 0 — validation skipped
	errors = fc.ValidateLineWidths("Hello World This Is Long", testFontID, 0)
	if len(errors) != 0 {
		t.Errorf("Expected no errors when maxWidth <= 0, got %d", len(errors))
	}
	errors = fc.ValidateLineWidths("Hello World This Is Long", testFontID, -1)
	if len(errors) != 0 {
		t.Errorf("Expected no errors when maxWidth < 0, got %d", len(errors))
	}

	// Single line exceeding limit (each char is 10px in TEST font, "Hello" = 50px)
	errors = fc.ValidateLineWidths("Hello", testFontID, 40)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
	if errors[0].LineIndex != 0 {
		t.Errorf("Expected LineIndex 0, got %d", errors[0].LineIndex)
	}
	if errors[0].PixelWidth != 50 {
		t.Errorf("Expected PixelWidth 50, got %d", errors[0].PixelWidth)
	}
	if errors[0].MaxWidth != 40 {
		t.Errorf("Expected MaxWidth 40, got %d", errors[0].MaxWidth)
	}

	// Multiple logical lines (real newlines), second exceeds.
	// Simulates AUTOSTRING: "Hi\n\nHello World\l\n"
	// Line 0: "Hi\n" → content "Hi" = 20px
	// Line 1: "Hello World\l" → content "Hello World" = 110px
	errors = fc.ValidateLineWidths("Hi\\n\nHello World\\l\n", testFontID, 100)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}
	if errors[0].LineIndex != 1 {
		t.Errorf("Expected LineIndex 1, got %d", errors[0].LineIndex)
	}
	if errors[0].PixelWidth != 110 {
		t.Errorf("Expected PixelWidth 110, got %d", errors[0].PixelWidth)
	}

	// Both lines exceed
	errors = fc.ValidateLineWidths("Hello World\\n\nHello World\\l\n", testFontID, 100)
	if len(errors) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(errors))
	}
	if errors[0].LineIndex != 0 {
		t.Errorf("Expected first error LineIndex 0, got %d", errors[0].LineIndex)
	}
	if errors[1].LineIndex != 1 {
		t.Errorf("Expected second error LineIndex 1, got %d", errors[1].LineIndex)
	}

	// Manual escape line breaks split into sub-segments.
	// "Hi\nWorld" splits into "Hi" (20px) and "World" (50px).
	// With maxWidth 40, only "World" exceeds.
	// "World" starts at byte offset 4 (after "Hi\n"), rune offset 4, length 5.
	errors = fc.ValidateLineWidths(`Hi\nWorld`, testFontID, 40)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for manual-break sub-segment, got %d", len(errors))
	}
	if errors[0].LineIndex != 0 {
		t.Errorf("Expected LineIndex 0 (parent logical line), got %d", errors[0].LineIndex)
	}
	if errors[0].PixelWidth != 50 {
		t.Errorf("Expected PixelWidth 50, got %d", errors[0].PixelWidth)
	}
	if errors[0].LineText != "World" {
		t.Errorf("Expected LineText 'World', got '%s'", errors[0].LineText)
	}
	if errors[0].CharOffset != 4 {
		t.Errorf("Expected CharOffset 4, got %d", errors[0].CharOffset)
	}
	if errors[0].Utf8CharOffset != 4 {
		t.Errorf("Expected Utf8CharOffset 4, got %d", errors[0].Utf8CharOffset)
	}
	if errors[0].CharLength != 5 {
		t.Errorf("Expected CharLength 5, got %d", errors[0].CharLength)
	}
	if errors[0].Utf8CharLength != 5 {
		t.Errorf("Expected Utf8CharLength 5, got %d", errors[0].Utf8CharLength)
	}

	// Spaces are counted individually (not collapsed).
	// "A  B" = 4 chars * 10px = 40px. Both spaces count.
	errors = fc.ValidateLineWidths("A  B", testFontID, 30)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for spaced text, got %d", len(errors))
	}
	if errors[0].PixelWidth != 40 {
		t.Errorf("Expected PixelWidth 40 (spaces counted), got %d", errors[0].PixelWidth)
	}

	// Trailing spaces contribute to width.
	// "Hi   " on first logical line = 5 chars * 10px = 50px
	errors = fc.ValidateLineWidths("Hi   \\n\nOk\\l\n", testFontID, 40)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for trailing spaces, got %d", len(errors))
	}
	if errors[0].LineIndex != 0 {
		t.Errorf("Expected LineIndex 0, got %d", errors[0].LineIndex)
	}
	if errors[0].PixelWidth != 50 {
		t.Errorf("Expected PixelWidth 50, got %d", errors[0].PixelWidth)
	}

	// Leading spaces contribute to width.
	// "   Hi" on second logical line = 5 chars * 10px = 50px
	errors = fc.ValidateLineWidths("Ok\\n\n   Hi\\l\n", testFontID, 40)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for leading spaces, got %d", len(errors))
	}
	if errors[0].LineIndex != 1 {
		t.Errorf("Expected LineIndex 1, got %d", errors[0].LineIndex)
	}
	if errors[0].PixelWidth != 50 {
		t.Errorf("Expected PixelWidth 50, got %d", errors[0].PixelWidth)
	}

	// Control codes contribute their width (100px in TEST font).
	// "{PLAYER}Hi" = 100 + 20 = 120px
	errors = fc.ValidateLineWidths("{PLAYER}Hi", testFontID, 110)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for control code text, got %d", len(errors))
	}
	if errors[0].PixelWidth != 120 {
		t.Errorf("Expected PixelWidth 120, got %d", errors[0].PixelWidth)
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

	fc := FontConfig{}

	for i, tt := range tests {
		resultValue, resultWidth := fc.processControlCodes(tt.inputText, testFontID)
		if resultValue != tt.expectedValue {
			t.Errorf("TestProcessControlCodes Test %d: Expected Value '%s', but Got '%s'", i, tt.expectedValue, resultValue)
		}
		if resultWidth != tt.expectedWidth {
			t.Errorf("TestProcessControlCodes Test %d: Expected Width '%d', but Got '%d'", i, tt.expectedWidth, resultWidth)
		}
	}
}
