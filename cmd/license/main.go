// Command license renders and verifies canonical license instances.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/t33n-software/license-hub/internal/adapters/cli"
	"github.com/t33n-software/license-hub/internal/adapters/fs"
	"github.com/t33n-software/license-hub/internal/application"
	"github.com/t33n-software/license-hub/internal/domain/contract"
)

var (
	version               = "devel"
	commit                = "unknown"
	date                  = "unknown"
	exitProcess           = os.Exit
	commandArgs           = os.Args
	lookupEnv             = os.LookupEnv
	stdinReader io.Reader = os.Stdin
	// stdinIsTerminal is the terminal-detection seam; the production rule is
	// detectTerminal. The terminal detection rule is owned by
	// docs/conventions/cli/interaction/README.md.
	stdinIsTerminal = detectTerminal
	newService      = func() *application.LicenseService {
		return application.NewLicenseService(fs.New())
	}
)

func main() {
	exitProcess(run(commandArgs[1:], os.Stdout, os.Stderr, newService()))
}

// run executes one CLI call against the contract registry: root help and
// version, the help command, and the four registered commands.
//
// Convention: docs/conventions/cli/help/README.md,
// docs/conventions/cli/identity/README.md
func run(arguments []string, stdout io.Writer, stderr io.Writer, service *application.LicenseService) int {
	root := scanRoot(arguments)
	if root.command == "" {
		return runRoot(root, stdout, stderr)
	}
	if root.command == "help" {
		return runHelp(root, stdout, stderr)
	}
	command, ok := contract.CommandByName(root.command)
	if !ok {
		renderer := cli.NewRenderer(stdout, stderr, contract.OutputHuman, false)
		renderer.WriteError(cli.ErrorRecord{
			Code:        contract.ErrUsage,
			Field:       "command",
			Actual:      root.command,
			Expected:    cli.OrJoin(commandNames()),
			Rule:        "unknown command",
			Example:     contract.BinaryName + " render --help",
			Remediation: "run '" + contract.BinaryName + " --help' for the command list",
		})
		return contract.ExitUsage
	}
	if root.showHelp || helpRequested(root.args) {
		cli.WriteLeafHelp(stdout, command)
		return contract.ExitSuccess
	}
	return runCommand(command, root, stdout, stderr, service)
}

// rootCall is the parsed root-level call shape: the global flag tokens before
// the command, the command token, and the command arguments.
type rootCall struct {
	preFlags    []string
	command     string
	args        []string
	showHelp    bool
	showVersion bool
}

// scanRoot splits the arguments into the root flags, the command token, and
// the command arguments. The flags --help, -h, and --version are root
// requests; every other flag token is forwarded to the command's own flag
// set, which binds the global flags as well.
func scanRoot(arguments []string) rootCall {
	root := rootCall{}
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]
		if root.command != "" {
			root.args = append(root.args, arg)
			continue
		}
		switch {
		case arg == "--help" || arg == "-h":
			root.showHelp = true
		case arg == "--version":
			root.showVersion = true
		case strings.HasPrefix(arg, "-"):
			root.preFlags = append(root.preFlags, arg)
			if arg == "--output" && i+1 < len(arguments) {
				i++
				root.preFlags = append(root.preFlags, arguments[i])
			}
		default:
			root.command = arg
		}
	}
	return root
}

// runRoot executes a root-level call without a command: the machine-readable
// version query or one of the help forms.
//
// Convention: docs/conventions/cli/identity/README.md
func runRoot(root rootCall, stdout io.Writer, stderr io.Writer) int {
	set := flag.NewFlagSet(contract.BinaryName, flag.ContinueOnError)
	set.SetOutput(io.Discard)
	bound := cli.Bind(set, contract.GlobalFlags(), lookupEnv)
	if record := bound.Parse(root.preFlags); record != nil {
		renderer := cli.NewRenderer(stdout, stderr, requestedMode(root.preFlags), false)
		renderer.WriteError(*record)
		return contract.ExitUsage
	}
	renderer := cli.NewRenderer(stdout, stderr, bound.String("output"), bound.Bool("quiet"))
	if root.showVersion {
		renderer.WriteResult(versionResult())
		return contract.ExitSuccess
	}
	if root.showHelp {
		cli.WriteRootHelp(stdout, contract.Commands(), contract.GlobalFlags())
		return contract.ExitSuccess
	}
	cli.WriteRootHelp(stderr, contract.Commands(), contract.GlobalFlags())
	return contract.ExitUsage
}

