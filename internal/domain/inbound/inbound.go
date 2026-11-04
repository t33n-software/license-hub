// Package inbound models the inbound license declarations of a tenant: the
// listed standard licenses a derived project carries alongside its own
// license family, each with the upstream copyright notice it must retain.
package inbound

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/t33n-software/license-hub/internal/domain/placeholder"
)

// Key is the tenant values key carrying the inbound license declarations.
const Key = "INBOUND_LICENSES"

// NoticesFile is the canonical third-party notices document.
const NoticesFile = "THIRD_PARTY_NOTICES.md"

// Entry is one declared inbound license: the listed SPDX identifier, the
// canonical template path of the license text, and the upstream copyright
// notice the derived project retains.
type Entry struct {
	ID              string `json:"id"`
	Template        string `json:"template"`
	CopyrightYear   string `json:"copyright_year"`
	CopyrightHolder string `json:"copyright_holder"`
}

var (
	// idPattern matches the canonical form of a listed SPDX license
	// identifier; the custom LicenseRef- form is refused separately.
	idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*$`)
	// yearPattern matches a single year or a year range.
	yearPattern = regexp.MustCompile(`^[0-9]{4}(-[0-9]{1,4})?$`)
)

// Parse decodes and validates an INBOUND_LICENSES array. A nil or blank
// document means no declaration. Every malformed shape is refused
// fail-closed: an empty array, trailing data, an unknown entry field, a
// custom identifier, a duplicate identifier, and every missing or malformed
// entry field. The result is sorted by identifier so every rendered
// document is deterministic.
func Parse(data []byte) ([]Entry, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var entries []Entry
	if err := decoder.Decode(&entries); err != nil {
		return nil, fmt.Errorf("parse inbound licenses: %w", err)
	}
	if decoder.More() {
		return nil, fmt.Errorf("parse inbound licenses: trailing data after the array")
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("parse inbound licenses: the array is empty")
	}
	seen := make(map[string]bool, len(entries))
	for i := range entries {
		if err := normalize(&entries[i]); err != nil {
			return nil, err
		}
		if seen[entries[i].ID] {
			return nil, fmt.Errorf("inbound license %q is declared more than once", entries[i].ID)
		}
		seen[entries[i].ID] = true
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries, nil
}

// normalize trims every field of one entry and validates its canonical form.
func normalize(entry *Entry) error {
	entry.ID = strings.TrimSpace(entry.ID)
	entry.Template = strings.TrimSpace(entry.Template)
	entry.CopyrightYear = strings.TrimSpace(entry.CopyrightYear)
	entry.CopyrightHolder = strings.TrimSpace(entry.CopyrightHolder)
	switch {
	case entry.ID == "":
		return fmt.Errorf("inbound license entry is missing id")
	case !idPattern.MatchString(entry.ID):
		return fmt.Errorf("inbound license id %q is not a listed SPDX identifier form", entry.ID)
	case strings.HasPrefix(entry.ID, "LicenseRef-"):
		return fmt.Errorf("inbound license id %q is a custom identifier; inbound declarations list standard licenses only", entry.ID)
	case entry.Template == "":
		return fmt.Errorf("inbound license %s is missing template", entry.ID)
	case !strings.HasSuffix(entry.Template, ".hbs"):
		return fmt.Errorf("inbound license %s references a non-canonical template: %s", entry.ID, entry.Template)
	case !yearPattern.MatchString(entry.CopyrightYear):
		return fmt.Errorf("inbound license %s carries an invalid copyright_year: %q", entry.ID, entry.CopyrightYear)
	case entry.CopyrightHolder == "" || strings.ContainsAny(entry.CopyrightHolder, "\n\r"):
		return fmt.Errorf("inbound license %s carries an invalid copyright_holder", entry.ID)
	}
	return nil
}

// LicensePath resolves the REUSE text path of one entry below dir.
func LicensePath(dir string, entry Entry) string {
	return filepath.Join(dir, "LICENSES", entry.ID+".txt")
}

// NoticesPath resolves the canonical notices document below dir.
func NoticesPath(dir string) string {
	return filepath.Join(dir, NoticesFile)
}

// CopyrightLine renders the retained upstream copyright line of one entry.
func CopyrightLine(entry Entry) string {
	return "Copyright (c) " + entry.CopyrightYear + " " + entry.CopyrightHolder
}

// Text renders the canonical license text of one entry from its template.
// Only the two copyright anchors are substituted; every remaining
// placeholder is a fail-closed refusal.
func Text(template string, entry Entry) (string, error) {
	out := strings.ReplaceAll(template, "{{COPYRIGHT_YEAR}}", entry.CopyrightYear)
	out = strings.ReplaceAll(out, "{{COPYRIGHT_HOLDER}}", entry.CopyrightHolder)
	if unresolved := placeholder.Unresolved(out); len(unresolved) > 0 {
		return "", fmt.Errorf("inbound template of %s carries unresolved anchors: %s", entry.ID, strings.Join(unresolved, ", "))
	}
	return out, nil
}

// NoticesDocument renders the canonical third-party notices document of
// the sorted entries.
func NoticesDocument(entries []Entry) string {
	var builder strings.Builder
	builder.WriteString("# Third-Party Notices\n\n")
	builder.WriteString("This project retains the inbound licenses listed below. Every license text is\n" +
		"committed under `LICENSES/`; the copyright lines are the notices retained for\n" +
		"the corresponding upstream works.\n")
	for _, entry := range entries {
		builder.WriteString("\n## " + entry.ID + "\n\n")
		builder.WriteString(CopyrightLine(entry) + "\n\n")
		builder.WriteString("License text: `LICENSES/" + entry.ID + ".txt`\n")
	}
	return builder.String()
}
