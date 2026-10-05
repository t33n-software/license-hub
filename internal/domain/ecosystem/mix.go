// The Elixir ecosystem adapter: it derives and aligns the license metadata
// surface of a mix.exs manifest from the declaration chain and the license
// lock. The declaration is the truth, the manifest field is its deterministic
// projection, and the license lock is the single source of truth of the
// license type. The Hex package configuration is the package keyword list of
// the project configuration, and its licenses entry is a bracket list of
// SPDX license identifiers; LicenseRef-<idstring> is the documented entry
// form for custom licenses included in the package.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// ElixirLanguage is the seam toolchain language that selects the Elixir
// matrix row.
const ElixirLanguage = "elixir"

// The Elixir license metadata surface: the manifest file and the keyword
// entries that carry the projected license declaration.
const (
	ElixirManifestName = "mix.exs"
	ElixirPackageKey   = "package"
	ElixirLicensesKey  = "licenses"
)

// ErrInvalidElixirSurface marks a mix.exs the scanner cannot lex or anchor:
// a structural break anywhere, a licenses entry outside the package keyword
// list, a package configuration that is not a keyword list, or a document
// without a package configuration to anchor the insertion.
var ErrInvalidElixirSurface = errors.New("invalid mix.exs surface")

// ErrMultipleElixirLicenses marks a licenses entry that declares more than
// one license identifier. The projection carries exactly one license family;
// the resolution is an explicit tenant decision.
var ErrMultipleElixirLicenses = errors.New("mix.exs declares multiple license entries")

// ErrAmbiguousElixirPackage marks a mix.exs that declares the package
// configuration more than once.
var ErrAmbiguousElixirPackage = errors.New("mix.exs declares the package configuration more than once")

// ErrAmbiguousElixirLicense marks a mix.exs that declares the licenses entry
// more than once.
var ErrAmbiguousElixirLicense = errors.New("mix.exs declares the licenses entry more than once")

// ErrInvalidElixirLicenseValue marks a licenses value that is not a bracket
// list of license string literals.
var ErrInvalidElixirLicenseValue = errors.New("mix.exs licenses value is not a license entry list")

// ElixirProjection derives the expected mix.exs licenses entry from the
// merged tenant values: a declared SPDX identifier projects itself; the
// file-based custom family projects the SPDX sideload entry
// LicenseRef-<LICENSE_ID> — the documented non-SPDX entry form keeps the
// license family identity. The derivation never reads the manifest.
func ElixirProjection(merged map[string]string) string {
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		return identifier
	}
	return "LicenseRef-" + merged["LICENSE_ID"]
}

// ElixirLicenseState classifies the licenses entry of a scanned mix.exs.
type ElixirLicenseState int

const (
	// ElixirLicenseMissing marks an absent package configuration or an
	// absent licenses entry.
	ElixirLicenseMissing ElixirLicenseState = iota
	// ElixirLicenseList marks the target form: a bracket list of license
	// string literals.
	ElixirLicenseList
	// ElixirLicenseInvalid marks a declared value that is not a bracket
	// list.
	ElixirLicenseInvalid
)

// ElixirLicenseSurface is the license-relevant view of a scanned mix.exs.
type ElixirLicenseSurface struct {
	// State classifies the licenses entry.
	State ElixirLicenseState
	// Value carries the declared license entry: the single entry of a
	// one-element list.
	Value string
	// Entries carries the decoded entries of the list.
	Entries []string
}

// elixirValueKind classifies the parsed form of a licenses value.
type elixirValueKind int

const (
	elixirValueList elixirValueKind = iota
	elixirValueInvalid
)

// elixirEntry carries one decoded string entry of a licenses list with the
// byte span of its literal.
type elixirEntry struct {
	value string
	start int
	end   int
}

// elixirLicensesSite carries the parsed licenses entry of a mix.exs scan.
type elixirLicensesSite struct {
	kind       elixirValueKind
	entries    []elixirEntry
	valueStart int
	valueEnd   int
	entryStart int
	entryEnd   int
}

// elixirScopeFrame carries one open bracket scope of a scan; the frame that
// is the value of a package keyword entry is marked, so a licenses entry is
// provably a direct member of the package configuration.
type elixirScopeFrame struct {
	kind      byte
	isPackage bool
}

// elixirScan carries the license-relevant skeleton of a scanned mix.exs: the
// bracket index of the package configuration (-1 when absent) and the parsed
// licenses entry (nil when absent).
type elixirScan struct {
	packageBracket int
	licenses       *elixirLicensesSite
}

