package packaging

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bindingManifest mirrors the tenant binding manifest (repo-bindings/v1) for
// the self-consistency proofs of the canonical adoption. The home-side proof
// against the canonical masters is owned by the verify-canonical tool; these
// tests bind the tenant files to the manifest.
type bindingManifest struct {
	Home struct {
		Repository string `json:"repository"`
		SHA        string `json:"sha"`
	} `json:"home"`
	Callers []struct {
		File   string `json:"file"`
		Master string `json:"master"`
		SHA256 string `json:"sha256"`
	} `json:"callers"`
	Files struct {
		Lefthook      fileBinding `json:"lefthook"`
		Gitattributes fileBinding `json:"gitattributes"`
		Gitignore     fileBinding `json:"gitignore"`
		Dependabot    fileBinding `json:"dependabot"`
	} `json:"files"`
	Codeowners struct {
		Path         string `json:"path"`
		DefaultOwner string `json:"defaultOwner"`
	} `json:"codeowners"`
}

type fileBinding struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func readBindingManifest(t *testing.T) bindingManifest {
	t.Helper()
	var manifest bindingManifest
	if err := json.Unmarshal([]byte(readRepositoryFile(t, "repo-bindings.json")), &manifest); err != nil {
		t.Fatalf("repo-bindings.json is not valid JSON: %v", err)
	}
	if manifest.Home.Repository != "t33n-software/repository-governance" {
		t.Fatalf("the manifest binds home %q", manifest.Home.Repository)
	}
	return manifest
}

// hashRepositoryFile hashes the LF-normalized repository file; the canonical
// .gitattributes makes the checkout LF, and the normalization keeps the
// derivation tolerant as the second line of defense.
func hashRepositoryFile(t *testing.T, path string) string {
	t.Helper()
	normalized := strings.ReplaceAll(readRepositoryFile(t, path), "\r\n", "\n")
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

func TestCanonicalCallersMatchTheBindingManifest(t *testing.T) {
	manifest := readBindingManifest(t)
	want := map[string]string{
		".github/workflows/ci.yml":                "hosting-platforms/github/workflows/callers/go/ci-full.yml",
		".github/workflows/codeql.yml":            "hosting-platforms/github/workflows/callers/go/codeql.yml",
		".github/workflows/dependency-review.yml": "hosting-platforms/github/workflows/callers/go/dependency-review.yml",
	}
	if len(manifest.Callers) != len(want) {
		t.Fatalf("the manifest carries %d callers, want %d", len(manifest.Callers), len(want))
	}
	for _, caller := range manifest.Callers {
		master, found := want[caller.File]
		if !found {
			t.Fatalf("the manifest carries an unexpected caller %q", caller.File)
		}
		if caller.Master != master {
			t.Fatalf("caller %q binds master %q, want %q", caller.File, caller.Master, master)
		}
		if hash := hashRepositoryFile(t, caller.File); hash != caller.SHA256 {
			t.Fatalf("the tenant caller %s hashes to %s, want the bound %s", caller.File, hash, caller.SHA256)
		}
		content := readRepositoryFile(t, caller.File)
		if !strings.Contains(content, "uses: "+manifest.Home.Repository+"/.github/workflows/reusable-") {
			t.Fatalf("the tenant caller %s does not reference a home payload", caller.File)
		}
		if !strings.Contains(content, "@"+manifest.Home.SHA) {
			t.Fatalf("the tenant caller %s does not pin the bound home SHA", caller.File)
		}
		if !strings.Contains(content, `branches: [main, develop, "release/**", "support/**"]`) {
			t.Fatalf("the tenant caller %s does not cover every shared line", caller.File)
		}
	}
}

func TestCanonicalFileFamilyMatchesTheBindingManifest(t *testing.T) {
	manifest := readBindingManifest(t)
	for _, topic := range []fileBinding{
		manifest.Files.Lefthook,
		manifest.Files.Gitattributes,
		manifest.Files.Dependabot,
	} {
		if hash := hashRepositoryFile(t, topic.Path); hash != topic.SHA256 {
			t.Fatalf("the canonical file %s hashes to %s, want the bound %s", topic.Path, hash, topic.SHA256)
		}
	}
	// The gitignore topic is prefix-mode in the home verifier: the canonical
	// core is a verbatim prefix and project additions live below the mark.
	gitignore := readRepositoryFile(t, manifest.Files.Gitignore.Path)
	if !strings.HasSuffix(gitignore, "# -- project additions below this line --\n") {
		t.Fatal("the gitignore does not carry the canonical core with the project-block mark")
	}

	codeowners := readRepositoryFile(t, manifest.Codeowners.Path)
	if !strings.Contains(codeowners, "* "+manifest.Codeowners.DefaultOwner) {
		t.Fatalf("the ownership file does not bind the default owner %q", manifest.Codeowners.DefaultOwner)
	}
}

func TestConformanceWorkflowBindsTheVerifier(t *testing.T) {
	manifest := readBindingManifest(t)
	content := readRepositoryFile(t, ".github/workflows/canonical-conformance.yml")
	for _, required := range []string{
		"permissions: {}",
		"name: Canonical conformance",
		"uses: " + manifest.Home.Repository + "/.github/actions/verify-canonical-files@" + manifest.Home.SHA,
		`branches: [main, develop, "release/**", "support/**"]`,
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("the canonical conformance workflow does not contain %q", required)
		}
	}
}

