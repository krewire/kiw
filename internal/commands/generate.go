package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/krewire/libs/core"
)

// RegisterGenerate registers flags for the generate command.
func RegisterGenerate(fs *flag.FlagSet) {
	fs.String("kind", "", "generator kind: handler, migration, spec")
	fs.String("name", "", "name for the generated artifact")
}

// RunGenerate runs code generators (KWF-N4K8Q, future).
func RunGenerate(fs *flag.FlagSet) core.ExitCode {
	kind := flagValue(fs, "kind")
	name := flagValue(fs, "name")

	if kind == "" {
		fmt.Fprintln(os.Stderr, "kiw generate: --kind required (handler, migration, spec)")
		return core.ExitCodeUsage
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "kiw generate: --name required")
		return core.ExitCodeUsage
	}

	switch kind {
	case "handler":
		fmt.Printf("generate: HTTP handler %q (placeholder)\n", name)
	case "migration":
		fmt.Printf("generate: database migration %q (placeholder)\n", name)
	case "spec":
		fmt.Printf("generate: spec document %q (placeholder)\n", name)
	default:
		fmt.Fprintf(os.Stderr, "kiw generate: unknown kind %q — supported: handler, migration, spec\n", kind)
		return core.ExitCodeUsage
	}
	return core.ExitCodeSuccess
}
