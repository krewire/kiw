package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/krewire/kiw/internal/packages"
	"github.com/krewire/libs/core"
)

func RegisterAdd(fs *flag.FlagSet) {
	// No flags yet; version is encoded in package@version syntax.
	// Future: --dev, --save-dev for npm dev deps.
}

func RunAdd(fs *flag.FlagSet) core.ExitCode {
	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: kiw add <package[@version]> [package[@version] ...]")
		fmt.Fprintln(os.Stderr, "  package@latest  latest version")
		fmt.Fprintln(os.Stderr, "  package@1.2.3   specific version")
		fmt.Fprintln(os.Stderr, "examples: kiw add twcss, kiw add tailwindcss@3.4.1, kiw add github.com/foo/bar@v1.2.3")
		return core.ExitCodeUsage
	}
	root, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	chain := packages.DefaultChain()
	for _, raw := range args {
		spec, err := packages.ParseSpec(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid package %q: %v\n", raw, err)
			return core.ExitCodeUsage
		}
		resolved, err := chain.Resolve(spec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot resolve %q: %v\n", raw, err)
			return core.ExitCodeFailure
		}
		version := spec.EffectiveVersion()
		slog.Info("adding package", "package", spec.Name, "version", version, "kind", resolved.Kind, "resolver", resolved.Installer.Describe())
		fmt.Printf("→ adding %s (%s) ...\n", spec.Name+"@"+version, resolved.Kind)
		if err := resolved.Installer.Add(root, version); err != nil {
			fmt.Fprintf(os.Stderr, "failed to add %q: %v\n", raw, err)
			return core.ExitCodeFailure
		}
		fmt.Printf("✓ added %s\n", spec.Name+"@"+version)
	}
	return core.ExitCodeSuccess
}