// elixirMaxInterpolationDepth bounds the nesting the skeleton scanner walks;
// deeper documents are refused fail-closed instead of being misread.
const elixirMaxInterpolationDepth = 64

// InspectElixirLicense lexes a mix.exs file and classifies its licenses
// entry. A file that declares the package configuration or the licenses
// entry more than once, carries a licenses entry outside the package keyword
// list, or breaks structurally anywhere is refused fail-closed.
func InspectElixirLicense(content string) (ElixirLicenseSurface, error) {
	scan, err := scanElixir(content)
	if err != nil {
		return ElixirLicenseSurface{}, err
	}
	if scan.licenses == nil {
		return ElixirLicenseSurface{State: ElixirLicenseMissing}, nil
	}
	site := scan.licenses
	if site.kind == elixirValueInvalid {
		return ElixirLicenseSurface{State: ElixirLicenseInvalid}, nil
	}
	if len(site.entries) == 0 {
		return ElixirLicenseSurface{}, ErrInvalidElixirLicenseValue
	}
	values := make([]string, 0, len(site.entries))
	for _, entry := range site.entries {
		values = append(values, entry.value)
	}
	if len(values) == 1 {
		return ElixirLicenseSurface{State: ElixirLicenseList, Value: values[0], Entries: values}, nil
	}
	return ElixirLicenseSurface{State: ElixirLicenseList, Entries: values}, nil
}

// AlignElixirLicense returns the mix.exs content with its licenses entry
// aligned to target. The projected form is the single-entry list: a diverging
// single entry has its literal span replaced, the list formatting is
// preserved, and a missing entry is inserted as the first member of the
// package keyword list. A list that declares more than one entry, an empty
// list, a value that is not a bracket list, and a document without a package
// configuration are refused fail-closed, because the projection carries
// exactly one license family and the insertion needs a provable anchor. Every
// byte outside the replaced literal span or the inserted entry is preserved
// exactly; a re-run over an aligned file is a no-op.
func AlignElixirLicense(content, target string) (string, bool, error) {
	if err := validateElixirTarget(target); err != nil {
		return "", false, err
	}
	scan, err := scanElixir(content)
	if err != nil {
		return "", false, err
	}
	if scan.licenses == nil {
		return insertElixirLicense(content, target, scan)
	}
	site := scan.licenses
	if site.kind == elixirValueInvalid {
		return "", false, ErrInvalidElixirLicenseValue
	}
	if len(site.entries) == 0 {
		return "", false, ErrInvalidElixirLicenseValue
	}
	if len(site.entries) > 1 {
		return "", false, ErrMultipleElixirLicenses
	}
	if site.entries[0].value == target {
		return content, false, nil
	}
	return content[:site.entryStart] + elixirQuote(target) + content[site.entryEnd:], true, nil
}

// elixirQuote renders the double-quoted Elixir string literal of the target
// value. The target guard guarantees an escape-free literal.
func elixirQuote(target string) string {
	return "\"" + target + "\""
}

// elixirLicenseList renders the one-entry bracket list of the target value.
func elixirLicenseList(target string) string {
	return "[" + elixirQuote(target) + "]"
}

// validateElixirTarget refuses alignment targets the Elixir string literal
// contract cannot carry: a non-UTF-8 value, an empty value, and a value that
// is not representable as an escape-free, interpolation-free double-quoted
// Elixir string literal.
func validateElixirTarget(target string) error {
	if !utf8.ValidString(target) || target == "" {
		return ErrInvalidTarget
	}
	for _, r := range target {
		switch r {
		case '"', '\\', '#':
			return fmt.Errorf("%w: the license entry is not representable as an escape-free, interpolation-free Elixir string literal", ErrInvalidTarget)
		}
	}
	return nil
}

// insertElixirLicense inserts the missing licenses entry as the first member
// of the package keyword list. The anchor is the package bracket the scanner
// proved, and the insertion is deterministic, so the scanner's fail-closed
// guards carry the anchor provability; an anchor without a proven package
// configuration is refused fail-closed.
func insertElixirLicense(content, target string, scan *elixirScan) (string, bool, error) {
	if scan.packageBracket < 0 {
		return "", false, fmt.Errorf("%w: the document carries no package configuration to anchor the insertion", ErrInvalidElixirSurface)
	}
	p := scan.packageBracket + 1
	insertion := "licenses: " + elixirLicenseList(target)
	j := skipElixirBlank(content, p)
	if j < len(content) && content[j] != ']' {
		insertion += ","
		if j == p && content[j] != '#' {
			insertion += " "
		}
	}
	return content[:p] + insertion + content[p:], true, nil
}

