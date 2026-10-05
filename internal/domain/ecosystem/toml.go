// The shared TOML skeleton lexer of the ecosystem adapters: it lexes every
// table header, every key, and the value form of every key of a TOML
// document. The scanner is the contract surface of the adapters; it refuses
// every form it cannot interpret instead of silently misreading it. The
// adapter-specific key checks (duplicate keys, shadowed license surfaces)
// stay with their adapters and are invoked through the key check callback.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// ErrInvalidTOMLSurface marks a document the scanner cannot interpret
// structurally.
var ErrInvalidTOMLSurface = errors.New("invalid TOML surface")

// tomlKeyCheck is the adapter-specific key check invoked for every parsed
// key of a document scan.
type tomlKeyCheck func(table *tomlTable, key *tomlKeyValue) error

// scanTOMLDocument lexes the table skeleton of a TOML document: every table
// header, every key, and the value form of every key. A document the scanner
// cannot interpret structurally is refused fail-closed, and the adapter
// check receives every key of its table.
func scanTOMLDocument(content string, check tomlKeyCheck) ([]*tomlTable, error) {
	tables := []*tomlTable{}
	defined := map[string]bool{}
	current := &tomlTable{name: ""}
	tables = append(tables, current)
	i := 0
	for i < len(content) {
		i = skipTOMLBlank(content, i)
		if i >= len(content) {
			break
		}
		if content[i] == '[' {
			table, next, err := parseTOMLHeader(content, i, defined)
			if err != nil {
				return nil, err
			}
			tables = append(tables, table)
			current = table
			i = next
			continue
		}
		key, next, err := parseTOMLKeyValue(content, i)
		if err != nil {
			return nil, err
		}
		if err := check(current, key); err != nil {
			return nil, err
		}
		current.keys = append(current.keys, *key)
		i = next
	}
	return tables, nil
}

// tomlValueKind classifies the parsed form of a TOML value.
type tomlValueKind int

const (
	tomlValueString tomlValueKind = iota
	tomlValueArray
	tomlValueInlineTable
	tomlValueAtom
)

type tomlArrayItem struct {
	isString bool
	decoded  string
}

type tomlValue struct {
	kind  tomlValueKind
	str   string
	items []tomlArrayItem
}

type tomlKeyValue struct {
	key        string
	segments   []string
	keyStart   int // the index of the first byte of the key
	valueStart int
	valueEnd   int
	lineEnd    int // the index just past the key's line terminator
	kind       tomlValueKind
	str        string
	items      []tomlArrayItem
}

type tomlTable struct {
	name      string
	array     bool
	start     int
	headerEnd int // the index just past the header line terminator
	keys      []tomlKeyValue
}

// parseTOMLHeader parses one table header and returns the index just past
// its line terminator. A repeated non-array table is refused fail-closed.
func parseTOMLHeader(content string, start int, defined map[string]bool) (*tomlTable, int, error) {
	i := start + 1
	array := false
	if i < len(content) && content[i] == '[' {
		array = true
		i++
	}
	segments := []string{}
	for {
		i = skipTOMLSpaces(content, i)
		segment, next, err := parseTOMLKeySegment(content, i)
		if err != nil {
			return nil, 0, err
		}
		segments = append(segments, segment)
		i = skipTOMLSpaces(content, next)
		if i < len(content) && content[i] == '.' {
			i++
			continue
		}
		break
	}
	if i >= len(content) || content[i] != ']' {
		return nil, 0, fmt.Errorf("%w: malformed table header", ErrInvalidTOMLSurface)
	}
	i++
	if array {
		if i >= len(content) || content[i] != ']' {
			return nil, 0, fmt.Errorf("%w: malformed array table header", ErrInvalidTOMLSurface)
		}
		i++
	}
	i = skipTOMLSpaces(content, i)
	if i < len(content) && content[i] == '#' {
		i = skipTOMLComment(content, i)
	}
	table := &tomlTable{name: strings.Join(segments, "."), array: array, start: start}
	switch {
	case i >= len(content):
		table.headerEnd = len(content)
	case content[i] == '\n':
		table.headerEnd = i + 1
		i++
	case content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n':
		table.headerEnd = i + 2
		i += 2
	default:
		return nil, 0, fmt.Errorf("%w: trailing content after the table header", ErrInvalidTOMLSurface)
	}
	if !array {
		if defined[table.name] {
			return nil, 0, fmt.Errorf("%w: duplicate table [%s]", ErrInvalidTOMLSurface, table.name)
		}
		defined[table.name] = true
	}
	return table, i, nil
}

