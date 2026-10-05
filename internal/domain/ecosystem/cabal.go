// The Haskell ecosystem adapter: it derives and aligns the license metadata
// surface of a .cabal package description from the declaration chain and the
// license lock. The declaration is the truth, the manifest fields are its
// deterministic projection, and the license lock is the single source of
// truth of the license type. The projected surface is the pair of the
// top-level license field (an SPDX expression, case-sensitive since
// cabal-version: 2.2, where LicenseRef-<idstring> is the documented form for
// custom licenses included in the package) and the top-level license-file
// field carrying the canonical license text at the repository root; the
// license-files list form with exactly the projected file is the value-equal
// alternative spelling of the same declaration.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// HaskellLanguage is the seam toolchain language that selects the Haskell
// matrix row.
const HaskellLanguage = "haskell"

// The Haskell license metadata surface: the manifest suffix the adapter
// discovers, the license fields of the top-level package description, and
// the file target of the projected pair.
const (
	CabalSuffix          = ".cabal"
	CabalLicenseKey      = "license"
	CabalLicenseFileKey  = "license-file"
	CabalLicenseFilesKey = "license-files"
	CabalFileTarget      = "LICENSE"
)

// ErrInvalidCabalSurface marks a .cabal the scanner cannot lex or anchor: a
// structural break anywhere, indented content before any top-level property,
// or a license surface that lives inside a component section instead of the
// top-level package description.
var ErrInvalidCabalSurface = errors.New("invalid .cabal surface")

// ErrAmbiguousCabalLicense marks a package description that declares the
// license field more than once.
var ErrAmbiguousCabalLicense = errors.New(".cabal declares the license field more than once")

// ErrAmbiguousCabalLicenseFile marks a package description that declares the
// license file surface more than once: a duplicate license-file field, a
// duplicate license-files field, or both forms together. The projection
// carries exactly one file form; the resolution is an explicit tenant
// decision.
var ErrAmbiguousCabalLicenseFile = errors.New(".cabal declares the license file surface more than once")

// ErrMultipleCabalLicenseFiles marks a license-files field that declares
// more than one license file entry. The projection carries exactly one
// license text; the resolution is an explicit tenant decision.
var ErrMultipleCabalLicenseFiles = errors.New(".cabal declares multiple license file entries")

// ErrInvalidCabalLicenseValue marks a declared license field that carries no
// value; the alignment contract cannot align an empty surface.
var ErrInvalidCabalLicenseValue = errors.New(".cabal license value is empty")

// CabalProjection derives the expected .cabal license field from the merged
// tenant values: a declared SPDX identifier projects itself; the file-based
// custom family projects the SPDX sideload expression LicenseRef-<LICENSE_ID>
// — the documented non-SPDX expression form keeps the license family
// identity. The derivation never reads the manifest.
func CabalProjection(merged map[string]string) string {
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		return identifier
	}
	return "LicenseRef-" + merged["LICENSE_ID"]
}

// CabalManifestNames filters a directory listing down to the .cabal manifest
// candidates, in sorted order. The suffix match is case-insensitive.
func CabalManifestNames(entries []string) []string {
	names := []string{}
	for _, entry := range entries {
		if strings.HasSuffix(strings.ToLower(entry), CabalSuffix) {
			names = append(names, entry)
		}
	}
	slices.Sort(names)
	return names
}

// CabalLicenseState classifies the license field of a scanned .cabal.
type CabalLicenseState int

const (
	// CabalLicenseMissing marks an absent license field.
	CabalLicenseMissing CabalLicenseState = iota
	// CabalLicenseValue marks the target form: a decoded license value.
	CabalLicenseValue
	// CabalLicenseInvalid marks a declared license field that carries no
	// value.
	CabalLicenseInvalid
)

// CabalFileState classifies the license file surface of a scanned .cabal.
type CabalFileState int

const (
	// CabalFileMissing marks an absent license file surface.
	CabalFileMissing CabalFileState = iota
	// CabalFileSingle marks the target form: the license-file field.
	CabalFileSingle
	// CabalFileList marks the license-files list form; the declared value of
	// a one-entry list is its single entry.
	CabalFileList
	// CabalFileInvalid marks a declared file field that carries no value.
	CabalFileInvalid
)

// CabalLicenseSurface is the license-relevant view of a scanned .cabal.
type CabalLicenseSurface struct {
	// State classifies the license field.
	State CabalLicenseState
	// Value carries the decoded license value.
	Value string
	// FileState classifies the license file surface.
	FileState CabalFileState
	// FileValue carries the decoded license-file value.
	FileValue string
	// FileEntries carries the decoded entries of the license-files list.
	FileEntries []string
}

