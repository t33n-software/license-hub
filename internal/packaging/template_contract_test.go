package packaging

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/t33n-software/license-hub/internal/domain/placeholder"
	"github.com/t33n-software/license-hub/internal/domain/render"
)

// templateFixtureValues carries every contract anchor so the render gate can
// prove that no template in the inventory leaves an unresolved placeholder.
var templateFixtureValues = map[string]string{
	"PROJECT_NAME":            "example-project",
	"LICENSE_ID":              "example-project-Test-1.0",
	"COPYRIGHT_YEAR":          "2026",
	"CANONICAL_SOURCE_URL":    "https://github.com/t33n-software/example-project",
	"COPYRIGHT_HOLDER":        "CyberT33N",
	"GOVERNING_LAW":           "the Federal Republic of Germany",
	"VENUE":                   "Germany",
	"PERMISSION_CONTACT":      "https://github.com/t33n-software",
	"SPDX_LICENSE_IDENTIFIER": "MIT",
}

var knownTemplateFamilies = map[string]bool{
	"permissive":               true,
	"weak-copyleft":            true,
	"strong-copyleft":          true,
	"network-copyleft":         true,
	"source-available":         true,
	"proprietary":              true,
	"custom":                   true,
	"public-domain-dedication": true,
	"non-software":             true,
	"multi-licensing":          true,
}

var (
	templateNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.-]*-(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.hbs$`)
	docsLinkPattern     = regexp.MustCompile(`\]\(([^)]*docs/licensing/[^)]+\.md)\)`)
	registryRowPattern  = regexp.MustCompile("\x60(templates/[^\x60]+\\.hbs)\x60")
)

// templateInventory walks the templates tree and returns every template file
// as a slash-separated repository-relative path.
func templateInventory(t *testing.T) []string {
	t.Helper()
	root := repositoryPath("templates")
	templates := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hbs") {
			return nil
		}
		rel, err := filepath.Rel(repositoryPath(), path)
		if err != nil {
			t.Fatalf("Rel(%q) error = %v", path, err)
		}
		templates = append(templates, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir(templates) error = %v", err)
	}
	if len(templates) == 0 {
		t.Fatal("the templates inventory is empty")
	}
	return templates
}

func TestTemplatesRenderWithoutUnresolvedPlaceholders(t *testing.T) {
	for _, template := range templateInventory(t) {
		content := readRepositoryFile(t, template)
		rendered := render.Execute(content, templateFixtureValues)
		if unresolved := placeholder.Unresolved(rendered); len(unresolved) > 0 {
			t.Fatalf("template %s leaves unresolved placeholders: %s", template, strings.Join(unresolved, ", "))
		}
	}
}

func TestTemplateDirectoriesCarryChangelogAndDocsReference(t *testing.T) {
	for _, template := range templateInventory(t) {
		dir := filepath.Dir(template)
		name := filepath.Base(template)

		if !templateNamePattern.MatchString(name) {
			t.Fatalf("template %s does not match the <Name>-<semver>.hbs naming contract", template)
		}

		family := filepath.Base(filepath.Dir(dir))
		if !knownTemplateFamilies[family] {
			t.Fatalf("template %s lives in the unknown family %q", template, family)
		}

		if _, err := os.Stat(repositoryPath(filepath.Join(dir, "CHANGELOG.md"))); err != nil {
			t.Fatalf("template directory %s misses CHANGELOG.md: %v", dir, err)
		}

		readmePath := repositoryPath(filepath.Join(dir, "README.md"))
		readme, err := os.ReadFile(readmePath)
		if err != nil {
			t.Fatalf("template directory %s misses the sibling README.md documentation seam: %v", dir, err)
		}
		links := docsLinkPattern.FindStringSubmatch(string(readme))
		if links == nil {
			t.Fatalf("the sibling README of %s does not reference a docs/licensing document", dir)
		}
		target := filepath.Clean(repositoryPath(filepath.Join(dir, filepath.FromSlash(links[1]))))
		if _, err := os.Stat(target); err != nil {
			t.Fatalf("the sibling README of %s references %s, which does not exist: %v", dir, links[1], err)
		}
	}
}

func TestTemplateRegistryCoversEveryTemplate(t *testing.T) {
	registry := readRepositoryFile(t, filepath.Join("templates", "README.md"))
	for _, template := range templateInventory(t) {
		if !strings.Contains(registry, template) {
			t.Fatalf("templates/README.md does not list the active template %s", template)
		}
	}
	for _, match := range registryRowPattern.FindAllStringSubmatch(registry, -1) {
		if _, err := os.Stat(repositoryPath(filepath.FromSlash(match[1]))); err != nil {
			t.Fatalf("templates/README.md lists %s, which does not exist: %v", match[1], err)
		}
	}
}