// parseTOMLKeyValue parses one key-value line and returns the index just
// past its line terminator.
func parseTOMLKeyValue(content string, start int) (*tomlKeyValue, int, error) {
	i := start
	segments := []string{}
	for {
		i = skipTOMLSpaces(content, i)
		segment, next, err := parseTOMLKeySegment(content, i)
		if err != nil {
			return nil, 0, err
		}
		segments = append(segments, segment)
		i = skipTOMLSpaces(content, next)
		if i < len(content) && content[i] == '.' {
			i++
			continue
		}
		break
	}
	name := strings.Join(segments, ".")
	if i >= len(content) || content[i] != '=' {
		return nil, 0, fmt.Errorf("%w: expected '=' after the key %q", ErrInvalidTOMLSurface, name)
	}
	i = skipTOMLSpaces(content, i+1)
	valueStart := i
	value, valueEnd, err := parseTOMLValue(content, i)
	if err != nil {
		return nil, 0, err
	}
	i = skipTOMLSpaces(content, valueEnd)
	if i < len(content) && content[i] == '#' {
		i = skipTOMLComment(content, i)
	}
	lineEnd := i
	switch {
	case i >= len(content):
	case content[i] == '\n':
		i++
		lineEnd = i
	case content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n':
		i += 2
		lineEnd = i
	default:
		return nil, 0, fmt.Errorf("%w: trailing content after the value of %q", ErrInvalidTOMLSurface, name)
	}
	return &tomlKeyValue{
		key:        name,
		segments:   segments,
		keyStart:   start,
		valueStart: valueStart,
		valueEnd:   valueEnd,
		lineEnd:    lineEnd,
		kind:       value.kind,
		str:        value.str,
		items:      value.items,
	}, i, nil
}

// parseTOMLKeySegment parses one dotted-key segment: a bare key, a basic
// string, or a literal string.
func parseTOMLKeySegment(content string, start int) (string, int, error) {
	if start >= len(content) {
		return "", 0, fmt.Errorf("%w: expected a key", ErrInvalidTOMLSurface)
	}
	switch content[start] {
	case '"':
		decoded, end, err := lexTOMLBasicString(content, start)
		if err != nil {
			return "", 0, err
		}
		return decoded, end, nil
	case '\'':
		decoded, end, err := lexTOMLLiteralString(content, start)
		if err != nil {
			return "", 0, err
		}
		return decoded, end, nil
	}
	i := start
	for i < len(content) && isTOMLBareKeyByte(content[i]) {
		i++
	}
	if i == start {
		return "", 0, fmt.Errorf("%w: expected a key", ErrInvalidTOMLSurface)
	}
	return content[start:i], i, nil
}

func isTOMLBareKeyByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
}

