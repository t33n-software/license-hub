// Package application composes the license render and verify use cases.
package application

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/t33n-software/license-hub/internal/domain/digest"
	"github.com/t33n-software/license-hub/internal/domain/ecosystem"
	"github.com/t33n-software/license-hub/internal/domain/lockfile"
	"github.com/t33n-software/license-hub/internal/domain/placeholder"
	"github.com/t33n-software/license-hub/internal/domain/render"
	"github.com/t33n-software/license-hub/internal/domain/values"
)

// ErrMissingValues marks an incomplete values document; the CLI maps it to
// the VALUE_INVALID error code of the structured error contract.
//
// Convention: docs/conventions/cli/output/README.md
var ErrMissingValues = errors.New("missing required values")

// ErrUnresolvedPlaceholders marks a render whose anchors stay unresolved; the
// CLI maps it to the VALUE_INVALID error code of the structured error
// contract.
//
// Convention: docs/conventions/cli/output/README.md
var ErrUnresolvedPlaceholders = errors.New("unresolved placeholders")

// ErrInvalidEcosystemSurface marks a declared ecosystem surface that cannot
// be aligned or proven; the CLI maps it to the VALUE_INVALID error class.
//
// Convention: docs/conventions/cli/output/README.md
var ErrInvalidEcosystemSurface = errors.New("invalid ecosystem license surface")

// FileSystem is the storage port consumed by the license use cases.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	// ListNames lists the non-directory entry names of a directory.
	ListNames(dir string) ([]string, error)
}

// LicenseService renders and verifies canonical license instances.
type LicenseService struct {
	fs FileSystem
}

// NewLicenseService binds the service to the storage port.
func NewLicenseService(fs FileSystem) *LicenseService {
	return &LicenseService{fs: fs}
}

// RenderRequest describes one instance render.
type RenderRequest struct {
	TemplatePath    string
	OrgDefaultsPath string
	ValuesPath      string
	OutDir          string
}

// RenderResult reports the rendered artifacts, the pinned template digest,
// and the ecosystem license surfaces the render aligned or skipped.
type RenderResult struct {
	Written []string
	Digest  string
	// Aligned carries the ecosystem manifests whose license field the render
	// aligned to the lock projection.
	Aligned []string
	// Skipped carries the declared ecosystem surfaces the render could not
	// align, with the reason.
	Skipped []string
}

// preparedRender carries the validated inputs of one render: the canonical
// content, the template bytes, the resolved instance targets, and the
// prepared ecosystem alignments.
type preparedRender struct {
	content    string
	template   []byte
	targets    []string
	alignments []*ecosystemAlignment
}

// ecosystemAlignment carries the prepared ecosystem license-field alignment
// of one render: the aligned manifest content and whether bytes change, or
// the reason the declared surface cannot be aligned.
type ecosystemAlignment struct {
	manifestPath string
	content      string
	changed      bool
	skip         string
}

// ecosystemRenderRow binds one ecosystem matrix row to its render-side
// alignment: the declared seam language, the manifest surface, and the
// lock-projection alignment function. A discovery row carries no fixed
// manifest name; it discovers its manifest candidates from the directory
// listing of the render target.
type ecosystemRenderRow struct {
	language      string
	manifestName  string
	manifestLabel string
	align         func(name, content string, merged map[string]string) (string, bool, error)
	discover      func(names []string) []string
}

var npmRenderRow = ecosystemRenderRow{
	language:      ecosystem.NpmLanguage,
	manifestName:  ecosystem.NpmManifestName,
	manifestLabel: "npm manifest",
	align: func(_ string, content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignNpmLicense(content, ecosystem.NpmProjection(merged))
	},
}

var pythonRenderRow = ecosystemRenderRow{
	language:      ecosystem.PythonLanguage,
	manifestName:  ecosystem.PythonManifestName,
	manifestLabel: "python manifest",
	align: func(_ string, content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignPythonLicense(content, ecosystem.PythonProjection(merged), ecosystem.PythonLicenseFilesTarget)
	},
}

var cargoRenderRow = ecosystemRenderRow{
	language:      ecosystem.CargoLanguage,
	manifestName:  ecosystem.CargoManifestName,
	manifestLabel: "cargo manifest",
	align: func(_ string, content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignCargoLicense(content, ecosystem.CargoProjection(merged))
	},
}

