// The Rust ecosystem adapter: it derives and aligns the license metadata
// surface of a Cargo.toml manifest from the declaration chain and the
// license lock. The declaration is the truth, the manifest field is its
// deterministic projection, and the license lock is the single source of
// truth of the license type. The cargo license keys are mutually exclusive:
// "license", an SPDX 2.3 expression, is declared in lieu of "license-file",
// the path to the license text inside the package.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// CargoLanguage is the seam toolchain language that selects the Rust
// matrix row.
const CargoLanguage = "rust"

// The Rust license metadata surface: the manifest file and the mutually
// exclusive license keys of the [package] table.
const (
	CargoManifestName   = "Cargo.toml"
	CargoLicenseField   = "license"
	CargoLicenseFileKey = "license-file"
)

// CargoLicenseFileTarget is the projected license-file path: the canonical
// license text at the repository root.
const CargoLicenseFileTarget = "LICENSE"

// ErrInvalidCargoSurface marks a manifest whose dotted-key forms would
// create or shadow the license surface outside the canonical [package]
// table.
var ErrInvalidCargoSurface = errors.New("invalid Cargo.toml surface")

// ErrAmbiguousCargoField marks a table that declares a key more than once.
var ErrAmbiguousCargoField = errors.New("ambiguous Cargo.toml key")

// ErrExclusiveCargoLicense marks a [package] table that declares both
// license keys. The keys are mutually exclusive; the surface cannot be
// aligned or proven deterministically, and the resolution is an explicit
// tenant decision.
var ErrExclusiveCargoLicense = errors.New("Cargo.toml declares both exclusive license keys")

// ErrInvalidCargoLicenseValue marks a license value that is not a string
// expression.
var ErrInvalidCargoLicenseValue = errors.New("Cargo.toml license value is not a string expression")

// ErrInvalidCargoLicenseFile marks a license-file value that is not a
// string.
var ErrInvalidCargoLicenseFile = errors.New("Cargo.toml license-file value is not a string")

// ErrNoPackageTable marks a manifest without the [package] table; the
// license surface cannot be aligned deterministically without it.
var ErrNoPackageTable = errors.New("Cargo.toml carries no [package] table")

// CargoLicenseForm carries the projected license field of the lock: the
// exclusive key name and its string value.
type CargoLicenseForm struct {
	// Field is the projected key: CargoLicenseField or CargoLicenseFileKey.
	Field string
	// Value carries the projected string value.
	Value string
}

// String renders the projected field in its manifest assignment form.
func (form CargoLicenseForm) String() string {
	return form.Field + " = " + tomlQuote(form.Value)
}

// CargoProjection derives the expected Cargo.toml license field from the
// merged tenant values: a declared SPDX identifier projects the license
// expression key; the file-based custom family projects the license-file
// key pointing at the committed LICENSE text at the repository root. The
// derivation never reads the manifest.
func CargoProjection(merged map[string]string) CargoLicenseForm {
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		return CargoLicenseForm{Field: CargoLicenseField, Value: identifier}
	}
	return CargoLicenseForm{Field: CargoLicenseFileKey, Value: CargoLicenseFileTarget}
}

// CargoLicenseFieldState classifies the license surface of the [package]
// table.
type CargoLicenseFieldState int

const (
	// CargoLicenseMissing marks an absent license surface.
	CargoLicenseMissing CargoLicenseFieldState = iota
	// CargoLicenseExpression marks the license expression key form.
	CargoLicenseExpression
	// CargoLicenseFileForm marks the license-file key form.
	CargoLicenseFileForm
	// CargoLicenseInvalid marks a present key whose value is not a string.
	CargoLicenseInvalid
)

// CargoLicenseSurface is the license-relevant view of a Cargo.toml.
type CargoLicenseSurface struct {
	// State classifies the license surface.
	State CargoLicenseFieldState
	// Value carries the decoded string value of the declared key.
	Value string
}

// InspectCargoLicense parses a Cargo.toml manifest and classifies the
// license surface of its [package] table. A manifest that declares both
// exclusive license keys is refused fail-closed.
func InspectCargoLicense(content string) (CargoLicenseSurface, error) {
	result := CargoLicenseSurface{State: CargoLicenseMissing}
	surface, err := scanCargoPackage(content)
	if err != nil {
		return CargoLicenseSurface{}, err
	}
	if surface == nil {
		return result, nil
	}
	license, licenseFile := surface.license, surface.licenseFile
	switch {
	case license != nil && licenseFile != nil:
		return CargoLicenseSurface{}, ErrExclusiveCargoLicense
	case license != nil:
		if license.kind != tomlValueString {
			result.State = CargoLicenseInvalid
		} else {
			result.State = CargoLicenseExpression
			result.Value = license.str
		}
	case licenseFile != nil:
		if licenseFile.kind != tomlValueString {
			result.State = CargoLicenseInvalid
		} else {
			result.State = CargoLicenseFileForm
			result.Value = licenseFile.str
		}
	}
	return result, nil
}

