package commands

import (
	"fmt"
	"strings"

	"flag"

	"github.com/krewire/kiw/internal/buildinfo"
	"github.com/krewire/kiw/internal/version"
	"github.com/krewire/libs/core"
)

const (
	ModFramework = "github.com/krewire/framework"
	ModLibs      = "github.com/krewire/libs"
)

func RunVersion(_ *flag.FlagSet) core.ExitCode {
	fmt.Printf("Krewire v%s\n", version.Version)
	fmt.Printf("Krewire Framework %s\n", qualifiedVersion(ModFramework))
	return core.ExitCodeSuccess
}

func moduleVersion(path string) string {
	v, _ := buildinfo.ResolveVersion(path)
	return v
}

func humanVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	if v == "" || v == "devel" || v == buildinfo.DevelVersion {
		return "dev"
	}
	return "v" + v
}

// qualifiedVersion renders the module version for display: "v0.5.1" for a
// released tag, "v0.5.1 (dev)" when resolved from workspace sources, or "dev"
// when unknown. The (dev) qualifier lets operators tell non-release builds
// apart before making production upgrade decisions.
func qualifiedVersion(path string) string {
	v, fromSource := buildinfo.ResolveVersion(path)
	human := humanVersion(v)
	if fromSource && human != "dev" {
		return human + " (dev)"
	}
	return human
}

func resolveVersions() (framework, libs string) {
	fw, _ := buildinfo.ResolveVersion(ModFramework)
	lb, _ := buildinfo.ResolveVersion(ModLibs)
	return fw, lb
}
