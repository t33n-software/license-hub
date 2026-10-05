// The .NET ecosystem adapter: it derives and aligns the license metadata
// surfaces of the NuGet packaging manifests from the declaration chain and
// the license lock. The declaration is the truth, the manifest field is its
// deterministic projection, and the license lock is the single source of
// truth of the license type. The row carries two parallel manifest
// surfaces: the nuspec `<license>` element (type "expression" or "file")
// and the MSBuild pack properties (`PackageLicenseExpression` /
// `PackageLicenseFile`); the licenseUrl declaration forms are deprecated
// and are proven, never rewritten.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

// NuGetLanguage is the seam toolchain language that selects the .NET/NuGet
// matrix row.
const NuGetLanguage = "nuget"

// The .NET license metadata surfaces: the manifest suffixes the adapter
// discovers, the projected license file path, and the captured element and
// attribute names.
const (
	NuspecSuffix              = ".nuspec"
	MSBuildSuffix             = ".csproj"
	NuGetLicenseFileTarget    = "LICENSE"
	NuspecLicenseElement      = "license"
	NuspecURLElement          = "licenseUrl"
	NuspecTypeAttribute       = "type"
	NuspecExpressionType      = "expression"
	NuspecFileType            = "file"
	MSBuildExpressionProperty = "PackageLicenseExpression"
	MSBuildFileProperty       = "PackageLicenseFile"
	MSBuildURLProperty        = "PackageLicenseUrl"
)

// ErrInvalidNuGetName marks a manifest name the adapter cannot attribute to
// one of its surfaces.
var ErrInvalidNuGetName = errors.New("manifest name is not a nuget license surface")

// ErrInvalidNuspecSurface marks a manifest that is not a well-formed nuspec
// surface.
var ErrInvalidNuspecSurface = errors.New("invalid nuspec surface")

// ErrMultipleNuspecLicenses marks a metadata element that declares more
// than one license element. The projection carries exactly one license
// family; the resolution is an explicit tenant decision.
var ErrMultipleNuspecLicenses = errors.New("nuspec declares multiple license elements")

// ErrInvalidNuspecLicenseType marks a license element whose type attribute
// is neither the expression nor the file form.
var ErrInvalidNuspecLicenseType = errors.New("nuspec license element carries no valid type")

// ErrInvalidMSBuildSurface marks a manifest that is not a well-formed
// MSBuild project surface.
var ErrInvalidMSBuildSurface = errors.New("invalid msbuild surface")

// ErrMultipleMSBuildLicenses marks a project that declares the license
// surface more than once across its property groups. The projection carries
// exactly one license family; the resolution is an explicit tenant decision.
var ErrMultipleMSBuildLicenses = errors.New("msbuild declares multiple license properties")

// NuGetLicenseForm carries the projected license surface of the lock: the
// form kind and its value.
type NuGetLicenseForm struct {
	// Expression is true when the projection is the SPDX expression form;
	// it is false for the license-file form of the file-based custom family.
	Expression bool
	// Value carries the projected value: the SPDX expression or the license
	// file path.
	Value string
}

// nuspecType carries the projected nuspec type attribute value.
func (form NuGetLicenseForm) nuspecType() string {
	if form.Expression {
		return NuspecExpressionType
	}
	return NuspecFileType
}

// NuspecLabel renders the projected license surface in its nuspec element
// form.
func (form NuGetLicenseForm) NuspecLabel() string {
	return "<license type=\"" + form.nuspecType() + "\">" + xmlEscape(form.Value) + "</license>"
}

// propertyName carries the projected MSBuild property name of the form.
func (form NuGetLicenseForm) propertyName() string {
	if form.Expression {
		return MSBuildExpressionProperty
	}
	return MSBuildFileProperty
}

// MSBuildLabel renders the projected license surface in its MSBuild property
// form.
func (form NuGetLicenseForm) MSBuildLabel() string {
	name := form.propertyName()
	return "<" + name + ">" + xmlEscape(form.Value) + "</" + name + ">"
}

