// The Ruby ecosystem adapter: it derives and aligns the license metadata
// surface of a gemspec file from the declaration chain and the license lock.
// The declaration is the truth, the manifest field is its deterministic
// projection, and the license lock is the single source of truth of the
// license type. The gemspec carries the license declaration in two
// value-equal assignment forms — the singular string form
// spec.license = "MIT" and the plural array form spec.licenses = ["MIT"] —
// and every entry is a single SPDX identifier: compound expressions and
// entries longer than 64 characters are not supported per entry, and
// LicenseRef-<idstring> is the documented non-SPDX entry form.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// RubyLanguage is the seam toolchain language that selects the Ruby/gemspec
// matrix row.
const RubyLanguage = "ruby"

// The Ruby license metadata surface: the manifest suffix the adapter
// discovers, and the RubyGems entry limit of a single SPDX identifier.
const (
	GemspecSuffix         = ".gemspec"
	GemspecMaxEntryLength = 64
)

// ErrInvalidGemspecSurface marks a gemspec the scanner cannot lex or anchor:
// a structural break anywhere, an uninterpretable or chained assignment
// receiver, an assignment without a value, or a missing specification block
// for the insertion anchor.
var ErrInvalidGemspecSurface = errors.New("invalid gemspec surface")

// ErrMultipleGemspecLicenses marks a license array that declares more than
// one entry. The projection carries exactly one license family; the
// resolution is an explicit tenant decision.
var ErrMultipleGemspecLicenses = errors.New("gemspec declares multiple license entries")

// ErrAmbiguousGemspecLicense marks a gemspec that declares the license
// assignment more than once, in any mix of the singular and the plural form.
var ErrAmbiguousGemspecLicense = errors.New("gemspec declares the license assignment more than once")

// ErrInvalidGemspecLicenseValue marks a license value that is not a license
// literal: a non-literal value, an escape or an interpolation inside the
// literal, an empty array, or a dynamic suffix other than .freeze.
var ErrInvalidGemspecLicenseValue = errors.New("gemspec license value is not a license literal")

// GemspecProjection derives the expected gemspec license entry from the
// merged tenant values: a declared SPDX identifier projects itself; the
// file-based custom family projects the SPDX sideload entry
// LicenseRef-<LICENSE_ID> — the documented non-SPDX entry form keeps the
// license family identity. The derivation never reads the manifest.
func GemspecProjection(merged map[string]string) string {
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		return identifier
	}
	return "LicenseRef-" + merged["LICENSE_ID"]
}

// GemspecManifestNames filters a directory listing down to the gemspec
// manifest candidates, in sorted order. The suffix match is case-insensitive.
func GemspecManifestNames(entries []string) []string {
	names := []string{}
	for _, entry := range entries {
		if strings.HasSuffix(strings.ToLower(entry), GemspecSuffix) {
			names = append(names, entry)
		}
	}
	slices.Sort(names)
	return names
}

// GemspecLicenseState classifies the license assignment of a scanned gemspec.
type GemspecLicenseState int

const (
	// GemspecLicenseMissing marks an absent license assignment.
	GemspecLicenseMissing GemspecLicenseState = iota
	// GemspecLicenseString marks the target form: the singular string
	// assignment.
	GemspecLicenseString
	// GemspecLicenseArray marks the plural array form; the declared value of
	// a one-element array is its single entry.
	GemspecLicenseArray
	// GemspecLicenseInvalid marks a declared value that is not a license
	// literal.
	GemspecLicenseInvalid
)

// GemspecLicenseSurface is the license-relevant view of a scanned gemspec.
type GemspecLicenseSurface struct {
	// State classifies the license assignment.
	State GemspecLicenseState
	// Value carries the declared license entry: the string form itself or the
	// single entry of a one-element array.
	Value string
	// Entries carries the decoded entries when the form is the array.
	Entries []string
}

// rubyAssignmentKind classifies the parsed value expression of one license
// assignment site.
type rubyAssignmentKind int

const (
	rubyValueString rubyAssignmentKind = iota
	rubyValueArray
	rubyValueInvalid
)

// rubyAssignment carries one captured license assignment site: the receiver
// identifier, the decoded value, and the byte spans of the assignment from
// the receiver start to the value end.
type rubyAssignment struct {
	receiver   string
	kind       rubyAssignmentKind
	value      string
	entries    []string
	start      int
	valueStart int
	valueEnd   int
}