func TestReleaseGatesWorkflowValidatesTheReleaseConfiguration(t *testing.T) {
	content := readRepositoryFile(t, ".github/workflows/release-gates.yml")
	for _, required := range []string{
		"name: Release gates",
		"  release-config:\n",
		"name: Release configuration",
		"goreleaser/goreleaser-action@5daf1e915a5f0af01ddbcd89a43b8061ff4f1a89",
		"args: check",
		`branches: [main, develop, "release/**", "support/**"]`,
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("the release-gates workflow does not contain %q", required)
		}
	}
}

// lifecycleFamilyBinding binds the tenant lifecycle callers to the canonical
// release-lifecycle family of the git-governance home: the exact payload pin
// on the merged home line plus the LF-stable content hash of every caller as
// recorded by the home caller hash record at that pin.
var lifecycleFamilyBinding = struct {
	homeSHA string
	callers map[string]string
}{
	homeSHA: "77da3857d8db3f3567522651750980e087c0a6eb",
	callers: map[string]string{
		"release-control.yml":                "5b339a721f8d31560090eb33f71d434686f17789ef0dc0309fe264b093604d4a",
		"release-reconciliation.yml":         "7fc7eb6efe4a71aceeee2e1c2dec32e25506d4ed59af6f85c36a571c7a6a12da",
		"execute-protected-line-request.yml": "7d166c87338cf6be30acbfcc4f01a807c464c0f0e2d9d0468e79ba229e9053b3",
		"tag-promoted-release.yml":           "578ffd6c5498545c22a3639a66cb49ed1b44964c1eb162e839f470266123bada",
		"publish-release-artifacts.yml":      "a790964db50b6bde5bfaca0b7358951952915cc2183300c97e36e83289521bd2",
		"hotfix-delivery.yml":                "2b5ddf06163f0453465a67dede9adeb0ffcc70ff2733d413f5d997d5a2e60d97",
		"hotfix-propagation.yml":             "476085ba51dbfb6f1d004943c5e603c95a8fa08d99371d58b96014d689f96208",
	},
}

func TestLifecycleCallersBindTheGovernedFamily(t *testing.T) {
	for _, name := range []string{
		"release-control.yml",
		"release-reconciliation.yml",
		"execute-protected-line-request.yml",
		"tag-promoted-release.yml",
		"publish-release-artifacts.yml",
		"hotfix-delivery.yml",
		"hotfix-propagation.yml",
	} {
		want, bound := lifecycleFamilyBinding.callers[name]
		if !bound {
			t.Fatalf("no governed family binding recorded for %q", name)
		}
		path := filepath.Join(".github", "workflows", name)
		if hash := hashRepositoryFile(t, path); hash != want {
			t.Fatalf("the lifecycle caller %s hashes to %s, want the governed family hash %s", path, hash, want)
		}
		content := readRepositoryFile(t, path)
		reference := "uses: t33n-software/git-governance/.github/workflows/reusable-" + name + "@" + lifecycleFamilyBinding.homeSHA
		if !strings.Contains(content, reference) {
			t.Fatalf("the lifecycle caller %s does not pin the governed family at the merged home line", path)
		}
	}
	for _, legacy := range []string{"recover-protected-line-request.yml", "release.yml"} {
		path := filepath.Join(".github", "workflows", legacy)
		if _, err := os.Stat(repositoryPath(path)); !os.IsNotExist(err) {
			t.Fatalf("the legacy lifecycle lane %s must be absent: the governed family carries the bound executor recovery mode and the release delivery", path)
		}
	}
}

