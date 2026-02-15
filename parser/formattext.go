package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// TextReplacement defines a single text substitution rule that is applied to
// all string literals during parsing.
type TextReplacement struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	IsRegex     bool   `json:"regex"`
}

// compiledTextReplacement holds the pre-processed form of a TextReplacement.
type compiledTextReplacement struct {
	literal     string
	regex       *regexp.Regexp
	replacement string
}

// FontConfig holds the configuration for various supported fonts, as well as
// the default font.
type FontConfig struct {
	DefaultFontID    string            `json:"defaultFontId"`
	Fonts            map[string]Fonts  `json:"fonts"`
	TextReplacements []TextReplacement `json:"textReplacements"`
	compiled         []compiledTextReplacement
}

type Fonts struct {
	Widths             map[string]int `json:"widths"`
	CursorOverlapWidth int            `json:"cursorOverlapWidth"`
	MaxLineLength      int            `json:"maxLineLength"`
	NumLines           int            `json:"numLines"`
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

	if err := config.compileReplacements(); err != nil {
		return config, err
	}

	return config, err
}

// compileReplacements pre-compiles text replacement rules. Literal patterns
// are stored as-is; regex patterns are compiled into *regexp.Regexp.
func (fc *FontConfig) compileReplacements() error {
	fc.compiled = make([]compiledTextReplacement, len(fc.TextReplacements))
	for i, tr := range fc.TextReplacements {
		fc.compiled[i].replacement = tr.Replacement
		if tr.IsRegex {
			re, err := regexp.Compile(tr.Pattern)
			if err != nil {
				return fmt.Errorf("invalid regex pattern in font config's textReplacements[%d]: %s", i, err)
			}
			fc.compiled[i].regex = re
		} else {
			fc.compiled[i].literal = tr.Pattern
		}
	}
	return nil
}

// ApplyTextReplacements applies all configured text substitution rules to the
// given string, in the order they are defined in the configuration.
func (fc *FontConfig) ApplyTextReplacements(text string) string {
	for _, cr := range fc.compiled {
		if cr.regex != nil {
			text = cr.regex.ReplaceAllString(text, cr.replacement)
		} else {
			text = strings.ReplaceAll(text, cr.literal, cr.replacement)
		}
	}
	return text
}

// LineTooLongError describes a single line that exceeds the maximum pixel width.
type LineTooLongError struct {
	LineIndex      int
	LineText       string
	PixelWidth     int
	MaxWidth       int
	CharOffset     int
	Utf8CharOffset int
	CharLength     int
	Utf8CharLength int
}

// ValidateLineWidths checks each line of text against maxWidth and returns
// a list of lines that exceed the limit. Logical lines are delimited by
// real newline characters (inserted by the lexer for AUTOSTRINGs).
// Within each logical line, manual line-break escape sequences (\n, \l,
// \p) further subdivide the text; each sub-segment is validated
// independently but reported under the parent logical line's index.
//
// Unlike FormatText (which collapses spaces during word-wrapping), this
// function counts every space character toward the pixel width, since the
// text is rendered as-is in the game.
func (fc *FontConfig) ValidateLineWidths(text, fontID string, maxWidth int) []LineTooLongError {
	if maxWidth <= 0 {
		return nil
	}

	// Split on real newlines to get logical lines. For AUTOSTRINGs the
	// lexer inserts a real newline after each escape sequence, e.g.
	// "Line one\n\nLine two\l\n" splits into ["Line one\n", "Line two\l", ""].
	lines := strings.Split(text, "\n")

	var errors []LineTooLongError
	for i, line := range lines {
		content := stripTrailingLineBreak(line)

		// Split on any remaining manual line-break escapes and validate
		// each sub-segment independently.
		segments := splitOnLineBreakEscapes(content)
		for _, seg := range segments {
			width := fc.computeLinePixelWidth(seg.text, fontID)
			if width > maxWidth && len(seg.text) > 0 {
				errors = append(errors, LineTooLongError{
					LineIndex:      i,
					LineText:       seg.text,
					PixelWidth:     width,
					MaxWidth:       maxWidth,
					CharOffset:     seg.byteOffset,
					Utf8CharOffset: seg.runeOffset,
					CharLength:     len(seg.text),
					Utf8CharLength: seg.runeLength,
				})
			}
		}
	}

	return errors
}

func stripTrailingLineBreak(line string) string {
	if len(line) >= 2 {
		tail := line[len(line)-2:]
		if tail == `\n` || tail == `\l` || tail == `\p` {
			return line[:len(line)-2]
		}
	}
	return line
}