// NuGetProjection derives the expected .NET license surface from the merged
// tenant values: a declared SPDX identifier projects the expression form;
// the file-based custom family projects the license-file form pointing at
// the committed LICENSE text at the repository root. The derivation never
// reads the manifest.
func NuGetProjection(merged map[string]string) NuGetLicenseForm {
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		return NuGetLicenseForm{Expression: true, Value: identifier}
	}
	return NuGetLicenseForm{Expression: false, Value: NuGetLicenseFileTarget}
}

// NuGetManifestNames filters a directory listing down to the NuGet manifest
// candidates, in sorted order. The suffix match is case-insensitive.
func NuGetManifestNames(entries []string) []string {
	names := []string{}
	for _, entry := range entries {
		lower := strings.ToLower(entry)
		if strings.HasSuffix(lower, NuspecSuffix) || strings.HasSuffix(lower, MSBuildSuffix) {
			names = append(names, entry)
		}
	}
	slices.Sort(names)
	return names
}

// NuGetSurfaceState classifies the declared license form of a scanned
// surface.
type NuGetSurfaceState int

const (
	// NuGetSurfaceMissing marks an absent license surface.
	NuGetSurfaceMissing NuGetSurfaceState = iota
	// NuGetSurfaceDeclared marks the target form: the expression or the
	// license-file declaration.
	NuGetSurfaceDeclared
	// NuGetSurfaceInvalid marks a declared license element whose type
	// attribute is neither the expression nor the file form.
	NuGetSurfaceInvalid
)

// NuGetSurface is the license-relevant view of one discovered NuGet
// manifest.
type NuGetSurface struct {
	// State classifies the declared license surface.
	State NuGetSurfaceState
	// Expression is true when the declared form is the expression type.
	Expression bool
	// Value carries the decoded declared value.
	Value string
	// FormCount carries the number of declared license surfaces; a count
	// above one is refused fail-closed by the alignment and is an explicit
	// tenant decision for the verify lane.
	FormCount int
	// DeprecatedURLPresent reports the deprecated licenseUrl declaration
	// form.
	DeprecatedURLPresent bool
}

// InspectNuGetSurface parses one discovered NuGet manifest and classifies
// its license-relevant surface. The manifest name selects the scanner; a
// name the adapter cannot attribute is refused fail-closed.
func InspectNuGetSurface(name, content string) (NuGetSurface, error) {
	switch {
	case strings.HasSuffix(strings.ToLower(name), NuspecSuffix):
		return inspectNuspecLicense(content)
	case strings.HasSuffix(strings.ToLower(name), MSBuildSuffix):
		return inspectMSBuildLicense(content)
	default:
		return NuGetSurface{}, ErrInvalidNuGetName
	}
}

// AlignNuGetSurface returns the manifest content with its license surface
// aligned to the projected license form. The manifest name selects the
// surface scanner. Every byte outside the replaced spans or the inserted
// blocks is preserved exactly; a re-run over an aligned manifest is a
// no-op. A surface that declares its license form more than once is
// refused fail-closed.
func AlignNuGetSurface(name, content string, form NuGetLicenseForm) (string, bool, error) {
	if !utf8.ValidString(form.Value) || !xmlRepresentable(form.Value) {
		return "", false, ErrInvalidTarget
	}
	switch {
	case strings.HasSuffix(strings.ToLower(name), NuspecSuffix):
		return alignNuspecLicense(content, form)
	case strings.HasSuffix(strings.ToLower(name), MSBuildSuffix):
		return alignMSBuildLicense(content, form)
	default:
		return "", false, ErrInvalidNuGetName
	}
}

