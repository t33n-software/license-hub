// Package ecosystem derives and aligns the license metadata surfaces of
// package-ecosystem manifests from the declaration chain and the license
// lock. The declaration is the truth, the manifest field is its
// deterministic projection, and the license lock is the single source of
// truth of the license type.
//
// Convention: spec/ecosystem-license-metadata.md
package ecosystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// NpmLanguage is the seam toolchain language that selects the npm matrix row.
const NpmLanguage = "node-typescript"

// The npm license metadata surface: the manifest file and the field that
// carries the projected license declaration.
const (
	NpmManifestName = "package.json"
	NpmLicenseField = "license"
)

// SeamFileName is the governed seam document that declares the toolchain
// identity of a project.
const SeamFileName = "git-governance.quality.json"

// NpmTargetForm is the bound organization target form for the npm license
// field of the file-based custom family: the npm-canonical reference that
// points the registry and every scanner at the committed LICENSE file at the
// repository root.
const NpmTargetForm = "SEE LICENSE IN LICENSE"

// ErrInvalidSurface marks a declared surface document that is not a
// scannable JSON surface.
var ErrInvalidSurface = errors.New("invalid JSON surface")

// ErrAmbiguousField marks a manifest that declares the license member more
// than once at the top level.
var ErrAmbiguousField = errors.New("ambiguous license member")

// ErrNonStringField marks a license member whose value is not a JSON string.
var ErrNonStringField = errors.New("license member is not a JSON string")

// SeamLanguage extracts the declared toolchain language of the governed seam
// document. A seam without a toolchain block declares no language.
func SeamLanguage(data []byte) (string, error) {
	var seam struct {
		Toolchain struct {
			Language string `json:"language"`
		} `json:"toolchain"`
	}
	if err := json.Unmarshal(data, &seam); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidSurface, err)
	}
	return strings.TrimSpace(seam.Toolchain.Language), nil
}

// NpmProjection derives the expected npm license field value from the merged
// tenant values: a declared SPDX identifier projects itself; the file-based
// custom family projects the bound target form. The derivation never reads
// the manifest.
func NpmProjection(merged map[string]string) string {
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		return identifier
	}
	return NpmTargetForm
}

// LicenseFieldState classifies the license member of a parsed manifest.
type LicenseFieldState int

const (
	LicenseFieldMissing LicenseFieldState = iota
	LicenseFieldString
	LicenseFieldNonString
)

// InspectNpmLicense parses an npm manifest and classifies its top-level
// license member.
func InspectNpmLicense(content string) (string, LicenseFieldState, error) {
	members, _, err := parseDocument(content)
	if err != nil {
		return "", LicenseFieldMissing, err
	}
	index := -1
	for i, member := range members {
		if member.key == NpmLicenseField {
			if index >= 0 {
				return "", LicenseFieldMissing, ErrAmbiguousField
			}
			index = i
		}
	}
	if index < 0 {
		return "", LicenseFieldMissing, nil
	}
	member := members[index]
	if !member.isString {
		return "", LicenseFieldNonString, nil
	}
	return member.str, LicenseFieldString, nil
}

