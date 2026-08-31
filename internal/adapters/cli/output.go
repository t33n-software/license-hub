package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/t33n-software/license-hub/internal/domain/contract"
)

// Field is one ordered key-value fact of a command result. The human and the
// machine form render content-identically from it.
//
// Convention: docs/conventions/cli/output/README.md
type Field struct {
	Key    string   // machine key of the JSON form
	Label  string   // human label of the human form
	Values []string // one or more values; a single value renders as a scalar
}

// Result is one command outcome.
type Result struct {
	Command string
	Status  string
	Message string
	Fields  []Field
}

// ErrorRecord is one structured, coded error of the error contract.
//
// Convention: docs/conventions/cli/output/README.md
type ErrorRecord struct {
	Code        contract.ErrorCode
	Field       string
	Actual      string
	Expected    string
	Rule        string
	Example     string
	Remediation string
}

// Renderer renders results and errors into the human or the machine channel.
// In the machine mode every document goes to the primary output; in the
// human mode successes go to the primary output and failures to the error
// output. The quiet mode suppresses successful human output only; errors are
// never suppressed.
//
// Convention: docs/conventions/cli/output/README.md
type Renderer struct {
	out    io.Writer
	errOut io.Writer
	mode   string
	quiet  bool
}

// NewRenderer binds the renderer to the output channels and the output mode.
func NewRenderer(out, errOut io.Writer, mode string, quiet bool) Renderer {
	return Renderer{out: out, errOut: errOut, mode: mode, quiet: quiet}
}

// WriteResult writes a success or plan result.
func (r Renderer) WriteResult(result Result) {
	if r.mode == contract.OutputJSON {
		writeJSON(r.out, resultDocument(result))
		return
	}
	if r.quiet {
		return
	}
	writeHumanResult(r.out, result)
}

// WriteGovernance writes a governance rejection.
func (r Renderer) WriteGovernance(result Result) {
	if r.mode == contract.OutputJSON {
		writeJSON(r.out, resultDocument(result))
		return
	}
	writeHumanResult(r.errOut, result)
}

// WriteError writes a structured error record.
func (r Renderer) WriteError(record ErrorRecord) {
	if r.mode == contract.OutputJSON {
		writeJSON(r.out, errorDocument(record))
		return
	}
	writeHumanError(r.errOut, record)
}

// writeHumanResult renders the fields and the message of a result.
func writeHumanResult(w io.Writer, result Result) {
	for _, field := range result.Fields {
		for _, value := range field.Values {
			fmt.Fprintf(w, "%s %s\n", field.Label, value)
		}
	}
	if result.Message != "" {
		fmt.Fprintln(w, result.Message)
	}
}

// writeHumanError renders the structured error record in the human form.
func writeHumanError(w io.Writer, record ErrorRecord) {
	fmt.Fprintf(w, "Error [%s]\n", record.Code)
	fmt.Fprintf(w, "Field: %s\n", orDash(record.Field))
	fmt.Fprintf(w, "Actual value: %s\n", orDash(record.Actual))
	fmt.Fprintf(w, "Expected: %s\n", orDash(record.Expected))
	fmt.Fprintf(w, "Rule: %s\n", orDash(record.Rule))
	fmt.Fprintf(w, "Valid example: %s\n", orDash(record.Example))
	fmt.Fprintf(w, "How to fix it: %s\n", orDash(record.Remediation))
}

// resultDocument builds the machine-readable document of a result.
func resultDocument(result Result) map[string]any {
	doc := map[string]any{
		"schemaVersion": contract.SchemaVersion,
		"command":       result.Command,
		"status":        result.Status,
	}
	if result.Message != "" {
		doc["message"] = result.Message
	}
	for _, field := range result.Fields {
		if len(field.Values) == 1 {
			doc[field.Key] = field.Values[0]
		} else {
			doc[field.Key] = field.Values
		}
	}
	return doc
}

// errorDocument builds the machine-readable document of an error record.
func errorDocument(record ErrorRecord) map[string]any {
	return map[string]any{
		"schemaVersion": contract.SchemaVersion,
		"status":        "error",
		"error": map[string]any{
			"code":        string(record.Code),
			"field":       record.Field,
			"actual":      record.Actual,
			"expected":    record.Expected,
			"rule":        record.Rule,
			"example":     record.Example,
			"remediation": record.Remediation,
		},
	}
}

// writeJSON writes one document as one JSON object.
func writeJSON(w io.Writer, doc map[string]any) {
	// The document carries only strings, integers, and string slices, so
	// marshalling cannot fail; the error return is dropped deliberately.
	encoded, _ := json.Marshal(doc)
	fmt.Fprintf(w, "%s\n", encoded)
}

// orDash renders an empty optional field as a dash.
func orDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
