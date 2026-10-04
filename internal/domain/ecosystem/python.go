// The Python ecosystem adapter: it derives and aligns the license metadata
// surface of a pyproject.toml manifest (PEP 621/639) from the declaration
// chain and the license lock. The declaration is the truth, the manifest
// field is its deterministic projection, and the license lock is the single
// source of truth of the license type.
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

// PythonLanguage is the seam toolchain language that selects the Python
// matrix row.
const PythonLanguage = "python"

// The Python license metadata surface: the manifest file, the license
// expression field, the license text glob field, and the classifiers field
// whose License :: entries are the deprecated license declaration form.
const (
	PythonManifestName      = "pyproject.toml"
	PythonLicenseField      = "license"
	PythonLicenseFilesField = "license-files"
	PythonClassifiersField  = "classifiers"
)

// pythonClassifierPrefix marks the deprecated PEP 639 license classifier
// entries inside the classifiers field.
const pythonClassifierPrefix = "License ::"

// PythonLicenseFilesTarget is the projected license-files glob list: the
// canonical license text at the repository root.
var PythonLicenseFilesTarget = []string{"LICENSE"}

// ErrInvalidPythonSurface marks a manifest that is not a scannable
// pyproject.toml surface.
var ErrInvalidPythonSurface = errors.New("invalid pyproject.toml surface")

// ErrAmbiguousPythonField marks a [project] table that declares a
// license-relevant key more than once.
var ErrAmbiguousPythonField = errors.New("ambiguous pyproject.toml license key")

// ErrInvalidPythonLicenseValue marks a license value that is neither a
// string expression nor the deprecated table form.
var ErrInvalidPythonLicenseValue = errors.New("pyproject.toml license value is not a string expression or table")

// ErrInvalidPythonLicenseFiles marks a license-files value that is not an
// array of strings.
var ErrInvalidPythonLicenseFiles = errors.New("pyproject.toml license-files value is not an array of strings")

// ErrNoProjectTable marks a manifest without the [project] table; the
// license surface cannot be aligned deterministically without it.
var ErrNoProjectTable = errors.New("pyproject.toml carries no [project] table")

// PythonLicenseFieldState classifies the license field of the [project]
// table.
type PythonLicenseFieldState int

const (
	// PythonLicenseMissing marks an absent license field.
	PythonLicenseMissing PythonLicenseFieldState = iota
	// PythonLicenseString marks the target form: a license expression string.
	PythonLicenseString
	// PythonLicenseTable marks the deprecated PEP 621 table form.
	PythonLicenseTable
	// PythonLicenseInvalid marks a value that is neither a string nor a
	// table.
	PythonLicenseInvalid
)

// PythonLicenseSurface is the license-relevant view of a pyproject.toml.
type PythonLicenseSurface struct {
	// State classifies the license field.
	State PythonLicenseFieldState
	// Value carries the decoded license expression when the state is
	// PythonLicenseString.
	Value string
	// LicenseFilesSet reports whether the license-files key exists.
	LicenseFilesSet bool
	// LicenseFiles carries the decoded globs when the form is valid.
	LicenseFiles []string
	// LicenseFilesValid reports whether license-files is an array of
	// strings.
	LicenseFilesValid bool
	// DeprecatedClassifiers carries the deprecated License :: classifier
	// entries.
	DeprecatedClassifiers []string
}

// PythonProjection derives the expected pyproject.toml license field value
// from the merged tenant values: a declared SPDX identifier projects
// itself; the file-based custom family projects the SPDX sideload form
// LicenseRef-<LICENSE_ID> (PEP 639). The derivation never reads the
// manifest.
func PythonProjection(merged map[string]string) string {
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		return identifier
	}
	return "LicenseRef-" + merged["LICENSE_ID"]
}

// PythonLicenseFilesForm renders a license-files glob list in its manifest
// form.
func PythonLicenseFilesForm(files []string) string {
	return pythonArray(files)
}