// InspectGemspecLicense lexes a gemspec file and classifies its license
// assignment. A file that declares the assignment more than once, in any mix
// of the singular and the plural form, is refused fail-closed.
func InspectGemspecLicense(content string) (GemspecLicenseSurface, error) {
	sites, err := scanGemspec(content)
	if err != nil {
		return GemspecLicenseSurface{}, err
	}
	if len(sites) > 1 {
		return GemspecLicenseSurface{}, ErrAmbiguousGemspecLicense
	}
	if len(sites) == 0 {
		return GemspecLicenseSurface{State: GemspecLicenseMissing}, nil
	}
	site := sites[0]
	switch site.kind {
	case rubyValueString:
		return GemspecLicenseSurface{State: GemspecLicenseString, Value: site.value}, nil
	case rubyValueArray:
		if len(site.entries) == 0 {
			return GemspecLicenseSurface{}, ErrInvalidGemspecLicenseValue
		}
		if len(site.entries) == 1 {
			return GemspecLicenseSurface{State: GemspecLicenseArray, Value: site.entries[0], Entries: site.entries}, nil
		}
		return GemspecLicenseSurface{State: GemspecLicenseArray, Entries: site.entries}, nil
	default:
		return GemspecLicenseSurface{State: GemspecLicenseInvalid}, nil
	}
}

// AlignGemspecLicense returns the gemspec content with its license assignment
// aligned to target. The projected form is the singular string assignment: a
// diverging string value is value-replaced, a diverging one-element array is
// normalized to the singular assignment with the observed receiver preserved,
// and the one-element array carrying the projected value is the value-equal
// alternative spelling of the same declaration. An array that declares more
// than one entry and a file that declares the assignment more than once are
// refused fail-closed, because the projection carries exactly one license
// family and the resolution is an explicit tenant decision. Every byte
// outside the replaced value span, the replaced assignment span, or the
// inserted line is preserved exactly; a re-run over an aligned file is a
// no-op. A non-literal license value is refused fail-closed.
func AlignGemspecLicense(content, target string) (string, bool, error) {
	if err := validateGemspecTarget(target); err != nil {
		return "", false, err
	}
	sites, err := scanGemspec(content)
	if err != nil {
		return "", false, err
	}
	if len(sites) > 1 {
		return "", false, ErrAmbiguousGemspecLicense
	}
	if len(sites) == 0 {
		return insertGemspecLicense(content, target)
	}
	site := sites[0]
	switch site.kind {
	case rubyValueString:
		if site.value == target {
			return content, false, nil
		}
		return content[:site.valueStart] + rubyLicenseLiteral(target) + content[site.valueEnd:], true, nil
	case rubyValueArray:
		if len(site.entries) == 0 {
			return "", false, ErrInvalidGemspecLicenseValue
		}
		if len(site.entries) > 1 {
			return "", false, ErrMultipleGemspecLicenses
		}
		if site.entries[0] == target {
			return content, false, nil
		}
		replacement := site.receiver + ".license = " + rubyLicenseLiteral(target)
		return content[:site.start] + replacement + content[site.valueEnd:], true, nil
	default:
		return "", false, ErrInvalidGemspecLicenseValue
	}
}

// rubyLicenseLiteral renders the double-quoted Ruby string literal of the
// target value. The target guard guarantees an escape-free literal.
func rubyLicenseLiteral(target string) string {
	return "\"" + target + "\""
}

// validateGemspecTarget refuses alignment targets the RubyGems entry contract
// cannot carry: a non-UTF-8 value, an empty entry, an entry beyond the
// 64-character limit, a compound expression (whitespace inside the entry),
// and a value that is not representable as an escape-free,
// interpolation-free Ruby string literal.
func validateGemspecTarget(target string) error {
	if !utf8.ValidString(target) || target == "" || len(target) > GemspecMaxEntryLength {
		return ErrInvalidTarget
	}
	for _, r := range target {
		switch {
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			return fmt.Errorf("%w: the license entry carries whitespace (compound expressions are not supported per entry)", ErrInvalidTarget)
		case r == '"' || r == '\\' || r == '#':
			return fmt.Errorf("%w: the license entry is not representable as an escape-free Ruby string literal", ErrInvalidTarget)
		}
	}
	return nil
}

