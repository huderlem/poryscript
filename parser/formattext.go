package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/huderlem/poryscript/genconfig"
	"github.com/huderlem/poryscript/types"
)

// FontWidthsConfig holds the pixel widths of characters in various game fonts.
type FontWidthsConfig struct {
	Fonts             map[string]map[string]int `json:"fonts"`
	DefaultFontIDGen2 string                    `json:"defaultFontId_gen2"`
	DefaultFontIDGen3 string                    `json:"defaultFontId"`
}

// LoadFontWidths reads a font width config JSON file.
func LoadFontWidths(filepath string) (FontWidthsConfig, error) {
	var config FontWidthsConfig
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		return config, err
	}

	return config, err
}

// GetDefaultFontID returns the key for the default font id
// of the given gen config.
func (fw *FontWidthsConfig) GetDefaultFontID(gen types.Gen) string {
	switch gen {
	case types.GEN2:
		return fw.DefaultFontIDGen2
	case types.GEN3:
		return fw.DefaultFontIDGen3
	default:
		return ""
	}
}

const testFontID = "TEST"

// FormatText automatically inserts line breaks into text
// according to in-game text box widths.
func (fw *FontWidthsConfig) FormatText(text string, maxWidth int, fontID string, gen types.Gen) (string, error) {
	if !fw.isFontIDValid(fontID) && len(fontID) > 0 && fontID != testFontID {
		validFontIDs := make([]string, len(fw.Fonts))
		i := 0
		for k := range fw.Fonts {
			validFontIDs[i] = k
			i++
		}
		return "", fmt.Errorf("Unknown fontID '%s' used in format(). List of valid fontIDs are '%s'", fontID, validFontIDs)
	}

	text = strings.ReplaceAll(text, "\n", " ")

	var formattedSb strings.Builder
	var curLineSb strings.Builder
	curWidth := 0
	isFirstLine := true
	isFirstWord := true
	pos := 0
	for pos < len(text) {
		endPos, word, err := fw.getNextWord(text[pos:], gen)
		fmt.Println(word)
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
			wordWidth += fw.getWordPixelWidth(word, fontID, gen)
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

func (fw *FontWidthsConfig) getNextWord(text string, gen types.Gen) (int, string, error) {
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
				if _, ok := genconfig.TextControlCodeStarts[gen][char]; ok {
					controlCodeLevel++
				} else if _, ok := genconfig.TextControlCodeEnds[gen][char]; ok {
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

func (fw *FontWidthsConfig) isLineBreak(word string) bool {
	return word == `\n` || word == `\l` || word == `\p`
}

func (fw *FontWidthsConfig) isParagraphBreak(word string) bool {
	return word == `\p`
}

func (fw *FontWidthsConfig) getWordPixelWidth(word string, fontID string, gen types.Gen) int {
	word, wordWidth := fw.processControlCodes(word, fontID, gen)
	for _, r := range word {
		wordWidth += fw.getRunePixelWidth(r, fontID)
	}
	return wordWidth
}

func (fw *FontWidthsConfig) processControlCodes(word string, fontID string, gen types.Gen) (string, int) {
	width := 0
	starts := []string{}
	for r := range genconfig.TextControlCodeStarts[gen] {
		starts = append(starts, string(r))
	}
	ends := []string{}
	for r := range genconfig.TextControlCodeEnds[gen] {
		ends = append(ends, string(r))
	}
	re := regexp.MustCompile(fmt.Sprintf("[%s][^}]*[%s]", strings.Join(starts, ""), strings.Join(ends, "")))
	positions := re.FindAllStringIndex(word, -1)
	for _, pos := range positions {
		code := word[pos[0]:pos[1]]
		width += fw.getControlCodePixelWidth(code, fontID)
	}
	strippedWord := re.ReplaceAllString(word, "")
	return strippedWord, width
}

func (fw *FontWidthsConfig) getRunePixelWidth(r rune, fontID string) int {
	if fontID == testFontID {
		return 10
	}
	return fw.readWidthFromFontConfig(string(r), fontID)
}

func (fw *FontWidthsConfig) getControlCodePixelWidth(code string, fontID string) int {
	if fontID == testFontID {
		return 100
	}
	return fw.readWidthFromFontConfig(code, fontID)
}

func (fw *FontWidthsConfig) isFontIDValid(fontID string) bool {
	_, ok := fw.Fonts[fontID]
	return ok
}

const fallbackWidth = 0

func (fw *FontWidthsConfig) readWidthFromFontConfig(value string, fontID string) int {
	font, ok := fw.Fonts[fontID]
	if !ok {
		return fallbackWidth
	}
	width, ok := font[value]
	if !ok {
		defaultWidth, ok := font["default"]
		if !ok {
			return fallbackWidth
		}
		return defaultWidth
	}
	return width
}
