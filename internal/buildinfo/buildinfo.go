// Package buildinfo resolves module versions for the devtool: from Go build
// metadata first, falling back to the modules' own declared versions when the
// binary was built inside a Go workspace (go.work), where metadata records
// "(devel)" instead of a tag.
package buildinfo

import (
	"runtime/debug"

	"github.com/krewire/framework"
	"github.com/krewire/libs/core"
)

const (
	modFramework = "github.com/krewire/framework"
	modLibs      = "github.com/krewire/libs"
)

// DevelVersion is the version Go records for modules built from source
// rather than a released tag.
const DevelVersion = "(devel)"

// ModuleVersion returns the version of the module at path as recorded in the
// build info, or "" when the module is not part of the build.
func ModuleVersion(path string) string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, dep := range bi.Deps {
		if dep.Path == path {
			return dep.Version
		}
	}
	return ""
}

// KnownVersion returns the version declared by the workspace-local Krewire
// module at path, or "" for unknown paths.
func KnownVersion(path string) string {
	switch path {
	case modFramework:
		return framework.Version.String()
	case modLibs:
		return core.CurrentVersion.String()
	}
	return ""
}

// ResolveVersion returns the effective version of the module at path and how
// it was resolved. fromSource is true when the binary was built from
// workspace sources (metadata records "(devel)") and the version comes from
// the module's declared constant rather than a released tag. version is ""
// when unknown.
func ResolveVersion(path string) (version string, fromSource bool) {
	v := ModuleVersion(path)
	if v != "" && v != DevelVersion {
		return v, false
	}
	if known := KnownVersion(path); known != "" {
		return known, true
	}
	return "", false
}
