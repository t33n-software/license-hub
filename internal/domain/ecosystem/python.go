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

// ErrInvalidPythonSurface marks a manifest whose dotted-key forms would
// create or shadow the license surface outside the canonical [project]
// table.
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
	return tomlStringArray(files)
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
		case tomlValueString:
			result.State = PythonLicenseString
			result.Value = key.str
		case tomlValueInlineTable:
			result.State = PythonLicenseTable
		default:
			result.State = PythonLicenseInvalid
		}
	}
	if key := surface.licenseFiles; key != nil {
		result.LicenseFilesSet = true
		if key.kind == tomlValueArray {
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
	if key := surface.classifiers; key != nil && key.kind == tomlValueArray {
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
	edits := []spanEdit{}
	switch {
	case surface.license == nil:
		// The insertion cases below carry the license field.
	case surface.license.kind == tomlValueString:
		if surface.license.str != target {
			edits = append(edits, spanEdit{
				start: surface.license.valueStart,
				end:   surface.license.valueEnd,
				text:  tomlQuote(target),
			})
		}
	case surface.license.kind == tomlValueInlineTable:
		edits = append(edits, spanEdit{
			start: surface.license.valueStart,
			end:   surface.license.valueEnd,
			text:  tomlQuote(target),
		})
	default:
		return "", false, ErrInvalidPythonLicenseValue
	}
	switch {
	case surface.licenseFiles == nil:
		// The insertion cases below carry the license-files field.
	case surface.licenseFiles.kind != tomlValueArray:
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
			edits = append(edits, spanEdit{
				start: surface.licenseFiles.valueStart,
				end:   surface.licenseFiles.valueEnd,
				text:  tomlStringArray(files),
			})
		}
	}
	switch {
	case surface.license == nil && surface.licenseFiles == nil:
		edits = append(edits, spanLineInsertion(content, surface.headerEnd,
			"license = "+tomlQuote(target)+"\nlicense-files = "+tomlStringArray(files)+"\n"))
	case surface.license == nil:
		edits = append(edits, spanLineInsertion(content, surface.headerEnd,
			"license = "+tomlQuote(target)+"\n"))
	case surface.licenseFiles == nil:
		edits = append(edits, spanLineInsertion(content, surface.license.lineEnd,
			"license-files = "+tomlStringArray(files)+"\n"))
	}
	if len(edits) == 0 {
		return content, false, nil
	}
	return applySpanEdits(content, edits), true, nil
}

type pythonProjectSurface struct {
	start        int
	headerEnd    int
	license      *tomlKeyValue
	licenseFiles *tomlKeyValue
	classifiers  *tomlKeyValue
}

// scanPythonProject lexes a pyproject.toml document and returns the
// license-relevant surface of its [project] table. A manifest without a
// [project] table returns a nil surface; a document the scanner cannot
// interpret structurally is refused fail-closed.
func scanPythonProject(content string) (*pythonProjectSurface, error) {
	tables, err := scanTOMLDocument(content, checkPythonKey)
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

// checkPythonKey refuses the duplicate and shadow forms the scanner cannot
// interpret: a repeated key inside one table, a dotted key inside [project]
// whose first segment names a license-relevant key, and any root
// declaration of the project name itself.
func checkPythonKey(table *tomlTable, key *tomlKeyValue) error {
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