// cabalField carries one scanned top-level license field: the observed key
// spelling, the byte spans of the property block and its value, the decoded
// value, and the decoded entries of the list form.
type cabalField struct {
	key        string
	propStart  int
	blockEnd   int // exclusive index of the newline that terminates the property block
	valueStart int
	valueEnd   int
	value      string
	entries    []string
	invalid    bool
}

// cabalScan carries the license-relevant skeleton of a scanned .cabal: the
// license fields (nil when absent) and the insertion anchor, the exclusive
// index just past the first top-level property block (-1 when the document
// carries none).
type cabalScan struct {
	license      *cabalField
	licenseFile  *cabalField
	licenseFiles *cabalField
	anchor       int
}

// InspectCabalLicense lexes a .cabal file and classifies its license fields.
// A file that declares the license field or the license file surface more
// than once, carries a license surface inside a component section, or breaks
// structurally anywhere is refused fail-closed.
func InspectCabalLicense(content string) (CabalLicenseSurface, error) {
	scan, err := scanCabal(content)
	if err != nil {
		return CabalLicenseSurface{}, err
	}
	surface := CabalLicenseSurface{}
	switch {
	case scan.license == nil:
		surface.State = CabalLicenseMissing
	case scan.license.invalid:
		surface.State = CabalLicenseInvalid
	default:
		surface.State = CabalLicenseValue
		surface.Value = scan.license.value
	}
	switch {
	case scan.licenseFile == nil && scan.licenseFiles == nil:
		surface.FileState = CabalFileMissing
	case scan.licenseFile != nil:
		if scan.licenseFile.invalid {
			surface.FileState = CabalFileInvalid
		} else {
			surface.FileState = CabalFileSingle
			surface.FileValue = scan.licenseFile.value
		}
	case scan.licenseFiles.invalid:
		surface.FileState = CabalFileInvalid
	default:
		surface.FileState = CabalFileList
		surface.FileEntries = scan.licenseFiles.entries
	}
	return surface, nil
}

// cabalEdit carries one byte-span replacement of an alignment.
type cabalEdit struct {
	start int
	end   int
	text  string
}

// AlignCabalLicense returns the .cabal content with its license surface
// aligned to target: the license field is value-replaced or inserted, the
// license file surface is value-replaced, normalized from a diverging
// license-files list to the projected license-file form, or inserted. The
// license-files list carrying exactly the projected file is the value-equal
// alternative spelling of the same declaration and stays unchanged. Every
// byte outside the replaced value span, the replaced property block, or the
// inserted lines is preserved exactly; a re-run over an aligned file is a
// no-op. A list that declares more than one file entry, a file surface
// declared more than once, a license surface inside a component section, and
// a document without a top-level property to anchor the insertion are
// refused fail-closed.
func AlignCabalLicense(content, target string) (string, bool, error) {
	if err := validateCabalTarget(target); err != nil {
		return "", false, err
	}
	scan, err := scanCabal(content)
	if err != nil {
		return "", false, err
	}
	edits := []cabalEdit{}
	switch {
	case scan.license == nil:
		// inserted together with the file surface at the anchor
	case scan.license.invalid:
		return "", false, ErrInvalidCabalLicenseValue
	case scan.license.value != target:
		edits = append(edits, cabalEdit{start: scan.license.valueStart, end: scan.license.valueEnd, text: target})
	}
	fileInsertion := false
	switch {
	case scan.licenseFile == nil && scan.licenseFiles == nil:
		fileInsertion = true
	case scan.licenseFile != nil:
		if scan.licenseFile.invalid {
			return "", false, ErrInvalidCabalLicenseValue
		}
		if scan.licenseFile.value != CabalFileTarget {
			edits = append(edits, cabalEdit{start: scan.licenseFile.valueStart, end: scan.licenseFile.valueEnd, text: CabalFileTarget})
		}
	case scan.licenseFiles.invalid:
		return "", false, ErrInvalidCabalLicenseValue
	case len(scan.licenseFiles.entries) > 1:
		return "", false, ErrMultipleCabalLicenseFiles
	case scan.licenseFiles.entries[0] != CabalFileTarget:
		edits = append(edits, cabalEdit{
			start: scan.licenseFiles.propStart,
			end:   scan.licenseFiles.blockEnd,
			text:  CabalLicenseFileKey + ": " + CabalFileTarget,
		})
	}
	if scan.license == nil || fileInsertion {
		if scan.anchor < 0 {
			return "", false, fmt.Errorf("%w: the document carries no top-level property to anchor the insertion", ErrInvalidCabalSurface)
		}
		insertion := ""
		if scan.license == nil {
			insertion += "\n" + CabalLicenseKey + ": " + target
		}
		if fileInsertion {
			insertion += "\n" + CabalLicenseFileKey + ": " + CabalFileTarget
		}
		edits = append(edits, cabalEdit{start: scan.anchor, end: scan.anchor, text: insertion})
	}
	if len(edits) == 0 {
		return content, false, nil
	}
	slices.SortFunc(edits, func(a, b cabalEdit) int { return b.start - a.start })
	result := content
	for _, edit := range edits {
		result = result[:edit.start] + edit.text + result[edit.end:]
	}
	return result, true, nil
}

