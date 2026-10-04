// Package application composes the license render and verify use cases.
package application

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
// prepared ecosystem alignment.
type preparedRender struct {
	content  string
	template []byte
	targets  []string
	npm      *npmAlignment
}

// npmAlignment carries the prepared ecosystem license-field alignment of one
// render: the aligned manifest content and whether bytes change, or the
// reason the declared surface cannot be aligned.
type npmAlignment struct {
	manifestPath string
	content      string
	changed      bool
	skip         string
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
	npm, err := s.prepareNpmAlignment(req, merged)
	if err != nil {
		return preparedRender{}, err
	}
	return preparedRender{
		content:  content,
		template: template,
		targets:  instancePaths(req.OutDir, merged),
		npm:      npm,
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
	if prepared.npm != nil && prepared.npm.changed {
		if err := s.fs.WriteFile(prepared.npm.manifestPath, []byte(prepared.npm.content)); err != nil {
			return RenderResult{}, fmt.Errorf("write %s: %w", prepared.npm.manifestPath, err)
		}
		result.Aligned = append(result.Aligned, prepared.npm.manifestPath)
	}
	if prepared.npm != nil && prepared.npm.skip != "" {
		result.Skipped = append(result.Skipped, prepared.npm.skip)
	}
	return result, nil
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
	if prepared.npm != nil && prepared.npm.changed {
		plan.Alignments = append(plan.Alignments, "align "+prepared.npm.manifestPath+" to the lock projection")
	}
	if prepared.npm != nil && prepared.npm.skip != "" {
		plan.Skips = append(plan.Skips, prepared.npm.skip)
	}
	return plan, nil
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
	npmViolations, err := s.verifyNpm(req, merged)
	if err != nil {
		return nil, err
	}
	violations = append(violations, npmViolations...)
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

// prepareNpmAlignment derives the ecosystem license-field alignment of one
// render from the declaration seam and the merged tenant values. A missing
// seam or a non-npm declaration leaves the surface untouched; a declared npm
// project without a manifest is reported as a skip that the verify lane
// carries as a fail-closed finding.
//
// Convention: spec/ecosystem-license-metadata.md
func (s *LicenseService) prepareNpmAlignment(req RenderRequest, merged map[string]string) (*npmAlignment, error) {
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
	if language != ecosystem.NpmLanguage {
		return nil, nil
	}
	manifestPath := filepath.Join(req.OutDir, ecosystem.NpmManifestName)
	data, err := s.fs.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &npmAlignment{
				manifestPath: manifestPath,
				skip:         "npm manifest is absent: the declared license surface cannot be aligned",
			}, nil
		}
		return nil, fmt.Errorf("read npm manifest %s: %w", manifestPath, err)
	}
	aligned, changed, err := ecosystem.AlignNpmLicense(string(data), ecosystem.NpmProjection(merged))
	if err != nil {
		return nil, fmt.Errorf("%w: align npm manifest %s: %v", ErrInvalidEcosystemSurface, manifestPath, err)
	}
	return &npmAlignment{manifestPath: manifestPath, content: aligned, changed: changed}, nil
}

// verifyNpm proves the declared npm license surface fail-closed in both
// directions: a declared manifest that is missing or diverging is a finding,
// a manifest without a covering declaration is a shadow-surface finding, and
// an unscannable surface is a proof failure.
//
// Convention: spec/ecosystem-license-metadata.md
func (s *LicenseService) verifyNpm(req VerifyRequest, merged map[string]string) ([]string, error) {
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
	manifestPath := filepath.Join(req.Dir, ecosystem.NpmManifestName)
	data, manifestErr := s.fs.ReadFile(manifestPath)
	switch {
	case errors.Is(manifestErr, os.ErrNotExist):
		if language == ecosystem.NpmLanguage {
			return []string{
				"declared npm ecosystem manifest missing: " + manifestPath +
					" (declared language: \"" + ecosystem.NpmLanguage + "\")",
			}, nil
		}
		return nil, nil
	case manifestErr != nil:
		return nil, fmt.Errorf("read npm manifest %s: %w", manifestPath, manifestErr)
	}
	if language != ecosystem.NpmLanguage {
		detail := "no ecosystem declaration found (" + ecosystem.SeamFileName + " is absent)"
		if seamErr == nil {
			detail = "declared language: \"" + language + "\""
		}
		return []string{
			"npm license metadata surface without covering declaration: " + manifestPath + " (" + detail + ")",
		}, nil
	}
	target := ecosystem.NpmProjection(merged)
	value, state, err := ecosystem.InspectNpmLicense(string(data))
	if err != nil {
		return []string{"npm license surface cannot be proven: " + manifestPath + " (" + err.Error() + ")"}, nil
	}
	switch state {
	case ecosystem.LicenseFieldMissing:
		return []string{"npm license field is missing: " + manifestPath + " (expected \"" + target + "\")"}, nil
	case ecosystem.LicenseFieldNonString:
		return []string{"npm license field is not a JSON string: " + manifestPath + " (expected \"" + target + "\")"}, nil
	}
	if value != target {
		return []string{
			"npm license field diverges from the lock projection: observed \"" + value +
				"\", expected \"" + target + "\" (run the render to align)",
		}, nil
	}
	return nil, nil
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