// AlignCargoLicense returns the manifest content with the license surface of
// the [package] table aligned to the projected license form. The cargo
// license keys are mutually exclusive: a form switch replaces the declared
// key line, and a manifest that declares both keys is refused fail-closed.
// Every byte outside the replaced key-to-value span or the inserted line is
// preserved exactly; a re-run over an aligned manifest is a no-op. A
// non-string license value is refused fail-closed.
func AlignCargoLicense(content string, form CargoLicenseForm) (string, bool, error) {
	if !utf8.ValidString(form.Value) {
		return "", false, ErrInvalidTarget
	}
	surface, err := scanCargoPackage(content)
	if err != nil {
		return "", false, err
	}
	if surface == nil {
		return "", false, ErrNoPackageTable
	}
	license, licenseFile := surface.license, surface.licenseFile
	switch {
	case license != nil && licenseFile != nil:
		return "", false, ErrExclusiveCargoLicense
	case license != nil:
		if license.kind != tomlValueString {
			return "", false, ErrInvalidCargoLicenseValue
		}
		if form.Field == CargoLicenseField {
			if license.str == form.Value {
				return content, false, nil
			}
			return content[:license.valueStart] + tomlQuote(form.Value) + content[license.valueEnd:], true, nil
		}
		return replaceCargoKey(content, license, form), true, nil
	case licenseFile != nil:
		if licenseFile.kind != tomlValueString {
			return "", false, ErrInvalidCargoLicenseFile
		}
		if form.Field == CargoLicenseFileKey {
			if licenseFile.str == form.Value {
				return content, false, nil
			}
			return content[:licenseFile.valueStart] + tomlQuote(form.Value) + content[licenseFile.valueEnd:], true, nil
		}
		return replaceCargoKey(content, licenseFile, form), true, nil
	default:
		insertion := form.String() + "\n"
		return applyTOMLEdits(content, []tomlEdit{tomlLineInsertion(content, surface.headerEnd, insertion)}), true, nil
	}
}

// replaceCargoKey replaces the key-to-value span of a declared license key
// with the projected assignment; the form switch keeps every other byte,
// including a trailing comment after the value, exactly.
func replaceCargoKey(content string, key *tomlKeyValue, form CargoLicenseForm) string {
	return content[:key.keyStart] + form.String() + content[key.valueEnd:]
}

type cargoPackageSurface struct {
	start       int
	headerEnd   int
	license     *tomlKeyValue
	licenseFile *tomlKeyValue
}

// scanCargoPackage lexes a Cargo.toml document and returns the
// license-relevant surface of its [package] table. A manifest without a
// [package] table returns a nil surface; a document the scanner cannot
// interpret structurally is refused fail-closed.
func scanCargoPackage(content string) (*cargoPackageSurface, error) {
	tables, err := scanTOMLDocument(content, checkCargoKey)
	if err != nil {
		return nil, err
	}
	for _, table := range tables {
		if table.name != "package" || table.array {
			continue
		}
		surface := &cargoPackageSurface{start: table.start, headerEnd: table.headerEnd}
		for index := range table.keys {
			key := &table.keys[index]
			switch key.key {
			case CargoLicenseField:
				surface.license = key
			case CargoLicenseFileKey:
				surface.licenseFile = key
			}
		}
		return surface, nil
	}
	return nil, nil
}

// checkCargoKey refuses the duplicate and shadow forms the scanner cannot
// interpret: a repeated key inside one table, a dotted key inside [package]
// whose first segment names a license-relevant key, and any root
// declaration of the package name itself.
func checkCargoKey(table *tomlTable, key *tomlKeyValue) error {
	for _, existing := range table.keys {
		if existing.key == key.key {
			return fmt.Errorf("%w: duplicate key %q in [%s]", ErrAmbiguousCargoField, key.key, table.name)
		}
	}
	if table.name == "package" && !table.array && len(key.segments) > 1 {
		switch key.segments[0] {
		case CargoLicenseField, CargoLicenseFileKey:
			return fmt.Errorf("%w: dotted key %q shadows the %q surface in [package]",
				ErrInvalidCargoSurface, key.key, key.segments[0])
		}
	}
	if table.name == "" && len(key.segments) > 0 && key.segments[0] == "package" {
		return fmt.Errorf("%w: the package table must be declared with the [package] header", ErrInvalidCargoSurface)
	}
	return nil
}