// InspectPythonLicense parses a pyproject.toml manifest and classifies its
// license-relevant surface: the license field, the license-files globs, and
// the deprecated License :: classifier entries.
func InspectPythonLicense(content string) (PythonLicenseSurface, error) {
	result := PythonLicenseSurface{State: PythonLicenseMissing}
	surface, err := scanPythonProject(content)
	if err != nil {
		return PythonLicenseSurface{}, err
	}
	if surface == nil {
		return result, nil
	}
	if key := surface.license; key != nil {
		switch key.kind {
		case pythonValueString:
			result.State = PythonLicenseString
			result.Value = key.str
		case pythonValueInlineTable:
			result.State = PythonLicenseTable
		default:
			result.State = PythonLicenseInvalid
		}
	}
	if key := surface.licenseFiles; key != nil {
		result.LicenseFilesSet = true
		if key.kind == pythonValueArray {
			result.LicenseFilesValid = true
			for _, item := range key.items {
				if !item.isString {
					result.LicenseFilesValid = false
					result.LicenseFiles = nil
					break
				}
				result.LicenseFiles = append(result.LicenseFiles, item.decoded)
			}
		}
	}
	if key := surface.classifiers; key != nil && key.kind == pythonValueArray {
		for _, item := range key.items {
			if item.isString && strings.HasPrefix(item.decoded, pythonClassifierPrefix) {
				result.DeprecatedClassifiers = append(result.DeprecatedClassifiers, item.decoded)
			}
		}
	}
	return result, nil
}

// AlignPythonLicense returns the manifest content with the license surface
// of the [project] table aligned to the projected license expression and
// the license-files glob list. Every byte outside the replaced value spans
// or the inserted lines is preserved exactly; a re-run over an aligned
// manifest is a no-op. The deprecated license table form is aligned to the
// string expression form; every other non-string license value is refused
// fail-closed. The classifiers array is never rewritten.
func AlignPythonLicense(content, target string, files []string) (string, bool, error) {
	if !utf8.ValidString(target) {
		return "", false, ErrInvalidTarget
	}
	for _, file := range files {
		if !utf8.ValidString(file) {
			return "", false, ErrInvalidTarget
		}
	}
	surface, err := scanPythonProject(content)
	if err != nil {
		return "", false, err
	}
	if surface == nil {
		return "", false, ErrNoProjectTable
	}
	edits := []pythonEdit{}
	switch {
	case surface.license == nil:
		// The insertion cases below carry the license field.
	case surface.license.kind == pythonValueString:
		if surface.license.str != target {
			edits = append(edits, pythonEdit{
				start: surface.license.valueStart,
				end:   surface.license.valueEnd,
				text:  pythonQuote(target),
			})
		}
	case surface.license.kind == pythonValueInlineTable:
		edits = append(edits, pythonEdit{
			start: surface.license.valueStart,
			end:   surface.license.valueEnd,
			text:  pythonQuote(target),
		})
	default:
		return "", false, ErrInvalidPythonLicenseValue
	}
	switch {
	case surface.licenseFiles == nil:
		// The insertion cases below carry the license-files field.
	case surface.licenseFiles.kind != pythonValueArray:
		return "", false, ErrInvalidPythonLicenseFiles
	default:
		observed := []string{}
		for _, item := range surface.licenseFiles.items {
			if !item.isString {
				return "", false, ErrInvalidPythonLicenseFiles
			}
			observed = append(observed, item.decoded)
		}
		if !slices.Equal(observed, files) {
			edits = append(edits, pythonEdit{
				start: surface.licenseFiles.valueStart,
				end:   surface.licenseFiles.valueEnd,
				text:  pythonArray(files),
			})
		}
	}
	switch {
	case surface.license == nil && surface.licenseFiles == nil:
		edits = append(edits, pythonLineInsertion(content, surface.headerEnd,
			"license = "+pythonQuote(target)+"\nlicense-files = "+pythonArray(files)+"\n"))
	case surface.license == nil:
		edits = append(edits, pythonLineInsertion(content, surface.headerEnd,
			"license = "+pythonQuote(target)+"\n"))
	case surface.licenseFiles == nil:
		edits = append(edits, pythonLineInsertion(content, surface.license.lineEnd,
			"license-files = "+pythonArray(files)+"\n"))
	}
	if len(edits) == 0 {
		return content, false, nil
	}
	return applyPythonEdits(content, edits), true, nil
}

// pythonEdit carries one byte-span replacement or insertion.
type pythonEdit struct {
	start int
	end   int
	text  string
}

// applyPythonEdits applies the edits from the highest offset down so the
// earlier spans stay valid.
func applyPythonEdits(content string, edits []pythonEdit) string {
	slices.SortFunc(edits, func(a, b pythonEdit) int { return b.start - a.start })
	result := content
	for _, edit := range edits {
		result = result[:edit.start] + edit.text + result[edit.end:]
	}
	return result
}