type lineSegment struct {
	text       string
	byteOffset int
	runeOffset int
	runeLength int
}

// splitOnLineBreakEscapes splits text on manual line-break escape sequences
// (\n, \l, \p) that appear outside of control codes ({...}).
// Returns the sub-segments between the breaks with their offsets.
func splitOnLineBreakEscapes(text string) []lineSegment {
	var segments []lineSegment
	controlCodeLevel := 0
	escape := false
	segStartByte := 0
	segStartRune := 0
	runeIndex := 0
	for i, ch := range text {
		if escape {
			if controlCodeLevel == 0 && (ch == 'n' || ch == 'l' || ch == 'p') {
				// The escape started at i - 1 (the backslash byte).
				seg := text[segStartByte : i-1]
				segments = append(segments, lineSegment{
					text:       seg,
					byteOffset: segStartByte,
					runeOffset: segStartRune,
					runeLength: runeIndex - 1 - segStartRune, // exclude backslash
				})
				segStartByte = i + 1
				segStartRune = runeIndex + 1
			}
			escape = false
			runeIndex++
			continue
		}
		if ch == '\\' && controlCodeLevel == 0 {
			escape = true
			runeIndex++
			continue
		}
		if ch == '{' {
			controlCodeLevel++
		} else if ch == '}' && controlCodeLevel > 0 {
			controlCodeLevel--
		}
		runeIndex++
	}
	// Append the final segment.
	seg := text[segStartByte:]
	segments = append(segments, lineSegment{
		text:       seg,
		byteOffset: segStartByte,
		runeOffset: segStartRune,
		runeLength: runeIndex - segStartRune,
	})
	return segments
}

// computeLinePixelWidth computes the pixel width of a single line of text,
// counting every character (including spaces) individually.
func (fc *FontConfig) computeLinePixelWidth(line, fontID string) int {
	width := 0
	controlCodeLevel := 0
	var controlCodeSb strings.Builder
	escape := false

	for _, ch := range line {
		if escape {
			// Non-line-break escape (line breaks already stripped/skipped).
			width += fc.getRunePixelWidth('\\', fontID)
			width += fc.getRunePixelWidth(ch, fontID)
			escape = false
			continue
		}

		if ch == '\\' && controlCodeLevel == 0 {
			escape = true
			continue
		}

		if ch == '{' {
			controlCodeLevel++
			controlCodeSb.WriteRune(ch)
			continue
		}
		if ch == '}' && controlCodeLevel > 0 {
			controlCodeSb.WriteRune(ch)
			controlCodeLevel--
			if controlCodeLevel == 0 {
				width += fc.getControlCodePixelWidth(controlCodeSb.String(), fontID)
				controlCodeSb.Reset()
			}
			continue
		}
		if controlCodeLevel > 0 {
			controlCodeSb.WriteRune(ch)
			continue
		}

		width += fc.getRunePixelWidth(ch, fontID)
	}

	return width
}

const testFontID = "TEST"

