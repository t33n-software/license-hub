// The shared XML law primitives of the ecosystem adapters: the char-data
// escape form and the XML character-class guard. Both XML manifest adapters
// (Maven and .NET) project values into XML char data, so the law lives once
// under neutral naming.
//
// Convention: spec/ecosystem-license-metadata.md

package ecosystem

import (
	"encoding/xml"
	"strings"
)

// xmlEscape renders the XML-escaped char-data form of a target value. The
// strings.Builder writer cannot fail, so the escape error is provably nil.
func xmlEscape(value string) string {
	var builder strings.Builder
	_ = xml.EscapeText(&builder, []byte(value))
	return builder.String()
}

// xmlRepresentable reports whether every rune of the value is representable
// in XML char data; a value outside the XML character classes cannot be
// projected faithfully and is refused fail-closed.
func xmlRepresentable(value string) bool {
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

// xmlExpansion builds the span edit that expands a self-closing start tag in
// place and inserts the block as its content; the reconstructed end tag
// carries only the element name, never the attributes of the start tag.
func xmlExpansion(content string, tagStart, tagEnd int, block string) spanEdit {
	open := strings.TrimSuffix(content[tagStart:tagEnd], "/>")
	name := open[1:]
	if cut := strings.IndexAny(name, " \t\n\r"); cut >= 0 {
		name = name[:cut]
	}
	return spanEdit{start: tagStart, end: tagEnd, text: open + ">" + block + "</" + name + ">"}
}