var mavenRenderRow = ecosystemRenderRow{
	language:      ecosystem.MavenLanguage,
	manifestName:  ecosystem.MavenManifestName,
	manifestLabel: "maven manifest",
	align: func(_ string, content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignMavenLicense(content, ecosystem.MavenProjection(merged))
	},
}

var nugetRenderRow = ecosystemRenderRow{
	language:      ecosystem.NuGetLanguage,
	manifestLabel: "nuget manifest",
	align: func(name, content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignNuGetSurface(name, content, ecosystem.NuGetProjection(merged))
	},
	discover: ecosystem.NuGetManifestNames,
}

var composerRenderRow = ecosystemRenderRow{
	language:      ecosystem.ComposerLanguage,
	manifestName:  ecosystem.ComposerManifestName,
	manifestLabel: "composer manifest",
	align: func(_ string, content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignComposerLicense(content, ecosystem.ComposerProjection(merged))
	},
}

var gemspecRenderRow = ecosystemRenderRow{
	language:      ecosystem.RubyLanguage,
	manifestLabel: "gemspec manifest",
	align: func(_ string, content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignGemspecLicense(content, ecosystem.GemspecProjection(merged))
	},
	discover: ecosystem.GemspecManifestNames,
}

// prepare reads and validates every input of a render without writing
// anything.
func (s *LicenseService) prepare(req RenderRequest) (preparedRender, error) {
	template, err := s.fs.ReadFile(req.TemplatePath)
	if err != nil {
		return preparedRender{}, fmt.Errorf("read template: %w", err)
	}
	merged, err := s.mergedValues(req.OrgDefaultsPath, req.ValuesPath)
	if err != nil {
		return preparedRender{}, err
	}
	content := render.Execute(string(template), merged)
	if unresolved := placeholder.Unresolved(content); len(unresolved) > 0 {
		return preparedRender{}, fmt.Errorf("%w: %s", ErrUnresolvedPlaceholders, strings.Join(unresolved, ", "))
	}
	prepared := preparedRender{
		content:  content,
		template: template,
		targets:  instancePaths(req.OutDir, merged),
	}
	for _, row := range []ecosystemRenderRow{npmRenderRow, pythonRenderRow, cargoRenderRow, mavenRenderRow, nugetRenderRow, composerRenderRow, gemspecRenderRow} {
		alignments, err := s.prepareEcosystemAlignment(req, merged, row)
		if err != nil {
			return preparedRender{}, err
		}
		prepared.alignments = append(prepared.alignments, alignments...)
	}
	return prepared, nil
}

// Render renders the canonical template into the LICENSE and LICENSES/
// artifacts of the target directory and aligns the declared ecosystem
// license surfaces.
func (s *LicenseService) Render(req RenderRequest) (RenderResult, error) {
	prepared, err := s.prepare(req)
	if err != nil {
		return RenderResult{}, err
	}
	for _, target := range prepared.targets {
		if err := s.fs.WriteFile(target, []byte(prepared.content)); err != nil {
			return RenderResult{}, fmt.Errorf("write %s: %w", target, err)
		}
	}
	result := RenderResult{Written: prepared.targets, Digest: digest.SHA256(prepared.template)}
	for _, alignment := range prepared.alignments {
		if err := s.writeAlignment(alignment, &result); err != nil {
			return RenderResult{}, err
		}
	}
	return result, nil
}

// writeAlignment writes one prepared ecosystem alignment and reports it in
// the render result.
func (s *LicenseService) writeAlignment(alignment *ecosystemAlignment, result *RenderResult) error {
	if alignment == nil {
		return nil
	}
	if alignment.changed {
		if err := s.fs.WriteFile(alignment.manifestPath, []byte(alignment.content)); err != nil {
			return fmt.Errorf("write %s: %w", alignment.manifestPath, err)
		}
		result.Aligned = append(result.Aligned, alignment.manifestPath)
	}
	if alignment.skip != "" {
		result.Skipped = append(result.Skipped, alignment.skip)
	}
	return nil
}

// PlanResult reports what a render would write, without writing it.
type PlanResult struct {
	Targets []string
	Digest  string
	// Alignments previews the ecosystem license-field alignments the render
	// would perform.
	Alignments []string
	// Skips previews the declared ecosystem surfaces the render cannot align.
	Skips []string
}

