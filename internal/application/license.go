// Package application composes the license render and verify use cases.
package application

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/t33n-software/license-hub/internal/domain/digest"
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

// RenderResult reports the rendered artifacts and the pinned template digest.
type RenderResult struct {
	Written []string
	Digest  string
}

// preparedRender carries the validated inputs of one render: the canonical
// content, the template bytes, and the resolved instance targets.
type preparedRender struct {
	content  string
	template []byte
	targets  []string
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
	return preparedRender{
		content:  content,
		template: template,
		targets:  instancePaths(req.OutDir, merged),
	}, nil
}

// Render renders the canonical template into the LICENSE and LICENSES/
// artifacts of the target directory.
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
	return RenderResult{Written: prepared.targets, Digest: digest.SHA256(prepared.template)}, nil
}

// PlanResult reports what a render would write, without writing it.
type PlanResult struct {
	Targets []string
	Digest  string
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
	return PlanResult{Targets: prepared.targets, Digest: digest.SHA256(prepared.template)}, nil
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
// render of the pinned template.
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
