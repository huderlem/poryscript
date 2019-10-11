package parser

import (
	"strings"
)

// FormatText automatically inserts line breaks into text
// according to in-game text box widths.
func FormatText(text string, maxWidth int, fontID string) string {
	var formattedSb strings.Builder
	var curLineSb strings.Builder
	curWidth := 0
	isFirstLine := true
	isFirstWord := true
	pos := 0
	for pos < len(text) {
		endPos, word := getNextWord(text[pos:])
		if len(word) == 0 {
			break
		}
		pos += endPos
		if isLineBreak(word) {
			curWidth = 0
			formattedSb.WriteString(curLineSb.String())
			formattedSb.WriteString(word)
			formattedSb.WriteByte('\n')
			if isParagraphBreak(word) {
				isFirstLine = true
			} else {
				isFirstLine = false
			}
			isFirstWord = true
			curLineSb.Reset()
		} else {
			wordWidth := 0
			if !isFirstWord {
				wordWidth += getRunePixelWidth(' ', fontID)
			}
			for _, r := range word {
				wordWidth += getRunePixelWidth(r, fontID)
			}
			if curWidth+wordWidth > maxWidth && curLineSb.Len() > 0 {
				formattedSb.WriteString(curLineSb.String())
				if isFirstLine {
					formattedSb.WriteString(`\n`)
					isFirstLine = false
				} else {
					formattedSb.WriteString(`\l`)
				}
				formattedSb.WriteByte('\n')
				isFirstWord = false
				curLineSb.Reset()
				curLineSb.WriteString(word)
				curWidth = wordWidth
			} else {
				curWidth += wordWidth
				if !isFirstWord {
					curLineSb.WriteByte(' ')
				}
				curLineSb.WriteString(word)
				isFirstWord = false
			}
		}
	}

	if curLineSb.Len() > 0 {
		formattedSb.WriteString(curLineSb.String())
	}

	return formattedSb.String()
}

func getNextWord(text string) (int, string) {
	escape := false
	endPos := 0
	startPos := 0
	foundNonSpace := false
	foundRegularRune := false
	endOnNext := false
	for pos, char := range text {
		if endOnNext {
			return pos, text[startPos:pos]
		}
		if escape && (char == 'l' || char == 'n' || char == 'p') {
			if foundRegularRune {
				return endPos, text[startPos:endPos]
			}
			endOnNext = true
		} else if char == '\\' {
			escape = true
			if !foundRegularRune {
				startPos = pos
			}
			foundNonSpace = true
			endPos = pos
		} else {
			if char == ' ' {
				if foundNonSpace {
					return pos, text[startPos:pos]
				}
			} else {
				if !foundNonSpace {
					startPos = pos
				}
				foundRegularRune = true
				foundNonSpace = true
			}
			escape = false
		}
	}
	if !foundNonSpace {
		return len(text), ""
	}
	return len(text), text[startPos:]
}

func isLineBreak(word string) bool {
	return word == `\n` || word == `\l` || word == `\p`
}

func isParagraphBreak(word string) bool {
	return word == `\p`
}

func getRunePixelWidth(r rune, fontID string) int {
	if fontID == "TEST" {
		return 10
	}
	return 8
}