// PlanRender computes the render plan without mutating anything.
//
// Convention: docs/conventions/cli/interaction/README.md (the dry-run duty of
// every mutating command)
func (s *LicenseService) PlanRender(req RenderRequest) (PlanResult, error) {
	prepared, err := s.prepare(req)
	if err != nil {
		return PlanResult{}, err
	}
	plan := PlanResult{Targets: prepared.targets, Digest: digest.SHA256(prepared.template)}
	for _, alignment := range prepared.alignments {
		planAlignment(alignment, &plan)
	}
	return plan, nil
}

// planAlignment previews one prepared ecosystem alignment in the dry-run
// plan.
func planAlignment(alignment *ecosystemAlignment, plan *PlanResult) {
	if alignment == nil {
		return
	}
	if alignment.changed {
		plan.Alignments = append(plan.Alignments, "align "+alignment.manifestPath+" to the lock projection")
	}
	if alignment.skip != "" {
		plan.Skips = append(plan.Skips, alignment.skip)
	}
}

// VerifyRequest describes one instance verification.
type VerifyRequest struct {
	TemplatePath    string
	OrgDefaultsPath string
	ValuesPath      string
	LockPath        string
	Dir             string
}

// Verify executes the tenant drift guard and reports every violation. An
// empty violation list means the committed instance matches the canonical
// render of the pinned template and the declared ecosystem license surfaces
// match their lock projections.
func (s *LicenseService) Verify(req VerifyRequest) ([]string, error) {
	template, err := s.fs.ReadFile(req.TemplatePath)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	violations := make([]string, 0)
	if req.LockPath != "" {
		lockViolations, err := s.verifyLock(req.LockPath, template)
		if err != nil {
			return nil, err
		}
		violations = append(violations, lockViolations...)
	}
	merged, err := s.mergedValues(req.OrgDefaultsPath, req.ValuesPath)
	if err != nil {
		return nil, err
	}
	content := render.Execute(string(template), merged)
	if unresolved := placeholder.Unresolved(content); len(unresolved) > 0 {
		violations = append(violations, "unresolved placeholders: "+strings.Join(unresolved, ", "))
	}
	for _, target := range instancePaths(req.Dir, merged) {
		committed, err := s.fs.ReadFile(target)
		if err != nil {
			violations = append(violations, "missing rendered file: "+target)
			continue
		}
		if string(committed) != content {
			violations = append(violations, "rendered file drifted from canonical render: "+target)
		}
	}
	for _, row := range []ecosystemVerifyRow{npmVerifyRow, pythonVerifyRow, cargoVerifyRow, mavenVerifyRow, nugetVerifyRow, composerVerifyRow, gemspecVerifyRow} {
		rowViolations, err := s.verifyEcosystem(req, merged, row)
		if err != nil {
			return nil, err
		}
		violations = append(violations, rowViolations...)
	}
	return violations, nil
}

// TemplateDigest computes the canonical digest of a template file.
func (s *LicenseService) TemplateDigest(path string) (string, error) {
	template, err := s.fs.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}
	return digest.SHA256(template), nil
}

// prepareEcosystemAlignment derives the ecosystem license-field alignments
// of one render from the declaration seam and the merged tenant values. A
// missing seam or a foreign declaration leaves the surface untouched; a
// declared project without a manifest is reported as a skip that the verify
// lane carries as a fail-closed finding. A discovery row aligns every
// manifest candidate it finds in the render target directory.
//
// Convention: spec/ecosystem-license-metadata.md
func (s *LicenseService) prepareEcosystemAlignment(req RenderRequest, merged map[string]string, row ecosystemRenderRow) ([]*ecosystemAlignment, error) {
	seamPath := filepath.Join(req.OutDir, ecosystem.SeamFileName)
	seamData, err := s.fs.ReadFile(seamPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read ecosystem declaration %s: %w", seamPath, err)
	}
	language, err := ecosystem.SeamLanguage(seamData)
	if err != nil {
		return nil, fmt.Errorf("%w: read ecosystem declaration %s: %v", ErrInvalidEcosystemSurface, seamPath, err)
	}
	if language != row.language {
		return nil, nil
	}
	if row.discover == nil {
		alignment, err := s.prepareNamedAlignment(filepath.Join(req.OutDir, row.manifestName), row, merged)
		if err != nil {
			return nil, err
		}
		return []*ecosystemAlignment{alignment}, nil
	}
	names, err := s.fs.ListNames(req.OutDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("list %s manifests in %s: %w", row.manifestLabel, req.OutDir, err)
		}
		names = nil
	}
	candidates := row.discover(names)
	if len(candidates) == 0 {
		return []*ecosystemAlignment{{
			manifestPath: req.OutDir,
			skip:         row.manifestLabel + " is absent: the declared license surface cannot be aligned",
		}}, nil
	}
	alignments := []*ecosystemAlignment{}
	for _, name := range candidates {
		alignment, err := s.prepareNamedAlignment(filepath.Join(req.OutDir, name), row, merged)
		if err != nil {
			return nil, err
		}
		alignments = append(alignments, alignment)
	}
	return alignments, nil
}