// runHelp executes the help command: the second access form of the universal
// help contract.
func runHelp(root rootCall, stdout io.Writer, stderr io.Writer) int {
	if len(root.args) == 0 {
		cli.WriteRootHelp(stdout, contract.Commands(), contract.GlobalFlags())
		return contract.ExitSuccess
	}
	command, ok := contract.CommandByName(root.args[0])
	if !ok {
		renderer := cli.NewRenderer(stdout, stderr, contract.OutputHuman, false)
		renderer.WriteError(cli.ErrorRecord{
			Code:        contract.ErrUsage,
			Field:       "command",
			Actual:      root.args[0],
			Expected:    cli.OrJoin(commandNames()),
			Rule:        "unknown command",
			Example:     contract.BinaryName + " help render",
			Remediation: "run '" + contract.BinaryName + " --help' for the command list",
		})
		return contract.ExitUsage
	}
	cli.WriteLeafHelp(stdout, command)
	return contract.ExitSuccess
}

// runCommand parses the command flags against the registry domains and
// dispatches to the use case.
func runCommand(command contract.Command, root rootCall, stdout io.Writer, stderr io.Writer, service *application.LicenseService) int {
	args := slices.Concat(root.preFlags, root.args)
	set := flag.NewFlagSet(command.Name, flag.ContinueOnError)
	set.SetOutput(io.Discard)
	bound := cli.Bind(set, slices.Concat(command.Flags, contract.GlobalFlags()), lookupEnv)
	if record := bound.Parse(args); record != nil {
		renderer := cli.NewRenderer(stdout, stderr, requestedMode(args), false)
		renderer.WriteError(*record)
		return contract.ExitUsage
	}
	renderer := cli.NewRenderer(stdout, stderr, bound.String("output"), bound.Bool("quiet"))
	switch command.Name {
	case "render":
		return execRender(bound, renderer, stdout, service)
	case "verify":
		return execVerify(bound, renderer, service)
	case "digest":
		return execDigest(bound, renderer, service)
	default:
		// The registry owns exactly four commands and CommandByName guards the
		// dispatch, so the default branch is the version command.
		renderer.WriteResult(versionResult())
		return contract.ExitSuccess
	}
}

// execRender executes the mutating render command under the interaction
// contract: --dry-run previews the plan, and the mutation requires --yes in a
// non-interactive context or an explicit confirmation on a terminal.
//
// Convention: docs/conventions/cli/interaction/README.md,
// docs/conventions/cli/security/README.md
func execRender(bound *cli.BoundFlags, renderer cli.Renderer, stdout io.Writer, service *application.LicenseService) int {
	req := application.RenderRequest{
		TemplatePath:    bound.String("template"),
		OrgDefaultsPath: bound.String("org-defaults"),
		ValuesPath:      bound.String("values"),
		OutDir:          bound.String("out"),
	}
	if bound.Bool("dry-run") {
		plan, err := service.PlanRender(req)
		if err != nil {
			return failService(renderer, "render", err)
		}
		renderer.WriteResult(planResult(plan))
		return contract.ExitSuccess
	}
	if !bound.Bool("yes") {
		if !stdinIsTerminal() {
			renderer.WriteError(cli.ErrorRecord{
				Code:        contract.ErrConfirmationRequired,
				Field:       "yes",
				Rule:        "render mutates the working tree and requires an explicit confirmation in a non-interactive context",
				Example:     contract.BinaryName + " render --template templates/<family>/<template>/<Name>-<semver>.hbs --org-defaults org-defaults.json --values license.values.json --yes",
				Remediation: "pass --yes to confirm the mutation, or --dry-run to preview the plan",
			})
			return contract.ExitUsage
		}
		plan, err := service.PlanRender(req)
		if err != nil {
			return failService(renderer, "render", err)
		}
		renderer.WriteResult(planResult(plan))
		fmt.Fprint(stdout, "Apply the render? [y/N] ")
		answer, _ := bufio.NewReader(stdinReader).ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "y" && answer != "yes" {
			renderer.WriteError(cli.ErrorRecord{
				Code:        contract.ErrConfirmationRequired,
				Field:       "yes",
				Rule:        "the confirmation was declined; no files were written",
				Example:     contract.BinaryName + " render ... --yes",
				Remediation: "rerun and confirm the plan, or pass --yes in a non-interactive context",
			})
			return contract.ExitUsage
		}
	}
	result, err := service.Render(req)
	if err != nil {
		return failService(renderer, "render", err)
	}
	renderer.WriteResult(cli.Result{
		Command: "render",
		Status:  "ok",
		Fields: []cli.Field{
			{Key: "written", Label: "wrote", Values: result.Written},
			{Key: "digest", Label: "template digest", Values: []string{result.Digest}},
		},
	})
	return contract.ExitSuccess
}

