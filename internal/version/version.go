// Package version defines the kiw CLI version (module github.com/krewire/kiw).
package version

import "github.com/krewire/libs/core"

// Version is the semantic version of the kiw devtool.
var Version = core.MustParseVersion("0.3.3")

// VersionString returns the version as a string for UI display.
func VersionString() string { return Version.String() }

// EcosystemRequires declares the minimum versions of the modules kiw depends on.
// This is the authoritative compatibility contract for kiw; `kiw compat` validates
// it against every other module's `<module>/version.go` declaration.
var EcosystemRequires = map[core.ModuleName]core.Version{
	core.ModuleFramework: core.MustParseVersion("0.3.1"),
	core.ModuleLibs:      core.MustParseVersion("0.3.0"),
	core.ModuleMdbind:    core.MustParseVersion("0.2.0"),
	core.ModuleGuild:     core.MustParseVersion("0.1.0"),
	core.ModuleShip:      core.MustParseVersion("0.0.0"),
}