// validateCabalTarget refuses alignment targets the .cabal token contract
// cannot carry: a non-UTF-8 value, an empty value, a compound expression
// (whitespace inside the value), and a value that is not representable as an
// unquoted .cabal token.
func validateCabalTarget(target string) error {
	if !utf8.ValidString(target) || target == "" {
		return ErrInvalidTarget
	}
	for _, r := range target {
		switch {
		case unicode.IsSpace(r):
			return fmt.Errorf("%w: the license value carries whitespace (compound expressions are not supported)", ErrInvalidTarget)
		case r == '"' || r == '\\' || r == '{' || r == '}':
			return fmt.Errorf("%w: the license value is not representable as an unquoted .cabal token", ErrInvalidTarget)
		}
	}
	if strings.HasPrefix(target, "--") {
		return fmt.Errorf("%w: the license value would lex as a line comment", ErrInvalidTarget)
	}
	return nil
}

// scanCabal lexes the license-relevant skeleton of a .cabal file line by
// line: line comments and blank lines are transparent, column-0 property
// lines are top-level fields, every other column-0 line opens a component
// section, and indented lines attach as continuations of the open top-level
// field or stay section content. A structural break anywhere — an
// unterminated quoted value, an unterminated brace list, indented content
// before any top-level property, or a license surface inside a component
// section — refuses the whole surface fail-closed.
func scanCabal(content string) (*cabalScan, error) {
	scan := &cabalScan{anchor: -1}
	var firstProp *cabalField
	lastProp := (*cabalField)(nil)
	section := false
	i := 0
	for i < len(content) {
		end := i
		for end < len(content) && content[end] != '\n' {
			end++
		}
		nl := -1
		if end < len(content) {
			nl = end
		}
		li := i
		for li < end && (content[li] == ' ' || content[li] == '\t') {
			li++
		}
		trimEnd := end
		for trimEnd > li && (content[trimEnd-1] == ' ' || content[trimEnd-1] == '\t' || content[trimEnd-1] == '\r') {
			trimEnd--
		}
		indented := li > i
		blank := li >= trimEnd
		comment := !blank && content[li] == '-' && li+1 < trimEnd && content[li+1] == '-'
		switch {
		case blank || comment:
			// transparent: neither a field boundary nor section content
		case !indented:
			key, colon := cabalPropertyKey(content, li, trimEnd)
			if colon < 0 {
				section = true
				lastProp = nil
				break
			}
			field, err := cabalOpenField(content, key, i, colon+1, trimEnd, nl, end)
			if err != nil {
				return nil, err
			}
			if err := scanCabalField(scan, field); err != nil {
				return nil, err
			}
			if firstProp == nil {
				firstProp = field
			}
			lastProp = field
			section = false
		case section:
			key, colon := cabalPropertyKey(content, li, trimEnd)
			if colon >= 0 && isCabalLicenseKey(key) {
				return nil, fmt.Errorf("%w: the license surface lives inside a component section", ErrInvalidCabalSurface)
			}
		case lastProp == nil:
			return nil, fmt.Errorf("%w: the document opens with indented content before any top-level property", ErrInvalidCabalSurface)
		default:
			if err := cabalContinueField(content, lastProp, li, trimEnd, nl, end); err != nil {
				return nil, err
			}
		}
		if nl >= 0 {
			i = nl + 1
		} else {
			i = end
		}
	}
	if firstProp != nil {
		scan.anchor = firstProp.blockEnd
	}
	return scan, nil
}

// cabalOpenField opens one top-level license field: the value span of the
// key line is parsed and the block end is recorded.
func cabalOpenField(content string, key string, propStart, valueStart, trimEnd, nl, lineEnd int) (*cabalField, error) {
	vs, ve, tokens, err := cabalValueTokens(content, valueStart, trimEnd)
	if err != nil {
		return nil, err
	}
	field := &cabalField{
		key:        key,
		propStart:  propStart,
		valueStart: vs,
		valueEnd:   ve,
		value:      strings.Join(tokens, " "),
		entries:    tokens,
		invalid:    len(tokens) == 0,
	}
	if nl >= 0 {
		field.blockEnd = nl
	} else {
		field.blockEnd = lineEnd
	}
	return field, nil
}

