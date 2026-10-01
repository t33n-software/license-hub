package cli

import (
	"strings"
	"testing"

	"github.com/t33n-software/license-hub/internal/domain/contract"
)

// Convention: docs/conventions/cli/testing/README.md

func TestWriteRootHelpRendersTheFullRegistry(t *testing.T) {
	var out strings.Builder
	WriteRootHelp(&out, contract.Commands(), contract.GlobalFlags())
	help := out.String()

	for _, want := range []string{
		contract.BinaryName + " — " + contract.ToolSummary,
		"license <command> [flags]",
		"license --version [--output json]",
		"Discovery: every command owns a complete help area",
		"Exit codes: 0 success; 1 execution failure; 2 usage error; 3 governance rejection.",
		"Every command works offline.",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("root help misses %q:\n%s", want, help)
		}
	}
	for _, command := range contract.Commands() {
		if !strings.Contains(help, command.Name) || !strings.Contains(help, command.Summary) {
			t.Fatalf("root help misses the command %q:\n%s", command.Name, help)
		}
	}
	for _, flag := range contract.GlobalFlags() {
		if !strings.Contains(help, "--"+flag.Name) || !strings.Contains(help, flag.Env) {
			t.Fatalf("root help misses the global flag %q:\n%s", flag.Name, help)
		}
	}
}

func TestWriteRootHelpOmitsTheOfflineClaimWhenACommandNeedsNetwork(t *testing.T) {
	commands := []contract.Command{{
		Name:      "sync",
		Summary:   "synchronize with a remote",
		Offline:   false,
		Stability: contract.StabilityStable,
	}}
	var out strings.Builder
	WriteRootHelp(&out, commands, contract.GlobalFlags())
	if strings.Contains(out.String(), "Every command works offline.") {
		t.Fatalf("root help must not claim offline capability:\n%s", out.String())
	}
}

func TestWriteLeafHelpRendersTheFullInputContractOfEveryCommand(t *testing.T) {
	for _, command := range contract.Commands() {
		var out strings.Builder
		WriteLeafHelp(&out, command)
		help := out.String()

		for _, want := range []string{
			contract.BinaryName + " " + command.Name + " — " + command.Summary,
			"Usage:",
			"Global flags:",
			"Examples:",
			"Exit codes: 0 success; 1 execution failure; 2 usage error; 3 governance rejection.",
			"This command works offline.",
			"Stability: stable",
		} {
			if !strings.Contains(help, want) {
				t.Fatalf("leaf help of %q misses %q:\n%s", command.Name, want, help)
			}
		}
		for _, flag := range command.Flags {
			for _, want := range []string{"--" + flag.Name, flag.Env, flag.Example} {
				if !strings.Contains(help, want) {
					t.Fatalf("leaf help of %q misses %q of flag %q:\n%s", command.Name, want, flag.Name, help)
				}
			}
			if flag.Class == contract.ClassStructuralReference {
				if !strings.Contains(help, flag.Form) || !strings.Contains(help, flag.Resolution) {
					t.Fatalf("leaf help of %q misses form or resolution of %q:\n%s", command.Name, flag.Name, help)
				}
			}
			if flag.Required && !strings.Contains(help, "required") {
				t.Fatalf("leaf help of %q does not mark %q as required:\n%s", command.Name, flag.Name, help)
			}
		}
		if command.Mutating {
			if !strings.Contains(help, "mutates state") {
				t.Fatalf("leaf help of %q misses the mutation note:\n%s", command.Name, help)
			}
		} else if !strings.Contains(help, "This command is read-only.") {
			t.Fatalf("leaf help of %q misses the read-only note:\n%s", command.Name, help)
		}
	}
}

func TestExamplesDeriveTheCanonicalCallsFromTheRegistry(t *testing.T) {
	render, _ := contract.CommandByName("render")
	got := examples(render)
	want := []string{
		"license render --template templates/custom/norepublish/NoRepublish-1.0.0.hbs --org-defaults org-defaults.json --values license.values.json --yes",
		"license render --template templates/custom/norepublish/NoRepublish-1.0.0.hbs --org-defaults org-defaults.json --values license.values.json --dry-run",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("examples(render) = %v, want %v", got, want)
	}
	version, _ := contract.CommandByName("version")
	if got := examples(version); len(got) != 1 || got[0] != "license version" {
		t.Fatalf("examples(version) = %v", got)
	}
}

