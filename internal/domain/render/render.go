// Package render executes placeholder substitution for canonical license templates.
package render

import "strings"

// Execute replaces every {{KEY}} occurrence in template with the matching value
// from values. Placeholders without a matching key stay untouched so the
// placeholder gate can detect them afterward.
func Execute(template string, values map[string]string) string {
	out := template
	for key, value := range values {
		out = strings.ReplaceAll(out, "{{"+key+"}}", value)
	}
	return out
}
