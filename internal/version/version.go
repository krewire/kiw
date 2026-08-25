// Package version defines the kiw CLI version (module github.com/krewire/kiw).
package version

import "github.com/krewire/libs/core"

// Version is the semantic version of the kiw devtool.
var Version = core.MustParseVersion("0.3.0")

// VersionString returns the version as a string for UI display.
func VersionString() string { return Version.String() }

// EcosystemRequires declares compatibility.
var EcosystemRequires = map[core.ModuleName]core.Version{
	core.ModuleFramework: core.MustParseVersion("0.3.1"),
	core.ModuleLibs:      core.MustParseVersion("0.3.0"),
	core.ModuleGuild:     core.MustParseVersion("0.1.0"),
}
