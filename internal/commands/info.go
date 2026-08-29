package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/gomod"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/kiw/internal/version"
	"github.com/krewire/libs/core"
	"github.com/krewire/libs/term"
)

func RunInfo(_ *flag.FlagSet) core.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	mod, _ := gomod.Find(dir)
	cfg, _ := config.Load(dir)
	res, err := shape.Detect(dir, cfg.Kind())
	if err != nil {
		return fail(err)
	}

	tm := term.NewTerminal()
	dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }
	boldDim := func(s string) string {
		return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleBold, term.StyleDim})
	}
	cyan := func(s string) string { return tm.Paint(s, term.ColorCyan, nil) }
	green := func(s string) string { return tm.Paint(s, term.ColorGreen, nil) }
	yellow := func(s string) string { return tm.Paint(s, term.ColorYellow, nil) }

	fmt.Println(boldDim("─ Environment ─────────────────────────────────"))
	printKV(tm, "CLI", cyan("Krewire v"+version.Version.String()), dim)
	printKV(tm, "Framework", cyan("Krewire Framework "+qualifiedVersion(ModFramework)), dim)
	printKV(tm, "Libraries", cyan(ModLibs+" "+qualifiedVersion(ModLibs)), dim)
	printKV(tm, "Go", dim(runtime.Version()+" ")+yellow("("+runtime.GOOS+"/"+runtime.GOARCH+")"), dim)
	printKV(tm, "Env", green(resolvedEnvLabel(cfg)), dim)
	printKV(tm, "Debug", dim(fmt.Sprintf("%t", cfg.ResolveDebug("", false))), dim)

	fmt.Println()
	fmt.Println(boldDim("─ Project ────────────────────────────────────"))
	printKV(tm, "Directory", dim(dir), dim)
	modulePath := dim("<none>")
	usesKrewire := dim("no")
	if mod != nil {
		modulePath = cyan(mod.Path)
		if mod.UsesKrewire() {
			usesKrewire = green("yes")
		} else {
			usesKrewire = yellow("no")
		}
	}
	printKV(tm, "Module path", modulePath, dim)
	printKV(tm, "Built on Krewire", usesKrewire, dim)
	kindStr := cyan(res.String())
	if res.Kind == shape.KindNone {
		kindStr = yellow(res.String())
	}
	printKV(tm, "Project kind", kindStr, dim)
	if res.Marker != "" {
		printKV(tm, "Detected by", dim(res.Marker), dim)
	}

	if len(cfg.Scripts) > 0 {
		fmt.Println()
		fmt.Println(boldDim("─ Scripts ──────────────────────────────────"))
		keys := make([]string, 0, len(cfg.Scripts))
		for k := range cfg.Scripts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			printKV(tm, k, dim(cfg.Scripts[k]), dim)
		}
	}

	if res.Kind == shape.KindApp {
		fmt.Println()
		fmt.Println(boldDim("─ Directories ────────────────────────────────"))
		printAppDirsStyled(dir, cfg, tm)
	}
	return core.ExitCodeSuccess
}

func printKV(tm *term.Terminal, key, value string, dim func(string) string) {
	_ = tm
	fmt.Printf("  %-16s %s\n", dim(key), value)
}

func printAppDirsStyled(root string, cfg *config.Config, tm *term.Terminal) {
	dirs := resolveDirs(cfg)
	dim := func(s string) string { return tm.Paint(s, term.ColorDefault, []term.Style{term.StyleDim}) }
	for name, path := range dirs {
		full := filepath.Join(root, path)
		status := tm.Paint("exists", term.ColorGreen, nil)
		if _, err := os.Stat(full); err != nil {
			status = tm.Paint("missing", term.ColorYellow, nil)
		}
		fmt.Printf("  %-16s %s %s\n", dim(name+":"), path, "("+status+")")
	}
}

// resolvedEnvLabel renders the effective environment for display, marking an
// invalid krewire.yaml value instead of failing the report (KWL-K4T7W).
func resolvedEnvLabel(cfg *config.Config) string {
	e, err := cfg.ResolveEnv("")
	if err == nil {
		return e.String()
	}
	if cfg != nil && cfg.Env != "" {
		return cfg.Env + " (invalid)"
	}
	return core.DefaultEnv.String()
}

func printAppDirs(root string, cfg *config.Config) {
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

func resolveDirs(cfg *config.Config) map[string]string {
	d := cfg.Project.Dirs
	return map[string]string{
		"cmd":      firstNonEmpty(d.Cmd, "cmd/app"),
		"internal": firstNonEmpty(d.Internal, "internal"),
		"web":      firstNonEmpty(d.Web, "web"),
		"public":   firstNonEmpty(d.Public, "public"),
	}
}
