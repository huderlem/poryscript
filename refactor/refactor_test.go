package refactor

import (
	"testing"

	"github.com/huderlem/poryscript/lexer"
	"github.com/huderlem/poryscript/token"
)

func TestDetectStringStyle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected StringStyle
	}{
		{
			name:     "single line string",
			input:    `"Hello, world."`,
			expected: StyleSingleLine,
		},
		{
			name:     "single line with break codes",
			input:    `"Hello\nworld\lthird line.\pNew paragraph."`,
			expected: StyleSingleLine,
		},
		{
			name: "auto string",
			input: `"Hello, I'm the first line,
                     and I'm the second line."`,
			expected: StyleAuto,
		},
		{
			name: "concatenated strings",
			input: `"Hello, line 1.\n"
                    "Line 2."`,
			expected: StyleConcatenated,
		},
		{
			name: "concatenated strings with indentation",
			input: `"Hello, line 1.\n"
                    "Line 2.\l"
                    "Line 3."`,
			expected: StyleConcatenated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectStringStyle(tt.input)
			if got != tt.expected {
				t.Errorf("DetectStringStyle() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestConvertAutoToConcatenated(t *testing.T) {
	input := `"Hello, I'm the first line,
	           and I'm the second line,
		       and this is the third.

		       This is a new paragraph because
		       of the blank line above."`
	indent := "       "

	got, err := ConvertString(input, StyleConcatenated, indent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "\"Hello, I'm the first line,\\n\"\n" +
		indent + "\"and I'm the second line,\\l\"\n" +
		indent + "\"and this is the third.\\p\"\n" +
		indent + "\"This is a new paragraph because\\n\"\n" +
		indent + "\"of the blank line above.\""

	if got != expected {
		t.Errorf("ConvertString(auto→concat):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertAutoToSingleLine(t *testing.T) {
	input := `"Hello, I'm the first line,
		       and I'm the second line,
		       and this is the third.

		       This is a new paragraph."`

	got, err := ConvertString(input, StyleSingleLine, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"Hello, I'm the first line,\nand I'm the second line,\land this is the third.\pThis is a new paragraph."`

	if got != expected {
		t.Errorf("ConvertString(auto→single):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertConcatenatedToAuto(t *testing.T) {
	input := `"Hello, I'm the first line.\n"
		      "and I'm the second line,\l"
		      "and this is the third.\p"
		      "This is a new paragraph."`
	indent := "        "

	got, err := ConvertString(input, StyleAuto, indent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "\"Hello, I'm the first line.\n" +
		indent + "and I'm the second line,\n" +
		indent + "and this is the third.\n" +
		indent + "\n" +
		indent + "This is a new paragraph.\""

	if got != expected {
		t.Errorf("ConvertString(concat→auto):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertConcatenatedToSingleLine(t *testing.T) {
	input := `"Hello, line 1.\n"
              "Line 2.\l"
			  "Line 3."`

	got, err := ConvertString(input, StyleSingleLine, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"Hello, line 1.\nLine 2.\lLine 3."`

	if got != expected {
		t.Errorf("ConvertString(concat→single):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertSingleLineToConcatenated(t *testing.T) {
	input := `"Hello, line 1.\nLine 2.\lLine 3.\pNew paragraph."`
	indent := "       "

	got, err := ConvertString(input, StyleConcatenated, indent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "\"Hello, line 1.\\n\"\n" +
		indent + "\"Line 2.\\l\"\n" +
		indent + "\"Line 3.\\p\"\n" +
		indent + "\"New paragraph.\""

	if got != expected {
		t.Errorf("ConvertString(single→concat):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertSingleLineToAuto(t *testing.T) {
	input := `"Hello, line 1.\nLine 2.\lLine 3.\pNew paragraph."`
	indent := "    "

	got, err := ConvertString(input, StyleAuto, indent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "\"Hello, line 1.\n" +
		indent + "Line 2.\n" +
		indent + "Line 3.\n" +
		indent + "\n" +
		indent + "New paragraph.\""

	if got != expected {
		t.Errorf("ConvertString(single→auto):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertSameStyleIsNoOp(t *testing.T) {
	inputs := []struct {
		source string
		style  StringStyle
	}{
		{`"Hello, world."`, StyleSingleLine},
		{"\"Hello\nworld\"", StyleAuto},
		{"\"Hello\\n\"\n\"world\"", StyleConcatenated},
	}

	for _, tt := range inputs {
		got, err := ConvertString(tt.source, tt.style, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tt.source {
			t.Errorf("ConvertString(same style) changed the text:\ngot:  %q\nwant: %q", got, tt.source)
		}
	}
}

func TestConvertNoBreakCodes(t *testing.T) {
	// A plain string with no break codes should produce equivalent output in all styles.
	input := `"Hello, world."`

	for _, style := range []StringStyle{StyleAuto, StyleConcatenated, StyleSingleLine} {
		got, err := ConvertString(input, style, "    ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != input {
			t.Errorf("ConvertString(no breaks → style %d):\ngot:  %q\nwant: %q", style, got, input)
		}
	}
}

func TestConvertPreservesNonBreakEscapes(t *testing.T) {
	// Escape sequences that are NOT break codes should be preserved.
	input := `"Hello {PLAYER}.\nYou found \\100 coins!"`

	got, err := ConvertString(input, StyleConcatenated, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"Hello {PLAYER}.\n"
"You found \\100 coins!"`

	if got != expected {
		t.Errorf("ConvertString(preserves escapes):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertBreakCodeMidText(t *testing.T) {
	// Break code in the middle of text content.
	input := `"Hello\nWorld"`

	got, err := ConvertString(input, StyleConcatenated, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `"Hello\n"
"World"`

	if got != expected {
		t.Errorf("ConvertString(mid-text break):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertMultipleConsecutivePageBreaks(t *testing.T) {
	input := `"Line 1.\pLine 2.\pLine 3."`

	got, err := ConvertString(input, StyleAuto, "    ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Each \p should produce a blank line in auto string.
	expected := `"Line 1.
    
    Line 2.
    
    Line 3."`

	if got != expected {
		t.Errorf("ConvertString(multi page breaks):\ngot:\n%s\n\nwant:\n%s", got, expected)
	}
}

func TestConvertUnicodeStrings(t *testing.T) {
	// Single-line with unicode characters including CJK, emoji, and accented chars.
	t.Run("single to concatenated preserves unicode", func(t *testing.T) {
		input := `"こんにちは世界\n第二行テスト\ltroisième ligne"`
		got, err := ConvertString(input, StyleConcatenated, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "\"こんにちは世界\\n\"\n\"第二行テスト\\l\"\n\"troisième ligne\""
		if got != expected {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, expected)
		}
	})

	t.Run("concatenated to single-line preserves unicode", func(t *testing.T) {
		input := "\"café\\n\"\n\"naïve\""
		got, err := ConvertString(input, StyleSingleLine, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `"café\nnaïve"`
		if got != expected {
			t.Errorf("got: %q, want: %q", got, expected)
		}
	})

	t.Run("single to auto preserves unicode", func(t *testing.T) {
		input := `"你好\p再见"`
		got, err := ConvertString(input, StyleAuto, "    ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "\"你好\n    \n    再见\""
		if got != expected {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, expected)
		}
	})

	t.Run("auto to single-line preserves unicode", func(t *testing.T) {
		input := "\"première ligne\n    deuxième ligne\""
		got, err := ConvertString(input, StyleSingleLine, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `"première ligne\ndeuxième ligne"`
		if got != expected {
			t.Errorf("got: %q, want: %q", got, expected)
		}
	})

	t.Run("unicode with game placeholders", func(t *testing.T) {
		input := `"{COLOR RED}アイテム「ポーション」\nを手に入れた！"`
		got, err := ConvertString(input, StyleConcatenated, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "\"{COLOR RED}アイテム「ポーション」\\n\"\n\"を手に入れた！\""
		if got != expected {
			t.Errorf("got:\n%s\n\nwant:\n%s", got, expected)
		}
	})
}

func TestExtractQuotedSegmentsUnicode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "CJK characters",
			input:    `"こんにちは"`,
			expected: []string{"こんにちは"},
		},
		{
			name:     "accented characters",
			input:    `"café" "naïve"`,
			expected: []string{"café", "naïve"},
		},
		{
			name:     "emoji",
			input:    `"hello 🌍"`,
			expected: []string{"hello 🌍"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractQuotedSegments(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d segments, want %d", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("segment[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSplitOnBreakCodesUnicode(t *testing.T) {
	got := splitOnBreakCodes(`こんにちは\n世界\pさようなら`)
	if len(got) != 3 {
		t.Fatalf("got %d segments, want 3", len(got))
	}
	if got[0].text != "こんにちは" || got[0].brk != breakNewline {
		t.Errorf("segment[0] = %+v", got[0])
	}
	if got[1].text != "世界" || got[1].brk != breakPage {
		t.Errorf("segment[1] = %+v", got[1])
	}
	if got[2].text != "さようなら" {
		t.Errorf("segment[2] = %+v", got[2])
	}
}

func TestExtractQuotedSegments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single segment",
			input:    `"hello"`,
			expected: []string{"hello"},
		},
		{
			name:     "two segments",
			input:    `"hello" "world"`,
			expected: []string{"hello", "world"},
		},
		{
			name:     "segments with newline between",
			input:    "\"hello\"\n\"world\"",
			expected: []string{"hello", "world"},
		},
		{
			name:     "segment with escaped quote",
			input:    `"he\"llo"`,
			expected: []string{`he\"llo`},
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractQuotedSegments(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("extractQuotedSegments() returned %d segments, want %d", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("segment[%d] = %q, want %q", i, got[i], tt.expected[i])
				}
			}
		})
	}
}

func TestSplitOnBreakCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []segment
	}{
		{
			name:     "no breaks",
			input:    "hello world",
			expected: []segment{{text: "hello world"}},
		},
		{
			name:  "single newline break",
			input: `hello\nworld`,
			expected: []segment{
				{text: "hello", brk: breakNewline},
				{text: "world"},
			},
		},
		{
			name:  "all three breaks",
			input: `line1\nline2\lline3\pline4`,
			expected: []segment{
				{text: "line1", brk: breakNewline},
				{text: "line2", brk: breakScroll},
				{text: "line3", brk: breakPage},
				{text: "line4"},
			},
		},
		{
			name:     "escaped backslash not treated as break",
			input:    `hello\\nworld`,
			expected: []segment{{text: `hello\\nworld`}},
		},
		{
			name:  "other escape preserved",
			input: `hello\tworld\nbye`,
			expected: []segment{
				{text: `hello\tworld`, brk: breakNewline},
				{text: "bye"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitOnBreakCodes(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("splitOnBreakCodes() returned %d segments, want %d", len(got), len(tt.expected))
			}
			for i := range got {
				if got[i].text != tt.expected[i].text {
					t.Errorf("segment[%d].text = %q, want %q", i, got[i].text, tt.expected[i].text)
				}
				if i < len(tt.expected)-1 && got[i].brk != tt.expected[i].brk {
					t.Errorf("segment[%d].brk = %d, want %d", i, got[i].brk, tt.expected[i].brk)
				}
			}
		})
	}
}

func tokenize(content string) []token.Token {
	l := lexer.New(content)
	var tokens []token.Token
	for {
		t := l.NextToken()
		if t.Type == token.EOF {
			break
		}
		tokens = append(tokens, t)
	}
	return tokens
}

func TestFindStringTokenAtPosition(t *testing.T) {
	t.Run("single-line string", func(t *testing.T) {
		content := `script MyScript {
    msgbox("Hello, world.")
}`
		tokens := tokenize(content)

		// Cursor on the opening quote (line 1, col 11).
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 11)
		if !found {
			t.Fatal("expected to find string token at opening quote")
		}
		if tok.Type != token.STRING {
			t.Errorf("expected STRING token, got %s", tok.Type)
		}

		// Cursor in the middle of the string.
		tok, _, found = FindStringTokenAtPosition(tokens, 1, 18)
		if !found {
			t.Fatal("expected to find string token in middle of string")
		}
		if tok.Type != token.STRING {
			t.Errorf("expected STRING token, got %s", tok.Type)
		}

		// Cursor on the closing quote.
		tok, _, found = FindStringTokenAtPosition(tokens, 1, tok.EndUtf8CharIndex)
		if !found {
			t.Fatal("expected to find string token at closing quote")
		}

		// Cursor before the string (on 'msgbox').
		_, _, found = FindStringTokenAtPosition(tokens, 1, 4)
		if found {
			t.Error("should not find string token before the string")
		}

		// Cursor after the string (past closing quote).
		_, _, found = FindStringTokenAtPosition(tokens, 1, 27)
		if found {
			t.Error("should not find string token after the string")
		}

		// Cursor on a different line entirely.
		_, _, found = FindStringTokenAtPosition(tokens, 0, 5)
		if found {
			t.Error("should not find string token on the script keyword line")
		}
	})

	t.Run("multi-line auto string", func(t *testing.T) {
		content := "text MyText {\n    \"Hello first line\n     second line\n     third line\"\n}"
		tokens := tokenize(content)

		// Cursor on the first line of the string.
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 6)
		if !found {
			t.Fatal("expected to find string token on first line")
		}
		if tok.Type != token.AUTOSTRING {
			t.Errorf("expected AUTOSTRING token, got %s", tok.Type)
		}

		// Cursor on a middle line.
		_, _, found = FindStringTokenAtPosition(tokens, 2, 10)
		if !found {
			t.Fatal("expected to find string token on middle line")
		}

		// Cursor on the last line of the string.
		_, _, found = FindStringTokenAtPosition(tokens, 3, 10)
		if !found {
			t.Fatal("expected to find string token on last line")
		}

		// Cursor on the closing brace line.
		_, _, found = FindStringTokenAtPosition(tokens, 4, 0)
		if found {
			t.Error("should not find string token on closing brace line")
		}
	})

	t.Run("no string tokens", func(t *testing.T) {
		content := "script MyScript {\n    lock\n}"
		tokens := tokenize(content)
		_, _, found := FindStringTokenAtPosition(tokens, 1, 4)
		if found {
			t.Error("should not find string token in script with no strings")
		}
	})

	t.Run("multiple strings picks correct one", func(t *testing.T) {
		content := "script MyScript {\n    msgbox(\"first\")\n    msgbox(\"second\")\n}"
		tokens := tokenize(content)

		// Cursor in first string.
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 13)
		if !found {
			t.Fatal("expected to find first string token")
		}
		if tok.Literal != "first" {
			t.Errorf("expected literal 'first', got %q", tok.Literal)
		}

		// Cursor in second string.
		tok, _, found = FindStringTokenAtPosition(tokens, 2, 13)
		if !found {
			t.Fatal("expected to find second string token")
		}
		if tok.Literal != "second" {
			t.Errorf("expected literal 'second', got %q", tok.Literal)
		}
	})

	t.Run("cursor just outside single-line string bounds", func(t *testing.T) {
		content := `    "hello"`
		tokens := tokenize(content)

		// One character before the opening quote.
		_, _, found := FindStringTokenAtPosition(tokens, 0, 3)
		if found {
			t.Error("should not find string token one char before opening quote")
		}
	})
}

func TestExtractTokenSourceText(t *testing.T) {
	t.Run("single-line string", func(t *testing.T) {
		content := `script MyScript {
    msgbox("Hello, world.")
}`
		tokens := tokenize(content)
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 12)
		if !found {
			t.Fatal("expected to find string token")
		}
		got := ExtractTokenSourceText(content, tok)
		expected := `"Hello, world."`
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("multi-line auto string", func(t *testing.T) {
		content := "text MyText {\n    \"first line\n     second line\"\n}"
		tokens := tokenize(content)
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 5)
		if !found {
			t.Fatal("expected to find string token")
		}
		got := ExtractTokenSourceText(content, tok)
		expected := "\"first line\n     second line\""
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("string with leading content on line", func(t *testing.T) {
		content := `script MyScript {
    msgbox("test string")
}`
		tokens := tokenize(content)
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 12)
		if !found {
			t.Fatal("expected to find string token")
		}
		got := ExtractTokenSourceText(content, tok)
		expected := `"test string"`
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("invalid token line returns empty", func(t *testing.T) {
		content := "hello"
		tok := token.Token{
			LineNumber:    100, // way out of bounds
			EndLineNumber: 100,
		}
		got := ExtractTokenSourceText(content, tok)
		if got != "" {
			t.Errorf("expected empty string for out-of-bounds token, got %q", got)
		}
	})

	t.Run("multi-line string with three lines", func(t *testing.T) {
		content := "text MyText {\n    \"line one\n     line two\n     line three\"\n}"
		tokens := tokenize(content)
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 5)
		if !found {
			t.Fatal("expected to find string token")
		}
		got := ExtractTokenSourceText(content, tok)
		expected := "\"line one\n     line two\n     line three\""
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

	t.Run("extracted text round-trips through DetectStringStyle", func(t *testing.T) {
		content := "text MyText {\n    \"first line\n     second line\"\n}"
		tokens := tokenize(content)
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 5)
		if !found {
			t.Fatal("expected to find string token")
		}
		sourceText := ExtractTokenSourceText(content, tok)
		style := DetectStringStyle(sourceText)
		if style != StyleAuto {
			t.Errorf("expected StyleAuto, got %d", style)
		}
	})

	t.Run("single-line extracted text round-trips through DetectStringStyle", func(t *testing.T) {
		content := "script MyScript {\n    msgbox(\"Hello\\nWorld\")\n}"
		tokens := tokenize(content)
		tok, _, found := FindStringTokenAtPosition(tokens, 1, 12)
		if !found {
			t.Fatal("expected to find string token")
		}
		sourceText := ExtractTokenSourceText(content, tok)
		style := DetectStringStyle(sourceText)
		if style != StyleSingleLine {
			t.Errorf("expected StyleSingleLine, got %d", style)
		}
	})
}

func TestIsFormatStringToken(t *testing.T) {
	t.Run("string inside format()", func(t *testing.T) {
		content := "text MyText {\n    format(\"Hello world\")\n}"
		tokens := tokenize(content)
		_, idx, found := FindStringTokenAtPosition(tokens, 1, 12)
		if !found {
			t.Fatal("expected to find string token")
		}
		if !IsFormatStringToken(tokens, idx) {
			t.Error("expected string inside format() to be detected")
		}
	})

	t.Run("string not inside format()", func(t *testing.T) {
		content := "text MyText {\n    msgbox(\"Hello world\")\n}"
		tokens := tokenize(content)
		_, idx, found := FindStringTokenAtPosition(tokens, 1, 12)
		if !found {
			t.Fatal("expected to find string token")
		}
		if IsFormatStringToken(tokens, idx) {
			t.Error("string inside msgbox() should not be detected as format string")
		}
	})

	t.Run("bare string in text block", func(t *testing.T) {
		content := "text MyText {\n    \"Hello world\"\n}"
		tokens := tokenize(content)
		_, idx, found := FindStringTokenAtPosition(tokens, 1, 5)
		if !found {
			t.Fatal("expected to find string token")
		}
		if IsFormatStringToken(tokens, idx) {
			t.Error("bare string should not be detected as format string")
		}
	})

	t.Run("format with string type prefix", func(t *testing.T) {
		content := "text MyText {\n    format(ascii\"Hello world\")\n}"
		tokens := tokenize(content)
		_, idx, found := FindStringTokenAtPosition(tokens, 1, 16)
		if !found {
			t.Fatal("expected to find string token")
		}
		if !IsFormatStringToken(tokens, idx) {
			t.Error("expected string inside format(ascii...) to be detected")
		}
	})
}