// parseTOMLValue parses one TOML value and returns the index just past it.
func parseTOMLValue(content string, start int) (tomlValue, int, error) {
	if start >= len(content) {
		return tomlValue{}, 0, fmt.Errorf("%w: expected a value", ErrInvalidTOMLSurface)
	}
	switch content[start] {
	case '"':
		if strings.HasPrefix(content[start:], `"""`) {
			decoded, end, err := lexTOMLMultilineBasicString(content, start)
			if err != nil {
				return tomlValue{}, 0, err
			}
			return tomlValue{kind: tomlValueString, str: decoded}, end, nil
		}
		decoded, end, err := lexTOMLBasicString(content, start)
		if err != nil {
			return tomlValue{}, 0, err
		}
		return tomlValue{kind: tomlValueString, str: decoded}, end, nil
	case '\'':
		if strings.HasPrefix(content[start:], `'''`) {
			decoded, end, err := lexTOMLMultilineLiteralString(content, start)
			if err != nil {
				return tomlValue{}, 0, err
			}
			return tomlValue{kind: tomlValueString, str: decoded}, end, nil
		}
		decoded, end, err := lexTOMLLiteralString(content, start)
		if err != nil {
			return tomlValue{}, 0, err
		}
		return tomlValue{kind: tomlValueString, str: decoded}, end, nil
	case '[':
		items, end, err := lexTOMLArray(content, start)
		if err != nil {
			return tomlValue{}, 0, err
		}
		return tomlValue{kind: tomlValueArray, items: items}, end, nil
	case '{':
		end, err := lexTOMLInlineTable(content, start)
		if err != nil {
			return tomlValue{}, 0, err
		}
		if strings.ContainsAny(content[start:end], "\n\r") {
			return tomlValue{}, 0, fmt.Errorf("%w: newlines are not allowed inside an inline table", ErrInvalidTOMLSurface)
		}
		return tomlValue{kind: tomlValueInlineTable}, end, nil
	default:
		i := start
		for i < len(content) && !isTOMLValueDelimiter(content[i]) {
			i++
		}
		if i == start {
			return tomlValue{}, 0, fmt.Errorf("%w: expected a value", ErrInvalidTOMLSurface)
		}
		return tomlValue{kind: tomlValueAtom}, i, nil
	}
}

func isTOMLValueDelimiter(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', ',', ']', '}', '#':
		return true
	}
	return false
}

// lexTOMLArray lexes a TOML array and records every item with its string
// state.
func lexTOMLArray(content string, start int) ([]tomlArrayItem, int, error) {
	i := start + 1
	items := []tomlArrayItem{}
	for {
		i = skipTOMLBlank(content, i)
		if i >= len(content) {
			return nil, 0, fmt.Errorf("%w: unterminated array", ErrInvalidTOMLSurface)
		}
		if content[i] == ']' {
			return items, i + 1, nil
		}
		value, next, err := parseTOMLValue(content, i)
		if err != nil {
			return nil, 0, err
		}
		if value.kind == tomlValueString {
			items = append(items, tomlArrayItem{isString: true, decoded: value.str})
		} else {
			items = append(items, tomlArrayItem{})
		}
		i = skipTOMLBlank(content, next)
		if i >= len(content) {
			return nil, 0, fmt.Errorf("%w: unterminated array", ErrInvalidTOMLSurface)
		}
		switch content[i] {
		case ',':
			i++
		case ']':
			return items, i + 1, nil
		default:
			return nil, 0, fmt.Errorf("%w: expected ',' or ']' in an array", ErrInvalidTOMLSurface)
		}
	}
}

// lexTOMLInlineTable lexes a TOML inline table; the newline check of the
// caller applies the TOML inline-table law after the lex.
func lexTOMLInlineTable(content string, start int) (int, error) {
	i := start + 1
	for {
		i = skipTOMLSpaces(content, i)
		if i >= len(content) || content[i] == '\n' || content[i] == '\r' {
			return 0, fmt.Errorf("%w: unterminated or multi-line inline table", ErrInvalidTOMLSurface)
		}
		if content[i] == '}' {
			return i + 1, nil
		}
		_, next, err := parseTOMLKeySegment(content, i)
		if err != nil {
			return 0, err
		}
		i = skipTOMLSpaces(content, next)
		if i >= len(content) || content[i] != '=' {
			return 0, fmt.Errorf("%w: expected '=' in an inline table", ErrInvalidTOMLSurface)
		}
		i = skipTOMLSpaces(content, i+1)
		_, valueEnd, err := parseTOMLValue(content, i)
		if err != nil {
			return 0, err
		}
		i = skipTOMLSpaces(content, valueEnd)
		if i >= len(content) || content[i] == '\n' || content[i] == '\r' {
			return 0, fmt.Errorf("%w: unterminated or multi-line inline table", ErrInvalidTOMLSurface)
		}
		switch content[i] {
		case ',':
			i++
		case '}':
			return i + 1, nil
		default:
			return 0, fmt.Errorf("%w: expected ',' or '}' in an inline table", ErrInvalidTOMLSurface)
		}
	}
}

