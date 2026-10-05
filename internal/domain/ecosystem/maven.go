// The Maven ecosystem adapter: it derives and aligns the license metadata
// surface of a pom.xml manifest from the declaration chain and the license
// lock. The declaration is the truth, the manifest field is its
// deterministic projection, and the license lock is the single source of
// truth of the license type. The pom surface is the licenses element of the
// project root: every license element carries the name of the license and
// the URL of its text; an SPDX identifier as the name is the recommended
// form, and child poms inherit the declared licenses.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// MavenLanguage is the seam toolchain language that selects the JVM/Maven
// matrix row.
const MavenLanguage = "maven"

// The Maven license metadata surface: the manifest file and the elements of
// the project-level licenses block.
const (
	MavenManifestName = "pom.xml"
	MavenNameElement  = "name"
	MavenURLElement   = "url"
)

// ErrInvalidMavenSurface marks a manifest that is not a well-formed pom.xml
// surface.
var ErrInvalidMavenSurface = errors.New("invalid pom.xml surface")

// ErrAmbiguousMavenField marks a surface element that the scanner cannot
// attribute deterministically: a repeated licenses element or a repeated
// name or url element inside one license element.
var ErrAmbiguousMavenField = errors.New("ambiguous pom.xml license element")

// ErrMultipleMavenLicenses marks a licenses element that declares more than
// one license element. The projection carries exactly one license family;
// the resolution is an explicit tenant decision.
var ErrMultipleMavenLicenses = errors.New("pom.xml declares multiple license elements")

// MavenLicenseForm carries the projected pom.xml license surface: the name
// of the license family and the URL of its canonical text.
type MavenLicenseForm struct {
	// Name carries the projected license name: a declared SPDX identifier
	// projects itself; the file-based custom family projects the LICENSE_ID
	// form.
	Name string
	// URL carries the canonical source URL of the tenant values.
	URL string
}

// MavenProjection derives the expected pom.xml license surface from the
// merged tenant values. The derivation never reads the manifest.
func MavenProjection(merged map[string]string) MavenLicenseForm {
	name := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"])
	if name == "" {
		name = merged["LICENSE_ID"]
	}
	return MavenLicenseForm{Name: name, URL: merged["CANONICAL_SOURCE_URL"]}
}

// MavenLicenseSurface is the license-relevant view of a pom.xml.
type MavenLicenseSurface struct {
	// LicensesPresent reports whether the project declares a licenses
	// element.
	LicensesPresent bool
	// LicenseCount carries the number of declared license elements.
	LicenseCount int
	// NamePresent reports whether the first license element declares a name.
	NamePresent bool
	// Name carries the decoded name text when present.
	Name string
	// URLPresent reports whether the first license element declares a url.
	URLPresent bool
	// URL carries the decoded url text when present.
	URL string
}

// InspectMavenLicense parses a pom.xml manifest and classifies the license
// surface of its project root. A manifest that is not well-formed XML is
// refused fail-closed.
func InspectMavenLicense(content string) (MavenLicenseSurface, error) {
	surface, err := scanMavenPOM(content)
	if err != nil {
		return MavenLicenseSurface{}, err
	}
	result := MavenLicenseSurface{LicenseCount: surface.licenseCount}
	result.LicensesPresent = surface.licenses != nil
	if license := surface.license; license != nil {
		result.NamePresent = license.name != nil
		if license.name != nil {
			result.Name = license.name.text
		}
		result.URLPresent = license.url != nil
		if license.url != nil {
			result.URL = license.url.text
		}
	}
	return result, nil
}

// AlignMavenLicense returns the manifest content with the license surface of
// the project root aligned to the projected license form. Every byte
// outside the replaced char-data spans, the expanded self-closing elements,
// or the inserted blocks is preserved exactly; a re-run over an aligned
// manifest is a no-op. A manifest that declares multiple license elements
// is refused fail-closed.
func AlignMavenLicense(content string, form MavenLicenseForm) (string, bool, error) {
	if !utf8.ValidString(form.Name) || !utf8.ValidString(form.URL) ||
		!mavenRepresentable(form.Name) || !mavenRepresentable(form.URL) {
		return "", false, ErrInvalidTarget
	}
	surface, err := scanMavenPOM(content)
	if err != nil {
		return "", false, err
	}
	if surface.projectEnd == 0 {
		return "", false, fmt.Errorf("%w: the document carries no project root", ErrInvalidMavenSurface)
	}
	if surface.licenses == nil {
		block := "\n    <licenses>\n        <license>\n            <name>" + mavenEscape(form.Name) +
			"</name>\n            <url>" + mavenEscape(form.URL) + "</url>\n        </license>\n    </licenses>"
		if surface.projectSelfClosing {
			open := strings.TrimSuffix(content[surface.projectTagStart:surface.projectEnd], "/>")
			replacement := open + ">" + block + "</" + open[len("<"):] + ">"
			return applySpanEdits(content, []spanEdit{{start: surface.projectTagStart, end: surface.projectEnd, text: replacement}}), true, nil
		}
		return applySpanEdits(content, []spanEdit{{start: surface.projectEnd, end: surface.projectEnd, text: block}}), true, nil
	}
	if surface.licenseCount > 1 {
		return "", false, ErrMultipleMavenLicenses
	}
	license := surface.license
	edits := []spanEdit{}
	switch {
	case license == nil:
		block := "\n        <license><name>" + mavenEscape(form.Name) + "</name><url>" +
			mavenEscape(form.URL) + "</url></license>"
		edits = append(edits, mavenInsertion(content, surface.licenses, block))
	case license.name == nil && license.url == nil:
		block := "<name>" + mavenEscape(form.Name) + "</name><url>" + mavenEscape(form.URL) + "</url>"
		edits = append(edits, mavenInsertion(content, &license.mavenElement, block))
	default:
		edits = append(edits, mavenValueEdit(license.name, license.tagEnd, MavenNameElement, form.Name)...)
		edits = append(edits, mavenValueEdit(license.url, license.tagEnd, MavenURLElement, form.URL)...)
	}
	if len(edits) == 0 {
		return content, false, nil
	}
	return applySpanEdits(content, edits), true, nil
}