// pythonLineInsertion builds a whole-line insertion at offset; a position
// that does not follow a newline gets one so the inserted lines start on
// their own line.
func pythonLineInsertion(content string, offset int, text string) pythonEdit {
	if offset > 0 && offset <= len(content) && content[offset-1] != '\n' {
		text = "\n" + text
	}
	return pythonEdit{start: offset, end: offset, text: text}
}

// pythonQuote renders a TOML basic string.
func pythonQuote(value string) string {
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

// pythonArray renders a TOML array of strings.
func pythonArray(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, pythonQuote(item))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

type pythonValueKind int

const (
	pythonValueString pythonValueKind = iota
	pythonValueArray
	pythonValueInlineTable
	pythonValueAtom
)

type pythonArrayItem struct {
	isString bool
	decoded  string
}

type pythonValue struct {
	kind  pythonValueKind
	str   string
	items []pythonArrayItem
}

type pythonKeyValue struct {
	key        string
	segments   []string
	valueStart int
	valueEnd   int
	lineEnd    int // the index just past the key's line terminator
	kind       pythonValueKind
	str        string
	items      []pythonArrayItem
}

type pythonTable struct {
	name      string
	array     bool
	start     int
	headerEnd int // the index just past the header line terminator
	keys      []pythonKeyValue
}

type pythonProjectSurface struct {
	start        int
	headerEnd    int
	license      *pythonKeyValue
	licenseFiles *pythonKeyValue
	classifiers  *pythonKeyValue
}

// scanPythonProject lexes a pyproject.toml document and returns the
// license-relevant surface of its [project] table. A manifest without a
// [project] table returns a nil surface; a document the scanner cannot
// interpret structurally is refused fail-closed.
func scanPythonProject(content string) (*pythonProjectSurface, error) {
	tables, err := scanPythonDocument(content)
	if err != nil {
		return nil, err
	}
	for _, table := range tables {
		if table.name != "project" || table.array {
			continue
		}
		surface := &pythonProjectSurface{start: table.start, headerEnd: table.headerEnd}
		for index := range table.keys {
			key := &table.keys[index]
			switch key.key {
			case PythonLicenseField:
				surface.license = key
			case PythonLicenseFilesField:
				surface.licenseFiles = key
			case PythonClassifiersField:
				surface.classifiers = key
			}
		}
		return surface, nil
	}
	return nil, nil
}

// scanPythonDocument lexes the table skeleton of a TOML document: every
// table header, every key, and the value form of every key. The scanner is
// the contract surface of the adapter; it refuses every form it cannot
// interpret instead of silently misreading it.
func scanPythonDocument(content string) ([]*pythonTable, error) {
	tables := []*pythonTable{}
	defined := map[string]bool{}
	current := &pythonTable{name: ""}
	tables = append(tables, current)
	i := 0
	for i < len(content) {
		i = skipPythonBlank(content, i)
		if i >= len(content) {
			break
		}
		if content[i] == '[' {
			table, next, err := parsePythonHeader(content, i, defined)
			if err != nil {
				return nil, err
			}
			tables = append(tables, table)
			current = table
			i = next
			continue
		}
		key, next, err := parsePythonKeyValue(content, i)
		if err != nil {
			return nil, err
		}
		if err := checkPythonKey(current, key); err != nil {
			return nil, err
		}
		current.keys = append(current.keys, *key)
		i = next
	}
	return tables, nil
}

// checkPythonKey refuses the duplicate and shadow forms the scanner cannot
// interpret: a repeated key inside one table, a dotted key inside [project]
// whose first segment names a license-relevant key, and any root
// declaration of the project name itself.
func checkPythonKey(table *pythonTable, key *pythonKeyValue) error {
	for _, existing := range table.keys {
		if existing.key == key.key {
			return fmt.Errorf("%w: duplicate key %q in [%s]", ErrAmbiguousPythonField, key.key, table.name)
		}
	}
	if table.name == "project" && !table.array && len(key.segments) > 1 {
		switch key.segments[0] {
		case PythonLicenseField, PythonLicenseFilesField, PythonClassifiersField:
			return fmt.Errorf("%w: dotted key %q shadows the %q surface in [project]",
				ErrInvalidPythonSurface, key.key, key.segments[0])
		}
	}
	if table.name == "" && len(key.segments) > 0 && key.segments[0] == "project" {
		return fmt.Errorf("%w: the project table must be declared with the [project] header", ErrInvalidPythonSurface)
	}
	return nil
}

func parsePythonHeader(content string, start int, defined map[string]bool) (*pythonTable, int, error) {
	i := start + 1
	array := false
	if i < len(content) && content[i] == '[' {
		array = true
		i++
	}
	segments := []string{}
	for {
		i = skipPythonSpaces(content, i)
		segment, next, err := parsePythonKeySegment(content, i)
		if err != nil {
			return nil, 0, err
		}
		segments = append(segments, segment)
		i = skipPythonSpaces(content, next)
		if i < len(content) && content[i] == '.' {
			i++
			continue
		}
		break
	}
	if i >= len(content) || content[i] != ']' {
		return nil, 0, fmt.Errorf("%w: malformed table header", ErrInvalidPythonSurface)
	}
	i++
	if array {
		if i >= len(content) || content[i] != ']' {
			return nil, 0, fmt.Errorf("%w: malformed array table header", ErrInvalidPythonSurface)
		}
		i++
	}
	i = skipPythonSpaces(content, i)
	if i < len(content) && content[i] == '#' {
		i = skipPythonComment(content, i)
	}
	table := &pythonTable{name: strings.Join(segments, "."), array: array, start: start}
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
		return nil, 0, fmt.Errorf("%w: trailing content after the table header", ErrInvalidPythonSurface)
	}
	if !array {
		if defined[table.name] {
			return nil, 0, fmt.Errorf("%w: duplicate table [%s]", ErrInvalidPythonSurface, table.name)
		}
		defined[table.name] = true
	}
	return table, i, nil
}

