// Command license renders and verifies canonical license instances.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/t33n-software/license-hub/internal/adapters/fs"
	"github.com/t33n-software/license-hub/internal/application"
)

var (
	version     = "devel"
	commit      = "unknown"
	date        = "unknown"
	exitProcess = os.Exit
	commandArgs = os.Args
	newService  = func() *application.LicenseService {
		return application.NewLicenseService(fs.New())
	}
)

func main() {
	exitProcess(run(commandArgs[1:], os.Stdout, os.Stderr, newService()))
}

func run(arguments []string, stdout io.Writer, stderr io.Writer, service *application.LicenseService) int {
	if len(arguments) == 0 {
		fmt.Fprintln(stderr, "usage: license <render|verify|digest|version> [flags]")
		return 2
	}
	switch arguments[0] {
	case "render":
		return runRender(arguments[1:], stdout, stderr, service)
	case "verify":
		return runVerify(arguments[1:], stdout, stderr, service)
	case "digest":
		return runDigest(arguments[1:], stdout, stderr, service)
	case "version":
		fmt.Fprintf(stdout, "license %s (%s, %s)\n", version, commit, date)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", arguments[0])
		return 2
	}
}

func runRender(arguments []string, stdout io.Writer, stderr io.Writer, service *application.LicenseService) int {
	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	flags.SetOutput(stderr)
	templatePath := flags.String("template", "", "path to the canonical template")
	orgDefaultsPath := flags.String("org-defaults", "", "path to the organization defaults JSON")
	valuesPath := flags.String("values", "", "path to the tenant values JSON")
	outDir := flags.String("out", ".", "output directory of the rendered instance")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *templatePath == "" || *orgDefaultsPath == "" || *valuesPath == "" {
		fmt.Fprintln(stderr, "render requires --template, --org-defaults and --values")
		return 2
	}
	result, err := service.Render(application.RenderRequest{
		TemplatePath:    *templatePath,
		OrgDefaultsPath: *orgDefaultsPath,
		ValuesPath:      *valuesPath,
		OutDir:          *outDir,
	})
	if err != nil {
		fmt.Fprintln(stderr, "render:", err)
		return 1
	}
	for _, path := range result.Written {
		fmt.Fprintln(stdout, "wrote", path)
	}
	fmt.Fprintln(stdout, "template digest", result.Digest)
	return 0
}

func runVerify(arguments []string, stdout io.Writer, stderr io.Writer, service *application.LicenseService) int {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	templatePath := flags.String("template", "", "path to the canonical template")
	orgDefaultsPath := flags.String("org-defaults", "", "path to the organization defaults JSON")
	valuesPath := flags.String("values", "", "path to the tenant values JSON")
	lockPath := flags.String("lock", "", "optional path to the tenant lock file")
	dir := flags.String("dir", ".", "directory of the committed instance")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *templatePath == "" || *orgDefaultsPath == "" || *valuesPath == "" {
		fmt.Fprintln(stderr, "verify requires --template, --org-defaults and --values")
		return 2
	}
	violations, err := service.Verify(application.VerifyRequest{
		TemplatePath:    *templatePath,
		OrgDefaultsPath: *orgDefaultsPath,
		ValuesPath:      *valuesPath,
		LockPath:        *lockPath,
		Dir:             *dir,
	})
	if err != nil {
		fmt.Fprintln(stderr, "verify:", err)
		return 1
	}
	if len(violations) > 0 {
		for _, violation := range violations {
			fmt.Fprintln(stderr, "violation:", violation)
		}
		return 1
	}
	fmt.Fprintln(stdout, "license instance matches the canonical render")
	return 0
}

func runDigest(arguments []string, stdout io.Writer, stderr io.Writer, service *application.LicenseService) int {
	flags := flag.NewFlagSet("digest", flag.ContinueOnError)
	flags.SetOutput(stderr)
	templatePath := flags.String("template", "", "path to the canonical template")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if *templatePath == "" {
		fmt.Fprintln(stderr, "digest requires --template")
		return 2
	}
	result, err := service.TemplateDigest(*templatePath)
	if err != nil {
		fmt.Fprintln(stderr, "digest:", err)
		return 1
	}
	fmt.Fprintln(stdout, result)
	return 0
}
