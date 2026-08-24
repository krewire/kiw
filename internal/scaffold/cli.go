// Package-level file for the command-line application (VariantCLI): kiw
// init --cli scaffolds a CLI built on framework/tui + libs/core, mirroring
// the kiw devtool's own layout (root main.go + internal/commands).
package scaffold

import (
	"fmt"
	"os"
)

// equipCLI shapes the kernel into a command-line application. The root
// main.go becomes the tui.App harness, commands live under
// internal/commands, and go.mod is upgraded in place with pinned requires.
func equipCLI(opts EquipOptions) ([]string, error) {
	if opts.Module == "" {
		return nil, fmt.Errorf("equip cli: module path is required (read from go.mod)")
	}
	if opts.Name == "" {
		return nil, fmt.Errorf("equip cli: project name is required")
	}
	frameworkVersion := opts.FrameworkVersion
	if frameworkVersion == "" {
		frameworkVersion = "latest"
	}
	libsVersion := opts.LibsVersion
	if libsVersion == "" {
		libsVersion = "v0.1.0"
	}

	cwd, _ := os.Getwd()
	frameworkPath, libsPath := detectLocalPaths(cwd)

	files := []file{
		{goModFile, goModTemplate(opts.Module, frameworkVersion, libsVersion, frameworkPath, libsPath)},
		{krewireYaml, cliKrewireYamlTemplate(opts.Name)},
		{mainGo, cliMainTemplate(opts.Module, opts.Name)},
		{"internal/commands/commands.go", cliCommandsTemplate(opts.Name)},
		{"internal/config/config.go", cliConfigTemplate()},
		{"README.md", cliReadmeTemplate(opts.Name)},
		{gitignoreFile, gitignoreTemplate(opts.Name)},
	}
	return writeVariant(opts.Dir, files)
}

func cliKrewireYamlTemplate(name string) string {
	return "project:\n  name: " + name + "\n  kind: cli\n  version: 0.1.0\n"
}

func cliMainTemplate(module, name string) string {
	return fmt.Sprintf(`// Entry point: build the CLI app and run it.
package main

import (
	"fmt"
	"os"

	"github.com/krewire/framework/tui"
	"%s/internal/commands"
	"%s/internal/config"
)

const version = "0.1.0"

func main() {
	if _, err := config.LoadMetadata("krewire.yaml"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	app := tui.NewApp("%s", version).
		Command(tui.NewCommand("hello", "print a greeting", nil, commands.Hello))
	os.Exit(int(app.Run(os.Args[1:])))
}
`, module, module, name)
}

func cliCommandsTemplate(name string) string {
	return fmt.Sprintf(`// Package commands implements the %s CLI sub-commands.
package commands

import (
	"flag"
	"fmt"

	"github.com/krewire/libs/core"
)

// Hello prints a friendly greeting.
func Hello(_ *flag.FlagSet) core.ExitCode {
	fmt.Println("hello, %s")
	return core.ExitCodeSuccess
}
`, name, name)
}

func cliReadmeTemplate(name string) string {
	return fmt.Sprintf(`# %s

A Krewire CLI project scaffolded by the [kiw CLI](https://github.com/krewire/kiw) (binary `+"`kiw`"+`, module `+"`github.com/krewire/kiw`"+`).

Built on the [Krewire Framework](https://github.com/krewire/framework) and
[Krewire Libraries](https://github.com/krewire/libs).

## Getting Started

	go run . hello

## Building

	go build -o ./%s .

## Testing

	go test ./...
`, name, name)
}

func cliConfigTemplate() string {
	return "// Package config defines the typed CLI configuration\n" +
		"// krewire.yaml: metadata and project-level config (metadata only)\n" +
		"package config\n\n" +
		"import (\n" +
		"\t\"fmt\"\n" +
		"\t\"os\"\n\n" +
		"\trconfig \"github.com/krewire/libs/config\"\n" +
		"\t\"github.com/krewire/libs/validate\"\n" +
		")\n\n" +
		"// Metadata mirrors krewire.yaml.\n" +
		"type Metadata struct {\n" +
		"\tProject Project `yaml:\"project\"`\n" +
		"}\n\n" +
		"// Project holds the project section of krewire.yaml.\n" +
		"type Project struct {\n" +
		"\tName    string `yaml:\"name\" validate:\"required\"`\n" +
		"\tKind    string `yaml:\"kind\" validate:\"required\"`\n" +
		"\tVersion string `yaml:\"version\"`\n" +
		"}\n\n" +
		"// LoadMetadata reads krewire.yaml from path, overlays the environment, and returns a\n" +
		"// validated Metadata.\n" +
		"func LoadMetadata(path string) (*Metadata, error) {\n" +
		"\tcfg := &Metadata{}\n" +
		"\tif err := rconfig.Load(path, cfg); err != nil {\n" +
		"\t\treturn nil, err\n" +
		"\t}\n" +
		"\tif err := rconfig.Override(cfg, os.LookupEnv); err != nil {\n" +
		"\t\treturn nil, err\n" +
		"\t}\n" +
		"\tif err := validate.Struct(cfg); err != nil {\n" +
		"\t\treturn nil, fmt.Errorf(\"config: %w\", err)\n" +
		"\t}\n" +
		"\treturn cfg, nil\n" +
		"}\n"
}
