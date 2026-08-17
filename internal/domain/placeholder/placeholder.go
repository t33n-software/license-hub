// Package placeholder detects unresolved template placeholders.
package placeholder

import "regexp"

// pattern matches the canonical Handlebars-style anchors {{KEY}} with an
// uppercase snake-case key.
var pattern = regexp.MustCompile(`\{\{[A-Z0-9_]+\}\}`)

// Unresolved returns the distinct unresolved placeholders of content in
// first-occurrence order.
func Unresolved(content string) []string {
	matches := pattern.FindAllString(content, -1)
	seen := make(map[string]bool, len(matches))
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		if seen[match] {
			continue
		}
		seen[match] = true
		out = append(out, match)
	}
	return out
}
