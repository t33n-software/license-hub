// Package values merges organization defaults with tenant project values and
// validates the canonical substitution anchors.
package values

import (
	"encoding/json"
	"fmt"
	"sort"
)

// RequiredKeys returns the eight canonical substitution anchors of the
// license template contract.
func RequiredKeys() []string {
	return []string{
		"PROJECT_NAME",
		"LICENSE_ID",
		"COPYRIGHT_YEAR",
		"COPYRIGHT_HOLDER",
		"CANONICAL_SOURCE_URL",
		"PERMISSION_CONTACT",
		"GOVERNING_LAW",
		"VENUE",
	}
}

// Parse decodes a JSON object with string values.
func Parse(data []byte) (map[string]string, error) {
	var out map[string]string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse values JSON: %w", err)
	}
	return out, nil
}

// Merge overlays project values on organization defaults; the project value
// wins on conflict.
func Merge(orgDefaults, project map[string]string) map[string]string {
	out := make(map[string]string, len(orgDefaults)+len(project))
	for key, value := range orgDefaults {
		out[key] = value
	}
	for key, value := range project {
		out[key] = value
	}
	return out
}

// MissingRequired returns the sorted required anchors absent from values.
func MissingRequired(values map[string]string) []string {
	missing := make([]string, 0)
	for _, key := range RequiredKeys() {
		if _, ok := values[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}