// AlignNpmLicense returns the manifest content with the top-level license
// member aligned to target. Every byte outside the replaced value span or
// the inserted member is preserved exactly; a re-run over an aligned
// manifest is a no-op. A non-string license member is refused fail-closed.
func AlignNpmLicense(content, target string) (string, bool, error) {
	members, obj, err := parseDocument(content)
	if err != nil {
		return "", false, err
	}
	index := -1
	for i, member := range members {
		if member.key == NpmLicenseField {
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
	if !member.isString {
		return "", false, ErrNonStringField
	}
	if member.str == target {
		return content, false, nil
	}
	return content[:member.valueStart] + licenseValue(target) + content[member.valueEnd:], true, nil
}

// rootSpan carries the byte span of the root object of a document.
type rootSpan struct {
	start int // index of the opening brace
	end   int // index just past the closing brace
}

// memberSpan carries one top-level member of the root object.
type memberSpan struct {
	key        string // decoded key
	isString   bool   // the value is a JSON string
	str        string // decoded string value when isString
	valueStart int    // inclusive index of the raw value
	valueEnd   int    // exclusive index of the raw value
}

// parseDocument validates content as a JSON document and returns the member
// spans of its root object. The parser is a self-validating recursive
// descent scanner: every rejected form is an explicit error branch, so an
// accepted document is JSON that the standard decoder accepts.
func parseDocument(content string) ([]memberSpan, rootSpan, error) {
	start := skipSpace(content, 0)
	if start >= len(content) || content[start] != '{' {
		return nil, rootSpan{}, fmt.Errorf("%w: the root value is not an object", ErrInvalidSurface)
	}
	members := []memberSpan{}
	end, err := walkObject(content, start, func(member memberSpan) {
		members = append(members, member)
	})
	if err != nil {
		return nil, rootSpan{}, err
	}
	return members, rootSpan{start: start, end: end}, nil
}

// walkObject validates a JSON object starting at the opening brace and
// returns the index just past the closing brace. The record callback
// receives every top-level member span; nested walks record nothing.
func walkObject(content string, start int, record func(memberSpan)) (int, error) {
	i := skipSpace(content, start+1)
	if i < len(content) && content[i] == '}' {
		return i + 1, nil
	}
	for {
		if i >= len(content) || content[i] != '"' {
			return 0, ErrInvalidSurface
		}
		keyStart := i
		keyEnd, err := parseStringToken(content, i)
		if err != nil {
			return 0, err
		}
		i = skipSpace(content, keyEnd)
		if i >= len(content) || content[i] != ':' {
			return 0, ErrInvalidSurface
		}
		i = skipSpace(content, i+1)
		valueStart := i
		i, err = parseValue(content, i)
		if err != nil {
			return 0, err
		}
		if record != nil {
			member := memberSpan{valueStart: valueStart, valueEnd: i}
			if content[valueStart] == '"' {
				member.isString = true
				member.str = decodeJSONString(content[valueStart:i])
			}
			member.key = decodeJSONString(content[keyStart:keyEnd])
			record(member)
		}
		i = skipSpace(content, i)
		if i >= len(content) {
			return 0, ErrInvalidSurface
		}
		switch content[i] {
		case ',':
			i = skipSpace(content, i+1)
			if i < len(content) && content[i] == '}' {
				return 0, ErrInvalidSurface
			}
		case '}':
			return i + 1, nil
		default:
			return 0, ErrInvalidSurface
		}
	}
}

// walkArray validates a JSON array starting at the opening bracket and
// returns the index just past the closing bracket.
func walkArray(content string, start int) (int, error) {
	i := skipSpace(content, start+1)
	if i < len(content) && content[i] == ']' {
		return i + 1, nil
	}
	for {
		var err error
		i, err = parseValue(content, i)
		if err != nil {
			return 0, err
		}
		i = skipSpace(content, i)
		if i >= len(content) {
			return 0, ErrInvalidSurface
		}
		switch content[i] {
		case ',':
			i = skipSpace(content, i+1)
			if i < len(content) && content[i] == ']' {
				return 0, ErrInvalidSurface
			}
		case ']':
			return i + 1, nil
		default:
			return 0, ErrInvalidSurface
		}
	}
}

// parseValue validates one JSON value starting at i and returns the index
// just past it.
func parseValue(content string, i int) (int, error) {
	if i >= len(content) {
		return 0, ErrInvalidSurface
	}
	switch content[i] {
	case '"':
		return parseStringToken(content, i)
	case '{':
		return walkObject(content, i, nil)
	case '[':
		return walkArray(content, i)
	case 't':
		return parseLiteral(content, i, "true")
	case 'f':
		return parseLiteral(content, i, "false")
	case 'n':
		return parseLiteral(content, i, "null")
	default:
		return parseNumber(content, i)
	}
}

// parseStringToken validates a JSON string starting at the opening quote and
// returns the index just past the closing quote, including strict escape
// validation.
func parseStringToken(content string, start int) (int, error) {
	for j := start + 1; j < len(content); j++ {
		switch content[j] {
		case '\\':
			j++
			if j >= len(content) {
				return 0, ErrInvalidSurface
			}
			switch content[j] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				for k := 1; k <= 4; k++ {
					if j+k >= len(content) || !isHexDigit(content[j+k]) {
						return 0, ErrInvalidSurface
					}
				}
				j += 4
			default:
				return 0, ErrInvalidSurface
			}
		case '"':
			return j + 1, nil
		}
	}
	return 0, ErrInvalidSurface
}

// parseLiteral validates one of the three JSON literals.
func parseLiteral(content string, start int, literal string) (int, error) {
	if start+len(literal) <= len(content) && content[start:start+len(literal)] == literal {
		return start + len(literal), nil
	}
	return 0, ErrInvalidSurface
}

// parseNumber validates a JSON number through the number charset scan plus
// the exact standard decoder check, and returns the index just past it.
func parseNumber(content string, start int) (int, error) {
	j := start
	for j < len(content) && isNumberByte(content[j]) {
		j++
	}
	if j == start {
		return 0, ErrInvalidSurface
	}
	var number json.Number
	if err := json.Unmarshal([]byte(content[start:j]), &number); err != nil {
		return 0, ErrInvalidSurface
	}
	return j, nil
}

// licenseMember renders the license member of the target value.
func licenseMember(target string) string {
	return `"license": ` + licenseValue(target)
}

// licenseValue renders the JSON-encoded string form of the target value.
func licenseValue(target string) string {
	encoded, _ := json.Marshal(target)
	return string(encoded)
}

// decodeJSONString decodes a string token produced by parseStringToken. The
// decode of a validated token cannot fail.
func decodeJSONString(token string) string {
	var decoded string
	_ = json.Unmarshal([]byte(token), &decoded)
	return decoded
}

// skipSpace returns the index of the next non-whitespace byte.
func skipSpace(content string, i int) int {
	for i < len(content) && isSpaceByte(content[i]) {
		i++
	}
	return i
}

// isSpaceByte reports whether the byte is JSON whitespace.
func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// isHexDigit reports whether the byte is a hexadecimal digit.
func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// isNumberByte reports whether the byte may appear inside a JSON number
// token; the exact acceptance is decided by the decoder check.
func isNumberByte(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E':
		return true
	default:
		return false
	}
}