// scanElixir lexes the license-relevant skeleton of a mix.exs file: line
// comments, string literals, charlists, heredoc strings, and sigils are
// context, so a license entry inside them is never observed. The package
// keyword entry marks its value bracket, and the licenses entry is captured
// only as a direct member of that bracket. A structural break anywhere — an
// unterminated string, charlist, heredoc, sigil, interpolation, or bracket
// scope — and every non-canonical licenses position refuse the whole surface
// fail-closed.
func scanElixir(content string) (*elixirScan, error) {
	result := &elixirScan{packageBracket: -1}
	stack := []*elixirScopeFrame{}
	i := 0
	for i < len(content) {
		c := content[i]
		switch {
		case c == '#':
			i = skipElixirComment(content, i)
		case c == '"':
			end, err := skipElixirString(content, i)
			if err != nil {
				return nil, err
			}
			i = end
		case c == '\'':
			end, err := skipElixirCharlist(content, i)
			if err != nil {
				return nil, err
			}
			i = end
		case c == '~' && isElixirSigilStart(content, i):
			end, err := skipElixirSigil(content, i)
			if err != nil {
				return nil, err
			}
			i = end
		case c == '?' && i+1 < len(content) && isElixirIdentByte(content[i+1]):
			i += 2
		case c == '[':
			stack = append(stack, &elixirScopeFrame{kind: '['})
			i++
		case c == '%' && i+1 < len(content) && content[i+1] == '{':
			stack = append(stack, &elixirScopeFrame{kind: '%'})
			i += 2
		case c == '{':
			stack = append(stack, &elixirScopeFrame{kind: '{'})
			i++
		case c == '(':
			stack = append(stack, &elixirScopeFrame{kind: '('})
			i++
		case c == ']' || c == '}' || c == ')':
			if len(stack) == 0 {
				return nil, fmt.Errorf("%w: unbalanced %q", ErrInvalidElixirSurface, c)
			}
			stack = stack[:len(stack)-1]
			i++
		case isElixirIdentStart(c):
			start := i
			i = skipElixirIdent(content, i)
			if i < len(content) && content[i] == ':' && (i+1 >= len(content) || content[i+1] != ':') {
				switch content[start:i] {
				case ElixirPackageKey:
					next, err := scanElixirPackage(content, i, result, &stack)
					if err != nil {
						return nil, err
					}
					i = next
				case ElixirLicensesKey:
					next, err := scanElixirLicenses(content, i, result, stack)
					if err != nil {
						return nil, err
					}
					i = next
				}
			}
		default:
			i++
		}
	}
	if len(stack) > 0 {
		return nil, fmt.Errorf("%w: unterminated bracket scope", ErrInvalidElixirSurface)
	}
	return result, nil
}

// scanElixirPackage handles one package keyword entry: the value must be a
// bracket keyword list, the bracket is marked as the package scope, and a
// second package entry is refused fail-closed.
func scanElixirPackage(content string, colon int, result *elixirScan, stack *[]*elixirScopeFrame) (int, error) {
	if result.packageBracket >= 0 {
		return 0, ErrAmbiguousElixirPackage
	}
	j := skipElixirBlank(content, colon+1)
	if j >= len(content) || content[j] != '[' {
		return 0, fmt.Errorf("%w: the package configuration is not a keyword list", ErrInvalidElixirSurface)
	}
	result.packageBracket = j
	*stack = append(*stack, &elixirScopeFrame{kind: '[', isPackage: true})
	return j + 1, nil
}

