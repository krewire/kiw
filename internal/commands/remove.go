package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/krewire/kiw/internal/packages"
	"github.com/krewire/libs/core"
)

func RegisterRemove(fs *flag.FlagSet) {}

func RunRemove(fs *flag.FlagSet) core.ExitCode {
	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: kiw remove <package> [package ...]")
		fmt.Fprintln(os.Stderr, "examples: kiw remove twcss, kiw remove tailwindcss, kiw remove github.com/foo/bar")
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
			// For remove, allow bare names without version parsing strictness
			spec = packages.Spec{Name: raw, Raw: raw}
		}
		// Strip version for remove; only name matters
		spec.Version = ""
		resolved, err := chain.Resolve(spec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot resolve %q: %v\n", raw, err)
			return core.ExitCodeFailure
		}
		slog.Info("removing package", "package", spec.Name, "kind", resolved.Kind)
		fmt.Printf("→ removing %s (%s) ...\n", spec.Name, resolved.Kind)
		if err := resolved.Installer.Remove(root); err != nil {
			fmt.Fprintf(os.Stderr, "failed to remove %q: %v\n", raw, err)
			return core.ExitCodeFailure
		}
		fmt.Printf("✓ removed %s\n", spec.Name)
	}
	return core.ExitCodeSuccess
}
