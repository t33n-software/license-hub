package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/t33n-software/license-hub/internal/domain/contract"
)

// Convention: docs/conventions/cli/testing/README.md

func renderResult() Result {
	return Result{
		Command: "render",
		Status:  "ok",
		Fields: []Field{
			{Key: "written", Label: "wrote", Values: []string{"LICENSE", "LICENSES/LicenseRef-x-1.0.txt"}},
			{Key: "digest", Label: "template digest", Values: []string{"sha256:abc"}},
		},
	}
}

func TestWriteResultHumanRendersFieldsAndMessage(t *testing.T) {
	var out, errOut strings.Builder
	r := NewRenderer(&out, &errOut, contract.OutputHuman, false)
	r.WriteResult(renderResult())
	want := "wrote LICENSE\nwrote LICENSES/LicenseRef-x-1.0.txt\ntemplate digest sha256:abc\n"
	if out.String() != want {
		t.Fatalf("WriteResult() = %q, want %q", out.String(), want)
	}
	if errOut.String() != "" {
		t.Fatalf("WriteResult() wrote to the error channel: %q", errOut.String())
	}

	out.Reset()
	r.WriteResult(Result{Command: "verify", Status: "ok", Message: "license instance matches the canonical render"})
	if out.String() != "license instance matches the canonical render\n" {
		t.Fatalf("WriteResult() message = %q", out.String())
	}
}

func TestWriteResultQuietSuppressesOnlyTheHumanSuccess(t *testing.T) {
	var out strings.Builder
	r := NewRenderer(&out, &strings.Builder{}, contract.OutputHuman, true)
	r.WriteResult(renderResult())
	if out.String() != "" {
		t.Fatalf("WriteResult() quiet = %q, want empty", out.String())
	}

	out.Reset()
	r = NewRenderer(&out, &strings.Builder{}, contract.OutputJSON, true)
	r.WriteResult(renderResult())
	if out.String() == "" {
		t.Fatal("WriteResult() must never suppress the machine output")
	}
}

func TestWriteResultJSONRendersTheVersionedDocument(t *testing.T) {
	var out strings.Builder
	r := NewRenderer(&out, &strings.Builder{}, contract.OutputJSON, false)
	r.WriteResult(renderResult())
	var doc map[string]any
	if err := json.Unmarshal([]byte(out.String()), &doc); err != nil {
		t.Fatalf("WriteResult() JSON does not parse: %v", err)
	}
	if doc["schemaVersion"] != float64(contract.SchemaVersion) {
		t.Fatalf("schemaVersion = %v", doc["schemaVersion"])
	}
	if doc["command"] != "render" || doc["status"] != "ok" {
		t.Fatalf("document = %v", doc)
	}
	if written, ok := doc["written"].([]any); !ok || len(written) != 2 {
		t.Fatalf("written = %v, want a two-element array", doc["written"])
	}
	if doc["digest"] != "sha256:abc" {
		t.Fatalf("digest = %v, want a scalar", doc["digest"])
	}
	if _, present := doc["message"]; present {
		t.Fatalf("message must be omitted when empty: %v", doc)
	}
}

func TestWriteResultJSONCarriesTheMessage(t *testing.T) {
	var out strings.Builder
	r := NewRenderer(&out, &strings.Builder{}, contract.OutputJSON, false)
	r.WriteResult(Result{Command: "verify", Status: "ok", Message: "license instance matches the canonical render"})
	var doc map[string]any
	if err := json.Unmarshal([]byte(out.String()), &doc); err != nil {
		t.Fatalf("WriteResult() JSON does not parse: %v", err)
	}
	if doc["message"] != "license instance matches the canonical render" {
		t.Fatalf("message = %v", doc["message"])
	}
}

func TestWriteGovernanceRoutesHumanToErrOutAndJSONToOut(t *testing.T) {
	violations := Result{
		Command: "verify",
		Status:  "governance_rejected",
		Fields:  []Field{{Key: "violations", Label: "violation:", Values: []string{"drifted"}}},
	}
	var out, errOut strings.Builder
	r := NewRenderer(&out, &errOut, contract.OutputHuman, false)
	r.WriteGovernance(violations)
	if out.String() != "" || errOut.String() != "violation: drifted\n" {
		t.Fatalf("WriteGovernance() human = %q / %q", out.String(), errOut.String())
	}

	out.Reset()
	errOut.Reset()
	r = NewRenderer(&out, &errOut, contract.OutputJSON, false)
	r.WriteGovernance(violations)
	if errOut.String() != "" {
		t.Fatalf("WriteGovernance() JSON must not write to the error channel: %q", errOut.String())
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(out.String()), &doc); err != nil || doc["status"] != "governance_rejected" {
		t.Fatalf("WriteGovernance() JSON = %q, err = %v", out.String(), err)
	}
}

func TestWriteErrorHumanRendersEveryField(t *testing.T) {
	record := ErrorRecord{
		Code:        contract.ErrValueMissing,
		Field:       "template",
		Expected:    "a path",
		Rule:        "the flag is required",
		Example:     "--template t.hbs",
		Remediation: "pass --template",
	}
	var out, errOut strings.Builder
	r := NewRenderer(&out, &errOut, contract.OutputHuman, false)
	r.WriteError(record)
	want := "Error [VALUE_MISSING]\nField: template\nActual value: -\nExpected: a path\nRule: the flag is required\n" +
		"Valid example: --template t.hbs\nHow to fix it: pass --template\n"
	if errOut.String() != want {
		t.Fatalf("WriteError() = %q, want %q", errOut.String(), want)
	}
	if out.String() != "" {
		t.Fatalf("WriteError() wrote to the success channel: %q", out.String())
	}
}

func TestWriteErrorJSONCarriesTheIdenticalInformation(t *testing.T) {
	record := ErrorRecord{
		Code:        contract.ErrValueInvalid,
		Field:       "output",
		Actual:      "yaml",
		Expected:    "human or json",
		Rule:        "the value must be one of the documented set",
		Example:     "--output json",
		Remediation: "see the help",
	}
	var out strings.Builder
	r := NewRenderer(&out, &strings.Builder{}, contract.OutputJSON, false)
	r.WriteError(record)
	var doc map[string]any
	if err := json.Unmarshal([]byte(out.String()), &doc); err != nil {
		t.Fatalf("WriteError() JSON does not parse: %v", err)
	}
	if doc["status"] != "error" {
		t.Fatalf("status = %v", doc["status"])
	}
	errDoc, ok := doc["error"].(map[string]any)
	if !ok {
		t.Fatalf("error document missing: %v", doc)
	}
	for key, want := range map[string]string{
		"code":        "VALUE_INVALID",
		"field":       "output",
		"actual":      "yaml",
		"expected":    "human or json",
		"rule":        "the value must be one of the documented set",
		"example":     "--output json",
		"remediation": "see the help",
	} {
		if errDoc[key] != want {
			t.Fatalf("error[%s] = %v, want %q", key, errDoc[key], want)
		}
	}
}

func TestOrDashRendersEmptyFieldsAsADash(t *testing.T) {
	if got := orDash(""); got != "-" {
		t.Fatalf("orDash(\"\") = %q", got)
	}
	if got := orDash("x"); got != "x" {
		t.Fatalf("orDash(x) = %q", got)
	}
}
