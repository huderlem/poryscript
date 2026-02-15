package refactor

import (
	"fmt"
	"strings"

	"github.com/huderlem/poryscript/token"
)

// StringStyle represents one of the three source-level string formats in Poryscript.
type StringStyle int

const (
	// StyleAuto is a multi-line string where line-break commands (\n, \l, \p)
	// are automatically derived from newlines and blank lines in the source.
	StyleAuto StringStyle = iota
	// StyleConcatenated is multiple adjacent quoted strings, each with explicit
	// break codes (e.g., "line1\n" "line2\l" "line3").
	StyleConcatenated
	// StyleSingleLine is a single quoted string with explicit break codes inline.
	StyleSingleLine
)

type breakType int

const (
	breakNewline breakType = iota // \n
	breakScroll                   // \l
	breakPage                     // \p
)

type segment struct {
	text string
	brk  breakType
}

var breakEscapes = []struct {
	code string
	brk  breakType
}{
	{`\p`, breakPage},
	{`\n`, breakNewline},
	{`\l`, breakScroll},
}

func breakCode(b breakType) string {
	switch b {
	case breakNewline:
		return `\n`
	case breakScroll:
		return `\l`
	case breakPage:
		return `\p`
	}
	return ""
}

// DetectStringStyle examines raw source text (from opening " to closing ")
// and returns which string style it uses.
func DetectStringStyle(sourceText string) StringStyle {
	segments := extractQuotedSegments(sourceText)
	if len(segments) > 1 {
		return StyleConcatenated
	}
	if len(segments) == 1 && strings.ContainsAny(segments[0], "\r\n") {
		return StyleAuto
	}
	return StyleSingleLine
}

// ConvertString converts a Poryscript string from its current style to targetStyle.
// sourceText is the raw source text including quotes (e.g., `"line1\n" "line2"`).
// indent is the whitespace prefix for continuation lines in the output.
func ConvertString(sourceText string, targetStyle StringStyle, indent string) (string, error) {
	currentStyle := DetectStringStyle(sourceText)
	if currentStyle == targetStyle {
		return sourceText, nil
	}

	segments, err := parseToSegments(sourceText, currentStyle)
	if err != nil {
		return "", err
	}

	return emitFromSegments(segments, targetStyle, indent), nil
}

// extractQuotedSegments returns the raw content of each "..." segment in the source,
// preserving everything between the quotes (including newlines for auto strings).
func extractQuotedSegments(sourceText string) []string {
	var segments []string
	i := 0
	for i < len(sourceText) {
		if sourceText[i] == '"' {
			i++ // skip opening quote
			start := i
			for i < len(sourceText) && sourceText[i] != '"' {
				if sourceText[i] == '\\' && i+1 < len(sourceText) {
					i++ // skip escaped character
				}
				i++
			}
			segments = append(segments, sourceText[start:i])
			if i < len(sourceText) {
				i++ // skip closing quote
			}
		} else {
			i++
		}
	}
	return segments
}

// parseToSegments converts source text into the intermediate segment representation.
func parseToSegments(sourceText string, style StringStyle) ([]segment, error) {
	switch style {
	case StyleAuto:
		return parseAutoString(sourceText)
	case StyleConcatenated, StyleSingleLine:
		return parseManualString(sourceText)
	default:
		return nil, fmt.Errorf("unknown string style: %d", style)
	}
}