// execVerify executes the read-only verify command; violations are a
// governance rejection with their own exit class.
//
// Convention: docs/conventions/cli/output/README.md
func execVerify(bound *cli.BoundFlags, renderer cli.Renderer, service *application.LicenseService) int {
	violations, err := service.Verify(application.VerifyRequest{
		TemplatePath:    bound.String("template"),
		OrgDefaultsPath: bound.String("org-defaults"),
		ValuesPath:      bound.String("values"),
		LockPath:        bound.String("lock"),
		Dir:             bound.String("dir"),
	})
	if err != nil {
		return failService(renderer, "verify", err)
	}
	if len(violations) > 0 {
		renderer.WriteGovernance(cli.Result{
			Command: "verify",
			Status:  "governance_rejected",
			Fields:  []cli.Field{{Key: "violations", Label: "violation:", Values: violations}},
		})
		return contract.ExitGovernance
	}
	renderer.WriteResult(cli.Result{
		Command: "verify",
		Status:  "ok",
		Message: "license instance matches the canonical render",
	})
	return contract.ExitSuccess
}

// execDigest executes the read-only digest command.
func execDigest(bound *cli.BoundFlags, renderer cli.Renderer, service *application.LicenseService) int {
	result, err := service.TemplateDigest(bound.String("template"))
	if err != nil {
		return failService(renderer, "digest", err)
	}
	renderer.WriteResult(cli.Result{
		Command: "digest",
		Status:  "ok",
		Fields:  []cli.Field{{Key: "digest", Label: "digest", Values: []string{result}}},
	})
	return contract.ExitSuccess
}

// versionResult renders the version record from the build-injected version
// source; the human and the machine form carry identical information.
//
// Convention: docs/conventions/cli/identity/README.md
func versionResult() cli.Result {
	return cli.Result{
		Command: "version",
		Status:  "ok",
		Fields: []cli.Field{
			{Key: "version", Label: "version", Values: []string{version}},
			{Key: "commit", Label: "commit", Values: []string{commit}},
			{Key: "date", Label: "date", Values: []string{date}},
		},
	}
}

// planResult renders the dry-run plan of a render.
func planResult(plan application.PlanResult) cli.Result {
	return cli.Result{
		Command: "render",
		Status:  "plan",
		Message: "dry-run plan: no files written",
		Fields: []cli.Field{
			{Key: "planned", Label: "would write", Values: plan.Targets},
			{Key: "digest", Label: "template digest", Values: []string{plan.Digest}},
		},
	}
}

// failService maps a service error to its structured error record and its
// semantic exit code: input contract violations are usage errors, every other
// failure is an execution failure.
//
// Convention: docs/conventions/cli/output/README.md,
// docs/conventions/cli/errors/README.md
func failService(renderer cli.Renderer, command string, err error) int {
	record := cli.ErrorRecord{
		Code:        contract.ErrExecution,
		Field:       command,
		Rule:        err.Error(),
		Remediation: "verify the referenced paths and inputs, then rerun; run '" + contract.BinaryName + " " + command + " --help' for the input contract",
	}
	switch {
	case errors.Is(err, application.ErrMissingValues):
		record.Code = contract.ErrValueInvalid
		record.Field = "values"
		record.Remediation = "add the missing keys to the values documents; see docs/infrastructure/tenant-files.md"
	case errors.Is(err, application.ErrUnresolvedPlaceholders):
		record.Code = contract.ErrValueInvalid
		record.Field = "template"
		record.Remediation = "resolve every placeholder through the values documents; see docs/infrastructure/template-contract.md"
	}
	renderer.WriteError(record)
	if record.Code == contract.ErrExecution {
		return contract.ExitExecution
	}
	return contract.ExitUsage
}

// helpRequested reports whether the command arguments ask for the help area.
func helpRequested(args []string) bool {
	return slices.Contains(args, "--help") || slices.Contains(args, "-h")
}

// charDevice reports whether the file mode marks a character device.
func charDevice(mode os.FileMode) bool {
	return mode&os.ModeCharDevice != 0
}

// detectTerminal reports whether the standard input is an interactive
// terminal.
func detectTerminal() bool {
	info, err := os.Stdin.Stat()
	return err == nil && charDevice(info.Mode())
}

// requestedMode extracts the requested output mode for the error path of a
// failed parse; an unparsable or undocumented mode falls back to human.
func requestedMode(args []string) string {
	for i, arg := range args {
		if arg == "--output=json" {
			return contract.OutputJSON
		}
		if arg == "--output" && i+1 < len(args) && args[i+1] == "json" {
			return contract.OutputJSON
		}
	}
	return contract.OutputHuman
}

// commandNames lists the registered command names in their declared order.
func commandNames() []string {
	commands := contract.Commands()
	names := make([]string, 0, len(commands))
	for _, command := range commands {
		names = append(names, command.Name)
	}
	return names
}
