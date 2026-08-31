// Package cli is the delivery adapter that renders every consumer-facing
// channel — static help, results, and structured errors — from the canonical
// contract registry in internal/domain/contract. No value domain, rule text,
// or value list is duplicated here; everything renders from the registry.
//
// Convention: docs/conventions/cli/values/single-source-of-truth.md,
// docs/conventions/cli/help/README.md.
package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/t33n-software/license-hub/internal/domain/contract"
)

// WriteRootHelp renders the level-0 root help from the command registry.
//
// Convention: docs/conventions/cli/help/README.md
func WriteRootHelp(w io.Writer, commands []contract.Command, globals []contract.FlagDomain) {
	var b strings.Builder
	fmt.Fprintf(&b, "%s — %s\n", contract.BinaryName, contract.ToolSummary)
	b.WriteString("\nUsage:\n")
	fmt.Fprintf(&b, "  %s <command> [flags]\n", contract.BinaryName)
	fmt.Fprintf(&b, "  %s <command> --help\n", contract.BinaryName)
	fmt.Fprintf(&b, "  %s --version [--output json]\n", contract.BinaryName)
	b.WriteString("\nCommands:\n")
	width := 0
	for _, command := range commands {
		if len(command.Name) > width {
			width = len(command.Name)
		}
	}
	for _, command := range commands {
		fmt.Fprintf(&b, "  %-*s %s\n", width, command.Name, command.Summary)
	}
	b.WriteString("\nGlobal flags:\n")
	writeFlagLines(&b, globals)
	fmt.Fprintf(&b, "\nDiscovery: every command owns a complete help area — run '%s <command> --help'.\n", contract.BinaryName)
	writeExitCodeLine(&b)
	offline := true
	for _, command := range commands {
		if !command.Offline {
			offline = false
		}
	}
	if offline {
		b.WriteString("Every command works offline.\n")
	}
	io.WriteString(w, b.String())
}

// WriteLeafHelp renders the level-2 leaf help of one command from the
// registry: the full input contract with every flag domain, canonical
// examples, and the exit behavior.
//
// Convention: docs/conventions/cli/help/README.md
func WriteLeafHelp(w io.Writer, command contract.Command) {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s — %s\n", contract.BinaryName, command.Name, command.Summary)
	b.WriteString("\nUsage:\n  ")
	b.WriteString(usageLine(command))
	b.WriteString("\n")
	if len(command.Flags) > 0 {
		b.WriteString("\nFlags:\n")
		writeFlagLines(&b, command.Flags)
	}
	b.WriteString("\nGlobal flags:\n")
	writeFlagLines(&b, contract.GlobalFlags())
	b.WriteString("\nExamples:\n")
	for _, example := range examples(command) {
		fmt.Fprintf(&b, "  %s\n", example)
	}
	b.WriteString("\n")
	writeExitCodeLine(&b)
	if command.Offline {
		b.WriteString("This command works offline.\n")
	} else {
		b.WriteString("This command requires network access.\n")
	}
	if command.Mutating {
		b.WriteString("This command mutates state: pass --yes to confirm non-interactively, or --dry-run to preview the plan.\n")
	} else {
		b.WriteString("This command is read-only.\n")
	}
	fmt.Fprintf(&b, "Stability: %s\n", command.Stability)
	io.WriteString(w, b.String())
}

// usageLine derives the usage line from the required flags of the command.
func usageLine(command contract.Command) string {
	var b strings.Builder
	b.WriteString(contract.BinaryName + " " + command.Name)
	for _, flag := range command.Flags {
		if flag.Required {
			fmt.Fprintf(&b, " --%s <value>", flag.Name)
		}
	}
	b.WriteString(" [flags]")
	return b.String()
}