func TestWriteLeafHelpCoversTheSyntheticBoundaryCases(t *testing.T) {
	command := contract.Command{
		Name:      "sync",
		Summary:   "synchronize with a remote",
		Mutating:  false,
		Offline:   false,
		Stability: contract.StabilityExperimental,
		Flags:     nil,
	}
	var out strings.Builder
	WriteLeafHelp(&out, command)
	help := out.String()
	for _, want := range []string{
		"license sync [flags]",
		"This command requires network access.",
		"This command is read-only.",
		"Stability: experimental",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("leaf help of the synthetic command misses %q:\n%s", want, help)
		}
	}
	if strings.Contains(help, "Flags:\n") {
		t.Fatalf("leaf help of a flagless command must not render a flags section:\n%s", help)
	}
}

func TestFlagTextRendersEveryValueClass(t *testing.T) {
	base := contract.FlagDomain{
		Name:      "probe",
		TypeName:  "string",
		Summary:   "probe summary",
		Example:   "probe-example",
		Env:       "LICENSE_PROBE",
		Stability: contract.StabilityStable,
	}
	cases := []struct {
		mutate func(*contract.FlagDomain)
		want   []string
	}{
		{func(f *contract.FlagDomain) { f.Class = contract.ClassClosedEnum; f.Values = []string{"a", "b"} }, []string{"probe summary", "a or b"}},
		{func(f *contract.FlagDomain) { f.Class = contract.ClassShaped; f.Grammar = "release/<semver>" }, []string{"grammar: release/<semver>"}},
		{func(f *contract.FlagDomain) { f.Class = contract.ClassFreeConstrained; f.Rules = "1-100 lowercase" }, []string{"1-100 lowercase"}},
		{func(f *contract.FlagDomain) {
			f.Class = contract.ClassStructuralReference
			f.Form = "a path"
			f.Resolution = "resolved at runtime"
		}, []string{"form: a path", "resolved at runtime"}},
		{func(f *contract.FlagDomain) { f.Class = contract.ClassScalarBounded; f.Range = "positive duration" }, []string{"range: positive duration"}},
		{func(f *contract.FlagDomain) { f.Class = contract.ClassBooleanSwitch; f.Effect = "enables the probe" }, []string{"enables the probe"}},
		{func(f *contract.FlagDomain) { f.Class = contract.ClassCompositeToken; f.Grammar = "TOKEN=VALUE" }, []string{"transport: TOKEN=VALUE"}},
		{func(f *contract.FlagDomain) { f.Class = contract.ClassSecretReference; f.Form = "environment variable" }, []string{"reference forms: environment variable"}},
		{func(f *contract.FlagDomain) {
			f.Class = contract.ClassStructuralReference
			f.Form = "a path"
			f.Resolution = "resolved at runtime"
			f.Required = true
			f.Default = "."
			f.Stability = contract.StabilityExperimental
		}, []string{"required", "(default \".\")", "stability: experimental"}},
		{func(f *contract.FlagDomain) {
			f.Class = contract.ClassBooleanSwitch
			f.Effect = "enables the probe"
			f.Default = "false"
		}, []string{"(default false)"}},
	}
	for i, tc := range cases {
		flag := base
		tc.mutate(&flag)
		got := flagText(flag)
		for _, want := range tc.want {
			if !strings.Contains(got, want) {
				t.Fatalf("case %d: flagText() = %q, want substring %q", i, got, want)
			}
		}
		if !strings.Contains(got, "env: LICENSE_PROBE") || !strings.Contains(got, "example: probe-example") {
			t.Fatalf("case %d: flagText() misses env or example: %q", i, got)
		}
	}
}

func TestFlagHeadingDistinguishesSwitchesFromValueFlags(t *testing.T) {
	value := contract.FlagDomain{Name: "template", Class: contract.ClassStructuralReference, TypeName: "string"}
	if got := flagHeading(value); got != "--template string" {
		t.Fatalf("flagHeading() = %q", got)
	}
	boolean := contract.FlagDomain{Name: "quiet", Class: contract.ClassBooleanSwitch}
	if got := flagHeading(boolean); got != "--quiet" {
		t.Fatalf("flagHeading() = %q", got)
	}
}

func TestOrJoinRendersTheEnumeration(t *testing.T) {
	if got := OrJoin(nil); got != "" {
		t.Fatalf("OrJoin(nil) = %q", got)
	}
	if got := OrJoin([]string{"a"}); got != "a" {
		t.Fatalf("OrJoin([a]) = %q", got)
	}
	if got := OrJoin([]string{"a", "b"}); got != "a or b" {
		t.Fatalf("OrJoin([a b]) = %q", got)
	}
	if got := OrJoin([]string{"a", "b", "c"}); got != "a, b, or c" {
		t.Fatalf("OrJoin([a b c]) = %q", got)
	}
}