// parseAutoString parses a multi-line auto-formatted string into segments.
// It applies the same break-code derivation rules as the lexer.
func parseAutoString(sourceText string) ([]segment, error) {
	raw := extractQuotedSegments(sourceText)
	if len(raw) == 0 {
		return nil, fmt.Errorf("no quoted string found in source text")
	}

	// Split the single segment's content by actual newlines.
	lines := strings.Split(raw[0], "\n")
	// Normalize \r\n
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}

	// Strip leading whitespace from continuation lines and collect non-empty lines
	// with blank-line markers for paragraph detection.
	type lineInfo struct {
		text    string
		isBlank bool
	}
	var lineInfos []lineInfo
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if i == 0 {
			// First line: preserve as-is (no indentation stripping).
			lineInfos = append(lineInfos, lineInfo{text: line, isBlank: false})
		} else if trimmed == "" {
			lineInfos = append(lineInfos, lineInfo{text: "", isBlank: true})
		} else {
			lineInfos = append(lineInfos, lineInfo{text: trimmed, isBlank: false})
		}
	}

	// Remove trailing blank lines (from closing " on its own line).
	for len(lineInfos) > 1 && lineInfos[len(lineInfos)-1].isBlank {
		lineInfos = lineInfos[:len(lineInfos)-1]
	}

	// Build segments using the lexer's break-code rules.
	var segments []segment
	newParagraph := true

	for i, li := range lineInfos {
		if li.isBlank {
			continue
		}

		seg := segment{text: li.text}

		if i > 0 && len(segments) > 0 {
			// Determine what break code the previous segment should have.
			// Look back for blank lines between prev non-blank and this one.
			hasBlankBetween := false
			for j := i - 1; j >= 0; j-- {
				if lineInfos[j].isBlank {
					hasBlankBetween = true
					break
				}
				if !lineInfos[j].isBlank {
					break
				}
			}

			prevIdx := len(segments) - 1
			if hasBlankBetween {
				segments[prevIdx].brk = breakPage
				newParagraph = true
			} else if newParagraph {
				segments[prevIdx].brk = breakNewline
				newParagraph = false
			} else {
				segments[prevIdx].brk = breakScroll
			}
		}

		segments = append(segments, seg)
	}

	if len(segments) == 0 {
		segments = []segment{{text: ""}}
	}

	return segments, nil
}

// parseManualString parses concatenated or single-line strings into segments
// by splitting on \n, \l, \p escape codes.
func parseManualString(sourceText string) ([]segment, error) {
	raw := extractQuotedSegments(sourceText)
	if len(raw) == 0 {
		return nil, fmt.Errorf("no quoted string found in source text")
	}

	// Join all quoted segments into one text body.
	fullText := strings.Join(raw, "")

	return splitOnBreakCodes(fullText), nil
}

// splitOnBreakCodes splits text on \n, \l, \p escape sequences and returns segments.
func splitOnBreakCodes(text string) []segment {
	var segments []segment
	var current strings.Builder

	i := 0
	for i < len(text) {
		if text[i] == '\\' && i+1 < len(text) {
			ch := text[i+1]
			var brk breakType
			isBreak := false
			switch ch {
			case 'n':
				brk = breakNewline
				isBreak = true
			case 'l':
				brk = breakScroll
				isBreak = true
			case 'p':
				brk = breakPage
				isBreak = true
			}
			if isBreak {
				segments = append(segments, segment{text: current.String(), brk: brk})
				current.Reset()
				i += 2
				continue
			}
			// Not a break code — preserve the escape sequence as-is.
			current.WriteByte(text[i])
			current.WriteByte(text[i+1])
			i += 2
			continue
		}
		current.WriteByte(text[i])
		i++
	}

	// Last segment (after the final break code, or the only segment if no breaks).
	segments = append(segments, segment{text: current.String()})

	return segments
}

// emitFromSegments produces source text in the target style from segments.
func emitFromSegments(segments []segment, style StringStyle, indent string) string {
	switch style {
	case StyleAuto:
		return emitAutoString(segments, indent)
	case StyleConcatenated:
		return emitConcatenatedString(segments, indent)
	case StyleSingleLine:
		return emitSingleLineString(segments)
	}
	return ""
}

// emitAutoString produces a multi-line auto-formatted string.
func emitAutoString(segments []segment, indent string) string {
	if len(segments) == 0 {
		return `""`
	}

	var sb strings.Builder
	sb.WriteByte('"')
	for i, seg := range segments {
		if i > 0 {
			sb.WriteByte('\n')
			// For page breaks, insert a blank line.
			if segments[i-1].brk == breakPage {
				sb.WriteString(indent)
				sb.WriteByte('\n')
			}
			sb.WriteString(indent)
		}
		sb.WriteString(seg.text)
	}
	sb.WriteByte('"')
	return sb.String()
}

// emitConcatenatedString produces multiple adjacent quoted strings with explicit break codes.
func emitConcatenatedString(segments []segment, indent string) string {
	if len(segments) == 0 {
		return `""`
	}

	var sb strings.Builder
	for i, seg := range segments {
		if i > 0 {
			sb.WriteByte('\n')
			sb.WriteString(indent)
		}
		sb.WriteByte('"')
		sb.WriteString(seg.text)
		if i < len(segments)-1 {
			sb.WriteString(breakCode(seg.brk))
		}
		sb.WriteByte('"')
	}
	return sb.String()
}

