// Package-level file for the fullstack monolith (VariantApp): the app
// template set and equip wiring (Domain: app). kiw init dispatches here
// whenever no variant flag is given. Templates are split per bounded context:
// - app_templates.go: go.mod, krewire.yaml, main.go, README, local paths
// - app_internal.go: internal/app/config/http
// - app_web.go: web layouts/pages/theme
// - app_assets.go: assets/embed, public css/js
package scaffold

import (
	"fmt"
	"os"
)

// equipApp shapes the kernel into a fullstack monolith. The root main.go
// becomes the entry point (canonical layout KWF-CCI0N) and go.mod is
// upgraded in place with pinned framework and libs requires.
func equipApp(opts EquipOptions) ([]string, error) {
	if opts.Module == "" {
		return nil, fmt.Errorf("equip app: module path is required (read from go.mod)")
	}
	if opts.Name == "" {
		return nil, fmt.Errorf("equip app: project name is required")
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
		{krewireYaml, krewireYamlTemplate(opts.Name)},
		{mainGo, cmdMainTemplate(opts.Module)},
		{"internal/app/app.go", internalAppTemplate(opts.Module)},
		{"internal/config/config.go", internalConfigTemplate()},
		{"internal/http/http.go", internalHttpTemplate()},
		{"web/layouts/shell.go", webLayoutsShellTemplate()},
		{"web/pages/pages.go", webPagesTemplate()},
		{"web/theme/theme.go", webThemeTemplate()},
		{"assets/embed.go", assetsEmbedTemplate()},
		{"assets/public/app.css", publicCssTemplate()},
		{"assets/public/app.js", publicJsTemplate()},
		{"README.md", readmeTemplate(opts.Name)},
		{gitignoreFile, gitignoreTemplate(opts.Name)},
	}
	return writeVariant(opts.Dir, files)
}