func parsePythonKeyValue(content string, start int) (*pythonKeyValue, int, error) {
	i := start
	segments := []string{}
	for {
		i = skipPythonSpaces(content, i)
		segment, next, err := parsePythonKeySegment(content, i)
		if err != nil {
			return nil, 0, err
		}
		segments = append(segments, segment)
		i = skipPythonSpaces(content, next)
		if i < len(content) && content[i] == '.' {
			i++
			continue
		}
		break
	}
	name := strings.Join(segments, ".")
	if i >= len(content) || content[i] != '=' {
		return nil, 0, fmt.Errorf("%w: expected '=' after the key %q", ErrInvalidPythonSurface, name)
	}
	i = skipPythonSpaces(content, i+1)
	valueStart := i
	value, valueEnd, err := parsePythonValue(content, i)
	if err != nil {
		return nil, 0, err
	}
	i = skipPythonSpaces(content, valueEnd)
	if i < len(content) && content[i] == '#' {
		i = skipPythonComment(content, i)
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
		return nil, 0, fmt.Errorf("%w: trailing content after the value of %q", ErrInvalidPythonSurface, name)
	}
	return &pythonKeyValue{
		key:        name,
		segments:   segments,
		valueStart: valueStart,
		valueEnd:   valueEnd,
		lineEnd:    lineEnd,
		kind:       value.kind,
		str:        value.str,
		items:      value.items,
	}, i, nil
}

func parsePythonKeySegment(content string, start int) (string, int, error) {
	if start >= len(content) {
		return "", 0, fmt.Errorf("%w: expected a key", ErrInvalidPythonSurface)
	}
	switch content[start] {
	case '"':
		decoded, end, err := lexBasicString(content, start)
		if err != nil {
			return "", 0, err
		}
		return decoded, end, nil
	case '\'':
		decoded, end, err := lexLiteralString(content, start)
		if err != nil {
			return "", 0, err
		}
		return decoded, end, nil
	}
	i := start
	for i < len(content) && isPythonBareKeyByte(content[i]) {
		i++
	}
	if i == start {
		return "", 0, fmt.Errorf("%w: expected a key", ErrInvalidPythonSurface)
	}
	return content[start:i], i, nil
}

func isPythonBareKeyByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
}

