package contract

import (
	"reflect"
	"regexp"
	"slices"
	"testing"
)

// Convention: docs/conventions/cli/testing/README.md

var (
	flagNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	envNamePattern  = regexp.MustCompile(`^LICENSE_[A-Z0-9_]+$`)
)

func everyFlag() []FlagDomain {
	flags := slices.Clone(GlobalFlags())
	for _, command := range Commands() {
		flags = append(flags, command.Flags...)
	}
	return flags
}

func TestCommandsCoverExactlyTheContractSurface(t *testing.T) {
	commands := Commands()
	want := []string{"render", "verify", "digest", "version"}
	if len(commands) != len(want) {
		t.Fatalf("Commands() = %v, want %v", commands, want)
	}
	for i, command := range commands {
		if command.Name != want[i] {
			t.Fatalf("Commands()[%d] = %q, want %q", i, command.Name, want[i])
		}
	}
}

func TestEveryCommandDeclaresItsGovernanceMetadata(t *testing.T) {
	for _, command := range Commands() {
		if command.Summary == "" {
			t.Fatalf("command %q carries no one-line purpose", command.Name)
		}
		if command.Stability != StabilityStable {
			t.Fatalf("command %q stability = %q, want %q", command.Name, command.Stability, StabilityStable)
		}
		if !command.Offline {
			t.Fatalf("command %q must declare its offline capability", command.Name)
		}
	}
}

func TestMutatingSurfaceIsExplicit(t *testing.T) {
	mutating := map[string]bool{}
	for _, command := range Commands() {
		mutating[command.Name] = command.Mutating
	}
	if !mutating["render"] {
		t.Fatal("render must be declared mutating")
	}
	for _, name := range []string{"verify", "digest", "version"} {
		if mutating[name] {
			t.Fatalf("%s must stay read-only", name)
		}
	}
}

func TestEveryFlagCarriesExactlyOneClassAndTheNamingConventions(t *testing.T) {
	valid := map[ValueClass]bool{
		ClassClosedEnum:          true,
		ClassShaped:              true,
		ClassFreeConstrained:     true,
		ClassStructuralReference: true,
		ClassScalarBounded:       true,
		ClassBooleanSwitch:       true,
		ClassCompositeToken:      true,
		ClassSecretReference:     true,
	}
	for _, flag := range everyFlag() {
		if !flagNamePattern.MatchString(flag.Name) {
			t.Fatalf("flag name %q violates the lowercase kebab-case convention", flag.Name)
		}
		if !valid[flag.Class] {
			t.Fatalf("flag %q carries the unknown class %q", flag.Name, flag.Class)
		}
		if !envNamePattern.MatchString(flag.Env) {
			t.Fatalf("flag %q env %q violates the LICENSE_* naming convention", flag.Name, flag.Env)
		}
		if flag.Summary == "" {
			t.Fatalf("flag %q carries no one-line purpose", flag.Name)
		}
		if flag.Example == "" {
			t.Fatalf("flag %q carries no canonical example", flag.Name)
		}
		if flag.Stability == "" {
			t.Fatalf("flag %q carries no stability level", flag.Name)
		}
	}
}

