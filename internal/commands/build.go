package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/framework/ui"
	"github.com/krewire/framework/web/ssg"
	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/gomod"
	"github.com/krewire/libs/core"
	"github.com/krewire/mdbind/book"
)

// RegisterBuild registers flags for the build command.
func RegisterBuild(fs *flag.FlagSet) {
	fs.String("input", "", "manuscript directory (default manuscript)")
	fs.String("output", "", "output directory (default site)")
	fs.String("base", "", "URL base the site will be served under (default /)")
	fs.String("title", "", "site title (defaults to the project name)")
	fs.String("author", "", "site author")
	fs.String("theme", "", "theme mode: auto, light, dark, or off (default auto)")
}

// RunBuild builds the current project's website. It supports two declarative
// shapes: a krewire.yaml with an `ssg:` key built with the web/ssg generator,
// or a manuscript/ directory built with the mdbind site builder. No project
// ships its own site command anymore.
func RunBuild(fs *flag.FlagSet) core.ExitCode {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: "+err.Error())
		return core.ExitCodeUsage
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	switch {
	case hasDir(root, "pages"):
		return buildSSGFromFile(root, cfg, fs)
	case cfg.IsSSG():
		return buildSSGFromConfig(root, cfg, fs)
	case hasDir(root, "manuscript"):
		return buildManuscript(root, cfg, fs)
	default:
		fmt.Fprintln(os.Stderr, "kiw: no website found — add pages/*.kiw, an `ssg:` key to krewire.yaml, or a manuscript/ directory")
		return core.ExitCodeUsage
	}
}

// buildSSGFromFile builds the project's SSG site from the file-based layout
// (pages/, components/, layouts/, content/, public/) using ssg.LoadFromDir.
// krewire.yaml supplies metadata (title, description, theme) and output dir.
func buildSSGFromFile(root string, cfg *config.Config, fs *flag.FlagSet) core.ExitCode {
	output := firstNonEmpty(flagValue(fs, "output"), cfg.Output, "site")
	outDir := joinRoot(root, output, "site")
	slog.Info("building SSG site from file layout", "root", root, "output", outDir)
	site, err := ssg.LoadFromDir(root)
	if err != nil {
		return fail(err)
	}
	created, err := site.Build(outDir)
	if err != nil {
		return fail(err)
	}
	for _, p := range created {
		fmt.Println("created " + p)
	}
	return core.ExitCodeSuccess
}

// buildSSGFromConfig builds the project's SSG site from the `ssg:` section
// of krewire.yaml. Top-level fields (title, output, theme) are merged into the
// ssg.Config so they don't need to be repeated under ssg:.
func buildSSGFromConfig(root string, cfg *config.Config, fs *flag.FlagSet) core.ExitCode {
	ssgCfg := cfg.ToSSGConfig()
	output := firstNonEmpty(flagValue(fs, "output"), cfg.Output, "site")
	outDir := joinRoot(root, output, "site")
	slog.Info("building SSG site from krewire.yaml", "output", outDir)
	created, err := ssg.BuildFromConfig(ssgCfg, outDir)
	if err != nil {
		return fail(err)
	}
	for _, p := range created {
		fmt.Println("created " + p)
	}
	return core.ExitCodeSuccess
}

// buildManuscript renders the project's manuscript/ directory with mdbind.
// Settings come from krewire.yaml in the project root, overridden by flags.
func buildManuscript(root string, cfg *config.Config, fs *flag.FlagSet) core.ExitCode {
	title := firstNonEmpty(flagValue(fs, "title"), cfg.Title, moduleName(root))
	bcfg := book.Config{
		Input:      joinRoot(root, flagValue(fs, "input"), firstNonEmpty(cfg.Input, "manuscript")),
		Output:     joinRoot(root, flagValue(fs, "output"), firstNonEmpty(cfg.Output, "site")),
		Title:      title,
		Author:     firstNonEmpty(flagValue(fs, "author"), cfg.Author),
		BasePath:   firstNonEmpty(flagValue(fs, "base"), cfg.Base, "/"),
		NavLinks:   navFromConfig(cfg.Nav),
		FooterText: cfg.Footer,
		Theme:      themeFrom(fs, cfg.Theme),
	}
	slog.Info("building site with mdbind", "input", bcfg.Input, "output", bcfg.Output)
	created, err := book.Build(bcfg)
	if err != nil {
		return fail(err)
	}
	for _, path := range created {
		fmt.Println("created " + path)
	}
	return core.ExitCodeSuccess
}

// themeFrom resolves the theme from the --theme flag and krewire.yaml,
// defaulting to auto. A mode of off disables the switcher.
func themeFrom(fs *flag.FlagSet, cfg *config.Theme) *ui.Theme {
	mode := strings.ToLower(strings.TrimSpace(flagValue(fs, "theme")))
	if mode == "" && cfg != nil {
		mode = strings.ToLower(strings.TrimSpace(cfg.Default))
	}
	if mode == "" {
		mode = "auto"
	}
	switch mode {
	case "off", "none", "disabled":
		return nil
	}
	t := &ui.Theme{Default: mode}
	if cfg != nil {
		t.Light = cfg.Light.UI(ui.DefaultLightPalette)
		t.Dark = cfg.Dark.UI(ui.DefaultDarkPalette)
	}
	return t
}

// navFromConfig converts krewire.yaml nav links into book links.
func navFromConfig(links []config.Link) []book.Link {
	if len(links) == 0 {
		return nil
	}
	out := make([]book.Link, 0, len(links))
	for _, l := range links {
		out = append(out, book.Link{Text: l.Text, URL: l.URL})
	}
	return out
}

// firstNonEmpty returns the first non-empty string among the arguments.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// findRoot locates the project root walking up from the working directory.
// For app/cli/worker/service it is the Go module root (go.mod). For site/book
// which have no entry point, krewire.yaml or a declarative layout (pages/,
// manuscript/) is sufficient (KWF-DF3PL FRK-FLS-001/002, KWL-K1N2Q).
func findRoot() (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cur, err = filepath.Abs(cur)
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil && !info.IsDir() {
			return cur, nil
		}
		if info, err := os.Stat(filepath.Join(cur, "krewire.yaml")); err == nil && !info.IsDir() {
			return cur, nil
		}
		if info, err := os.Stat(filepath.Join(cur, "pages")); err == nil && info.IsDir() {
			return cur, nil
		}
		if info, err := os.Stat(filepath.Join(cur, "manuscript")); err == nil && info.IsDir() {
			return cur, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", fmt.Errorf("not inside a Go module — run 'kiw new <project>' first")
		}
		cur = parent
	}
}

// joinRoot resolves a flag value relative to the project root, falling back to
// def when the flag is unset.
func joinRoot(root, value, def string) string {
	if value == "" {
		value = def
	}
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(root, value)
}

// hasDir reports whether path exists inside root as a directory.
func hasDir(root, path string) bool {
	info, err := os.Stat(filepath.Join(root, path))
	return err == nil && info.IsDir()
}

// hasFile reports whether path exists inside root as a regular file.
func hasFile(root, path string) bool {
	info, err := os.Stat(filepath.Join(root, path))
	return err == nil && !info.IsDir()
}

// moduleName returns the last path element of the module declared in root's
// go.mod.
func moduleName(root string) string {
	mod, err := gomod.Read(filepath.Join(root, "go.mod"))
	if err != nil || mod.Path == "" {
		return filepath.Base(root)
	}
	parts := strings.Split(strings.TrimSuffix(mod.Path, "/"), "/")
	return parts[len(parts)-1]
}