// mavenInsertion builds the insertion of a block into an element: a
// self-closing element is expanded in place, an open element receives the
// block directly after its start tag.
func mavenInsertion(content string, element *mavenElement, block string) spanEdit {
	if element.selfClosing {
		open := strings.TrimSuffix(content[element.tagStart:element.tagEnd], "/>")
		return spanEdit{start: element.tagStart, end: element.tagEnd, text: open + ">" + block + "</" + open[len("<"):] + ">"}
	}
	return spanEdit{start: element.tagEnd, end: element.tagEnd, text: block}
}

// mavenValueEdit builds the value alignment of one name or url element: a
// missing element is inserted after the license start tag, a self-closing
// element is expanded, a diverging char-data span is replaced.
func mavenValueEdit(value *mavenValue, insertAt int, elementName, target string) []spanEdit {
	expanded := "<" + elementName + ">" + mavenEscape(target) + "</" + elementName + ">"
	switch {
	case value == nil:
		return []spanEdit{{start: insertAt, end: insertAt, text: expanded}}
	case value.selfClosing:
		return []spanEdit{{start: value.tagStart, end: value.tagEnd, text: expanded}}
	case value.text == target:
		return nil
	default:
		return []spanEdit{{start: value.dataStart, end: value.dataEnd, text: mavenEscape(target)}}
	}
}

// mavenEscape renders the XML-escaped char-data form of a target value. The
// strings.Builder writer cannot fail, so the escape error is provably nil.
func mavenEscape(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}

// mavenRepresentable reports whether every rune of the value is
// representable in XML char data; a value outside the XML character
// classes cannot be projected faithfully and is refused fail-closed.
func mavenRepresentable(value string) bool {
	for _, r := range value {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
		case r >= 0x20 && r <= 0xD7FF:
		case r >= 0xE000 && r <= 0xFFFD:
		case r >= 0x10000 && r <= 0x10FFFF:
		default:
			return false
		}
	}
	return true
}

// mavenElement carries the byte spans of one skeleton element: the full
// start-tag token span and its self-closing state.
type mavenElement struct {
	tagStart    int
	tagEnd      int // the index just past the start tag
	selfClosing bool
}

// mavenValue carries the byte spans of one name or url element: the full
// start-tag token span plus the char-data span that the alignment replaces.
type mavenValue struct {
	tagStart    int
	tagEnd      int
	dataStart   int
	dataEnd     int
	text        string
	selfClosing bool
}

type mavenLicense struct {
	mavenElement
	name *mavenValue
	url  *mavenValue
}

type mavenSurface struct {
	projectTagStart    int
	projectEnd         int
	projectSelfClosing bool
	licenses           *mavenElement
	license            *mavenLicense
	licenseCount       int
}

// scanMavenPOM scans the license-relevant skeleton of a pom.xml document
// through the standard XML decoder: the decoder validates the
// well-formedness fail-closed, and its documented input-offset contract
// tiles every byte to a token, so every element span is exact. Only the
// project root, its first licenses element, and the first license element
// with its name and url children are captured; every other element,
// comment, or declaration is context.
func scanMavenPOM(content string) (*mavenSurface, error) {
	decoder := xml.NewDecoder(strings.NewReader(content))
	surface := &mavenSurface{}
	stack := []string{}
	prevEnd := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidMavenSurface, err)
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
				if name != "project" {
					return nil, fmt.Errorf("%w: the root element is not project", ErrInvalidMavenSurface)
				}
				surface.projectTagStart = start
				surface.projectEnd = end
				surface.projectSelfClosing = selfClosing
			case depth == 2 && name == "licenses":
				if surface.licenses != nil {
					return nil, fmt.Errorf("%w: duplicate licenses element in project", ErrAmbiguousMavenField)
				}
				surface.licenses = &mavenElement{tagStart: start, tagEnd: end, selfClosing: selfClosing}
			case depth == 3 && name == "license" && stack[1] == "licenses":
				surface.licenseCount++
				surface.license = &mavenLicense{mavenElement: mavenElement{tagStart: start, tagEnd: end, selfClosing: selfClosing}}
			case depth == 4 && stack[1] == "licenses" && stack[2] == "license" && name == MavenNameElement:
				if surface.license.name != nil {
					return nil, fmt.Errorf("%w: duplicate name element in license", ErrAmbiguousMavenField)
				}
				surface.license.name = &mavenValue{tagStart: start, tagEnd: end, dataStart: end, selfClosing: selfClosing}
			case depth == 4 && stack[1] == "licenses" && stack[2] == "license" && name == MavenURLElement:
				if surface.license.url != nil {
					return nil, fmt.Errorf("%w: duplicate url element in license", ErrAmbiguousMavenField)
				}
				surface.license.url = &mavenValue{tagStart: start, tagEnd: end, dataStart: end, selfClosing: selfClosing}
			}
		case xml.CharData:
			if len(stack) == 4 && stack[1] == "licenses" && stack[2] == "license" && surface.license != nil {
				decoded := string(element)
				switch stack[3] {
				case MavenNameElement:
					if value := surface.license.name; value != nil {
						value.text += decoded
						value.dataEnd = end
					}
				case MavenURLElement:
					if value := surface.license.url; value != nil {
						value.text += decoded
						value.dataEnd = end
					}
				}
			}
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		}
	}
	return surface, nil
}