// prepareNamedAlignment prepares the license-field alignment of one
// manifest: the aligned content and whether bytes change, or the reason the
// declared surface cannot be aligned.
func (s *LicenseService) prepareNamedAlignment(manifestPath string, row ecosystemRenderRow, merged map[string]string) (*ecosystemAlignment, error) {
	data, err := s.fs.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ecosystemAlignment{
				manifestPath: manifestPath,
				skip:         row.manifestLabel + " is absent: the declared license surface cannot be aligned",
			}, nil
		}
		return nil, fmt.Errorf("read %s %s: %w", row.manifestLabel, manifestPath, err)
	}
	aligned, changed, err := row.align(filepath.Base(manifestPath), string(data), merged)
	if err != nil {
		return nil, fmt.Errorf("%w: align %s %s: %v", ErrInvalidEcosystemSurface, row.manifestLabel, manifestPath, err)
	}
	return &ecosystemAlignment{manifestPath: manifestPath, content: aligned, changed: changed}, nil
}

// ecosystemVerifyRow binds one ecosystem matrix row to its verify-side
// finding derivation. A discovery row carries no fixed manifest name and
// proves every manifest candidate it finds in the verified directory.
type ecosystemVerifyRow struct {
	language     string
	manifestName string
	label        string
	missingHint  string
	findings     func(name, content, manifestPath string, merged map[string]string) []string
	discover     func(names []string) []string
}

var npmVerifyRow = ecosystemVerifyRow{
	language:     ecosystem.NpmLanguage,
	manifestName: ecosystem.NpmManifestName,
	label:        "npm",
	findings:     npmFindings,
}

var pythonVerifyRow = ecosystemVerifyRow{
	language:     ecosystem.PythonLanguage,
	manifestName: ecosystem.PythonManifestName,
	label:        "python",
	findings:     pythonFindings,
}

var cargoVerifyRow = ecosystemVerifyRow{
	language:     ecosystem.CargoLanguage,
	manifestName: ecosystem.CargoManifestName,
	label:        "cargo",
	findings:     cargoFindings,
}

var mavenVerifyRow = ecosystemVerifyRow{
	language:     ecosystem.MavenLanguage,
	manifestName: ecosystem.MavenManifestName,
	label:        "maven",
	findings:     mavenFindings,
}

var nugetVerifyRow = ecosystemVerifyRow{
	language:    ecosystem.NuGetLanguage,
	label:       "nuget",
	missingHint: "*.csproj or *.nuspec",
	findings:    nugetFindings,
	discover:    ecosystem.NuGetManifestNames,
}

var composerVerifyRow = ecosystemVerifyRow{
	language:     ecosystem.ComposerLanguage,
	manifestName: ecosystem.ComposerManifestName,
	label:        "composer",
	findings:     composerFindings,
}

var gemspecVerifyRow = ecosystemVerifyRow{
	language:    ecosystem.RubyLanguage,
	label:       "gemspec",
	missingHint: "*.gemspec",
	findings:    gemspecFindings,
	discover:    ecosystem.GemspecManifestNames,
}