// cabalContinueField attaches one continuation line to the open field: the
// tokens join the decoded value and the block end advances.
func cabalContinueField(content string, field *cabalField, valueStart, trimEnd, nl, lineEnd int) error {
	_, ve, tokens, err := cabalValueTokens(content, valueStart, trimEnd)
	if err != nil {
		return err
	}
	field.entries = append(field.entries, tokens...)
	field.value = strings.Join(field.entries, " ")
	if len(tokens) > 0 {
		field.invalid = false
	}
	if ve > field.valueEnd {
		field.valueEnd = ve
	}
	if nl >= 0 {
		field.blockEnd = nl
	} else {
		field.blockEnd = lineEnd
	}
	return nil
}

// scanCabalField binds one opened field into the scan and refuses duplicate
// declarations fail-closed.
func scanCabalField(scan *cabalScan, field *cabalField) error {
	switch field.key {
	case CabalLicenseKey:
		if scan.license != nil {
			return ErrAmbiguousCabalLicense
		}
		scan.license = field
	case CabalLicenseFileKey, CabalLicenseFilesKey:
		if scan.licenseFile != nil || scan.licenseFiles != nil {
			return ErrAmbiguousCabalLicenseFile
		}
		if field.key == CabalLicenseFileKey {
			scan.licenseFile = field
		} else {
			scan.licenseFiles = field
		}
	}
	return nil
}

// isCabalLicenseKey reports whether the lowercase key names a license
// surface.
func isCabalLicenseKey(key string) bool {
	switch key {
	case CabalLicenseKey, CabalLicenseFileKey, CabalLicenseFilesKey:
		return true
	}
	return false
}

// cabalPropertyKey matches a column-0 property line and returns its
// lowercase key and the colon index (-1 when the line opens no property).
func cabalPropertyKey(content string, start, end int) (string, int) {
	if start >= end || !isCabalKeyStart(content[start]) {
		return "", -1
	}
	j := start + 1
	for j < end && isCabalKeyByte(content[j]) {
		j++
	}
	k := j
	for k < end && (content[k] == ' ' || content[k] == '\t') {
		k++
	}
	if k >= end || content[k] != ':' {
		return "", -1
	}
	return strings.ToLower(content[start:j]), k
}

// cabalValueTokens parses the value region of one property or continuation
// line: the comment-cut, whitespace-trimmed span, and the decoded tokens —
// the interior of a quoted value, the entries of a brace list, or the
// whitespace-split token run.
func cabalValueTokens(content string, valueStart, trimEnd int) (int, int, []string, error) {
	i := valueStart
	for i < trimEnd && (content[i] == ' ' || content[i] == '\t') {
		i++
	}
	rawEnd := trimEnd
	for j := i; j < rawEnd; j++ {
		if content[j] == '-' && j+1 < rawEnd && content[j+1] == '-' &&
			(j == i || content[j-1] == ' ' || content[j-1] == '\t') {
			rawEnd = j
			break
		}
	}
	e := rawEnd
	for e > i && (content[e-1] == ' ' || content[e-1] == '\t' || content[e-1] == '\r') {
		e--
	}
	if i >= e {
		return i, i, nil, nil
	}
	switch content[i] {
	case '"':
		quoteEnd := -1
		for j := i + 1; j < e; j++ {
			if content[j] == '\\' {
				return 0, 0, nil, fmt.Errorf("%w: the quoted license value carries escapes", ErrInvalidCabalSurface)
			}
			if content[j] == '"' {
				quoteEnd = j
				break
			}
		}
		if quoteEnd < 0 {
			return 0, 0, nil, fmt.Errorf("%w: unterminated quoted license value", ErrInvalidCabalSurface)
		}
		if strings.TrimLeft(content[quoteEnd+1:e], " \t") != "" {
			return 0, 0, nil, fmt.Errorf("%w: the quoted license value carries trailing tokens", ErrInvalidCabalSurface)
		}
		return i, quoteEnd + 1, []string{content[i+1 : quoteEnd]}, nil
	case '{':
		braceEnd := -1
		for j := i + 1; j < e; j++ {
			if content[j] == '}' {
				braceEnd = j
				break
			}
		}
		if braceEnd < 0 {
			return 0, 0, nil, fmt.Errorf("%w: unterminated brace list", ErrInvalidCabalSurface)
		}
		if strings.TrimLeft(content[braceEnd+1:e], " \t") != "" {
			return 0, 0, nil, fmt.Errorf("%w: the brace list carries trailing tokens", ErrInvalidCabalSurface)
		}
		return i, braceEnd + 1, strings.Fields(content[i+1 : braceEnd]), nil
	default:
		return i, e, strings.Fields(content[i:e]), nil
	}
}

// isCabalKeyStart reports whether the byte may start a .cabal field name.
func isCabalKeyStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isCabalKeyByte reports whether the byte may appear inside a .cabal field
// name.
func isCabalKeyByte(c byte) bool {
	return isCabalKeyStart(c) || (c >= '0' && c <= '9') || c == '-'
}
