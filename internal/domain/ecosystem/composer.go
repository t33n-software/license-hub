// The Composer ecosystem adapter: it derives and aligns the license metadata
// surface of a composer.json manifest from the declaration chain and the
// license lock. The declaration is the truth, the manifest field is its
// deterministic projection, and the license lock is the single source of
// truth of the license type. The composer license value is an SPDX license
// expression (a string) or an array of SPDX license expressions; the
// proprietary form is the documented non-SPDX value, and the compound
// expression forms (A or B, A and B) are legal string expressions.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// ComposerLanguage is the seam toolchain language that selects the PHP/
// Composer matrix row.
const ComposerLanguage = "composer"

// The Composer license metadata surface: the manifest file and the field
// that carries the projected license declaration.
const (
	ComposerManifestName = "composer.json"
	ComposerLicenseField = "license"
)

// ErrMultipleComposerLicenses marks a license array that declares more than
// one license expression. The projection carries exactly one license family;
// the resolution is an explicit tenant decision.
var ErrMultipleComposerLicenses = errors.New("composer.json declares multiple license entries")

// ErrInvalidComposerLicenseValue marks a license value that is neither a
// string expression nor an array of string expressions.
var ErrInvalidComposerLicenseValue = errors.New("composer.json license value is not a string or array of strings")

// ComposerProjection derives the expected composer.json license value from
// the merged tenant values: a declared SPDX identifier projects itself; the
// file-based custom family projects the SPDX sideload expression
// LicenseRef-<LICENSE_ID> — the composer field carries SPDX expressions, and
// the sideload form keeps the license family identity. The derivation never
// reads the manifest.
func ComposerProjection(merged map[string]string) string {
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		return identifier
	}
	return "LicenseRef-" + merged["LICENSE_ID"]
}

// ComposerLicenseState classifies the license member of a parsed manifest.
type ComposerLicenseState int

const (
	// ComposerLicenseMissing marks an absent license field.
	ComposerLicenseMissing ComposerLicenseState = iota
	// ComposerLicenseString marks the target form: a string expression.
	ComposerLicenseString
	// ComposerLicenseArray marks the array form; the declared value of a
	// one-element array is its single entry.
	ComposerLicenseArray
	// ComposerLicenseInvalid marks a value that is neither a string nor an
	// array of strings.
	ComposerLicenseInvalid
)

// ComposerLicenseSurface is the license-relevant view of a composer.json.
type ComposerLicenseSurface struct {
	// State classifies the license field.
	State ComposerLicenseState
	// Value carries the declared license expression: the string form itself
	// or the single entry of a one-element array.
	Value string
	// Entries carries the decoded entries when the form is the array.
	Entries []string
}

// InspectComposerLicense parses a composer.json manifest and classifies its
// top-level license member. A manifest that declares the license member more
// than once is refused fail-closed.
func InspectComposerLicense(content string) (ComposerLicenseSurface, error) {
	members, _, err := parseDocument(content)
	if err != nil {
		return ComposerLicenseSurface{}, err
	}
	index := -1
	for i, member := range members {
		if member.key == ComposerLicenseField {
			if index >= 0 {
				return ComposerLicenseSurface{}, ErrAmbiguousField
			}
			index = i
		}
	}
	if index < 0 {
		return ComposerLicenseSurface{State: ComposerLicenseMissing}, nil
	}
	member := members[index]
	if member.isString {
		return ComposerLicenseSurface{State: ComposerLicenseString, Value: member.str}, nil
	}
	if content[member.valueStart] != '[' {
		return ComposerLicenseSurface{State: ComposerLicenseInvalid}, nil
	}
	entries, err := scanComposerArray(content[member.valueStart:member.valueEnd])
	if err != nil {
		return ComposerLicenseSurface{}, err
	}
	if len(entries) > 1 {
		return ComposerLicenseSurface{State: ComposerLicenseArray, Entries: entries}, nil
	}
	return ComposerLicenseSurface{State: ComposerLicenseArray, Value: entries[0], Entries: entries}, nil
}

// AlignComposerLicense returns the manifest content with the top-level
// license member aligned to target. The projected form is the string
// expression: a diverging string value or a one-element array is aligned to
// the projected string form — the one-element array is the value-equal
// alternative spelling of the same declaration — and a license array that
// declares more than one expression is refused fail-closed, because the
// projection carries exactly one license family and the resolution is an
// explicit tenant decision. Every byte outside the replaced value span or
// the inserted member is preserved exactly; a re-run over an aligned
// manifest is a no-op. A non-string, non-array license value is refused
// fail-closed.
func AlignComposerLicense(content, target string) (string, bool, error) {
	if !utf8.ValidString(target) {
		return "", false, ErrInvalidTarget
	}
	members, obj, err := parseDocument(content)
	if err != nil {
		return "", false, err
	}
	index := -1
	for i, member := range members {
		if member.key == ComposerLicenseField {
			if index >= 0 {
				return "", false, ErrAmbiguousField
			}
			index = i
		}
	}
	if index < 0 {
		insertion := licenseMember(target)
		if len(members) > 0 {
			insertion += ","
		}
		return content[:obj.start+1] + insertion + content[obj.start+1:], true, nil
	}
	member := members[index]
	switch {
	case member.isString:
		if member.str == target {
			return content, false, nil
		}
	case content[member.valueStart] == '[':
		entries, err := scanComposerArray(content[member.valueStart:member.valueEnd])
		if err != nil {
			return "", false, err
		}
		if len(entries) > 1 {
			return "", false, ErrMultipleComposerLicenses
		}
		if len(entries) == 1 && entries[0] == target {
			return content, false, nil
		}
	default:
		return "", false, ErrInvalidComposerLicenseValue
	}
	return content[:member.valueStart] + licenseValue(target) + content[member.valueEnd:], true, nil
}

// scanComposerArray validates the raw value span of an array license member
// and returns its decoded string entries. An empty array, a non-string
// entry, and every malformed delimiter form are refused fail-closed.
func scanComposerArray(raw string) ([]string, error) {
	i := skipSpace(raw, 0)
	if i >= len(raw) || raw[i] != '[' {
		return nil, ErrInvalidComposerLicenseValue
	}
	i = skipSpace(raw, i+1)
	if i < len(raw) && raw[i] == ']' {
		return nil, ErrInvalidComposerLicenseValue
	}
	entries := []string{}
	for {
		if i >= len(raw) || raw[i] != '"' {
			return nil, ErrInvalidComposerLicenseValue
		}
		end, err := parseStringToken(raw, i)
		if err != nil {
			return nil, err
		}
		entries = append(entries, decodeJSONString(raw[i:end]))
		i = skipSpace(raw, end)
		if i >= len(raw) {
			return nil, ErrInvalidComposerLicenseValue
		}
		switch raw[i] {
		case ',':
			i = skipSpace(raw, i+1)
			if i < len(raw) && raw[i] == ']' {
				return nil, ErrInvalidComposerLicenseValue
			}
		case ']':
			return entries, nil
		default:
			return nil, ErrInvalidComposerLicenseValue
		}
	}
}
