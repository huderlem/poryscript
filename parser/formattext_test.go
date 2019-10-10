package parser

import "testing"

func TestFormatText(t *testing.T) {
	tests := []struct {
		inputText string
		expected  string
	}{
		{"Hello", "Hello"},
	}

	for i, tt := range tests {
		result := FormatText(tt.inputText)
		if result != tt.expected {
			t.Errorf("FormatText Test %d: Expected '%s', but Got '%s'", i, tt.expected, result)
		}
	}
}
