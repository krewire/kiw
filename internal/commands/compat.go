package commands

import (
	"flag"
	"fmt"

	"github.com/krewire/framework"
	"github.com/krewire/guild"
	"github.com/krewire/kiw/internal/version"
	"github.com/krewire/libs/core"
	"github.com/krewire/mdbind"
	"github.com/krewire/ship"
)

// RunCompat validates version compatibility across the whole Krewire ecosystem.
// It treats each module's `<module>/version.go` (`Version` + `EcosystemRequires`)
// as the authoritative compatibility contract and checks that every declared
// requirement is satisfied by the versions actually declared by the dependencies.
// This is the single entry point that turns the per-module version.go files into
// a repo-wide compatibility gate.
func RunCompat(_ *flag.FlagSet) core.ExitCode {
	// Actual: each module's own declared version (from its version.go).
	actual := map[core.ModuleName]core.Version{
		core.ModuleFramework: framework.Version,
		core.ModuleLibs:      core.CurrentVersion,
		core.ModuleKiw:       version.Version,
		core.ModuleMdbind:    mdbind.Version,
		core.ModuleGuild:     guild.Version,
		core.ModuleShip:      ship.Version,
	}

	// Requirements: each module's declared EcosystemRequires.
	reqs := map[core.ModuleName]map[core.ModuleName]core.Version{
		core.ModuleFramework: framework.EcosystemRequires,
		core.ModuleKiw:       version.EcosystemRequires,
		core.ModuleMdbind:    mdbind.EcosystemRequires,
		core.ModuleGuild:     guild.EcosystemRequires,
		core.ModuleShip:      ship.EcosystemRequires,
	}

	issues := core.CheckCompatibility(actual, reqs)
	if len(issues) == 0 {
		fmt.Println("ok: all module version.go declarations are mutually compatible")
		return core.ExitCodeSuccess
	}
	for _, err := range issues {
		fmt.Printf("  x %s\n", err.Error())
	}
	return core.ExitCodeFailure
}
