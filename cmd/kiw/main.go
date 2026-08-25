// The kiw devtool: composes the Krewire ecosystem from the command line.
// Binary is named "kiw" for fast typing; module remains github.com/krewire/kiw.
package main

import (
	"os"

	"github.com/krewire/framework/tui"
	"github.com/krewire/kiw/internal/commands"
	"github.com/krewire/kiw/internal/version"
)

func main() {
	app := tui.NewApp("kiw", version.VersionString()).
		Command(tui.NewCommand("version", "print the krewire CLI and framework versions", nil, commands.RunVersion)).
		Command(tui.NewCommand("info", "print environment and project information", nil, commands.RunInfo)).
		Command(tui.NewCommand("new", "scaffold a minimal Krewire project kernel", commands.RegisterNew, commands.RunNew)).
		Command(tui.NewCommand("build", "build the current project's website", commands.RegisterBuild, commands.RunBuild)).
		Command(tui.NewCommand("serve", "preview the current project's website over HTTP", commands.RegisterServe, commands.RunServe)).
		Command(tui.NewCommand("init", "equip the project with a variant: monolith, --site, --book, --cli, or --template", commands.RegisterInit, commands.RunInit)).
		Command(tui.NewCommand("test", "run the tests of the current project", commands.RegisterTest, commands.RunTest)).
		Command(tui.NewCommand("vet", "run go vet on the current project", nil, commands.RunVet)).
		Command(tui.NewCommand("fmt", "check formatting with gofmt/go fmt", commands.RegisterFmt, commands.RunFmt)).
		Command(tui.NewCommand("run", "build and run the current app", commands.RegisterRun, commands.RunRun)).
		Command(tui.NewCommand("dev", "run the current app in dev mode with auto-restart", commands.RegisterDev, commands.RunDev)).
		Command(tui.NewCommand("deploy", "stage deployable artifacts in dist/", commands.RegisterDeploy, commands.RunDeploy)).
		Command(tui.NewCommand("guild", "install the Guild AI agent template into a project", commands.RegisterGuild, commands.RunGuild)).
		Command(tui.NewCommand("ws", "workspace for monorepo, multi-repo, multi-project, microservices", commands.RegisterWs, commands.RunWs))

	os.Exit(int(app.Run(os.Args[1:])))
}