// scanElixirLicenses handles one licenses keyword entry: the entry must be a
// direct member of the package keyword list — outside it and nested deeper
// inside it are both refused fail-closed — and its value is parsed as a
// bracket list of license string literals.
func scanElixirLicenses(content string, colon int, result *elixirScan, stack []*elixirScopeFrame) (int, error) {
	if len(stack) == 0 {
		return 0, fmt.Errorf("%w: the licenses entry lives outside the package configuration", ErrInvalidElixirSurface)
	}
	top := stack[len(stack)-1]
	if !top.isPackage {
		for _, frame := range stack {
			if frame.isPackage {
				return 0, fmt.Errorf("%w: the licenses entry is nested inside the package configuration rather than a direct member", ErrInvalidElixirSurface)
			}
		}
		return 0, fmt.Errorf("%w: the licenses entry lives outside the package configuration", ErrInvalidElixirSurface)
	}
	if result.licenses != nil {
		return 0, ErrAmbiguousElixirLicense
	}
	j := skipElixirBlank(content, colon+1)
	if j >= len(content) || content[j] != '[' {
		result.licenses = &elixirLicensesSite{kind: elixirValueInvalid, valueStart: j, valueEnd: j}
		return j, nil
	}
	entries, end, err := parseElixirLicenseEntries(content, j)
	if err != nil {
		return 0, err
	}
	site := &elixirLicensesSite{kind: elixirValueList, entries: entries, valueStart: j, valueEnd: end}
	if len(entries) > 0 {
		site.entryStart = entries[0].start
		site.entryEnd = entries[0].end
	}
	result.licenses = site
	return end, nil
}

// parseElixirLicenseEntries parses the bracket list of a licenses value: a
// sequence of double-quoted, escape-free, interpolation-free string
// literals. The empty list, a non-string entry, an entry carrying escapes or
// interpolation, a malformed delimiter, and an unterminated list are refused
// fail-closed; the trailing comma is legal Elixir list syntax.
func parseElixirLicenseEntries(content string, start int) ([]elixirEntry, int, error) {
	entries := []elixirEntry{}
	i := start + 1
	for {
		i = skipElixirBlank(content, i)
		if i >= len(content) {
			return nil, 0, fmt.Errorf("%w: unterminated licenses list", ErrInvalidElixirSurface)
		}
		if content[i] == ']' {
			return entries, i + 1, nil
		}
		if content[i] != '"' {
			return nil, 0, fmt.Errorf("%w: the licenses list carries a non-string entry", ErrInvalidElixirLicenseValue)
		}
		entry, end, err := parseElixirLicenseString(content, i)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
		i = skipElixirBlank(content, end)
		if i >= len(content) {
			return nil, 0, fmt.Errorf("%w: unterminated licenses list", ErrInvalidElixirSurface)
		}
		switch content[i] {
		case ',':
			i++
		case ']':
			return entries, i + 1, nil
		default:
			return nil, 0, fmt.Errorf("%w: the licenses list carries a malformed delimiter", ErrInvalidElixirLicenseValue)
		}
	}
}

// parseElixirLicenseString decodes one string literal at a licenses entry
// position: the interior must be escape-free and interpolation-free so the
// decoded value provably round-trips.
func parseElixirLicenseString(content string, start int) (elixirEntry, int, error) {
	end, err := skipElixirString(content, start)
	if err != nil {
		return elixirEntry{}, 0, err
	}
	interior := content[start+1 : end-1]
	if strings.ContainsRune(interior, '\\') {
		return elixirEntry{}, 0, fmt.Errorf("%w: the license entry carries escapes", ErrInvalidElixirLicenseValue)
	}
	if strings.Contains(interior, "#{") {
		return elixirEntry{}, 0, fmt.Errorf("%w: the license entry carries interpolation", ErrInvalidElixirLicenseValue)
	}
	return elixirEntry{value: interior, start: start, end: end}, end, nil
}

// skipElixirString lexes one double-quoted string in context and returns the
// index just past its closing quote. The heredoc form, backslash escapes,
// and interpolation are consumed; a nesting depth beyond the scanner bound
// and an unterminated literal refuse fail-closed.
func skipElixirString(content string, start int) (int, error) {
	return skipElixirStringDepth(content, start, 0)
}

func skipElixirStringDepth(content string, start, depth int) (int, error) {
	if depth > elixirMaxInterpolationDepth {
		return 0, fmt.Errorf("%w: interpolation nesting exceeds the scanner bound", ErrInvalidElixirSurface)
	}
	if strings.HasPrefix(content[start:], `"""`) {
		return skipElixirHeredoc(content, start+3)
	}
	for i := start + 1; i < len(content); i++ {
		switch content[i] {
		case '\\':
			i++
		case '"':
			return i + 1, nil
		case '#':
			if strings.HasPrefix(content[i:], "#{") {
				end, err := skipElixirInterpolationDepth(content, i, depth+1)
				if err != nil {
					return 0, err
				}
				i = end - 1
			}
		}
	}
	return 0, fmt.Errorf("%w: unterminated string literal", ErrInvalidElixirSurface)
}