// insertGemspecLicense inserts the missing license assignment as the first
// statement of the Gem::Specification block, using the block parameter as the
// receiver and the block body's indentation. The inserted surface is
// re-scanned and must provably carry the target; an anchor that does not
// produce a provable assignment is refused fail-closed.
func insertGemspecLicense(content, target string) (string, bool, error) {
	receiver, pastPipe, lineStart, ok := findGemspecBlockAnchor(content)
	if !ok {
		return "", false, fmt.Errorf("%w: the document carries no Gem::Specification block to anchor the insertion", ErrInvalidGemspecSurface)
	}
	lineEnd := pastPipe
	for lineEnd < len(content) && content[lineEnd] != '\n' {
		lineEnd++
	}
	bodyIndent := rubyBlockBodyIndent(content, pastPipe, lineEnd, lineStart)
	assignment := bodyIndent + receiver + ".license = " + rubyLicenseLiteral(target)
	var aligned string
	if lineEnd < len(content) {
		aligned = content[:lineEnd+1] + assignment + "\n" + content[lineEnd+1:]
	} else {
		aligned = content + "\n" + assignment
	}
	sites, err := scanGemspec(aligned)
	if err != nil || len(sites) != 1 || sites[0].kind != rubyValueString || sites[0].value != target {
		return "", false, fmt.Errorf("%w: the insertion anchor does not produce a provable license assignment", ErrInvalidGemspecSurface)
	}
	return aligned, true, nil
}

// rubyBlockBodyIndent derives the indentation of the inserted assignment: the
// leading whitespace of the first non-blank block body line, or the do-line
// indentation plus two spaces when the block carries no body line.
func rubyBlockBodyIndent(content string, pastPipe, lineEnd, lineStart int) string {
	if lineEnd < len(content) {
		i := lineEnd + 1
		for i < len(content) {
			rowEnd := i
			for rowEnd < len(content) && content[rowEnd] != '\n' {
				rowEnd++
			}
			trimmed := strings.TrimLeft(content[i:rowEnd], " \t")
			if trimmed != "" {
				indent := content[i : i+len(content[i:rowEnd])-len(trimmed)]
				if trimmed == "end" {
					return indent + "  "
				}
				return indent
			}
			if rowEnd >= len(content) {
				break
			}
			i = rowEnd + 1
		}
	}
	doIndent := lineStart
	for doIndent < pastPipe && (content[doIndent] == ' ' || content[doIndent] == '\t') {
		doIndent++
	}
	return content[lineStart:doIndent] + "  "
}

// findGemspecBlockAnchor locates the first Gem::Specification block and
// returns its block parameter, the index just past the parameter pipe, and
// the start of the line carrying the block.
func findGemspecBlockAnchor(content string) (string, int, int, bool) {
	idx := strings.Index(content, "Gem::Specification.new")
	if idx < 0 {
		return "", 0, 0, false
	}
	lineStart := idx
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}
	j := idx + len("Gem::Specification.new")
	j = skipRubySpaceTab(content, j)
	if !strings.HasPrefix(content[j:], "do") {
		return "", 0, 0, false
	}
	j = skipRubySpaceTab(content, j+2)
	if j >= len(content) || content[j] != '|' {
		return "", 0, 0, false
	}
	j++
	start := j
	for j < len(content) && rubyIdentByte(content[j]) {
		j++
	}
	if j == start || j >= len(content) || content[j] != '|' {
		return "", 0, 0, false
	}
	return content[start:j], j + 1, lineStart, true
}