// FormatText automatically inserts line breaks into text
// according to in-game text box widths.
func (fc *FontConfig) FormatText(text string, maxWidth int, cursorOverlapWidth int, fontID string, numLines int) (string, error) {
	if !fc.isFontIDValid(fontID) && len(fontID) > 0 && fontID != testFontID {
		validFontIDs := make([]string, len(fc.Fonts))
		i := 0
		for k := range fc.Fonts {
			validFontIDs[i] = k
			i++
		}
		return "", fmt.Errorf("unknown fontID '%s' used in format(). List of valid fontIDs are '%s'", fontID, validFontIDs)
	}

	text = strings.ReplaceAll(text, "\n", " ")

	var formattedSb strings.Builder
	var curLineSb strings.Builder
	curWidth := 0
	curLineNum := 0
	isFirstWord := true
	spaceCharWidth := fc.getRunePixelWidth(' ', fontID)
	pos, word := fc.getNextWord(text)
	if len(word) == 0 {
		return formattedSb.String(), nil
	}

	for len(word) > 0 {
		endPos, nextWord := fc.getNextWord(text[pos:])
		pos += endPos
		if fc.isLineBreak(word) {
			curWidth = 0
			formattedSb.WriteString(curLineSb.String())
			if fc.isAutoLineBreak(word) {
				if curLineNum < numLines-1 {
					formattedSb.WriteString(`\n`)
				} else {
					formattedSb.WriteString(`\l`)
				}
			} else {
				formattedSb.WriteString(word)
			}
			formattedSb.WriteByte('\n')
			if fc.isParagraphBreak(word) {
				curLineNum = 0
			} else {
				curLineNum++
			}
			isFirstWord = true
			curLineSb.Reset()
		} else {
			wordWidth := fc.getWordPixelWidth(word, fontID)
			nextWordWidth := wordWidth
			if !isFirstWord {
				nextWordWidth += spaceCharWidth
			}
			nextWidth := curWidth + nextWordWidth
			// Technically, this isn't quite correct--especially if the cursorOverlapWidth is large, such that
			// it could span multiple words. The true solution would require optimistically trying to fit all
			// remaining words onto the same line, rather than only looking at the current word + cursor. However,
			// this is "good enough" and likely works for almost all actual use cases in practice.
			if len(nextWord) > 0 && (curLineNum >= numLines-1 || fc.isParagraphBreak(nextWord)) {
				nextWidth += cursorOverlapWidth
			}
			if nextWidth > maxWidth && curLineSb.Len() > 0 {
				formattedSb.WriteString(curLineSb.String())
				if fc.shouldUseLineFeed(curLineNum, numLines) {
					formattedSb.WriteString(`\l`)
				} else {
					formattedSb.WriteString(`\n`)
				}

				curLineNum++
				formattedSb.WriteByte('\n')
				isFirstWord = false
				curLineSb.Reset()
				curLineSb.WriteString(word)
				curWidth = wordWidth
			} else {
				curWidth += nextWordWidth
				if !isFirstWord {
					curLineSb.WriteByte(' ')
				}
				curLineSb.WriteString(word)
				isFirstWord = false
			}
		}
		word = nextWord
	}

	if curLineSb.Len() > 0 {
		formattedSb.WriteString(curLineSb.String())
	}

	return formattedSb.String(), nil
}

func (fc *FontConfig) shouldUseLineFeed(curLineNum int, numLines int) bool {
	return curLineNum >= numLines-1
}

func (fc *FontConfig) getNextWord(text string) (int, string) {
	escape := false
	endPos := 0
	startPos := 0
	foundNonSpace := false
	foundRegularRune := false
	endOnNext := false
	controlCodeLevel := 0
	for pos, char := range text {
		if endOnNext {
			return pos, text[startPos:pos]
		}
		if escape && (char == 'l' || char == 'n' || char == 'p' || char == 'N') {
			if foundRegularRune {
				return endPos, text[startPos:endPos]
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
					return pos, text[startPos:pos]
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
		return len(text), ""
	}
	return len(text), text[startPos:]
}

func (fc *FontConfig) isLineBreak(word string) bool {
	return word == `\n` || word == `\l` || word == `\p` || word == `\N`
}

func (fc *FontConfig) isAutoLineBreak(word string) bool {
	return word == `\N`
}

func (fc *FontConfig) isParagraphBreak(word string) bool {
	return word == `\p`
}

func (fc *FontConfig) getWordPixelWidth(word string, fontID string) int {
	word, wordWidth := fc.processControlCodes(word, fontID)
	for _, r := range word {
		wordWidth += fc.getRunePixelWidth(r, fontID)
	}
	return wordWidth
}

var controlCodeRegex = regexp.MustCompile(`{[^}]*}`)

func (fc *FontConfig) processControlCodes(word string, fontID string) (string, int) {
	width := 0
	positions := controlCodeRegex.FindAllStringIndex(word, -1)
	for _, pos := range positions {
		code := word[pos[0]:pos[1]]
		width += fc.getControlCodePixelWidth(code, fontID)
	}
	strippedWord := controlCodeRegex.ReplaceAllString(word, "")
	return strippedWord, width
}

func (fc *FontConfig) getRunePixelWidth(r rune, fontID string) int {
	if fontID == testFontID {
		return 10
	}
	return fc.getWidth(string(r), fontID)
}

func (fc *FontConfig) getControlCodePixelWidth(code string, fontID string) int {
	if fontID == testFontID {
		return 100
	}
	return fc.getWidth(code, fontID)
}

func (fc *FontConfig) isFontIDValid(fontID string) bool {
	_, ok := fc.Fonts[fontID]
	return ok
}

const fallbackWidth = 0

func (fc *FontConfig) getWidth(value string, fontID string) int {
	font, ok := fc.Fonts[fontID]
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