// verifyEcosystem proves one declared ecosystem surface fail-closed in both
// directions: a declared manifest that is missing or diverging is a finding,
// a manifest without a covering declaration is a shadow-surface finding, and
// an unscannable surface is a proof failure. A discovery row proves every
// manifest candidate it finds in the verified directory.
//
// Convention: spec/ecosystem-license-metadata.md
func (s *LicenseService) verifyEcosystem(req VerifyRequest, merged map[string]string, row ecosystemVerifyRow) ([]string, error) {
	seamPath := filepath.Join(req.Dir, ecosystem.SeamFileName)
	seamData, seamErr := s.fs.ReadFile(seamPath)
	language := ""
	switch {
	case seamErr == nil:
		parsed, err := ecosystem.SeamLanguage(seamData)
		if err != nil {
			return nil, fmt.Errorf("%w: read ecosystem declaration %s: %v", ErrInvalidEcosystemSurface, seamPath, err)
		}
		language = parsed
	case errors.Is(seamErr, os.ErrNotExist):
		// No declaration: only the shadow-surface guard applies.
	default:
		return nil, fmt.Errorf("read ecosystem declaration %s: %w", seamPath, seamErr)
	}
	if row.discover != nil {
		return s.verifyDiscoveredEcosystem(req, seamErr, language, merged, row)
	}
	manifestPath := filepath.Join(req.Dir, row.manifestName)
	data, manifestErr := s.fs.ReadFile(manifestPath)
	switch {
	case errors.Is(manifestErr, os.ErrNotExist):
		if language == row.language {
			return []string{
				"declared " + row.label + " ecosystem manifest missing: " + manifestPath +
					" (declared language: \"" + row.language + "\")",
			}, nil
		}
		return nil, nil
	case manifestErr != nil:
		return nil, fmt.Errorf("read %s manifest %s: %w", row.label, manifestPath, manifestErr)
	}
	if language != row.language {
		detail := "no ecosystem declaration found (" + ecosystem.SeamFileName + " is absent)"
		if seamErr == nil {
			detail = "declared language: \"" + language + "\""
		}
		return []string{
			row.label + " license metadata surface without covering declaration: " + manifestPath + " (" + detail + ")",
		}, nil
	}
	return row.findings(row.manifestName, string(data), manifestPath, merged), nil
}

// verifyDiscoveredEcosystem proves the manifest candidates of a discovery
// row fail-closed in both directions: the declared row without any
// candidate is a missing finding, every candidate without a covering
// declaration is a shadow-surface finding, and every candidate surface is
// proven against the lock projection.
func (s *LicenseService) verifyDiscoveredEcosystem(req VerifyRequest, seamErr error, language string, merged map[string]string, row ecosystemVerifyRow) ([]string, error) {
	names, err := s.fs.ListNames(req.Dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("list %s manifests in %s: %w", row.label, req.Dir, err)
		}
		names = nil
	}
	candidates := row.discover(names)
	if len(candidates) == 0 {
		if language == row.language {
			return []string{
				"declared " + row.label + " ecosystem manifest missing: no " + row.missingHint +
					" manifest is present under " + req.Dir + " (declared language: \"" + row.language + "\")",
			}, nil
		}
		return nil, nil
	}
	violations := []string{}
	for _, name := range candidates {
		manifestPath := filepath.Join(req.Dir, name)
		data, err := s.fs.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("read %s manifest %s: %w", row.label, manifestPath, err)
		}
		if language != row.language {
			detail := "no ecosystem declaration found (" + ecosystem.SeamFileName + " is absent)"
			if seamErr == nil {
				detail = "declared language: \"" + language + "\""
			}
			violations = append(violations, row.label+" license metadata surface without covering declaration: "+manifestPath+" ("+detail+")")
			continue
		}
		violations = append(violations, row.findings(name, string(data), manifestPath, merged)...)
	}
	return violations, nil
}

// npmFindings derives the npm license-surface findings from the inspected
// manifest and the lock projection.
func npmFindings(_ string, content, manifestPath string, merged map[string]string) []string {
	target := ecosystem.NpmProjection(merged)
	value, state, err := ecosystem.InspectNpmLicense(content)
	if err != nil {
		return []string{"npm license surface cannot be proven: " + manifestPath + " (" + err.Error() + ")"}
	}
	switch state {
	case ecosystem.LicenseFieldMissing:
		return []string{"npm license field is missing: " + manifestPath + " (expected \"" + target + "\")"}
	case ecosystem.LicenseFieldNonString:
		return []string{"npm license field is not a JSON string: " + manifestPath + " (expected \"" + target + "\")"}
	}
	if value != target {
		return []string{
			"npm license field diverges from the lock projection: observed \"" + value +
				"\", expected \"" + target + "\" (run the render to align)",
		}
	}
	return nil
}

