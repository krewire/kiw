package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/krewire/kiw/internal/scaffold"
	"github.com/krewire/libs/core"
)

func RegisterNew(fs *flag.FlagSet) {
	fs.String("module", "", "module path for the new project (defaults to the project name)")
	fs.String("dir", "", "directory to create the project in (defaults to the current directory)")
}

func RunNew(fs *flag.FlagSet) core.ExitCode {
	name := fs.Arg(0)
	if name == "" {
		fmt.Fprintln(os.Stderr, "usage: kiw new <project> [--module <module-path>] [--dir <parent-dir>]")
		return core.ExitCodeUsage
	}
	created, err := scaffold.New(scaffold.Options{
		Name:   name,
		Dir:    flagValue(fs, "dir"),
		Module: flagValue(fs, "module"),
	})
	if err != nil {
		return commandError(err)
	}
	slog.Info("scaffolded Krewire project kernel", "name", name, "files", len(created))
	for _, path := range created {
		fmt.Println("created " + path)
	}
	fmt.Println("next: cd " + name + " && kiw init [--site|--book|--cli|--template <git-url>] to equip a variant")
	return core.ExitCodeSuccess
}