// alignNuspecLicense aligns the license element of the metadata element: a
// missing element is inserted after the metadata start tag, a self-closing
// metadata element is expanded, the exclusive form switch replaces the
// declared element, and a diverging char-data span is replaced.
func alignNuspecLicense(content string, form NuGetLicenseForm) (string, bool, error) {
	scanned, err := scanNuspec(content)
	if err != nil {
		return "", false, err
	}
	if scanned.metadata == nil {
		return "", false, fmt.Errorf("%w: the document carries no metadata element", ErrInvalidNuspecSurface)
	}
	if scanned.licenseCount > 1 {
		return "", false, ErrMultipleNuspecLicenses
	}
	license := scanned.license
	switch {
	case license == nil:
		if scanned.metadata.selfClosing {
			return applySpanEdits(content, []spanEdit{xmlExpansion(content, scanned.metadata.tagStart, scanned.metadata.tagEnd, "\n        "+form.NuspecLabel())}), true, nil
		}
		return applySpanEdits(content, []spanEdit{{start: scanned.metadata.tagEnd, end: scanned.metadata.tagEnd, text: "\n        " + form.NuspecLabel()}}), true, nil
	case license.licenseType != NuspecExpressionType && license.licenseType != NuspecFileType:
		return "", false, ErrInvalidNuspecLicenseType
	case (license.licenseType == NuspecExpressionType) != form.Expression:
		// The license types are exclusive: the form switch replaces the
		// declared element.
		return applySpanEdits(content, []spanEdit{{start: license.tagStart, end: license.endTagEnd, text: form.NuspecLabel()}}), true, nil
	case license.selfClosing:
		return applySpanEdits(content, []spanEdit{{start: license.tagStart, end: license.tagEnd, text: form.NuspecLabel()}}), true, nil
	case license.value == form.Value:
		return content, false, nil
	default:
		return applySpanEdits(content, []spanEdit{{start: license.dataStart, end: license.dataEnd, text: xmlEscape(form.Value)}}), true, nil
	}
}

// alignMSBuildLicense aligns the license property of the first property
// group: a missing property is inserted into the first property group, a
// project without a property group receives a new one, the exclusive form
// switch replaces the declared property, and a diverging char-data span is
// replaced.
func alignMSBuildLicense(content string, form NuGetLicenseForm) (string, bool, error) {
	scanned, err := scanMSBuild(content)
	if err != nil {
		return "", false, err
	}
	if scanned.projectEnd == 0 {
		return "", false, fmt.Errorf("%w: the document carries no project root", ErrInvalidMSBuildSurface)
	}
	if scanned.licenseCount > 1 {
		return "", false, ErrMultipleMSBuildLicenses
	}
	property := scanned.licenseProperty
	switch {
	case property == nil:
		insertion := "\n        " + form.MSBuildLabel()
		if scanned.firstPropertyGroup != nil {
			if scanned.firstPropertyGroup.selfClosing {
				return applySpanEdits(content, []spanEdit{xmlExpansion(content, scanned.firstPropertyGroup.tagStart, scanned.firstPropertyGroup.tagEnd, insertion)}), true, nil
			}
			return applySpanEdits(content, []spanEdit{{start: scanned.firstPropertyGroup.tagEnd, end: scanned.firstPropertyGroup.tagEnd, text: insertion}}), true, nil
		}
		block := "\n    <PropertyGroup>" + insertion + "\n    </PropertyGroup>"
		if scanned.projectSelfClosing {
			return applySpanEdits(content, []spanEdit{xmlExpansion(content, scanned.projectTagStart, scanned.projectEnd, block)}), true, nil
		}
		return applySpanEdits(content, []spanEdit{{start: scanned.projectEnd, end: scanned.projectEnd, text: block}}), true, nil
	case property.name != form.propertyName():
		// The license properties are exclusive: the form switch replaces
		// the declared property.
		return applySpanEdits(content, []spanEdit{{start: property.tagStart, end: property.endTagEnd, text: form.MSBuildLabel()}}), true, nil
	case property.selfClosing:
		return applySpanEdits(content, []spanEdit{{start: property.tagStart, end: property.tagEnd, text: form.MSBuildLabel()}}), true, nil
	case property.value == form.Value:
		return content, false, nil
	default:
		return applySpanEdits(content, []spanEdit{{start: property.dataStart, end: property.dataEnd, text: xmlEscape(form.Value)}}), true, nil
	}
}

// nugetElement carries the byte spans of one captured skeleton element: the
// full start-tag token span, the span end just past the end tag, and its
// self-closing state.
type nugetElement struct {
	tagStart    int
	tagEnd      int // the index just past the start tag
	endTagEnd   int // the index just past the end tag
	selfClosing bool
}

// nuspecLicense carries the captured license element of the metadata
// element with its type attribute and its char-data value.
type nuspecLicense struct {
	nugetElement
	licenseType string
	value       string
	dataStart   int
	dataEnd     int
}

