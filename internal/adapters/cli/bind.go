package cli

import (
	"flag"
	"slices"
	"strconv"
	"strings"

	"github.com/t33n-software/license-hub/internal/domain/contract"
)

// BoundFlags binds the flag domains of one command to a flag.FlagSet and
// resolves the effective values with the organization-wide precedence
// flag > environment variable > default.
//
// Convention: docs/conventions/cli/configuration/README.md
type BoundFlags struct {
	set       *flag.FlagSet
	domains   []contract.FlagDomain
	lookupEnv func(string) (string, bool)
	values    map[string]*string
	bools     map[string]*bool
}

// Bind registers every domain on the set. The caller owns the set's output
// and usage handling; parse failures are reported as structured error
// records, never through the flag package's own output.
func Bind(set *flag.FlagSet, domains []contract.FlagDomain, lookupEnv func(string) (string, bool)) *BoundFlags {
	b := &BoundFlags{
		set:       set,
		domains:   domains,
		lookupEnv: lookupEnv,
		values:    make(map[string]*string, len(domains)),
		bools:     make(map[string]*bool, len(domains)),
	}
	for _, domain := range domains {
		if domain.Class == contract.ClassBooleanSwitch {
			b.bools[domain.Name] = set.Bool(domain.Name, domain.Default == "true", domain.Summary)
		} else {
			b.values[domain.Name] = set.String(domain.Name, domain.Default, domain.Summary)
		}
	}
	return b
}

// Parse parses args, applies the environment precedence, and validates the
// effective values against their domains. A non-nil result is the structured
// error record of the rejection; the call fails closed before any mutation.
//
// Convention: docs/conventions/cli/errors/README.md
func (b *BoundFlags) Parse(args []string) *ErrorRecord {
	if err := b.set.Parse(args); err != nil {
		return &ErrorRecord{
			Code:        contract.ErrUsage,
			Field:       "flags",
			Actual:      strings.Join(args, " "),
			Rule:        "the flags could not be parsed: " + err.Error(),
			Remediation: "run '" + contract.BinaryName + " " + b.set.Name() + " --help' for the full input contract",
		}
	}
	if record := b.applyEnvironment(); record != nil {
		return record
	}
	return b.validate()
}

// applyEnvironment fills every flag that was not set explicitly from its
// mapped environment variable.
func (b *BoundFlags) applyEnvironment() *ErrorRecord {
	seen := make(map[string]bool)
	b.set.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, domain := range b.domains {
		if seen[domain.Name] {
			continue
		}
		value, ok := b.lookupEnv(domain.Env)
		if !ok || value == "" {
			continue
		}
		if domain.Class == contract.ClassBooleanSwitch {
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				return &ErrorRecord{
					Code:        contract.ErrValueInvalid,
					Field:       domain.Name,
					Actual:      value,
					Expected:    "true or false",
					Rule:        "the environment variable " + domain.Env + " must carry a boolean",
					Example:     domain.Env + "=true",
					Remediation: "fix the environment variable " + domain.Env + " or pass --" + domain.Name,
				}
			}
			*b.bools[domain.Name] = parsed
			continue
		}
		*b.values[domain.Name] = value
	}
	return nil
}

// validate proves the effective values against their value domains.
func (b *BoundFlags) validate() *ErrorRecord {
	for _, domain := range b.domains {
		if domain.Class == contract.ClassBooleanSwitch {
			continue
		}
		value := *b.values[domain.Name]
		if domain.Required && value == "" {
			return &ErrorRecord{
				Code:     contract.ErrValueMissing,
				Field:    domain.Name,
				Expected: domain.Form,
				Rule:     "the flag is required and no value was supplied",
				Example:  "--" + domain.Name + " " + domain.Example,
				Remediation: "pass --" + domain.Name + " or set " + domain.Env + "; run '" + contract.BinaryName + " " +
					b.set.Name() + " --help' for the full input contract",
			}
		}
		if domain.Class == contract.ClassClosedEnum && !slices.Contains(domain.Values, value) {
			return &ErrorRecord{
				Code:        contract.ErrValueInvalid,
				Field:       domain.Name,
				Actual:      value,
				Expected:    OrJoin(domain.Values),
				Rule:        "the value must be one of the documented set",
				Example:     "--" + domain.Name + " " + domain.Values[0],
				Remediation: "run '" + contract.BinaryName + " " + b.set.Name() + " --help' for the accepted values",
			}
		}
	}
	return nil
}

// String returns the effective value of a value-taking flag.
func (b *BoundFlags) String(name string) string {
	return *b.values[name]
}

// Bool returns the effective value of a boolean-switch flag.
func (b *BoundFlags) Bool(name string) bool {
	return *b.bools[name]
}