// scanGemspec lexes the license-relevant skeleton of a gemspec file: line
// comments, string literals, percent literals, heredocs, and =begin/=end
// block comments are context, so a license assignment inside them is never
// observed. The license assignment sites of the form
// <receiver>.license = <value> and <receiver>.licenses = <value> are
// captured with their exact spans. A structural break anywhere — an
// unterminated string, percent literal, heredoc, or block comment — refuses
// the whole surface fail-closed.
func scanGemspec(content string) ([]rubyAssignment, error) {
	sites := []rubyAssignment{}
	lineStart := true
	i := 0
	for i < len(content) {
		c := content[i]
		switch {
		case c == '#':
			i = skipRubyComment(content, i)
		case c == '"' || c == '\'':
			end, err := parseRubyString(content, i)
			if err != nil {
				return nil, err
			}
			i = end
			lineStart = false
		case c == '%' && isRubyPercentLiteralStart(content, i):
			end, err := skipRubyPercentLiteral(content, i)
			if err != nil {
				return nil, err
			}
			i = end
			lineStart = false
		case c == '<' && isRubyHeredocStart(content, i):
			end, err := skipRubyHeredoc(content, i)
			if err != nil {
				return nil, err
			}
			i = end
			lineStart = false
		case c == '=' && lineStart && isRubyBlockCommentStart(content, i):
			end, err := skipRubyBlockComment(content, i)
			if err != nil {
				return nil, err
			}
			i = end
			lineStart = false
		case c == '.':
			site, next, matched, err := rubyAssignmentAt(content, i)
			if err != nil {
				return nil, err
			}
			if matched {
				sites = append(sites, site)
			}
			i = next
			lineStart = false
		case c == '\n':
			i++
			lineStart = true
		default:
			i++
			lineStart = false
		}
	}
	return sites, nil
}

// rubyAssignmentAt inspects one dot token: when it opens a license
// assignment of the form <receiver>.license = <value> or
// <receiver>.licenses = <value>, the site is returned with its parsed value
// and the index just past the value expression.
func rubyAssignmentAt(content string, dot int) (rubyAssignment, int, bool, error) {
	rest := content[dot+1:]
	var keyEnd int
	switch {
	case strings.HasPrefix(rest, "licenses") && !rubyIdentByteAt(content, dot+9):
		keyEnd = dot + 9
	case strings.HasPrefix(rest, "license") && !rubyIdentByteAt(content, dot+8):
		keyEnd = dot + 8
	default:
		return rubyAssignment{}, dot + 1, false, nil
	}
	recvStart := dot
	for recvStart > 0 && rubyIdentByte(content[recvStart-1]) {
		recvStart--
	}
	if recvStart == dot {
		return rubyAssignment{}, keyEnd, false, fmt.Errorf("%w: the license assignment receiver cannot be interpreted", ErrInvalidGemspecSurface)
	}
	if recvStart > 0 && content[recvStart-1] == '.' {
		return rubyAssignment{}, keyEnd, false, fmt.Errorf("%w: the license assignment receiver is chained", ErrInvalidGemspecSurface)
	}
	j := skipRubySpaceTab(content, keyEnd)
	if j >= len(content) || content[j] != '=' {
		return rubyAssignment{}, keyEnd, false, nil
	}
	if j+1 < len(content) && (content[j+1] == '=' || content[j+1] == '~' || content[j+1] == '>') {
		return rubyAssignment{}, keyEnd, false, nil
	}
	j = skipRubyAssignmentSpace(content, j+1)
	if j >= len(content) {
		return rubyAssignment{}, keyEnd, false, fmt.Errorf("%w: the license assignment carries no value", ErrInvalidGemspecSurface)
	}
	valueEnd, kind, value, entries, err := parseRubyLicenseValue(content, j)
	if err != nil {
		return rubyAssignment{}, keyEnd, false, err
	}
	valueEnd = skipRubyFreezeSuffix(content, valueEnd)
	if valueEnd < len(content) {
		boundary := skipRubySpaceTab(content, valueEnd)
		if boundary < len(content) {
			switch content[boundary] {
			case '\n', '\r', '#':
			default:
				return rubyAssignment{}, keyEnd, false, fmt.Errorf("%w: the license value carries a dynamic suffix", ErrInvalidGemspecLicenseValue)
			}
		}
	}
	return rubyAssignment{
		receiver:   content[recvStart:dot],
		kind:       kind,
		value:      value,
		entries:    entries,
		start:      recvStart,
		valueStart: j,
		valueEnd:   valueEnd,
	}, valueEnd, true, nil
}

// skipRubyAssignmentSpace skips whitespace and line comments between the
// assignment operator and its value expression.
func skipRubyAssignmentSpace(content string, i int) int {
	for {
		i = skipSpace(content, i)
		if i < len(content) && content[i] == '#' {
			i = skipRubyComment(content, i)
			continue
		}
		return i
	}
}