// msBuildProperty carries one captured license property with its element
// name and its char-data value.
type msBuildProperty struct {
	nugetElement
	name      string
	value     string
	dataStart int
	dataEnd   int
}

type nuspecSurface struct {
	metadata     *nugetElement
	license      *nuspecLicense
	licenseCount int
	licenseURL   *nugetElement
}

type msBuildSurface struct {
	projectTagStart    int
	projectEnd         int
	projectSelfClosing bool
	firstPropertyGroup *nugetElement
	licenseProperty    *msBuildProperty
	licenseCount       int
	licenseURL         *nugetElement
}

// inspectNuspecLicense parses a nuspec manifest and classifies the license
// surface of its metadata element. A manifest that is not well-formed XML
// or whose root element is not package is refused fail-closed.
func inspectNuspecLicense(content string) (NuGetSurface, error) {
	scanned, err := scanNuspec(content)
	if err != nil {
		return NuGetSurface{}, err
	}
	result := NuGetSurface{
		State:                NuGetSurfaceMissing,
		FormCount:            scanned.licenseCount,
		DeprecatedURLPresent: scanned.licenseURL != nil,
	}
	if scanned.license != nil {
		switch scanned.license.licenseType {
		case NuspecExpressionType:
			result.State = NuGetSurfaceDeclared
			result.Expression = true
			result.Value = scanned.license.value
		case NuspecFileType:
			result.State = NuGetSurfaceDeclared
			result.Value = scanned.license.value
		default:
			result.State = NuGetSurfaceInvalid
		}
	}
	return result, nil
}

// inspectMSBuildLicense parses an MSBuild project and classifies its
// license properties. A manifest that is not well-formed XML or whose root
// element is not Project is refused fail-closed.
func inspectMSBuildLicense(content string) (NuGetSurface, error) {
	scanned, err := scanMSBuild(content)
	if err != nil {
		return NuGetSurface{}, err
	}
	result := NuGetSurface{
		State:                NuGetSurfaceMissing,
		FormCount:            scanned.licenseCount,
		DeprecatedURLPresent: scanned.licenseURL != nil,
	}
	if scanned.licenseProperty != nil {
		result.State = NuGetSurfaceDeclared
		result.Expression = scanned.licenseProperty.name == MSBuildExpressionProperty
		result.Value = scanned.licenseProperty.value
	}
	return result, nil
}

// scanNuspec scans the license-relevant skeleton of a nuspec document
// through the standard XML decoder: the decoder validates the
// well-formedness fail-closed, and its documented input-offset contract
// tiles every byte to a token, so every element span is exact. Only the
// package root, its first metadata element, and the license-relevant
// children of the metadata element are captured; every other element,
// comment, or declaration is context.
func scanNuspec(content string) (*nuspecSurface, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	surface := &nuspecSurface{}
	stack := []string{}
	prevEnd := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidNuspecSurface, err)
		}
		end := int(decoder.InputOffset())
		start := prevEnd
		prevEnd = end
		switch element := token.(type) {
		case xml.StartElement:
			name := element.Name.Local
			stack = append(stack, name)
			raw := content[start:end]
			selfClosing := strings.HasSuffix(raw, "/>")
			depth := len(stack)
			switch {
			case depth == 1:
				if name != "package" {
					return nil, fmt.Errorf("%w: the root element is not package", ErrInvalidNuspecSurface)
				}
			case depth == 2 && name == "metadata":
				if surface.metadata != nil {
					return nil, fmt.Errorf("%w: duplicate metadata element in package", ErrInvalidNuspecSurface)
				}
				surface.metadata = &nugetElement{tagStart: start, tagEnd: end, selfClosing: selfClosing}
			case depth == 3 && stack[1] == "metadata" && name == NuspecLicenseElement:
				surface.licenseCount++
				if surface.license == nil {
					licenseType := ""
					for _, attr := range element.Attr {
						if attr.Name.Local == NuspecTypeAttribute {
							licenseType = attr.Value
							break
						}
					}
					endTagEnd := end
					if !selfClosing {
						endTagEnd = 0
					}
					surface.license = &nuspecLicense{
						nugetElement: nugetElement{tagStart: start, tagEnd: end, endTagEnd: endTagEnd, selfClosing: selfClosing},
						licenseType:  licenseType,
						dataStart:    end,
					}
				}
			case depth == 3 && stack[1] == "metadata" && name == NuspecURLElement:
				if surface.licenseURL != nil {
					return nil, fmt.Errorf("%w: duplicate licenseUrl element in metadata", ErrInvalidNuspecSurface)
				}
				surface.licenseURL = &nugetElement{tagStart: start, tagEnd: end, selfClosing: selfClosing}
			}
		case xml.CharData:
			if len(stack) == 3 && stack[1] == "metadata" && surface.licenseCount == 1 && stack[2] == NuspecLicenseElement && surface.license != nil {
				surface.license.value += string(element)
				surface.license.dataEnd = end
			}
		case xml.EndElement:
			if len(stack) == 3 && stack[1] == "metadata" && surface.licenseCount == 1 && stack[2] == NuspecLicenseElement && surface.license != nil && surface.license.endTagEnd == 0 {
				surface.license.endTagEnd = end
			}
			stack = stack[:len(stack)-1]
		}
	}
	return surface, nil
}

