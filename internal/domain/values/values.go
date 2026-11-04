// Package values merges organization defaults with tenant project values and
// validates the canonical substitution anchors.
package values

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/t33n-software/license-hub/internal/domain/inbound"
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

// Parse decodes a JSON object with string values. The INBOUND_LICENSES key
// carries a structured array; it is kept out of the flat substitution map
// and proven by the inbound domain instead.
func Parse(data []byte) (map[string]string, error) {
	raw, err := rawObject(data)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(raw))
	for key, value := range raw {
		if key == inbound.Key {
			continue
		}
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			return nil, fmt.Errorf("parse values JSON: key %q is not a string value", key)
		}
		out[key] = text
	}
	return out, nil
}

// InboundRaw returns the raw INBOUND_LICENSES document of a values object,
// or nil when the key is absent.
func InboundRaw(data []byte) ([]byte, error) {
	raw, err := rawObject(data)
	if err != nil {
		return nil, err
	}
	return raw[inbound.Key], nil
}

// rawObject decodes a values document into its raw member map.
func rawObject(data []byte) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse values JSON: %w", err)
	}
	return raw, nil
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