// skipRubyFreezeSuffix consumes the deterministic .freeze method suffix of a
// license value; whitespace around the dot is legal. Any other suffix stays
// unconsumed and is refused by the assignment end check.
func skipRubyFreezeSuffix(content string, i int) int {
	j := skipRubySpaceTab(content, i)
	if j >= len(content) || content[j] != '.' {
		return i
	}
	k := skipRubySpaceTab(content, j+1)
	if strings.HasPrefix(content[k:], "freeze") && !rubyIdentByteAt(content, k+6) {
		return k + 6
	}
	return i
}

// parseRubyLicenseValue parses the value expression of one license
// assignment: a string literal, a bracket array of string literals, a
// %w/%W word array, or — uninterpretable — one invalid token.
func parseRubyLicenseValue(content string, start int) (int, rubyAssignmentKind, string, []string, error) {
	switch c := content[start]; {
	case c == '"' || c == '\'':
		value, end, err := decodeRubyLicenseString(content, start)
		if err != nil {
			return 0, 0, "", nil, err
		}
		return end, rubyValueString, value, nil, nil
	case c == '[':
		entries, end, err := parseRubyBracketArray(content, start)
		if err != nil {
			return 0, 0, "", nil, err
		}
		return end, rubyValueArray, "", entries, nil
	case isRubyWordArrayStart(content, start):
		entries, end, err := parseRubyPercentArray(content, start)
		if err != nil {
			return 0, 0, "", nil, err
		}
		return end, rubyValueArray, "", entries, nil
	default:
		return skipRubyInvalidToken(content, start), rubyValueInvalid, "", nil, nil
	}
}

// decodeRubyLicenseString decodes one string literal at a license value
// position: the interior must be escape-free and interpolation-free so the
// decoded value provably round-trips.
func decodeRubyLicenseString(content string, start int) (string, int, error) {
	end, err := parseRubyString(content, start)
	if err != nil {
		return "", 0, err
	}
	interior := content[start+1 : end-1]
	if strings.ContainsRune(interior, '\\') {
		return "", 0, fmt.Errorf("%w: the license string carries escapes", ErrInvalidGemspecLicenseValue)
	}
	if content[start] == '"' && strings.Contains(interior, "#{") {
		return "", 0, fmt.Errorf("%w: the license string carries interpolation", ErrInvalidGemspecLicenseValue)
	}
	return interior, end, nil
}