func parsePythonValue(content string, start int) (pythonValue, int, error) {
	if start >= len(content) {
		return pythonValue{}, 0, fmt.Errorf("%w: expected a value", ErrInvalidPythonSurface)
	}
	switch content[start] {
	case '"':
		if strings.HasPrefix(content[start:], `"""`) {
			decoded, end, err := lexMultilineBasicString(content, start)
			if err != nil {
				return pythonValue{}, 0, err
			}
			return pythonValue{kind: pythonValueString, str: decoded}, end, nil
		}
		decoded, end, err := lexBasicString(content, start)
		if err != nil {
			return pythonValue{}, 0, err
		}
		return pythonValue{kind: pythonValueString, str: decoded}, end, nil
	case '\'':
		if strings.HasPrefix(content[start:], `'''`) {
			decoded, end, err := lexMultilineLiteralString(content, start)
			if err != nil {
				return pythonValue{}, 0, err
			}
			return pythonValue{kind: pythonValueString, str: decoded}, end, nil
		}
		decoded, end, err := lexLiteralString(content, start)
		if err != nil {
			return pythonValue{}, 0, err
		}
		return pythonValue{kind: pythonValueString, str: decoded}, end, nil
	case '[':
		items, end, err := lexPythonArray(content, start)
		if err != nil {
			return pythonValue{}, 0, err
		}
		return pythonValue{kind: pythonValueArray, items: items}, end, nil
	case '{':
		end, err := lexPythonInlineTable(content, start)
		if err != nil {
			return pythonValue{}, 0, err
		}
		if strings.ContainsAny(content[start:end], "\n\r") {
			return pythonValue{}, 0, fmt.Errorf("%w: newlines are not allowed inside an inline table", ErrInvalidPythonSurface)
		}
		return pythonValue{kind: pythonValueInlineTable}, end, nil
	default:
		i := start
		for i < len(content) && !isPythonValueDelimiter(content[i]) {
			i++
		}
		if i == start {
			return pythonValue{}, 0, fmt.Errorf("%w: expected a value", ErrInvalidPythonSurface)
		}
		return pythonValue{kind: pythonValueAtom}, i, nil
	}
}

func isPythonValueDelimiter(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', ',', ']', '}', '#':
		return true
	}
	return false
}

func lexPythonArray(content string, start int) ([]pythonArrayItem, int, error) {
	i := start + 1
	items := []pythonArrayItem{}
	for {
		i = skipPythonBlank(content, i)
		if i >= len(content) {
			return nil, 0, fmt.Errorf("%w: unterminated array", ErrInvalidPythonSurface)
		}
		if content[i] == ']' {
			return items, i + 1, nil
		}
		value, next, err := parsePythonValue(content, i)
		if err != nil {
			return nil, 0, err
		}
		if value.kind == pythonValueString {
			items = append(items, pythonArrayItem{isString: true, decoded: value.str})
		} else {
			items = append(items, pythonArrayItem{})
		}
		i = skipPythonBlank(content, next)
		if i >= len(content) {
			return nil, 0, fmt.Errorf("%w: unterminated array", ErrInvalidPythonSurface)
		}
		switch content[i] {
		case ',':
			i++
		case ']':
			return items, i + 1, nil
		default:
			return nil, 0, fmt.Errorf("%w: expected ',' or ']' in an array", ErrInvalidPythonSurface)
		}
	}
}

func lexPythonInlineTable(content string, start int) (int, error) {
	i := start + 1
	for {
		i = skipPythonSpaces(content, i)
		if i >= len(content) || content[i] == '\n' || content[i] == '\r' {
			return 0, fmt.Errorf("%w: unterminated or multi-line inline table", ErrInvalidPythonSurface)
		}
		if content[i] == '}' {
			return i + 1, nil
		}
		_, next, err := parsePythonKeySegment(content, i)
		if err != nil {
			return 0, err
		}
		i = skipPythonSpaces(content, next)
		if i >= len(content) || content[i] != '=' {
			return 0, fmt.Errorf("%w: expected '=' in an inline table", ErrInvalidPythonSurface)
		}
		i = skipPythonSpaces(content, i+1)
		_, valueEnd, err := parsePythonValue(content, i)
		if err != nil {
			return 0, err
		}
		i = skipPythonSpaces(content, valueEnd)
		if i >= len(content) || content[i] == '\n' || content[i] == '\r' {
			return 0, fmt.Errorf("%w: unterminated or multi-line inline table", ErrInvalidPythonSurface)
		}
		switch content[i] {
		case ',':
			i++
		case '}':
			return i + 1, nil
		default:
			return 0, fmt.Errorf("%w: expected ',' or '}' in an inline table", ErrInvalidPythonSurface)
		}
	}
}

