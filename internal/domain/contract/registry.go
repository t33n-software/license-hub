package contract

// This file is the canonical command and flag registry. Shared flag domains
// are declared exactly once as package-level variables and composed into the
// commands; endpoint subsets would be produced through declared filters on
// the shared domain, never through a hand copy.
//
// Convention: docs/conventions/cli/values/single-source-of-truth.md

// templateFlag is the canonical value domain of the --template flag, shared
// by every command that consumes a template path.
var templateFlag = FlagDomain{
	Name:       "template",
	Class:      ClassStructuralReference,
	TypeName:   "string",
	Summary:    "path to the canonical template",
	Form:       "repository-relative or absolute path to a <Name>-<semver>.hbs template file",
	Resolution: "existence and readability are resolved at runtime",
	Example:    "templates/custom/norepublish/NoRepublish-1.0.0.hbs",
	Env:        "LICENSE_TEMPLATE",
	Required:   true,
	Stability:  StabilityStable,
}

// orgDefaultsFlag is the canonical value domain of the --org-defaults flag.
var orgDefaultsFlag = FlagDomain{
	Name:       "org-defaults",
	Class:      ClassStructuralReference,
	TypeName:   "string",
	Summary:    "path to the organization defaults JSON",
	Form:       "path to a JSON object with string values",
	Resolution: "existence and readability are resolved at runtime",
	Example:    "org-defaults.json",
	Env:        "LICENSE_ORG_DEFAULTS",
	Required:   true,
	Stability:  StabilityStable,
}

// valuesFlag is the canonical value domain of the --values flag.
var valuesFlag = FlagDomain{
	Name:       "values",
	Class:      ClassStructuralReference,
	TypeName:   "string",
	Summary:    "path to the tenant values JSON",
	Form:       "path to a JSON object with string values",
	Resolution: "existence and readability are resolved at runtime",
	Example:    "license.values.json",
	Env:        "LICENSE_VALUES",
	Required:   true,
	Stability:  StabilityStable,
}

// outFlag is the canonical value domain of the --out flag.
var outFlag = FlagDomain{
	Name:       "out",
	Class:      ClassStructuralReference,
	TypeName:   "string",
	Summary:    "output directory of the rendered instance",
	Form:       "path to a directory",
	Resolution: "missing parent directories are created at runtime",
	Default:    ".",
	Example:    ".",
	Env:        "LICENSE_OUT",
	Stability:  StabilityStable,
}

// lockFlag is the canonical value domain of the --lock flag.
var lockFlag = FlagDomain{
	Name:       "lock",
	Class:      ClassStructuralReference,
	TypeName:   "string",
	Summary:    "path to the tenant lock file",
	Form:       "path to a license.lock.json document",
	Resolution: "existence and readability are resolved at runtime when the flag is given",
	Example:    "license.lock.json",
	Env:        "LICENSE_LOCK",
	Stability:  StabilityStable,
}

// dirFlag is the canonical value domain of the --dir flag.
var dirFlag = FlagDomain{
	Name:       "dir",
	Class:      ClassStructuralReference,
	TypeName:   "string",
	Summary:    "directory of the committed instance",
	Form:       "path to the directory that carries the committed LICENSE and LICENSES/ files",
	Resolution: "existence is resolved at runtime",
	Default:    ".",
	Example:    ".",
	Env:        "LICENSE_DIR",
	Stability:  StabilityStable,
}

// dryRunFlag is the canonical value domain of the --dry-run flag of every
// mutating command.
//
// Convention: docs/conventions/cli/interaction/README.md
var dryRunFlag = FlagDomain{
	Name:      "dry-run",
	Class:     ClassBooleanSwitch,
	Summary:   "show the plan without mutating",
	Effect:    "computes and prints the plan without writing any file",
	Default:   "false",
	Example:   "--dry-run",
	Env:       "LICENSE_DRY_RUN",
	Stability: StabilityStable,
}

// yesFlag is the canonical value domain of the --yes confirmation flag of
// every mutating command.
//
// Convention: docs/conventions/cli/interaction/README.md
var yesFlag = FlagDomain{
	Name:      "yes",
	Class:     ClassBooleanSwitch,
	Summary:   "confirm the mutation non-interactively",
	Effect:    "confirms the mutation without an interactive confirmation",
	Default:   "false",
	Example:   "--yes",
	Env:       "LICENSE_YES",
	Stability: StabilityStable,
}

// GlobalFlags returns the canonical global flags available on the root and on
// every command.
//
// Convention: docs/conventions/cli/output/README.md
func GlobalFlags() []FlagDomain {
	return []FlagDomain{
		{
			Name:      "output",
			Class:     ClassClosedEnum,
			TypeName:  "string",
			Summary:   "output mode",
			Values:    []string{OutputHuman, OutputJSON},
			Default:   OutputHuman,
			Example:   "json",
			Env:       "LICENSE_OUTPUT",
			Stability: StabilityStable,
		},
		{
			Name:      "quiet",
			Class:     ClassBooleanSwitch,
			Summary:   "suppress successful human output",
			Effect:    "suppresses the successful human output; errors are never suppressed",
			Default:   "false",
			Example:   "--quiet",
			Env:       "LICENSE_QUIET",
			Stability: StabilityStable,
		},
	}
}

// Commands returns the canonical command registry in its declared order.
func Commands() []Command {
	return []Command{
		{
			Name:      "render",
			Summary:   "render the canonical template into the LICENSE and LICENSES/ artifacts",
			Mutating:  true,
			Offline:   true,
			Stability: StabilityStable,
			Flags:     []FlagDomain{templateFlag, orgDefaultsFlag, valuesFlag, outFlag, dryRunFlag, yesFlag},
		},
		{
			Name:      "verify",
			Summary:   "prove the committed instance matches the canonical render",
			Mutating:  false,
			Offline:   true,
			Stability: StabilityStable,
			Flags:     []FlagDomain{templateFlag, orgDefaultsFlag, valuesFlag, lockFlag, dirFlag},
		},
		{
			Name:      "digest",
			Summary:   "print the canonical sha256 digest of a template file",
			Mutating:  false,
			Offline:   true,
			Stability: StabilityStable,
			Flags:     []FlagDomain{templateFlag},
		},
		{
			Name:      "version",
			Summary:   "print the binary version, commit, and build date",
			Mutating:  false,
			Offline:   true,
			Stability: StabilityStable,
			Flags:     nil,
		},
	}
}

// CommandByName resolves one registry entry by name.
func CommandByName(name string) (Command, bool) {
	for _, command := range Commands() {
		if command.Name == name {
			return command, true
		}
	}
	return Command{}, false
}

// GlobalFlagByName resolves one global flag domain by name.
func GlobalFlagByName(name string) (FlagDomain, bool) {
	for _, flag := range GlobalFlags() {
		if flag.Name == name {
			return flag, true
		}
	}
	return FlagDomain{}, false
}