// lexTOMLBasicString lexes a single-line basic string with its escapes.
func lexTOMLBasicString(content string, start int) (string, int, error) {
	var builder strings.Builder
	i := start + 1
	for {
		if i >= len(content) {
			return "", 0, fmt.Errorf("%w: unterminated string", ErrInvalidTOMLSurface)
		}
		c := content[i]
		switch {
		case c == '"':
			return builder.String(), i + 1, nil
		case c == '\\':
			decoded, next, err := lexTOMLEscape(content, i)
			if err != nil {
				return "", 0, err
			}
			builder.WriteString(decoded)
			i = next
		case c < 0x20 && c != '\t':
			return "", 0, fmt.Errorf("%w: control character inside a string", ErrInvalidTOMLSurface)
		default:
			builder.WriteByte(c)
			i++
		}
	}
}

// lexTOMLMultilineBasicString lexes a multi-line basic string with its
// immediate-newline trims and line-ending backslash form.
func lexTOMLMultilineBasicString(content string, start int) (string, int, error) {
	var builder strings.Builder
	i := start + 3
	if i < len(content) && content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n' {
		i += 2
	} else if i < len(content) && content[i] == '\n' {
		i++
	}
	for {
		if i >= len(content) {
			return "", 0, fmt.Errorf("%w: unterminated multi-line string", ErrInvalidTOMLSurface)
		}
		switch {
		case strings.HasPrefix(content[i:], `"""`):
			return builder.String(), i + 3, nil
		case content[i] == '\\':
			j := i + 1
			k := j
			sawNewline := false
			for k < len(content) && (content[k] == ' ' || content[k] == '\t' || content[k] == '\n' || content[k] == '\r') {
				if content[k] == '\n' || content[k] == '\r' {
					sawNewline = true
				}
				k++
			}
			if k > j && sawNewline {
				if k >= len(content) {
					return "", 0, fmt.Errorf("%w: unterminated multi-line string", ErrInvalidTOMLSurface)
				}
				i = k
				continue
			}
			decoded, next, err := lexTOMLEscape(content, i)
			if err != nil {
				return "", 0, err
			}
			builder.WriteString(decoded)
			i = next
		default:
			builder.WriteByte(content[i])
			i++
		}
	}
}

// lexTOMLLiteralString lexes a single-line literal string without escapes.
func lexTOMLLiteralString(content string, start int) (string, int, error) {
	i := start + 1
	for {
		if i >= len(content) {
			return "", 0, fmt.Errorf("%w: unterminated string", ErrInvalidTOMLSurface)
		}
		c := content[i]
		if c == '\'' {
			return content[start+1 : i], i + 1, nil
		}
		if c == '\n' || c == '\r' || c < 0x20 && c != '\t' {
			return "", 0, fmt.Errorf("%w: invalid character inside a literal string", ErrInvalidTOMLSurface)
		}
		i++
	}
}

// lexTOMLMultilineLiteralString lexes a multi-line literal string without
// escapes.
func lexTOMLMultilineLiteralString(content string, start int) (string, int, error) {
	i := start + 3
	if i < len(content) && content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n' {
		i += 2
	} else if i < len(content) && content[i] == '\n' {
		i++
	}
	begin := i
	for {
		if i >= len(content) {
			return "", 0, fmt.Errorf("%w: unterminated multi-line literal string", ErrInvalidTOMLSurface)
		}
		if strings.HasPrefix(content[i:], `'''`) {
			return content[begin:i], i + 3, nil
		}
		i++
	}
}