// parseRubyString lexes one Ruby string literal starting at its opening
// quote and returns the index just past the closing quote. Backslash escapes
// are skipped; both quote kinds span line breaks.
func parseRubyString(content string, start int) (int, error) {
	quote := content[start]
	for i := start + 1; i < len(content); i++ {
		switch content[i] {
		case '\\':
			i++
		case quote:
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("%w: unterminated string literal", ErrInvalidGemspecSurface)
}

// parseRubyBracketArray parses a bracket array of string literals. The empty
// array is returned as an empty entry list; a non-string entry, an escape or
// interpolation inside an entry, a malformed delimiter, and an unterminated
// array are refused fail-closed.
func parseRubyBracketArray(content string, start int) ([]string, int, error) {
	entries := []string{}
	i := start + 1
	for {
		i = skipSpace(content, i)
		if i >= len(content) {
			return nil, 0, fmt.Errorf("%w: unterminated license array", ErrInvalidGemspecSurface)
		}
		if content[i] == ']' {
			return entries, i + 1, nil
		}
		if content[i] != '"' && content[i] != '\'' {
			return nil, 0, fmt.Errorf("%w: the license array carries a non-string entry", ErrInvalidGemspecLicenseValue)
		}
		entry, end, err := decodeRubyLicenseString(content, i)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
		i = skipSpace(content, end)
		if i >= len(content) {
			return nil, 0, fmt.Errorf("%w: unterminated license array", ErrInvalidGemspecSurface)
		}
		switch content[i] {
		case ',':
			i++
		case ']':
			return entries, i + 1, nil
		default:
			return nil, 0, fmt.Errorf("%w: the license array carries a malformed delimiter", ErrInvalidGemspecLicenseValue)
		}
	}
}

// parseRubyPercentArray parses a %w/%W word array. The interior must be
// escape-free and interpolation-free; the words are split on whitespace.
func parseRubyPercentArray(content string, start int) ([]string, int, error) {
	open := content[start+2]
	close := rubyMatchingDelimiter(open)
	depth := 1
	interiorEnd := -1
	for j := start + 3; j < len(content); j++ {
		c := content[j]
		if c == '\\' {
			return nil, 0, fmt.Errorf("%w: the percent array carries escapes", ErrInvalidGemspecLicenseValue)
		}
		if c == open {
			depth++
		} else if c == close {
			depth--
			if depth == 0 {
				interiorEnd = j
				break
			}
		}
	}
	if interiorEnd < 0 {
		return nil, 0, fmt.Errorf("%w: unterminated percent array", ErrInvalidGemspecSurface)
	}
	interior := content[start+3 : interiorEnd]
	if strings.Contains(interior, "#{") {
		return nil, 0, fmt.Errorf("%w: the percent array carries interpolation", ErrInvalidGemspecLicenseValue)
	}
	words := strings.FieldsFunc(interior, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	return words, interiorEnd + 1, nil
}

// skipRubyInvalidToken consumes one uninterpretable value token: an
// identifier or number run, a symbol, a balanced group, or a single byte.
func skipRubyInvalidToken(content string, start int) int {
	c := content[start]
	switch {
	case rubyIdentByte(c):
		i := start
		for i < len(content) && (rubyIdentByte(content[i]) || content[i] == '.') {
			i++
		}
		return i
	case c == ':':
		i := start + 1
		for i < len(content) && rubyIdentByte(content[i]) {
			i++
		}
		return i
	case c == '{' || c == '(':
		open, close := c, byte('}')
		if c == '(' {
			close = ')'
		}
		depth := 0
		for i := start; i < len(content); i++ {
			if content[i] == open {
				depth++
			} else if content[i] == close {
				depth--
				if depth == 0 {
					return i + 1
				}
			}
		}
		return len(content)
	default:
		return start + 1
	}
}

// isRubyWordArrayStart reports whether the value position opens a %w/%W word
// array with a paired delimiter.
func isRubyWordArrayStart(content string, start int) bool {
	return start+2 < len(content) &&
		(content[start+1] == 'w' || content[start+1] == 'W') &&
		isRubyPercentDelimiter(content[start+2])
}

// isRubyPercentDelimiter reports whether the byte opens a paired percent
// literal delimiter.
func isRubyPercentDelimiter(c byte) bool {
	switch c {
	case '(', '[', '{', '<':
		return true
	}
	return false
}

// rubyMatchingDelimiter resolves the closing delimiter of a paired percent
// literal delimiter.
func rubyMatchingDelimiter(open byte) byte {
	switch open {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	default:
		return '>'
	}
}

// isRubyPercentLiteralStart reports whether the percent token at i opens a
// percent literal in context: an optional literal-kind letter followed by a
// non-alphanumeric delimiter.
func isRubyPercentLiteralStart(content string, i int) bool {
	if i+1 >= len(content) {
		return false
	}
	j := i + 1
	if rubyLetterByte(content[j]) {
		switch content[j] {
		case 'w', 'W', 'q', 'Q', 'r', 'i', 'I', 's', 'x':
			j++
		default:
			return false
		}
	}
	if j >= len(content) {
		return false
	}
	return !rubyIdentByte(content[j]) && content[j] != ' ' && content[j] != '\t' &&
		content[j] != '\n' && content[j] != '\r'
}

// skipRubyPercentLiteral lexes one percent literal in context and returns the
// index just past its closing delimiter. Paired delimiters count nesting;
// backslash escapes skip the next byte.
func skipRubyPercentLiteral(content string, start int) (int, error) {
	j := start + 1
	if rubyLetterByte(content[j]) {
		j++
	}
	open := content[j]
	close := open
	switch open {
	case '(':
		close = ')'
	case '[':
		close = ']'
	case '{':
		close = '}'
	case '<':
		close = '>'
	}
	paired := close != open
	depth := 1
	for i := j + 1; i < len(content); i++ {
		c := content[i]
		if c == '\\' {
			i++
			continue
		}
		if paired && c == open {
			depth++
		} else if c == close {
			depth--
			if depth == 0 {
				return i + 1, nil
			}
		}
	}
	return 0, fmt.Errorf("%w: unterminated percent literal", ErrInvalidGemspecSurface)
}

// isRubyHeredocStart reports whether the less-than token at i opens a
// heredoc: <<, an optional tilde or dash, and an identifier or quoted
// terminator.
func isRubyHeredocStart(content string, i int) bool {
	if i+1 >= len(content) || content[i+1] != '<' {
		return false
	}
	j := i + 2
	if j < len(content) && (content[j] == '~' || content[j] == '-') {
		j++
	}
	if j >= len(content) {
		return false
	}
	c := content[j]
	return rubyLetterByte(c) || c == '_' || c == '"' || c == '\''
}

// skipRubyHeredoc skips one heredoc body and returns the index of the
// terminator line's end (its newline, or the end of the content). The body
// ends at the first line whose trimmed content equals the terminator.
func skipRubyHeredoc(content string, i int) (int, error) {
	j := i + 2
	if j < len(content) && (content[j] == '~' || content[j] == '-') {
		j++
	}
	var terminator string
	if j < len(content) && (content[j] == '"' || content[j] == '\'') {
		quote := content[j]
		j++
		start := j
		for j < len(content) && content[j] != quote {
			j++
		}
		if j >= len(content) {
			return 0, fmt.Errorf("%w: unterminated heredoc terminator", ErrInvalidGemspecSurface)
		}
		terminator = content[start:j]
		j++
	} else {
		start := j
		for j < len(content) && rubyIdentByte(content[j]) {
			j++
		}
		terminator = content[start:j]
	}
	k := j
	for k < len(content) && content[k] != '\n' {
		k++
	}
	if k >= len(content) {
		return 0, fmt.Errorf("%w: unterminated heredoc", ErrInvalidGemspecSurface)
	}
	k++
	for {
		if k >= len(content) {
			return 0, fmt.Errorf("%w: unterminated heredoc", ErrInvalidGemspecSurface)
		}
		rowEnd := k
		for rowEnd < len(content) && content[rowEnd] != '\n' {
			rowEnd++
		}
		if strings.TrimSpace(content[k:rowEnd]) == terminator {
			return rowEnd, nil
		}
		k = rowEnd + 1
	}
}

// isRubyBlockCommentStart reports whether the line at i opens an =begin block
// comment.
func isRubyBlockCommentStart(content string, i int) bool {
	return strings.HasPrefix(content[i:], "=begin") && !rubyIdentByteAt(content, i+6)
}

// skipRubyBlockComment skips one =begin/=end block comment and returns the
// index just past the =end directive.
func skipRubyBlockComment(content string, i int) (int, error) {
	j := i
	for j < len(content) && content[j] != '\n' {
		j++
	}
	if j >= len(content) {
		return 0, fmt.Errorf("%w: unterminated =begin block comment", ErrInvalidGemspecSurface)
	}
	j++
	for {
		if j >= len(content) {
			return 0, fmt.Errorf("%w: unterminated =begin block comment", ErrInvalidGemspecSurface)
		}
		if strings.HasPrefix(content[j:], "=end") {
			return j + 4, nil
		}
		for j < len(content) && content[j] != '\n' {
			j++
		}
		if j >= len(content) {
			return 0, fmt.Errorf("%w: unterminated =begin block comment", ErrInvalidGemspecSurface)
		}
		j++
	}
}

// skipRubyComment skips one line comment and returns the index of its
// newline, or the end of the content.
func skipRubyComment(content string, i int) int {
	for i < len(content) && content[i] != '\n' {
		i++
	}
	return i
}

// rubyIdentByte reports whether the byte may appear inside a Ruby identifier.
func rubyIdentByte(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// rubyIdentByteAt reports whether the byte at i is a Ruby identifier byte;
// positions outside the content are boundaries.
func rubyIdentByteAt(content string, i int) bool {
	return i >= 0 && i < len(content) && rubyIdentByte(content[i])
}

// rubyLetterByte reports whether the byte is an ASCII letter.
func rubyLetterByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// skipRubySpaceTab skips spaces and tabs.
func skipRubySpaceTab(content string, i int) int {
	for i < len(content) && (content[i] == ' ' || content[i] == '\t') {
		i++
	}
	return i
}
