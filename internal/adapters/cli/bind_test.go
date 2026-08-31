package cli

import (
	"flag"
	"io"
	"testing"

	"github.com/t33n-software/license-hub/internal/domain/contract"
)

// Convention: docs/conventions/cli/testing/README.md

func envStub(env map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := env[key]
		return value, ok
	}
}

func bindCommand(t *testing.T, commandName string, env map[string]string) *BoundFlags {
	t.Helper()
	command, ok := contract.CommandByName(commandName)
	if !ok {
		t.Fatalf("command %q not registered", commandName)
	}
	set := flag.NewFlagSet(commandName, flag.ContinueOnError)
	set.SetOutput(io.Discard)
	return Bind(set, append(command.Flags, contract.GlobalFlags()...), envStub(env))
}

func TestParseAppliesTheFlagValues(t *testing.T) {
	bound := bindCommand(t, "digest", nil)
	if record := bound.Parse([]string{"--template", "t.hbs"}); record != nil {
		t.Fatalf("Parse() record = %+v", record)
	}
	if got := bound.String("template"); got != "t.hbs" {
		t.Fatalf("String(template) = %q", got)
	}
	if got := bound.String("output"); got != contract.OutputHuman {
		t.Fatalf("String(output) = %q, want the default %q", got, contract.OutputHuman)
	}
	if bound.Bool("quiet") {
		t.Fatal("Bool(quiet) = true, want the default false")
	}
}

func TestParseReportsAFlagSyntaxErrorAsUsageError(t *testing.T) {
	bound := bindCommand(t, "digest", nil)
	record := bound.Parse([]string{"--bogus"})
	if record == nil {
		t.Fatal("Parse() expected a record for an unknown flag")
	}
	if record.Code != contract.ErrUsage {
		t.Fatalf("Parse() code = %q, want %q", record.Code, contract.ErrUsage)
	}
	if record.Remediation == "" {
		t.Fatal("Parse() record carries no remediation")
	}
}

func TestEnvironmentPrecedenceIsFlagThenEnvironmentThenDefault(t *testing.T) {
	bound := bindCommand(t, "digest", map[string]string{"LICENSE_TEMPLATE": "env.hbs"})
	if record := bound.Parse(nil); record != nil {
		t.Fatalf("Parse() record = %+v", record)
	}
	if got := bound.String("template"); got != "env.hbs" {
		t.Fatalf("String(template) = %q, want the environment value", got)
	}

	bound = bindCommand(t, "digest", map[string]string{"LICENSE_TEMPLATE": "env.hbs"})
	if record := bound.Parse([]string{"--template", "flag.hbs"}); record != nil {
		t.Fatalf("Parse() record = %+v", record)
	}
	if got := bound.String("template"); got != "flag.hbs" {
		t.Fatalf("String(template) = %q, want the flag to beat the environment", got)
	}

	bound = bindCommand(t, "render", nil)
	args := []string{"--template", "t.hbs", "--org-defaults", "o.json", "--values", "v.json", "--yes"}
	if record := bound.Parse(args); record != nil {
		t.Fatalf("Parse() record = %+v", record)
	}
	if got := bound.String("out"); got != "." {
		t.Fatalf("String(out) = %q, want the declared default", got)
	}
}

func TestEnvironmentBooleanParsing(t *testing.T) {
	bound := bindCommand(t, "digest", map[string]string{"LICENSE_QUIET": "true"})
	if record := bound.Parse([]string{"--template", "t.hbs"}); record != nil {
		t.Fatalf("Parse() record = %+v", record)
	}
	if !bound.Bool("quiet") {
		t.Fatal("Bool(quiet) = false, want the environment value true")
	}

	bound = bindCommand(t, "digest", map[string]string{"LICENSE_QUIET": "maybe"})
	record := bound.Parse([]string{"--template", "t.hbs"})
	if record == nil {
		t.Fatal("Parse() expected a record for a non-boolean environment value")
	}
	if record.Code != contract.ErrValueInvalid || record.Field != "quiet" {
		t.Fatalf("Parse() record = %+v", record)
	}
}

func TestMissingRequiredFlagFailsClosed(t *testing.T) {
	bound := bindCommand(t, "digest", nil)
	record := bound.Parse(nil)
	if record == nil {
		t.Fatal("Parse() expected a record for the missing required flag")
	}
	if record.Code != contract.ErrValueMissing || record.Field != "template" {
		t.Fatalf("Parse() record = %+v", record)
	}
	if record.Example == "" || record.Remediation == "" {
		t.Fatalf("Parse() record misses example or remediation: %+v", record)
	}
}

func TestClosedEnumValidationAcceptsAndRejects(t *testing.T) {
	bound := bindCommand(t, "digest", nil)
	record := bound.Parse([]string{"--template", "t.hbs", "--output", "yaml"})
	if record == nil {
		t.Fatal("Parse() expected a record for an undocumented output mode")
	}
	if record.Code != contract.ErrValueInvalid || record.Expected != "human or json" {
		t.Fatalf("Parse() record = %+v", record)
	}

	bound = bindCommand(t, "digest", nil)
	if record := bound.Parse([]string{"--template", "t.hbs", "--output", "json"}); record != nil {
		t.Fatalf("Parse() record = %+v", record)
	}
	if got := bound.String("output"); got != contract.OutputJSON {
		t.Fatalf("String(output) = %q", got)
	}
}