// pythonFindings derives the Python license-surface findings from the
// inspected manifest and the lock projection: the license field, the
// license-files globs, and the deprecated License :: classifier guard.
func pythonFindings(_ string, content, manifestPath string, merged map[string]string) []string {
	target := ecosystem.PythonProjection(merged)
	surface, err := ecosystem.InspectPythonLicense(content)
	if err != nil {
		return []string{"python license surface cannot be proven: " + manifestPath + " (" + err.Error() + ")"}
	}
	violations := []string{}
	switch surface.State {
	case ecosystem.PythonLicenseMissing:
		violations = append(violations, "python license field is missing: "+manifestPath+" (expected \""+target+"\")")
	case ecosystem.PythonLicenseTable:
		violations = append(violations, "python license field uses the deprecated license table form: "+manifestPath+" (expected \""+target+"\")")
	case ecosystem.PythonLicenseInvalid:
		violations = append(violations, "python license field is not a license expression string: "+manifestPath+" (expected \""+target+"\")")
	case ecosystem.PythonLicenseString:
		if surface.Value != target {
			violations = append(violations, "python license field diverges from the lock projection: observed \""+surface.Value+"\", expected \""+target+"\" (run the render to align)")
		}
	}
	filesTarget := ecosystem.PythonLicenseFilesForm(ecosystem.PythonLicenseFilesTarget)
	switch {
	case !surface.LicenseFilesSet:
		violations = append(violations, "python license-files field is missing: "+manifestPath+" (expected "+filesTarget+")")
	case !surface.LicenseFilesValid:
		violations = append(violations, "python license-files field is not an array of strings: "+manifestPath)
	case !slices.Equal(surface.LicenseFiles, ecosystem.PythonLicenseFilesTarget):
		violations = append(violations, "python license-files field diverges from the lock projection: observed "+ecosystem.PythonLicenseFilesForm(surface.LicenseFiles)+", expected "+filesTarget+" (run the render to align)")
	}
	if len(surface.DeprecatedClassifiers) > 0 {
		violations = append(violations, fmt.Sprintf(
			"pyproject.toml carries the deprecated License :: classifiers: %s (%d entries; removal is an explicit tenant decision)",
			manifestPath, len(surface.DeprecatedClassifiers),
		))
	}
	return violations
}

// cargoFindings derives the Rust license-surface findings from the inspected
// manifest and the lock projection: the declared form (the exclusive
// license keys), its value, and the fail-closed proof state.
func cargoFindings(_ string, content, manifestPath string, merged map[string]string) []string {
	form := ecosystem.CargoProjection(merged)
	surface, err := ecosystem.InspectCargoLicense(content)
	if err != nil {
		return []string{"cargo license surface cannot be proven: " + manifestPath + " (" + err.Error() + ")"}
	}
	violations := []string{}
	switch surface.State {
	case ecosystem.CargoLicenseMissing:
		violations = append(violations, "cargo license field is missing: "+manifestPath+" (expected "+form.String()+")")
	case ecosystem.CargoLicenseInvalid:
		violations = append(violations, "cargo license field is not a license expression string: "+manifestPath+" (expected "+form.String()+")")
	case ecosystem.CargoLicenseExpression:
		if form.Field != ecosystem.CargoLicenseField || surface.Value != form.Value {
			violations = append(violations, "cargo license field diverges from the lock projection: observed "+ecosystem.CargoLicenseField+" = \""+surface.Value+"\", expected "+form.String()+" (run the render to align)")
		}
	case ecosystem.CargoLicenseFileForm:
		if form.Field != ecosystem.CargoLicenseFileKey || surface.Value != form.Value {
			violations = append(violations, "cargo license field diverges from the lock projection: observed "+ecosystem.CargoLicenseFileKey+" = \""+surface.Value+"\", expected "+form.String()+" (run the render to align)")
		}
	}
	return violations
}

// mavenFindings derives the Maven license-surface findings from the
// inspected manifest and the lock projection: the licenses element, its
// license count, and the name and url values of the first license element.
func mavenFindings(_ string, content, manifestPath string, merged map[string]string) []string {
	form := ecosystem.MavenProjection(merged)
	surface, err := ecosystem.InspectMavenLicense(content)
	if err != nil {
		return []string{"maven license surface cannot be proven: " + manifestPath + " (" + err.Error() + ")"}
	}
	violations := []string{}
	switch {
	case !surface.LicensesPresent:
		violations = append(violations, "maven licenses element is missing: "+manifestPath+" (expected <name> \""+form.Name+"\" and <url> \""+form.URL+"\")")
	case surface.LicenseCount > 1:
		violations = append(violations, "maven licenses element carries multiple license elements: "+manifestPath+" (the resolution is an explicit tenant decision)")
	case !surface.NamePresent:
		violations = append(violations, "maven license name element is missing: "+manifestPath+" (expected \""+form.Name+"\")")
	case surface.Name != form.Name:
		violations = append(violations, "maven license name diverges from the lock projection: observed \""+surface.Name+"\", expected \""+form.Name+"\" (run the render to align)")
	case !surface.URLPresent:
		violations = append(violations, "maven license url element is missing: "+manifestPath+" (expected \""+form.URL+"\")")
	case surface.URL != form.URL:
		violations = append(violations, "maven license url diverges from the lock projection: observed \""+surface.URL+"\", expected \""+form.URL+"\" (run the render to align)")
	}
	return violations
}