// scanMSBuild scans the license-relevant skeleton of an MSBuild project
// through the standard XML decoder: the decoder validates the
// well-formedness fail-closed, and its documented input-offset contract
// tiles every byte to a token, so every element span is exact. Only the
// Project root, its first property group, and the license-relevant
// properties are captured; every other element, comment, or declaration is
// context.
func scanMSBuild(content string) (*msBuildSurface, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	surface := &msBuildSurface{}
	stack := []string{}
	prevEnd := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidMSBuildSurface, err)
		}
		end := int(decoder.InputOffset())
		start := prevEnd
		prevEnd = end
		switch element := token.(type) {
		case xml.StartElement:
			name := element.Name.Local
			stack = append(stack, name)
			raw := content[start:end]
			selfClosing := strings.HasSuffix(raw, "/>")
			depth := len(stack)
			switch {
			case depth == 1:
				if name != "Project" {
					return nil, fmt.Errorf("%w: the root element is not Project", ErrInvalidMSBuildSurface)
				}
				surface.projectTagStart = start
				surface.projectEnd = end
				surface.projectSelfClosing = selfClosing
			case depth == 2 && name == "PropertyGroup":
				if surface.firstPropertyGroup == nil {
					surface.firstPropertyGroup = &nugetElement{tagStart: start, tagEnd: end, selfClosing: selfClosing}
				}
			case depth == 3 && stack[1] == "PropertyGroup" && (name == MSBuildExpressionProperty || name == MSBuildFileProperty):
				surface.licenseCount++
				if surface.licenseProperty == nil {
					endTagEnd := end
					if !selfClosing {
						endTagEnd = 0
					}
					surface.licenseProperty = &msBuildProperty{
						nugetElement: nugetElement{tagStart: start, tagEnd: end, endTagEnd: endTagEnd, selfClosing: selfClosing},
						name:         name,
						dataStart:    end,
					}
				}
			case depth == 3 && stack[1] == "PropertyGroup" && name == MSBuildURLProperty:
				if surface.licenseURL != nil {
					return nil, fmt.Errorf("%w: duplicate %s property", ErrInvalidMSBuildSurface, MSBuildURLProperty)
				}
				surface.licenseURL = &nugetElement{tagStart: start, tagEnd: end, selfClosing: selfClosing}
			}
		case xml.CharData:
			if len(stack) == 3 && stack[1] == "PropertyGroup" && surface.licenseCount == 1 && surface.licenseProperty != nil && stack[2] == surface.licenseProperty.name {
				surface.licenseProperty.value += string(element)
				surface.licenseProperty.dataEnd = end
			}
		case xml.EndElement:
			if len(stack) == 3 && stack[1] == "PropertyGroup" && surface.licenseCount == 1 && surface.licenseProperty != nil && stack[2] == surface.licenseProperty.name && surface.licenseProperty.endTagEnd == 0 {
				surface.licenseProperty.endTagEnd = end
			}
			stack = stack[:len(stack)-1]
		}
	}
	return surface, nil
}
