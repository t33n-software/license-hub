// Package contract is the canonical single source of truth for the CLI
// contract of the license tool: the command registry, the value-domain class
// of every flag, the error codes, and the semantic exit codes. Every
// consumer-facing channel renders from this package; no channel keeps its own
// copy.
//
// Convention: docs/conventions/cli/values/README.md,
// docs/conventions/cli/values/single-source-of-truth.md,
// docs/conventions/cli/output/README.md.
package contract

// BinaryName is the single stable binary name of the tool.
//
// Convention: docs/conventions/cli/identity/README.md
const BinaryName = "license"

// SchemaVersion is the version of the machine-readable output schema. The
// schema evolves additively within one version.
//
// Convention: docs/conventions/cli/output/README.md,
// docs/conventions/cli/lifecycle/README.md
const SchemaVersion = 1

// Output modes of the stable output contract.
//
// Convention: docs/conventions/cli/output/README.md
const (
	OutputHuman = "human"
	OutputJSON  = "json"
)

// Semantic exit codes of the CLI contract.
//
// Convention: docs/conventions/cli/output/README.md
const (
	ExitSuccess    = 0 // the command completed successfully
	ExitExecution  = 1 // the command failed at execution time
	ExitUsage      = 2 // the call itself was invalid
	ExitGovernance = 3 // a governance proof rejected the current state
)

// ToolSummary is the one-line purpose of the tool shown in the root help.
//
// Convention: docs/conventions/cli/help/README.md
const ToolSummary = "render and verify canonical license instances"

// ExitCodeNames maps every semantic exit code to its stable class name.
//
// Convention: docs/conventions/cli/output/README.md
var ExitCodeNames = map[int]string{
	ExitSuccess:    "success",
	ExitExecution:  "execution failure",
	ExitUsage:      "usage error",
	ExitGovernance: "governance rejection",
}

// ValueClass is one of the eight value-domain classes. Every flag and every
// positional argument is assigned to exactly one class, and the class
// determines the binding help duty.
//
// Convention: docs/conventions/cli/values/README.md
type ValueClass string

// The eight canonical value-domain classes.
const (
	ClassClosedEnum          ValueClass = "closed-enum"
	ClassShaped              ValueClass = "shaped"
	ClassFreeConstrained     ValueClass = "free-constrained"
	ClassStructuralReference ValueClass = "structural-reference"
	ClassScalarBounded       ValueClass = "scalar-bounded"
	ClassBooleanSwitch       ValueClass = "boolean-switch"
	ClassCompositeToken      ValueClass = "composite-token"
	ClassSecretReference     ValueClass = "secret-reference"
)

// Stability is the stability level of a command or flag, shown visibly in the
// help.
//
// Convention: docs/conventions/cli/lifecycle/README.md
type Stability string

// The canonical stability levels.
const (
	StabilityStable       Stability = "stable"
	StabilityExperimental Stability = "experimental"
	StabilityInternal     Stability = "internal"
)

// ErrorCode is a stable, coded error identifier of the structured error
// contract.
//
// Convention: docs/conventions/cli/output/README.md
type ErrorCode string

// The canonical error codes.
const (
	ErrUsage                ErrorCode = "USAGE_ERROR"
	ErrValueMissing         ErrorCode = "VALUE_MISSING"
	ErrValueInvalid         ErrorCode = "VALUE_INVALID"
	ErrConfirmationRequired ErrorCode = "CONFIRMATION_REQUIRED"
	ErrExecution            ErrorCode = "EXECUTION_FAILED"
	ErrGovernanceRejected   ErrorCode = "GOVERNANCE_REJECTED"
)

// FlagDomain is the canonical value-domain definition of one flag. Only the
// fields of the assigned class carry content; every other field is empty.
//
// Convention: docs/conventions/cli/values/README.md
type FlagDomain struct {
	Name       string     // flag name without leading dashes, lowercase kebab-case
	Class      ValueClass // the exactly-one value-domain class
	TypeName   string     // display type of a value-taking flag, empty for switches
	Summary    string     // one-line purpose
	Values     []string   // closed-enum: the complete accepted set of the endpoint
	Default    string     // default display, empty when no default exists
	Grammar    string     // shaped: the grammar template
	Rules      string     // free-constrained: the compact rule set
	Form       string     // structural-reference: the accepted form
	Resolution string     // structural-reference: the runtime resolution rule
	Range      string     // scalar-bounded: the value range with unit
	Effect     string     // boolean-switch: the effect of the set switch
	Example    string     // one canonical valid example
	Env        string     // environment variable of the organization-wide mapping
	Required   bool       // whether the call fails closed without the value
	Stability  Stability  // stability level shown in the help
}

// Command is the canonical registry entry of one command: a leaf command is a
// use-case boundary and owns a complete help area.
//
// Convention: docs/conventions/cli/help/README.md
type Command struct {
	Name      string       // command name, lowercase kebab-case
	Summary   string       // one-line purpose for the parent navigation
	Mutating  bool         // whether the command mutates state
	Offline   bool         // whether the command works without network
	Stability Stability    // stability level shown in the help
	Flags     []FlagDomain // the command-local flag domains
}