// nugetFindings derives the .NET license-surface findings from one
// discovered manifest and the lock projection. The manifest name selects
// the surface phrasing: the nuspec license element or the MSBuild license
// property.
func nugetFindings(name, content, manifestPath string, merged map[string]string) []string {
	form := ecosystem.NuGetProjection(merged)
	surface, err := ecosystem.InspectNuGetSurface(name, content)
	if err != nil {
		return []string{"nuget license surface cannot be proven: " + manifestPath + " (" + err.Error() + ")"}
	}
	if strings.HasSuffix(strings.ToLower(name), ecosystem.MSBuildSuffix) {
		return msBuildFindings(surface, manifestPath, form)
	}
	return nuspecFindings(surface, manifestPath, form)
}

// nuspecFindings derives the nuspec license-element findings: the declared
// form, its value, the multiple-declaration refusal, and the deprecated
// licenseUrl guard.
func nuspecFindings(surface ecosystem.NuGetSurface, manifestPath string, form ecosystem.NuGetLicenseForm) []string {
	expected := form.NuspecLabel()
	violations := []string{}
	if surface.FormCount > 1 {
		violations = append(violations, "nuget metadata element carries multiple license elements: "+manifestPath+" (the resolution is an explicit tenant decision)")
	} else {
		switch surface.State {
		case ecosystem.NuGetSurfaceMissing:
			violations = append(violations, "nuget license element is missing: "+manifestPath+" (expected "+expected+")")
		case ecosystem.NuGetSurfaceInvalid:
			violations = append(violations, "nuget license element carries no valid type: "+manifestPath+" (expected type=\"expression\" or type=\"file\")")
		case ecosystem.NuGetSurfaceDeclared:
			if surface.Expression != form.Expression || surface.Value != form.Value {
				observed := ecosystem.NuGetLicenseForm{Expression: surface.Expression, Value: surface.Value}.NuspecLabel()
				violations = append(violations, "nuget license element diverges from the lock projection: observed "+observed+", expected "+expected+" (run the render to align)")
			}
		}
	}
	if surface.DeprecatedURLPresent {
		violations = append(violations, "nuget manifest carries the deprecated licenseUrl element: "+manifestPath+" (removal is an explicit tenant decision)")
	}
	return violations
}

// msBuildFindings derives the MSBuild license-property findings: the
// declared form, its value, the multiple-declaration refusal, and the
// deprecated PackageLicenseUrl guard.
func msBuildFindings(surface ecosystem.NuGetSurface, manifestPath string, form ecosystem.NuGetLicenseForm) []string {
	expected := form.MSBuildLabel()
	violations := []string{}
	if surface.FormCount > 1 {
		violations = append(violations, "msbuild project carries multiple license properties: "+manifestPath+" (the resolution is an explicit tenant decision)")
	} else {
		switch surface.State {
		case ecosystem.NuGetSurfaceMissing:
			violations = append(violations, "msbuild license property is missing: "+manifestPath+" (expected "+expected+")")
		case ecosystem.NuGetSurfaceDeclared:
			if surface.Expression != form.Expression || surface.Value != form.Value {
				observed := ecosystem.NuGetLicenseForm{Expression: surface.Expression, Value: surface.Value}.MSBuildLabel()
				violations = append(violations, "msbuild license property diverges from the lock projection: observed "+observed+", expected "+expected+" (run the render to align)")
			}
		}
	}
	if surface.DeprecatedURLPresent {
		violations = append(violations, "msbuild manifest carries the deprecated PackageLicenseUrl property: "+manifestPath+" (removal is an explicit tenant decision)")
	}
	return violations
}