func TestClassSpecificHelpDutiesAreFulfilled(t *testing.T) {
	for _, flag := range everyFlag() {
		switch flag.Class {
		case ClassClosedEnum:
			if len(flag.Values) == 0 {
				t.Fatalf("closed-enum flag %q carries no value list", flag.Name)
			}
			if flag.Default == "" || !slices.Contains(flag.Values, flag.Default) {
				t.Fatalf("closed-enum flag %q default %q is not in the accepted set %v", flag.Name, flag.Default, flag.Values)
			}
		case ClassShaped:
			if flag.Grammar == "" {
				t.Fatalf("shaped flag %q carries no grammar", flag.Name)
			}
		case ClassFreeConstrained:
			if flag.Rules == "" {
				t.Fatalf("free-constrained flag %q carries no rule set", flag.Name)
			}
		case ClassStructuralReference:
			if flag.Form == "" || flag.Resolution == "" {
				t.Fatalf("structural-reference flag %q carries no form or resolution rule", flag.Name)
			}
		case ClassScalarBounded:
			if flag.Range == "" {
				t.Fatalf("scalar-bounded flag %q carries no value range", flag.Name)
			}
		case ClassBooleanSwitch:
			if flag.Effect == "" {
				t.Fatalf("boolean-switch flag %q carries no effect", flag.Name)
			}
			if flag.Default != "false" && flag.Default != "true" {
				t.Fatalf("boolean-switch flag %q default = %q, want a boolean default", flag.Name, flag.Default)
			}
		case ClassCompositeToken, ClassSecretReference:
			// No command of this tool uses these classes yet; the duties are
			// enforced here as soon as one does.
		}
	}
}

func TestSharedFlagDomainsAreIdenticalAcrossCommands(t *testing.T) {
	render, _ := CommandByName("render")
	verify, _ := CommandByName("verify")
	digest, _ := CommandByName("digest")
	if !reflect.DeepEqual(render.Flags[0], verify.Flags[0]) || !reflect.DeepEqual(verify.Flags[0], digest.Flags[0]) {
		t.Fatal("the --template domain drifts between commands")
	}
	if !reflect.DeepEqual(render.Flags[1], verify.Flags[1]) || !reflect.DeepEqual(render.Flags[2], verify.Flags[2]) {
		t.Fatal("the --org-defaults or --values domain drifts between commands")
	}
}

func TestRequiredFlagsMatchTheUseCaseContracts(t *testing.T) {
	required := func(command string) []string {
		entry, _ := CommandByName(command)
		out := make([]string, 0)
		for _, flag := range entry.Flags {
			if flag.Required {
				out = append(out, flag.Name)
			}
		}
		return out
	}
	for command, want := range map[string][]string{
		"render": {"template", "org-defaults", "values"},
		"verify": {"template", "org-defaults", "values"},
		"digest": {"template"},
	} {
		if got := required(command); !slices.Equal(got, want) {
			t.Fatalf("required(%s) = %v, want %v", command, got, want)
		}
	}
}

func TestMutatingCommandCarriesTheInteractionFlags(t *testing.T) {
	render, _ := CommandByName("render")
	names := make([]string, 0, len(render.Flags))
	for _, flag := range render.Flags {
		names = append(names, flag.Name)
	}
	for _, want := range []string{"dry-run", "yes"} {
		if !slices.Contains(names, want) {
			t.Fatalf("render misses the interaction flag %q", want)
		}
	}
}

func TestOutputFlagDomainIsTheClosedEnumOfTheOutputContract(t *testing.T) {
	output, ok := GlobalFlagByName("output")
	if !ok {
		t.Fatal("the global --output flag is not registered")
	}
	if output.Class != ClassClosedEnum {
		t.Fatalf("--output class = %q, want %q", output.Class, ClassClosedEnum)
	}
	if !slices.Equal(output.Values, []string{OutputHuman, OutputJSON}) {
		t.Fatalf("--output values = %v", output.Values)
	}
	if output.Default != OutputHuman {
		t.Fatalf("--output default = %q, want %q", output.Default, OutputHuman)
	}
}

func TestCommandByNameResolvesAndRejects(t *testing.T) {
	if _, ok := CommandByName("render"); !ok {
		t.Fatal("CommandByName(render) not found")
	}
	if _, ok := CommandByName("frobnicate"); ok {
		t.Fatal("CommandByName(frobnicate) must not resolve")
	}
}

func TestGlobalFlagByNameResolvesAndRejects(t *testing.T) {
	if _, ok := GlobalFlagByName("quiet"); !ok {
		t.Fatal("GlobalFlagByName(quiet) not found")
	}
	if _, ok := GlobalFlagByName("frobnicate"); ok {
		t.Fatal("GlobalFlagByName(frobnicate) must not resolve")
	}
}
