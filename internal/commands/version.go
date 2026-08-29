package commands

import (
	"fmt"
	"runtime"
	"strings"

	"flag"

	"github.com/krewire/kiw/internal/buildinfo"
	"github.com/krewire/kiw/internal/version"
	"github.com/krewire/libs/core"
	"github.com/krewire/libs/term"
)

const (
	ModFramework = "github.com/krewire/framework"
	ModLibs      = "github.com/krewire/libs"
)

func RunVersion(_ *flag.FlagSet) core.ExitCode {
	tm := term.NewTerminal()
	bold := func(s string) string { return tm.Paint(s, term.ColorCyan, []term.Style{term.StyleBold}) }
	dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
	yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, nil) }

	fmt.Printf("%s %s\n", bold("kiw"), dim("Krewire Devtool"))
	fmt.Printf("  %-12s %s\n", dim("CLI"), green("v"+version.Version.String()))
	fw := qualifiedVersion(ModFramework)
	fwColor := green(fw)
	if strings.Contains(fw, "dev") {
		fwColor = yellow(fw)
	}
	fmt.Printf("  %-12s %s %s\n", dim("Framework"), green("Krewire Framework"), fwColor)
	lb := qualifiedVersion(ModLibs)
	lbColor := green(lb)
	if strings.Contains(lb, "dev") {
		lbColor = yellow(lb)
	}
	fmt.Printf("  %-12s %s %s\n", dim("Libraries"), green(ModLibs), lbColor)
	fmt.Printf("  %-12s %s\n", dim("Go"), dim(runtime.Version()+" ("+runtime.GOOS+"/"+runtime.GOARCH+")"))
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