// composerFindings derives the Composer license-surface findings from the
// inspected manifest and the lock projection: the declared form (the string
// expression or the one-element array spelling), its value, and the
// fail-closed proof state.
func composerFindings(_ string, content, manifestPath string, merged map[string]string) []string {
	target := ecosystem.ComposerProjection(merged)
	surface, err := ecosystem.InspectComposerLicense(content)
	if err != nil {
		return []string{"composer license surface cannot be proven: " + manifestPath + " (" + err.Error() + ")"}
	}
	violations := []string{}
	switch {
	case surface.State == ecosystem.ComposerLicenseMissing:
		violations = append(violations, "composer license field is missing: "+manifestPath+" (expected \""+target+"\")")
	case surface.State == ecosystem.ComposerLicenseInvalid:
		violations = append(violations, "composer license field is not a license expression string or array: "+manifestPath+" (expected \""+target+"\")")
	case surface.State == ecosystem.ComposerLicenseArray && len(surface.Entries) > 1:
		violations = append(violations, "composer license field carries multiple license entries: "+manifestPath+" (the resolution is an explicit tenant decision)")
	case surface.Value != target:
		violations = append(violations, "composer license field diverges from the lock projection: observed \""+surface.Value+"\", expected \""+target+"\" (run the render to align)")
	}
	return violations
}

// gemspecFindings derives the Ruby license-surface findings from the
// inspected gemspec and the lock projection: the declared assignment (the
// singular string form or the value-equal plural spelling), its entry, and
// the fail-closed proof state.
func gemspecFindings(_ string, content, manifestPath string, merged map[string]string) []string {
	target := ecosystem.GemspecProjection(merged)
	surface, err := ecosystem.InspectGemspecLicense(content)
	if err != nil {
		return []string{"gemspec license surface cannot be proven: " + manifestPath + " (" + err.Error() + ")"}
	}
	violations := []string{}
	switch {
	case surface.State == ecosystem.GemspecLicenseMissing:
		violations = append(violations, "gemspec license field is missing: "+manifestPath+" (expected \""+target+"\")")
	case surface.State == ecosystem.GemspecLicenseInvalid:
		violations = append(violations, "gemspec license field is not a license expression string or array: "+manifestPath+" (expected \""+target+"\")")
	case surface.State == ecosystem.GemspecLicenseArray && len(surface.Entries) > 1:
		violations = append(violations, "gemspec license field carries multiple license entries: "+manifestPath+" (the resolution is an explicit tenant decision)")
	case surface.Value != target:
		violations = append(violations, "gemspec license field diverges from the lock projection: observed \""+surface.Value+"\", expected \""+target+"\" (run the render to align)")
	}
	return violations
}

func (s *LicenseService) verifyLock(path string, template []byte) ([]string, error) {
	data, err := s.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read lock file: %w", err)
	}
	lock, err := lockfile.Parse(data)
	if err != nil {
		return nil, err
	}
	if lock.Digest != digest.SHA256(template) {
		return []string{"template digest does not match pinned lock digest: " + path}, nil
	}
	return nil, nil
}

func (s *LicenseService) mergedValues(orgDefaultsPath, valuesPath string) (map[string]string, error) {
	orgDefaults, err := s.readValues(orgDefaultsPath)
	if err != nil {
		return nil, err
	}
	project, err := s.readValues(valuesPath)
	if err != nil {
		return nil, err
	}
	merged := values.Merge(orgDefaults, project)
	if missing := values.MissingRequired(merged); len(missing) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrMissingValues, strings.Join(missing, ", "))
	}
	return merged, nil
}

func (s *LicenseService) readValues(path string) (map[string]string, error) {
	data, err := s.fs.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read values %s: %w", path, err)
	}
	parsed, err := values.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("read values %s: %w", path, err)
	}
	return parsed, nil
}

// instancePaths resolves the canonical instance locations. The REUSE license
// text file carries the tenant-declared SPDX identifier when the values
// declare one; custom licenses without a listed identifier keep the
// LicenseRef-<LICENSE_ID> form.
func instancePaths(dir string, merged map[string]string) []string {
	stem := "LicenseRef-" + merged["LICENSE_ID"]
	if identifier := strings.TrimSpace(merged["SPDX_LICENSE_IDENTIFIER"]); identifier != "" {
		stem = identifier
	}
	return []string{
		filepath.Join(dir, "LICENSE"),
		filepath.Join(dir, "LICENSES", stem+".txt"),
	}
}
