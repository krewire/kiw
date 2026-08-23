package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func detectLocalPaths(startDir string) (string, string) {
	dir := startDir
	for {
		modPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(modPath); err == nil {
			if strings.Contains(string(data), "module github.com/krewire/kiw") {
				frameworkPath := filepath.Join(dir, "..", "framework")
				libsPath := filepath.Join(dir, "..", "libs")
				if _, err := os.Stat(frameworkPath); err == nil {
					if _, err := os.Stat(libsPath); err == nil {
						return frameworkPath, libsPath
					}
				}
				return "", ""
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", ""
}

func goModTemplate(module, frameworkVersion, libsVersion, frameworkPath, libsPath string) string {
	mod := fmt.Sprintf("module %s\n\ngo 1.22\n\nrequire (\n\t%s %s\n\t%s %s\n)",
		module, modFramework, strings.TrimSpace(frameworkVersion), modLibs, strings.TrimSpace(libsVersion))
	if frameworkPath != "" {
		mod += fmt.Sprintf("\n\nreplace %s => %s", modFramework, frameworkPath)
	}
	if libsPath != "" {
		mod += fmt.Sprintf("\nreplace %s => %s", modLibs, libsPath)
	}
	return mod + "\n"
}

func krewireYamlTemplate(name string) string {
	return fmt.Sprintf(`project:
  name: %s
  kind: app
  version: 0.1.0
`, name)
}

func cmdMainTemplate(module string) string {
	return fmt.Sprintf(`// Thin entry point: load config, build app, run.
package main

import (
	"fmt"
	"os"

	"%s/internal/app"
	"%s/internal/config"
)

func main() {
	meta, err := config.LoadMetadata("krewire.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	appCfg, err := config.LoadConfig("cfg.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	a, err := app.New(meta, appCfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := a.Run(meta.Addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, module, module)
}

func readmeTemplate(name string) string {
	return fmt.Sprintf(`# %s

A Krewire project scaffolded by the [kiw CLI](https://github.com/krewire/kiw) (binary `+"`kiw`"+`).

Built on the [Krewire Framework](https://github.com/krewire/framework) and
[Krewire Libraries](https://github.com/krewire/libs).

## Project Structure

This is a Krewire fullstack monolith:

	main.go       # thin entry point
	internal/     # app assembly, config, HTTP handlers
	web/          # UI sources: layouts, pages, theme
	public/       # static assets (embedded)

## Getting Started

	go run .

## Testing

	go test ./...
`, name)
}