// emitSingleLineString produces one quoted string with all break codes inline.
func emitSingleLineString(segments []segment) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for i, seg := range segments {
		sb.WriteString(seg.text)
		if i < len(segments)-1 {
			sb.WriteString(breakCode(seg.brk))
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// IsFormatStringToken reports whether the string token at index i in the
// token slice is inside a format() call.
func IsFormatStringToken(tokens []token.Token, i int) bool {
	if i < 2 {
		return false
	}
	prev := i - 1
	if tokens[prev].Type == token.STRINGTYPE {
		prev--
	}
	if prev < 1 {
		return false
	}
	return tokens[prev].Type == token.LPAREN && tokens[prev-1].Type == token.FORMAT
}

// FindStringTokenAtPosition returns the first STRING or AUTOSTRING token
// whose range contains the given position, along with its index in the
// tokens slice. The line and character parameters use 0-based indices
// (matching LSP conventions). Character is measured in UTF-8 code points.
func FindStringTokenAtPosition(tokens []token.Token, line, character int) (token.Token, int, bool) {
	for i, tok := range tokens {
		if !token.IsStringLikeToken(tok.Type) {
			continue
		}
		startLine := tok.LineNumber - 1
		endLine := tok.EndLineNumber - 1

		if line < startLine || line > endLine {
			continue
		}
		// Single-line token.
		if startLine == endLine {
			if character >= tok.StartUtf8CharIndex && character <= tok.EndUtf8CharIndex {
				return tok, i, true
			}
			continue
		}
		// Multi-line token.
		if line == startLine {
			if character >= tok.StartUtf8CharIndex {
				return tok, i, true
			}
			continue
		}
		if line == endLine {
			if character <= tok.EndUtf8CharIndex {
				return tok, i, true
			}
			continue
		}
		// Middle line — always inside.
		return tok, i, true
	}
	return token.Token{}, -1, false
}

// ExtractTokenSourceText extracts the raw source text for a token from the
// document content. It uses the token's byte-based character indices
// (StartCharIndex / EndCharIndex) and 1-based line numbers.
func ExtractTokenSourceText(content string, tok token.Token) string {
	lines := strings.Split(content, "\n")
	startLine := tok.LineNumber - 1
	endLine := tok.EndLineNumber - 1

	if startLine < 0 || startLine >= len(lines) || endLine < 0 || endLine >= len(lines) {
		return ""
	}

	if startLine == endLine {
		line := lines[startLine]
		start := tok.StartCharIndex
		end := tok.EndCharIndex
		if start > len(line) {
			start = len(line)
		}
		if end > len(line) {
			end = len(line)
		}
		return line[start:end]
	}

	// Multi-line: first line from StartCharIndex to end, middle lines fully,
	// last line to EndCharIndex.
	var sb strings.Builder
	firstLine := lines[startLine]
	start := tok.StartCharIndex
	if start > len(firstLine) {
		start = len(firstLine)
	}
	sb.WriteString(firstLine[start:])
	for i := startLine + 1; i < endLine; i++ {
		sb.WriteByte('\n')
		sb.WriteString(lines[i])
	}
	sb.WriteByte('\n')
	lastLine := lines[endLine]
	end := tok.EndCharIndex
	if end > len(lastLine) {
		end = len(lastLine)
	}
	sb.WriteString(lastLine[:end])
	return sb.String()
}

// ComputeConversionIndent returns the whitespace prefix for continuation
// lines when converting a string token to targetStyle. For auto strings,
// content aligns one character past the opening quote. For concatenated
// strings, each line's opening quote aligns with the first line's quote.
// Single-line strings don't use indentation, so an empty string is returned.
func ComputeConversionIndent(content string, tok token.Token, targetStyle StringStyle) string {
	if targetStyle == StyleSingleLine {
		return ""
	}
	lines := strings.Split(content, "\n")
	if tok.LineNumber-1 >= len(lines) {
		return ""
	}
	line := lines[tok.LineNumber-1]
	quoteCol := tok.StartCharIndex
	if quoteCol >= len(line) {
		return ""
	}
	if targetStyle == StyleConcatenated {
		return strings.Repeat(" ", quoteCol)
	}
	return strings.Repeat(" ", quoteCol+1)
}