func lexBasicString(content string, start int) (string, int, error) {
	var builder strings.Builder
	i := start + 1
	for {
		if i >= len(content) {
			return "", 0, fmt.Errorf("%w: unterminated string", ErrInvalidPythonSurface)
		}
		c := content[i]
		switch {
		case c == '"':
			return builder.String(), i + 1, nil
		case c == '\\':
			decoded, next, err := lexPythonEscape(content, i)
			if err != nil {
				return "", 0, err
			}
			builder.WriteString(decoded)
			i = next
		case c < 0x20 && c != '\t':
			return "", 0, fmt.Errorf("%w: control character inside a string", ErrInvalidPythonSurface)
		default:
			builder.WriteByte(c)
			i++
		}
	}
}

func lexMultilineBasicString(content string, start int) (string, int, error) {
	var builder strings.Builder
	i := start + 3
	if i < len(content) && content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n' {
		i += 2
	} else if i < len(content) && content[i] == '\n' {
		i++
	}
	for {
		if i >= len(content) {
			return "", 0, fmt.Errorf("%w: unterminated multi-line string", ErrInvalidPythonSurface)
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
					return "", 0, fmt.Errorf("%w: unterminated multi-line string", ErrInvalidPythonSurface)
				}
				i = k
				continue
			}
			decoded, next, err := lexPythonEscape(content, i)
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

func lexLiteralString(content string, start int) (string, int, error) {
	i := start + 1
	for {
		if i >= len(content) {
			return "", 0, fmt.Errorf("%w: unterminated string", ErrInvalidPythonSurface)
		}
		c := content[i]
		if c == '\'' {
			return content[start+1 : i], i + 1, nil
		}
		if c == '\n' || c == '\r' || c < 0x20 && c != '\t' {
			return "", 0, fmt.Errorf("%w: invalid character inside a literal string", ErrInvalidPythonSurface)
		}
		i++
	}
}

func lexMultilineLiteralString(content string, start int) (string, int, error) {
	i := start + 3
	if i < len(content) && content[i] == '\r' && i+1 < len(content) && content[i+1] == '\n' {
		i += 2
	} else if i < len(content) && content[i] == '\n' {
		i++
	}
	begin := i
	for {
		if i >= len(content) {
			return "", 0, fmt.Errorf("%w: unterminated multi-line literal string", ErrInvalidPythonSurface)
		}
		if strings.HasPrefix(content[i:], `'''`) {
			return content[begin:i], i + 3, nil
		}
		i++
	}
}

func lexPythonEscape(content string, backslash int) (string, int, error) {
	if backslash+1 >= len(content) {
		return "", 0, fmt.Errorf("%w: unterminated escape sequence", ErrInvalidPythonSurface)
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
		return lexPythonUnicodeEscape(content, backslash, 4)
	case 'U':
		return lexPythonUnicodeEscape(content, backslash, 8)
	default:
		return "", 0, fmt.Errorf("%w: unknown escape sequence", ErrInvalidPythonSurface)
	}
}

func lexPythonUnicodeEscape(content string, backslash int, digits int) (string, int, error) {
	if backslash+2+digits > len(content) {
		return "", 0, fmt.Errorf("%w: short unicode escape sequence", ErrInvalidPythonSurface)
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
			return "", 0, fmt.Errorf("%w: invalid unicode escape sequence", ErrInvalidPythonSurface)
		}
		value = value<<4 | digit
	}
	if value > 0x10FFFF || value >= 0xD800 && value <= 0xDFFF {
		return "", 0, fmt.Errorf("%w: invalid unicode escape sequence", ErrInvalidPythonSurface)
	}
	return string(rune(value)), backslash + 2 + digits, nil
}

func skipPythonBlank(content string, i int) int {
	for i < len(content) {
		switch content[i] {
		case ' ', '\t', '\n', '\r':
			i++
		case '#':
			i = skipPythonComment(content, i)
		default:
			return i
		}
	}
	return i
}

func skipPythonSpaces(content string, i int) int {
	for i < len(content) && (content[i] == ' ' || content[i] == '\t') {
		i++
	}
	return i
}

func skipPythonComment(content string, i int) int {
	for i < len(content) && content[i] != '\n' {
		i++
	}
	return i
}
