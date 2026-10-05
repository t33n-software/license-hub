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
	content  string
	template []byte
	targets  []string
	npm      *ecosystemAlignment
	python   *ecosystemAlignment
	cargo    *ecosystemAlignment
	maven    *ecosystemAlignment
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
// lock-projection alignment function.
type ecosystemRenderRow struct {
	language      string
	manifestName  string
	manifestLabel string
	align         func(content string, merged map[string]string) (string, bool, error)
}

var npmRenderRow = ecosystemRenderRow{
	language:      ecosystem.NpmLanguage,
	manifestName:  ecosystem.NpmManifestName,
	manifestLabel: "npm manifest",
	align: func(content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignNpmLicense(content, ecosystem.NpmProjection(merged))
	},
}

var pythonRenderRow = ecosystemRenderRow{
	language:      ecosystem.PythonLanguage,
	manifestName:  ecosystem.PythonManifestName,
	manifestLabel: "python manifest",
	align: func(content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignPythonLicense(content, ecosystem.PythonProjection(merged), ecosystem.PythonLicenseFilesTarget)
	},
}

var cargoRenderRow = ecosystemRenderRow{
	language:      ecosystem.CargoLanguage,
	manifestName:  ecosystem.CargoManifestName,
	manifestLabel: "cargo manifest",
	align: func(content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignCargoLicense(content, ecosystem.CargoProjection(merged))
	},
}

var mavenRenderRow = ecosystemRenderRow{
	language:      ecosystem.MavenLanguage,
	manifestName:  ecosystem.MavenManifestName,
	manifestLabel: "maven manifest",
	align: func(content string, merged map[string]string) (string, bool, error) {
		return ecosystem.AlignMavenLicense(content, ecosystem.MavenProjection(merged))
	},
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
	npm, err := s.prepareEcosystemAlignment(req, merged, npmRenderRow)
	if err != nil {
		return preparedRender{}, err
	}
	python, err := s.prepareEcosystemAlignment(req, merged, pythonRenderRow)
	if err != nil {
		return preparedRender{}, err
	}
	cargo, err := s.prepareEcosystemAlignment(req, merged, cargoRenderRow)
	if err != nil {
		return preparedRender{}, err
	}
	maven, err := s.prepareEcosystemAlignment(req, merged, mavenRenderRow)
	if err != nil {
		return preparedRender{}, err
	}
	return preparedRender{
		content:  content,
		template: template,
		targets:  instancePaths(req.OutDir, merged),
		npm:      npm,
		python:   python,
		cargo:    cargo,
		maven:    maven,
	}, nil
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
	for _, alignment := range []*ecosystemAlignment{prepared.npm, prepared.python, prepared.cargo, prepared.maven} {
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
	for _, alignment := range []*ecosystemAlignment{prepared.npm, prepared.python, prepared.cargo, prepared.maven} {
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
	for _, row := range []ecosystemVerifyRow{npmVerifyRow, pythonVerifyRow, cargoVerifyRow, mavenVerifyRow} {
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

// prepareEcosystemAlignment derives the ecosystem license-field alignment of
// one render from the declaration seam and the merged tenant values. A
// missing seam or a foreign declaration leaves the surface untouched; a
// declared project without a manifest is reported as a skip that the verify
// lane carries as a fail-closed finding.
//
// Convention: spec/ecosystem-license-metadata.md
func (s *LicenseService) prepareEcosystemAlignment(req RenderRequest, merged map[string]string, row ecosystemRenderRow) (*ecosystemAlignment, error) {
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
	manifestPath := filepath.Join(req.OutDir, row.manifestName)
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
	aligned, changed, err := row.align(string(data), merged)
	if err != nil {
		return nil, fmt.Errorf("%w: align %s %s: %v", ErrInvalidEcosystemSurface, row.manifestLabel, manifestPath, err)
	}
	return &ecosystemAlignment{manifestPath: manifestPath, content: aligned, changed: changed}, nil
}

// ecosystemVerifyRow binds one ecosystem matrix row to its verify-side
// finding derivation.
type ecosystemVerifyRow struct {
	language     string
	manifestName string
	label        string
	findings     func(content, manifestPath string, merged map[string]string) []string
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

// verifyEcosystem proves one declared ecosystem surface fail-closed in both
// directions: a declared manifest that is missing or diverging is a finding,
// a manifest without a covering declaration is a shadow-surface finding, and
// an unscannable surface is a proof failure.
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
	return row.findings(string(data), manifestPath, merged), nil
}

// npmFindings derives the npm license-surface findings from the inspected
// manifest and the lock projection.
func npmFindings(content, manifestPath string, merged map[string]string) []string {
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
func pythonFindings(content, manifestPath string, merged map[string]string) []string {
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
func cargoFindings(content, manifestPath string, merged map[string]string) []string {
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
func mavenFindings(content, manifestPath string, merged map[string]string) []string {
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