// examples derives the canonical, valid examples from the flag domains; a
// mutating command carries the confirmation form and the dry-run form.
//
// Convention: docs/conventions/cli/help/README.md (examples are valid and
// canonical)
func examples(command contract.Command) []string {
	var b strings.Builder
	b.WriteString(contract.BinaryName + " " + command.Name)
	for _, flag := range command.Flags {
		if flag.Required {
			fmt.Fprintf(&b, " --%s %s", flag.Name, flag.Example)
		}
	}
	base := b.String()
	if !command.Mutating {
		return []string{base}
	}
	return []string{base + " --yes", base + " --dry-run"}
}

// writeExitCodeLine renders the semantic exit codes from the contract.
func writeExitCodeLine(b *strings.Builder) {
	codes := []int{contract.ExitSuccess, contract.ExitExecution, contract.ExitUsage, contract.ExitGovernance}
	names := make([]string, 0, len(codes))
	for _, code := range codes {
		names = append(names, fmt.Sprintf("%d %s", code, contract.ExitCodeNames[code]))
	}
	fmt.Fprintf(b, "Exit codes: %s.\n", strings.Join(names, "; "))
}

// writeFlagLines renders every flag domain as one help entry; the rule
// segments wrap at a fixed continuation column under the text start.
func writeFlagLines(b *strings.Builder, flags []contract.FlagDomain) {
	headings := make([]string, len(flags))
	width := 0
	for i, flag := range flags {
		headings[i] = flagHeading(flag)
		if len(headings[i]) > width {
			width = len(headings[i])
		}
	}
	continuation := strings.Repeat(" ", 2+width+3)
	for i, flag := range flags {
		segments := strings.Split(flagText(flag), "; ")
		fmt.Fprintf(b, "  %-*s   %s\n", width, headings[i], segments[0])
		for _, segment := range segments[1:] {
			fmt.Fprintf(b, "%s%s\n", continuation, segment)
		}
	}
}

// flagHeading renders the flag name with its display type.
func flagHeading(flag contract.FlagDomain) string {
	if flag.Class == contract.ClassBooleanSwitch {
		return "--" + flag.Name
	}
	return "--" + flag.Name + " " + flag.TypeName
}

// flagText renders the binding help duty of the flag's value class.
//
// Convention: docs/conventions/cli/values/README.md
func flagText(flag contract.FlagDomain) string {
	var body string
	switch flag.Class {
	case contract.ClassClosedEnum:
		body = flag.Summary + "; " + OrJoin(flag.Values)
	case contract.ClassShaped:
		body = flag.Summary + "; grammar: " + flag.Grammar
	case contract.ClassFreeConstrained:
		body = flag.Summary + "; " + flag.Rules
	case contract.ClassStructuralReference:
		body = flag.Summary + "; form: " + flag.Form + "; " + flag.Resolution
	case contract.ClassScalarBounded:
		body = flag.Summary + "; range: " + flag.Range
	case contract.ClassBooleanSwitch:
		body = flag.Effect
	case contract.ClassCompositeToken:
		body = flag.Summary + "; transport: " + flag.Grammar
	case contract.ClassSecretReference:
		body = flag.Summary + "; reference forms: " + flag.Form
	}
	if flag.Required {
		body += "; required"
	}
	if flag.Default != "" {
		if flag.Class == contract.ClassBooleanSwitch {
			body += " (default " + flag.Default + ")"
		} else {
			body += fmt.Sprintf(" (default %q)", flag.Default)
		}
	}
	body += "; env: " + flag.Env
	body += "; example: " + flag.Example
	if flag.Stability != contract.StabilityStable {
		body += "; stability: " + string(flag.Stability)
	}
	return body
}

// OrJoin renders a closed-enum value set as an enumeration. The help and the
// error contract share this rendering of the canonical source.
//
// Convention: docs/conventions/cli/values/single-source-of-truth.md
func OrJoin(values []string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	case 2:
		return values[0] + " or " + values[1]
	}
	return strings.Join(values[:len(values)-1], ", ") + ", or " + values[len(values)-1]
}
