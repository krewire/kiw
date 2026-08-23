package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/krewire/kiw/internal/gomod"
	"github.com/krewire/kiw/internal/rvconf"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/kiw/internal/version"
	"github.com/krewire/libs/core"
)

func RunInfo(_ *flag.FlagSet) core.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	mod, _ := gomod.Find(dir)
	cfg, _ := rvconf.Load(dir)
	res, err := shape.Detect(dir, cfg.Kind())
	if err != nil {
		return fail(err)
	}

	fmt.Println("Environment")
	fmt.Printf("  CLI              Krewire v%s\n", version.Version)
	fmt.Printf("  Framework        Krewire Framework %s\n", qualifiedVersion(ModFramework))
	fmt.Printf("  Libraries        %s %s\n", ModLibs, qualifiedVersion(ModLibs))
	fmt.Printf("  Go               %s (%s/%s)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Printf("  Env              %s\n", resolvedEnvLabel(cfg))
	fmt.Printf("  Debug            %t\n", cfg.ResolveDebug("", false))

	modulePath := "<none>"
	usesKrewire := "no"
	if mod != nil {
		modulePath = mod.Path
		if mod.UsesKrewire() {
			usesKrewire = "yes"
		}
	}

	fmt.Println("Project")
	fmt.Printf("  Directory        %s\n", dir)
	fmt.Printf("  Module path      %s\n", modulePath)
	fmt.Printf("  Built on Krewire  %s\n", usesKrewire)
	fmt.Printf("  Project kind     %s\n", res)
	if res.Marker != "" {
		fmt.Printf("  Detected by      %s\n", res.Marker)
	}

	if res.Kind == shape.KindApp {
		printAppDirs(dir, cfg)
	}
	return core.ExitCodeSuccess
}

// resolvedEnvLabel renders the effective environment for display, marking an
// invalid krewire.yaml value instead of failing the report (KWL-K4T7W).
func resolvedEnvLabel(cfg *rvconf.Config) string {
	e, err := cfg.ResolveEnv("")
	if err == nil {
		return e.String()
	}
	if cfg != nil && cfg.Env != "" {
		return cfg.Env + " (invalid)"
	}
	return core.DefaultEnv.String()
}

func printAppDirs(root string, cfg *rvconf.Config) {
	dirs := resolveDirs(cfg)
	fmt.Println("Directories")
	for name, path := range dirs {
		full := filepath.Join(root, path)
		exists := "exists"
		if _, err := os.Stat(full); err != nil {
			exists = "missing"
		}
		fmt.Printf("  %-10s %s (%s)\n", name+":", path, exists)
	}
}

func resolveDirs(cfg *rvconf.Config) map[string]string {
	d := cfg.Project.Dirs
	return map[string]string{
		"cmd":      firstNonEmpty(d.Cmd, "cmd/app"),
		"internal": firstNonEmpty(d.Internal, "internal"),
		"web":      firstNonEmpty(d.Web, "web"),
		"public":   firstNonEmpty(d.Public, "public"),
	}
}