// skipElixirHeredoc lexes one heredoc string body and returns the index just
// past its closing triple quote.
func skipElixirHeredoc(content string, start int) (int, error) {
	for i := start; i < len(content); i++ {
		if content[i] == '\\' {
			i++
			continue
		}
		if strings.HasPrefix(content[i:], `"""`) {
			return i + 3, nil
		}
	}
	return 0, fmt.Errorf("%w: unterminated heredoc string", ErrInvalidElixirSurface)
}

// skipElixirCharlist lexes one single-quoted charlist in context and returns
// the index just past its closing quote.
func skipElixirCharlist(content string, start int) (int, error) {
	for i := start + 1; i < len(content); i++ {
		switch content[i] {
		case '\\':
			i++
		case '\'':
			return i + 1, nil
		}
	}
	return 0, fmt.Errorf("%w: unterminated charlist", ErrInvalidElixirSurface)
}

// skipElixirSigil lexes one sigil in context and returns the index just past
// its closing delimiter. Paired delimiters count nesting; backslash escapes
// skip the next byte; the heredoc sigil form is consumed like a heredoc.
func skipElixirSigil(content string, start int) (int, error) {
	j := start + 2
	open := content[j]
	if open == '"' && strings.HasPrefix(content[j:], `"""`) {
		return skipElixirHeredoc(content, j+3)
	}
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
	return 0, fmt.Errorf("%w: unterminated sigil", ErrInvalidElixirSurface)
}

// skipElixirInterpolationDepth lexes one interpolation body and returns the
// index just past its closing brace. Nested braces, strings, charlists, and
// interpolations are consumed with the shared depth bound.
func skipElixirInterpolationDepth(content string, hash, depth int) (int, error) {
	if depth > elixirMaxInterpolationDepth {
		return 0, fmt.Errorf("%w: interpolation nesting exceeds the scanner bound", ErrInvalidElixirSurface)
	}
	inner := 0
	for i := hash + 2; i < len(content); i++ {
		switch c := content[i]; c {
		case '{':
			inner++
		case '}':
			if inner == 0 {
				return i + 1, nil
			}
			inner--
		case '"':
			end, err := skipElixirStringDepth(content, i, depth+1)
			if err != nil {
				return 0, err
			}
			i = end - 1
		case '\'':
			end, err := skipElixirCharlist(content, i)
			if err != nil {
				return 0, err
			}
			i = end - 1
		case '#':
			if strings.HasPrefix(content[i:], "#{") {
				end, err := skipElixirInterpolationDepth(content, i, depth+1)
				if err != nil {
					return 0, err
				}
				i = end - 1
			}
		}
	}
	return 0, fmt.Errorf("%w: unterminated interpolation", ErrInvalidElixirSurface)
}

// isElixirSigilStart reports whether the tilde at i opens a sigil: a tilde,
// one letter, and a non-alphanumeric, non-whitespace delimiter.
func isElixirSigilStart(content string, i int) bool {
	if i+1 >= len(content) || !isElixirLetter(content[i+1]) {
		return false
	}
	j := i + 2
	if j >= len(content) {
		return false
	}
	return !isElixirIdentByte(content[j]) && content[j] != ' ' && content[j] != '\t' &&
		content[j] != '\n' && content[j] != '\r'
}

// skipElixirComment skips one line comment and returns the index of its
// newline, or the end of the content.
func skipElixirComment(content string, i int) int {
	for i < len(content) && content[i] != '\n' {
		i++
	}
	return i
}

// skipElixirBlank skips whitespace and line comments.
func skipElixirBlank(content string, i int) int {
	for i < len(content) {
		switch content[i] {
		case ' ', '\t', '\n', '\r':
			i++
		case '#':
			i = skipElixirComment(content, i)
		default:
			return i
		}
	}
	return i
}

// isElixirIdentStart reports whether the byte may start an Elixir identifier.
func isElixirIdentStart(c byte) bool {
	return c == '_' || isElixirLetter(c)
}

// isElixirIdentByte reports whether the byte may appear inside an Elixir
// identifier.
func isElixirIdentByte(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || isElixirLetter(c)
}

// isElixirLetter reports whether the byte is an ASCII letter.
func isElixirLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// skipElixirIdent lexes one identifier run, including the trailing question
// mark and exclamation mark forms, and returns the index just past it.
func skipElixirIdent(content string, i int) int {
	for i < len(content) && (isElixirIdentByte(content[i]) || content[i] == '?' || content[i] == '!') {
		i++
	}
	return i
}
