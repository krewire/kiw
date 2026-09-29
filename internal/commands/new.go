package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/krewire/kiw/internal/scaffold"
	"github.com/krewire/libs/core"
)

func RegisterNew(fs *flag.FlagSet) {
	fs.String("module", "", "module path for the new project (defaults to the project name)")
	fs.String("dir", "", "directory to create the project in (defaults to the current directory)")
	fs.Bool("site", false, "equip a static site (pages/, layouts/, components/)")
	fs.Bool("book", false, "equip a manuscript book (mdbind)")
	fs.Bool("cli", false, "equip a command-line application (framework/tui)")
	fs.Bool("app", false, "equip a fullstack monolith application")
	fs.String("template", "", "bootstrap from a remote git template (git URL)")
	fs.String("title", "", "site title for the site and book variants")
}

func RunNew(fs *flag.FlagSet) core.ExitCode {
	name := fs.Arg(0)
	if name == "" {
		fmt.Fprintln(os.Stderr, "usage: kiw new <project> [--site|--book|--cli|--app|--template <git-url>]")
		return core.ExitCodeUsage
	}

	site := flagBool(fs, "site")
	book := flagBool(fs, "book")
	cli := flagBool(fs, "cli")
	app := flagBool(fs, "app")
	templateURL := flagValue(fs, "template")

	if count := boolCount(site, book, cli, app, templateURL != ""); count > 1 {
		fmt.Fprintln(os.Stderr, "kiw new: choose at most one variant: --site, --book, --cli, --app, or --template")
		return core.ExitCodeUsage
	}

	parentDir := flagValue(fs, "dir")
	modulePath := flagValue(fs, "module")

	created, err := scaffold.New(scaffold.Options{
		Name:   name,
		Dir:    parentDir,
		Module: modulePath,
	})
	if err != nil {
		return commandError(err)
	}

	targetDir := filepath.Join(firstNonEmpty(parentDir, "."), name)

	hasVariant := site || book || cli || app || templateURL != ""
	if hasVariant {
		opts := scaffold.EquipOptions{
			Dir:         targetDir,
			Title:       flagValue(fs, "title"),
			TemplateURL: templateURL,
		}
		switch {
		case site:
			opts.Variant = scaffold.VariantStatic
			opts.Name = name
			opts.Title = firstNonEmpty(opts.Title, name)
		case book:
			opts.Variant = scaffold.VariantBook
			opts.Name = name
			opts.Title = firstNonEmpty(opts.Title, name)
		case cli:
			opts.Variant = scaffold.VariantCLI
			opts.Module = firstNonEmpty(modulePath, name)
			opts.Name = name
			fw, libs := resolveVersions()
			opts.FrameworkVersion = fw
			opts.LibsVersion = libs
		case templateURL != "":
			opts.Variant = scaffold.VariantTemplate
		default: // app
			opts.Variant = scaffold.VariantApp
			opts.Module = firstNonEmpty(modulePath, name)
			opts.Name = name
			fw, libs := resolveVersions()
			opts.FrameworkVersion = fw
			opts.LibsVersion = libs
		}

		eqCreated, err := scaffold.Equip(opts)
		if err != nil {
			if isScaffoldUsage(err) {
				return commandError(err)
			}
			return fail(err)
		}
		slog.Info("scaffolded and equipped Krewire project", "name", name, "variant", opts.Variant, "files", len(eqCreated))
		for _, path := range eqCreated {
			fmt.Println("created " + filepath.Join(name, path))
		}
		fmt.Printf("next: cd %s && kiw dev\n", name)
		return core.ExitCodeSuccess
	}

	slog.Info("scaffolded Krewire project kernel", "name", name, "files", len(created))
	for _, path := range created {
		fmt.Println("created " + path)
	}
	fmt.Println("next: cd " + name + " && kiw init [--site|--book|--cli|--template <git-url>] to equip a variant")
	return core.ExitCodeSuccess
}
