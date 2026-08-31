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
		WithDescription("One CLI for every workload — site, book, app, worker, service, infra").
		Command(tui.NewCommand("version", "print CLI and framework versions", nil, commands.RunVersion).WithGroup("inspect").WithExample("kiw version")).
		Command(tui.NewCommand("info", "print environment and project information", nil, commands.RunInfo).WithGroup("inspect").WithExample("kiw info")).
		Command(tui.NewCommand("compat", "check version compatibility across all modules", nil, commands.RunCompat).WithGroup("inspect").WithExample("kiw compat")).
		Command(tui.NewCommand("new", "scaffold a minimal Krewire project kernel", commands.RegisterNew, commands.RunNew).WithGroup("project").WithExample("kiw new my-site")).
		Command(tui.NewCommand("init", "equip the project with a variant: monolith, --site, --book, --cli, or --template", commands.RegisterInit, commands.RunInit).WithGroup("project").WithExample("kiw init --site")).
		Command(tui.NewCommand("build", "build the current project's website", commands.RegisterBuild, commands.RunBuild).WithGroup("build").WithExample("kiw build")).
		Command(tui.NewCommand("serve", "preview the current project's website over HTTP", commands.RegisterServe, commands.RunServe).WithGroup("build").WithExample("kiw serve --port 3000")).
		Command(tui.NewCommand("run", "build and run the current app, a Go file, or a task from krewire.yaml", commands.RegisterRun, commands.RunRun).WithGroup("develop").WithExample("kiw run [task|path/to/file.go] [-- args]")).
		Command(tui.NewCommand("dev", "run the current app in dev mode with auto-restart", commands.RegisterDev, commands.RunDev).WithGroup("develop").WithExample("kiw dev")).
		Command(tui.NewCommand("test", "run the tests of the current project", commands.RegisterTest, commands.RunTest).WithGroup("develop").WithExample("kiw test")).
		Command(tui.NewCommand("vet", "run go vet on the current project", nil, commands.RunVet).WithGroup("develop").WithExample("kiw vet")).
		Command(tui.NewCommand("fmt", "check formatting with gofmt/go fmt", commands.RegisterFmt, commands.RunFmt).WithGroup("develop").WithExample("kiw fmt --write")).
		Command(tui.NewCommand("add", "add a package or plugin (package@version, @latest for latest)", commands.RegisterAdd, commands.RunAdd).WithGroup("project").WithExample("kiw add twcss@latest")).
		Command(tui.NewCommand("remove", "remove a package or plugin", commands.RegisterRemove, commands.RunRemove).WithGroup("project").WithExample("kiw remove twcss")).
		Command(tui.NewCommand("deploy", "stage deployable artifacts in dist/", commands.RegisterDeploy, commands.RunDeploy).WithGroup("ship").WithExample("kiw deploy --preview")).
		Command(tui.NewCommand("worker", "run background workers / job queues", commands.RegisterWorker, commands.RunWorker).WithGroup("develop").WithExample("kiw worker")).
		Command(tui.NewCommand("dashboard", "local dev dashboard (services, logs, traces)", commands.RegisterDashboard, commands.RunDashboard).WithGroup("develop").WithExample("kiw dashboard")).
		Command(tui.NewCommand("ws", "workspace for monorepo, multi-repo, multi-project, microservices", commands.RegisterWs, commands.RunWs).WithGroup("ship").WithExample("kiw ws help")).
		Command(tui.NewCommand("guild", "install the Guild AI agent template into a project", commands.RegisterGuild, commands.RunGuild).WithGroup("ship").WithExample("kiw guild install")).
		Command(tui.NewCommand("release", "bump versions and stage a release across modules", commands.RegisterRelease, commands.RunRelease).WithGroup("ship").WithExample("kiw release framework --bump minor --apply")).
		Command(tui.NewCommand("generate", "generate code (handlers, migrations, specs)", commands.RegisterGenerate, commands.RunGenerate).WithGroup("project").WithExample("kiw generate --kind handler --name CreateUser"))

	os.Exit(int(app.Run(os.Args[1:])))
}
