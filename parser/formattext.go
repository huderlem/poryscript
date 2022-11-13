package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// FontConfig holds the pixel widths of characters in various game fonts.
type FontConfig struct {
	DefaultFontID string           `json:"defaultFontId"`
	Fonts         map[string]Fonts `json:"fonts"`
}

type Fonts struct {
	Widths        map[string]int `json:"widths"`
	MaxLineLength int            `json:"maxLineLength"`
}

// LoadFontConfig reads a font width config JSON file.
func LoadFontConfig(filepath string) (FontConfig, error) {
	var config FontConfig
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		return config, err
	}

	return config, err
}

const testFontID = "TEST"

// FormatText automatically inserts line breaks into text
// according to in-game text box widths.
func (fw *FontConfig) FormatText(text string, maxWidth int, fontID string) (string, error) {
	if !fw.isFontIDValid(fontID) && len(fontID) > 0 && fontID != testFontID {
		validFontIDs := make([]string, len(fw.Fonts))
		i := 0
		for k := range fw.Fonts {
			validFontIDs[i] = k
			i++
		}
		return "", fmt.Errorf("unknown fontID '%s' used in format(). List of valid fontIDs are '%s'", fontID, validFontIDs)
	}

	text = strings.ReplaceAll(text, "\n", " ")

	var formattedSb strings.Builder
	var curLineSb strings.Builder
	curWidth := 0
	isFirstLine := true
	isFirstWord := true
	pos := 0
	for pos < len(text) {
		endPos, word, err := fw.getNextWord(text[pos:])
		if err != nil {
			return "", err
		}
		if len(word) == 0 {
			break
		}
		pos += endPos
		if fw.isLineBreak(word) {
			curWidth = 0
			formattedSb.WriteString(curLineSb.String())
			formattedSb.WriteString(word)
			formattedSb.WriteByte('\n')
			if fw.isParagraphBreak(word) {
				isFirstLine = true
			} else {
				isFirstLine = false
			}
			isFirstWord = true
			curLineSb.Reset()
		} else {
			wordWidth := 0
			if !isFirstWord {
				wordWidth += fw.getRunePixelWidth(' ', fontID)
			}
			wordWidth += fw.getWordPixelWidth(word, fontID)
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

	return formattedSb.String(), nil
}

func (fw *FontConfig) getNextWord(text string) (int, string, error) {
	escape := false
	endPos := 0
	startPos := 0
	foundNonSpace := false
	foundRegularRune := false
	endOnNext := false
	controlCodeLevel := 0
	for pos, char := range text {
		if endOnNext {
			return pos, text[startPos:pos], nil
		}
		if escape && (char == 'l' || char == 'n' || char == 'p') {
			if foundRegularRune {
				return endPos, text[startPos:endPos], nil
			}
			endOnNext = true
		} else if char == '\\' && controlCodeLevel == 0 {
			escape = true
			if !foundRegularRune {
				startPos = pos
			}
			foundNonSpace = true
			endPos = pos
		} else {
			if char == ' ' {
				if foundNonSpace && controlCodeLevel == 0 {
					return pos, text[startPos:pos], nil
				}
			} else {
				if !foundNonSpace {
					startPos = pos
				}
				foundRegularRune = true
				foundNonSpace = true
				if char == '{' {
					controlCodeLevel++
				} else if char == '}' {
					if controlCodeLevel > 0 {
						controlCodeLevel--
					}
				}
			}
			escape = false
		}
	}
	if !foundNonSpace {
		return len(text), "", nil
	}
	return len(text), text[startPos:], nil
}

func (fw *FontConfig) isLineBreak(word string) bool {
	return word == `\n` || word == `\l` || word == `\p`
}

func (fw *FontConfig) isParagraphBreak(word string) bool {
	return word == `\p`
}

func (fw *FontConfig) getWordPixelWidth(word string, fontID string) int {
	word, wordWidth := fw.processControlCodes(word, fontID)
	for _, r := range word {
		wordWidth += fw.getRunePixelWidth(r, fontID)
	}
	return wordWidth
}

func (fw *FontConfig) processControlCodes(word string, fontID string) (string, int) {
	width := 0
	re := regexp.MustCompile(`{[^}]*}`)
	positions := re.FindAllStringIndex(word, -1)
	for _, pos := range positions {
		code := word[pos[0]:pos[1]]
		width += fw.getControlCodePixelWidth(code, fontID)
	}
	strippedWord := re.ReplaceAllString(word, "")
	return strippedWord, width
}

func (fw *FontConfig) getRunePixelWidth(r rune, fontID string) int {
	if fontID == testFontID {
		return 10
	}
	return fw.readWidthFromFontConfig(string(r), fontID)
}

func (fw *FontConfig) getControlCodePixelWidth(code string, fontID string) int {
	if fontID == testFontID {
		return 100
	}
	return fw.readWidthFromFontConfig(code, fontID)
}

func (fw *FontConfig) isFontIDValid(fontID string) bool {
	_, ok := fw.Fonts[fontID]
	return ok
}

const fallbackWidth = 0

func (fw *FontConfig) readWidthFromFontConfig(value string, fontID string) int {
	font, ok := fw.Fonts[fontID]
	if !ok {
		return fallbackWidth
	}
	width, ok := font.Widths[value]
	if !ok {
		defaultWidth, ok := font.Widths["default"]
		if !ok {
			return fallbackWidth
		}
		return defaultWidth
	}
	return width
}