// lexTOMLEscape decodes one escape sequence of a basic string.
func lexTOMLEscape(content string, backslash int) (string, int, error) {
	if backslash+1 >= len(content) {
		return "", 0, fmt.Errorf("%w: unterminated escape sequence", ErrInvalidTOMLSurface)
	}
	switch content[backslash+1] {
	case 'b':
		return "\b", backslash + 2, nil
	case 't':
		return "\t", backslash + 2, nil
	case 'n':
		return "\n", backslash + 2, nil
	case 'f':
		return "\f", backslash + 2, nil
	case 'r':
		return "\r", backslash + 2, nil
	case '"':
		return "\"", backslash + 2, nil
	case '\\':
		return "\\", backslash + 2, nil
	case 'u':
		return lexTOMLUnicodeEscape(content, backslash, 4)
	case 'U':
		return lexTOMLUnicodeEscape(content, backslash, 8)
	default:
		return "", 0, fmt.Errorf("%w: unknown escape sequence", ErrInvalidTOMLSurface)
	}
}

// lexTOMLUnicodeEscape decodes a unicode escape sequence and refuses the
// surrogate and overflow forms.
func lexTOMLUnicodeEscape(content string, backslash int, digits int) (string, int, error) {
	if backslash+2+digits > len(content) {
		return "", 0, fmt.Errorf("%w: short unicode escape sequence", ErrInvalidTOMLSurface)
	}
	value := 0
	for k := 0; k < digits; k++ {
		c := content[backslash+2+k]
		var digit int
		switch {
		case c >= '0' && c <= '9':
			digit = int(c - '0')
		case c >= 'a' && c <= 'f':
			digit = int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			digit = int(c-'A') + 10
		default:
			return "", 0, fmt.Errorf("%w: invalid unicode escape sequence", ErrInvalidTOMLSurface)
		}
		value = value<<4 | digit
	}
	if value > 0x10FFFF || value >= 0xD800 && value <= 0xDFFF {
		return "", 0, fmt.Errorf("%w: invalid unicode escape sequence", ErrInvalidTOMLSurface)
	}
	return string(rune(value)), backslash + 2 + digits, nil
}

func skipTOMLBlank(content string, i int) int {
	for i < len(content) {
		switch content[i] {
		case ' ', '\t', '\n', '\r':
			i++
		case '#':
			i = skipTOMLComment(content, i)
		default:
			return i
		}
	}
	return i
}

func skipTOMLSpaces(content string, i int) int {
	for i < len(content) && (content[i] == ' ' || content[i] == '\t') {
		i++
	}
	return i
}

func skipTOMLComment(content string, i int) int {
	for i < len(content) && content[i] != '\n' {
		i++
	}
	return i
}

// tomlEdit carries one byte-span replacement or insertion.
type tomlEdit struct {
	start int
	end   int
	text  string
}

// applyTOMLEdits applies the edits from the highest offset down so the
// earlier spans stay valid.
func applyTOMLEdits(content string, edits []tomlEdit) string {
	slices.SortFunc(edits, func(a, b tomlEdit) int { return b.start - a.start })
	result := content
	for _, edit := range edits {
		result = result[:edit.start] + edit.text + result[edit.end:]
	}
	return result
}

// tomlLineInsertion builds a whole-line insertion at offset; a position that
// does not follow a newline gets one so the inserted lines start on their
// own line.
func tomlLineInsertion(content string, offset int, text string) tomlEdit {
	if offset > 0 && offset <= len(content) && content[offset-1] != '\n' {
		text = "\n" + text
	}
	return tomlEdit{start: offset, end: offset, text: text}
}

// tomlQuote renders a TOML basic string.
func tomlQuote(value string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, r := range value {
		switch r {
		case '"':
			builder.WriteString(`\"`)
		case '\\':
			builder.WriteString(`\\`)
		case '\b':
			builder.WriteString(`\b`)
		case '\t':
			builder.WriteString(`\t`)
		case '\n':
			builder.WriteString(`\n`)
		case '\f':
			builder.WriteString(`\f`)
		case '\r':
			builder.WriteString(`\r`)
		default:
			if r < 0x20 {
				builder.WriteString(fmt.Sprintf(`\u%04X`, r))
			} else {
				builder.WriteRune(r)
			}
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

// tomlStringArray renders a TOML array of strings.
func tomlStringArray(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, tomlQuote(item))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