func TestModuleIdentityAndQualityContract(t *testing.T) {
	goMod := readRepositoryFile(t, "go.mod")
	for _, required := range []string{
		"module github.com/t33n-software/license-hub",
		"go 1.26",
		"toolchain go1.26.6",
	} {
		if !strings.Contains(goMod, required) {
			t.Fatalf("go.mod does not contain %q", required)
		}
	}

	quality := readRepositoryFile(t, "git-governance.quality.json")
	for _, required := range []string{
		`"schemaVersion": 4`,
		`"language": "go"`,
		`"version": "1.26.6"`,
		`"extends": []`,
		"full-local-build",
		"\"args\": [\n        \"tool\",\n        \"-modfile\",\n        \"tools/go.mod\",\n        \"quality-gate\"\n      ],",
		"./cmd/license",
	} {
		if !strings.Contains(quality, required) {
			t.Fatalf("git-governance.quality.json does not contain %q", required)
		}
	}
	for _, forbidden := range []string{
		"./cmd/build",
		"./cmd/check-coverage",
		`"defaults"`,
	} {
		if strings.Contains(quality, forbidden) {
			t.Fatalf("git-governance.quality.json must not contain %q: the canonical gate chain lives in the pinned orchestrator and the schema owns the family default", forbidden)
		}
	}
	for _, copy := range []string{"cmd/build", "cmd/check-coverage"} {
		if _, err := os.Stat(repositoryPath(copy)); !os.IsNotExist(err) {
			t.Fatalf("the chain copy %s must be absent: the canonical gate chain is referenced via the tool pin, never re-implemented per repository", copy)
		}
	}

	lefthook := readRepositoryFile(t, "lefthook.yml")
	if !strings.Contains(lefthook, "git-governance --interactive never validate pre-push --remote") {
		t.Fatal("lefthook.yml does not bind the canonical pre-push validation")
	}
}

func TestGoToolingModuleContract(t *testing.T) {
	toolsMod := readRepositoryFile(t, filepath.Join("tools", "go.mod"))
	for _, required := range []string{
		"module github.com/t33n-software/license-hub/tools",
		"toolchain go1.26.6",
		"github.com/evilmartians/lefthook/v2",
		"golang.org/x/vuln/cmd/govulncheck",
		"honnef.co/go/tools/cmd/staticcheck",
		"github.com/t33n-software/go-quality-authority/cmd/quality-gate",
		"github.com/t33n-software/go-quality-authority/cmd/check-coverage",
		"github.com/t33n-software/repository-governance/cmd/verify-canonical",
		"github.com/t33n-software/git-governance/cmd/git-governance",
	} {
		if !strings.Contains(toolsMod, required) {
			t.Fatalf("tools/go.mod does not contain %q", required)
		}
	}
	if _, err := os.Stat(repositoryPath("tools", "go.sum")); err != nil {
		t.Fatalf("tools/go.sum is missing: %v", err)
	}
}

func TestOrganizationRulesetAdoptionHasNoLocalLegacyDefinitions(t *testing.T) {
	if _, err := os.Stat(repositoryPath("docs", "hosting-platforms")); !os.IsNotExist(err) {
		t.Fatalf("legacy ruleset location must not exist")
	}

	conventions := readRepositoryFile(t, filepath.Join("docs", "conventions", "hosting-platforms", "github", "rule-sets", "README.md"))
	for _, required := range []string{
		"git-governance",
		"quality-gates=full",
		"~ALL",
	} {
		if !strings.Contains(conventions, required) {
			t.Fatalf("rule-set conventions README does not contain %q", required)
		}
	}
}

func readRepositoryFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(repositoryPath(filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(content)
}

func repositoryPath(parts ...string) string {
	return filepath.Join(append([]string{"..", ".."}, parts...)...)
}
