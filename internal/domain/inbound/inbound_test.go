package inbound

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const validEntry = `{"id":"MIT","template":"templates/permissive/mit/MIT-1.0.0.hbs","copyright_year":"2017","copyright_holder":"Example Author"}`

func TestParseNoDeclaration(t *testing.T) {
	for _, data := range [][]byte{nil, []byte(""), []byte("   \n")} {
		entries, err := Parse(data)
		if err != nil || entries != nil {
			t.Fatalf("Parse(%q) = %v, %v", data, entries, err)
		}
	}
}

func TestParseValidSorted(t *testing.T) {
	input := `[` + validEntry + `,{"id":"Apache-2.0","template":"templates/permissive/apache-2.0/Apache-2.0-1.0.0.hbs","copyright_year":"2004","copyright_holder":"Apache Foundation"}]`
	entries, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(entries) != 2 || entries[0].ID != "Apache-2.0" || entries[1].ID != "MIT" {
		t.Fatalf("Parse() = %v", entries)
	}
}

func TestParseAcceptsYearRange(t *testing.T) {
	input := `[{"id":"MIT","template":"m.hbs","copyright_year":"2020-22","copyright_holder":"M"}]`
	entries, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if entries[0].CopyrightYear != "2020-22" {
		t.Fatalf("Parse() = %v", entries)
	}
}

func TestParseTrimsFields(t *testing.T) {
	input := `[{"id":" MIT ","template":" mit.hbs ","copyright_year":"2017","copyright_holder":"  Example Author  "}]`
	entries, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	entry := entries[0]
	if entry.ID != "MIT" || entry.Template != "mit.hbs" || entry.CopyrightHolder != "Example Author" {
		t.Fatalf("Parse() = %+v", entry)
	}
}

func TestParseRejects(t *testing.T) {
	cases := map[string]string{
		"invalid json":        `{`,
		"non-array":           `"MIT"`,
		"empty array":         `[]`,
		"trailing data":       `[` + validEntry + `] {}`,
		"unknown field":       `[{"id":"MIT","template":"m.hbs","copyright_year":"2017","copyright_holder":"X","extra":1}]`,
		"missing id":          `[{"template":"m.hbs","copyright_year":"2017","copyright_holder":"X"}]`,
		"bad id":              `[{"id":"MIT!!","template":"m.hbs","copyright_year":"2017","copyright_holder":"X"}]`,
		"license ref id":      `[{"id":"LicenseRef-Custom","template":"m.hbs","copyright_year":"2017","copyright_holder":"X"}]`,
		"missing template":    `[{"id":"MIT","copyright_year":"2017","copyright_holder":"X"}]`,
		"non hbs template":    `[{"id":"MIT","template":"m.txt","copyright_year":"2017","copyright_holder":"X"}]`,
		"bad year":            `[{"id":"MIT","template":"m.hbs","copyright_year":"17","copyright_holder":"X"}]`,
		"year range too long": `[{"id":"MIT","template":"m.hbs","copyright_year":"2017-12345","copyright_holder":"X"}]`,
		"missing holder":      `[{"id":"MIT","template":"m.hbs","copyright_year":"2017"}]`,
		"multiline holder":    `[{"id":"MIT","template":"m.hbs","copyright_year":"2017","copyright_holder":"A\nB"}]`,
		"duplicate id":        `[` + validEntry + `,` + validEntry + `]`,
		"non-object entry":    `["MIT"]`,
		"typed field":         `[{"id":"MIT","template":"m.hbs","copyright_year":2017,"copyright_holder":"X"}]`,
	}
	for name, input := range cases {
		if entries, err := Parse([]byte(input)); err == nil {
			t.Fatalf("%s: expected error, got %v", name, entries)
		}
	}
}

func TestTextSubstitutesCopyrightAnchors(t *testing.T) {
	entry := Entry{ID: "MIT", CopyrightYear: "2017", CopyrightHolder: "D"}
	text, err := Text("Copyright (c) {{COPYRIGHT_YEAR}} {{COPYRIGHT_HOLDER}}\n", entry)
	if err != nil {
		t.Fatalf("Text() error = %v", err)
	}
	if text != "Copyright (c) 2017 D\n" {
		t.Fatalf("Text() = %q", text)
	}
}

func TestTextRejectsUnresolvedAnchors(t *testing.T) {
	entry := Entry{ID: "MIT", CopyrightYear: "2017", CopyrightHolder: "D"}
	if _, err := Text("Copyright {{COPYRIGHT_YEAR}} {{VENUE}}", entry); err == nil {
		t.Fatal("Text() expected error for unresolved anchors")
	}
}

func TestPathsAndCopyrightLine(t *testing.T) {
	entry := Entry{ID: "MIT", CopyrightYear: "2017", CopyrightHolder: "D"}
	if got := LicensePath("/t", entry); got != filepath.Join("/t", "LICENSES", "MIT.txt") {
		t.Fatalf("LicensePath() = %q", got)
	}
	if got := NoticesPath("/t"); got != filepath.Join("/t", NoticesFile) {
		t.Fatalf("NoticesPath() = %q", got)
	}
	if got := CopyrightLine(entry); got != "Copyright (c) 2017 D" {
		t.Fatalf("CopyrightLine() = %q", got)
	}
}

func TestNoticesDocumentGolden(t *testing.T) {
	entries := []Entry{
		{ID: "Apache-2.0", CopyrightYear: "2004", CopyrightHolder: "A"},
		{ID: "MIT", CopyrightYear: "2017", CopyrightHolder: "M"},
	}
	want := "# Third-Party Notices\n\n" +
		"This project retains the inbound licenses listed below. Every license text is\n" +
		"committed under `LICENSES/`; the copyright lines are the notices retained for\n" +
		"the corresponding upstream works.\n" +
		"\n## Apache-2.0\n\nCopyright (c) 2004 A\n\nLicense text: `LICENSES/Apache-2.0.txt`\n" +
		"\n## MIT\n\nCopyright (c) 2017 M\n\nLicense text: `LICENSES/MIT.txt`\n"
	if got := NoticesDocument(entries); got != want {
		t.Fatalf("NoticesDocument() =\n%q\nwant\n%q", got, want)
	}
}

func FuzzParse(f *testing.F) {
	f.Add(`[` + validEntry + `]`)
	f.Add(`[]`)
	f.Add(`{`)
	f.Add(`"MIT"`)
	f.Add(`[{"id":"LicenseRef-X","template":"m.hbs","copyright_year":"2017","copyright_holder":"X"}]`)
	f.Fuzz(func(t *testing.T, input string) {
		entries, err := Parse([]byte(input))
		if err != nil {
			if entries != nil {
				t.Fatal("error with entries")
			}
			return
		}
		if entries == nil {
			return
		}
		if !slices.IsSortedFunc(entries, func(a, b Entry) int { return strings.Compare(a.ID, b.ID) }) {
			t.Fatal("accepted unsorted result")
		}
	})
}
