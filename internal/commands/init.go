package commands

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/kiw/internal/gomod"
	"github.com/krewire/kiw/internal/scaffold"
	"github.com/krewire/libs/core"
)

// RegisterInit registers flags for the init command.
func RegisterInit(fs *flag.FlagSet) {
	fs.Bool("site", false, "equip a declarative static site (ssg: key in krewire.yaml)")
	fs.Bool("book", false, "equip a manuscript book (mdbind)")
	fs.Bool("cli", false, "equip a command-line application (framework/tui)")
	fs.String("template", "", "bootstrap from a remote git template (git URL)")
	fs.String("title", "", "site title for the site and book variants")
}

// RunInit equips the project in the current directory (or an optional
// positional target) with the requested variant. With no variant flag the
// project is equipped as a fullstack monolith. Config lives exclusively in
// krewire.yaml — no ssg.yaml is produced.
func RunInit(fs *flag.FlagSet) core.ExitCode {
	dir := fs.Arg(0)
	if dir == "" {
		dir = "."
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return fail(err)
	}

	templateURL := flagValue(fs, "template")
	site := flagBool(fs, "site")
	book := flagBool(fs, "book")
	cli := flagBool(fs, "cli")
	if count := boolCount(site, book, cli, templateURL != ""); count > 1 {
		fmt.Fprintln(os.Stderr, "kiw init: choose one variant: --site, --book, --cli, or --template")
		return core.ExitCodeUsage
	}

	opts := scaffold.EquipOptions{
		Dir:         dir,
		Title:       flagValue(fs, "title"),
		TemplateURL: templateURL,
	}
	switch {
	case site:
		opts.Variant = scaffold.VariantStatic
	case book:
		opts.Variant = scaffold.VariantBook
	case cli:
		opts.Variant = scaffold.VariantCLI
	case templateURL != "":
		opts.Variant = scaffold.VariantTemplate
	default:
		opts.Variant = scaffold.VariantApp
	}

	switch opts.Variant {
	case scaffold.VariantApp, scaffold.VariantCLI:
		mod, err := gomod.Find(dir)
		if err != nil {
			return usageOrFail(core.WithHint(
				core.UsageError("kiw init: not inside a Go module"),
				"run 'kiw new <project>' first, then 'kiw init --cli' (or another variant) inside it",
			))
		}
		opts.Module = mod.Path
		opts.Name = moduleBase(opts.Module)
		fw, libs := resolveVersions()
		opts.FrameworkVersion = fw
		opts.LibsVersion = libs
	case scaffold.VariantStatic, scaffold.VariantBook:
		name := filepath.Base(dir)
		opts.Name = name
		opts.Title = firstNonEmpty(opts.Title, name)
	}

	created, err := scaffold.Equip(opts)
	if err != nil {
		if isScaffoldUsage(err) {
			return commandError(err)
		}
		return fail(err)
	}
	slog.Info("equipped Krewire project", "dir", dir, "variant", opts.Variant, "files", len(created))
	for _, path := range created {
		fmt.Println("created " + path)
	}
	return core.ExitCodeSuccess
}

// moduleBase returns the last path segment of a module path, used as the
// project name for the app variant.
func moduleBase(module string) string {
	parts := strings.Split(strings.TrimSuffix(module, "/"), "/")
	return parts[len(parts)-1]
}

// flagBool returns the boolean value of a registered flag.
func flagBool(fs *flag.FlagSet, name string) bool {
	if f := fs.Lookup(name); f != nil {
		return f.Value.String() == "true"
	}
	return false
}

// boolCount returns how many of the given conditions are true.
func boolCount(vals ...bool) int {
	n := 0
	for _, v := range vals {
		if v {
			n++
		}
	}
	return n
}

// isScaffoldUsage reports whether err is a scaffold usage-class error that
// should exit with code 2.
func isScaffoldUsage(err error) bool {
	return errors.Is(err, scaffold.ErrProjectExists) ||
		errors.Is(err, scaffold.ErrNotEmpty) ||
		errors.Is(err, scaffold.ErrConflict)
}
