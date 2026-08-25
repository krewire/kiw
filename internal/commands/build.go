package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/krewire/framework/plugin"
	_ "github.com/krewire/framework/plugin"
	"github.com/krewire/framework/ui"
	"github.com/krewire/framework/web/ssg"
	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/gomod"
	"github.com/krewire/libs/core"
	"github.com/krewire/mdbind/book"
)

// RegisterBuild registers flags for the build command.
func RegisterBuild(fs *flag.FlagSet) {
	fs.String("input", "", "content directory (default content)")
	fs.String("output", "", "output directory (default .krewire/build)")
	fs.String("o", "", "output directory (shorthand for --output)")
	fs.String("include", "", "comma-separated content glob patterns to build (default **/*.md)")
	fs.String("exclude", "", "comma-separated content globs to skip (default **/README.md,**/readme.md)")
	fs.String("base", "", "URL base the site will be served under (default /)")
	fs.String("title", "", "site title (defaults to the project name)")
	fs.String("author", "", "site author")
	fs.String("theme", "", "theme mode: auto, light, dark, or off (default auto)")
}

// RunBuild builds the current project's website. It supports both declarative
// shapes — a krewire.yaml with an `ssg:` key or pages/*.kiw built with
// web/ssg, and content/**/*.md built with mdbind — and allows them to
// coexist for progressive enhancement (a docs site can start as book and
// grow a custom ssg site without rewrite). When both are present both are
// built into the same output (shared .krewire/build default); the book then
// suppresses its root TOC so the ssg landing page owns "/".
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
	hasPages := hasDir(root, "pages")
	hasSSG := cfg.IsSSG()
	hasBook := hasDir(root, "content") || hasDir(root, "manuscript")
	if !hasPages && !hasSSG && !hasBook {
		fmt.Fprintln(os.Stderr, "kiw: no website found — add pages/*.kiw, an `ssg:` key to krewire.yaml, or a content/ directory")
		return core.ExitCodeUsage
	}
	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
	pruneStale(outDir)
	var firstErr core.ExitCode = core.ExitCodeSuccess
	built := false
	if hasPages {
		if code := buildSSGFromFile(root, cfg, fs); code != core.ExitCodeSuccess {
			firstErr = code
		} else {
			built = true
		}
	} else if hasSSG {
		if code := buildSSGFromConfig(root, cfg, fs); code != core.ExitCodeSuccess {
			firstErr = code
		} else {
			built = true
		}
	}
	if hasBook {
		if code := buildManuscript(root, cfg, fs, hasPages || hasSSG); code != core.ExitCodeSuccess {
			if firstErr == core.ExitCodeSuccess {
				firstErr = code
			}
		} else {
			built = true
		}
	}
	if built {
		writeManifest(outDir, collectCreated(outDir))
	}
	if !built {
		return firstErr
	}
	return core.ExitCodeSuccess
}

// manifestName is the build manifest written into the output directory; it
// lists every file the last successful build produced so the next run can
// prune stale artifacts (e.g. pages dropped by new exclude rules).
const manifestName = ".kiw-build-manifest"

// readManifest returns the relative paths recorded in outDir's manifest.
func readManifest(outDir string) []string {
	data, err := os.ReadFile(filepath.Join(outDir, manifestName))
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, filepath.Clean(line))
		}
	}
	return out
}

// writeManifest records rel (slash-separated) paths as the current build
// output. Missing entries from a previous manifest were already pruned.
func writeManifest(outDir string, rel []string) {
	sort.Strings(rel)
	body := strings.Join(rel, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(outDir, manifestName), []byte(body), 0o644); err != nil {
		slog.Warn("kiw build: could not write build manifest", "err", err)
	}
}

// pruneStale deletes files recorded by the previous manifest that still
// exist, plus empty parent directories inside outDir. Only manifest-listed
// files are touched — anything else in the output directory is left alone.
func pruneStale(outDir string) {
	old := readManifest(outDir)
	if len(old) == 0 {
		return
	}
	for _, rel := range old {
		p := filepath.Join(outDir, filepath.FromSlash(rel))
		if err := os.Remove(p); err == nil {
			slog.Debug("pruned stale output", "file", rel)
		}
	}
	removeEmptyDirs(outDir)
	os.Remove(filepath.Join(outDir, manifestName))
}

// removeEmptyDirs prunes now-empty directories under root (deepest first).
func removeEmptyDirs(root string) {
	var dirs []string
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() && p != root {
			dirs = append(dirs, p)
		}
		return nil
	})
	sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	for _, d := range dirs {
		os.Remove(d)
	}
}

// collectCreated walks outDir and lists every regular file relative to it,
// excluding the manifest itself.
func collectCreated(outDir string) []string {
	var out []string
	filepath.WalkDir(outDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(outDir, p)
		if rerr != nil || rel == manifestName {
			return nil
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	return out
}

// buildSSGFromFile builds the project's SSG site from the file-based layout
// (pages/, components/, layouts/, content/, public/) using ssg.LoadFromDir.
// krewire.yaml supplies metadata (title, description, theme) and output dir.
// After the core site is built, any detected plugins (e.g., Tailwind via
// tailwind.config.js) are run — they may write additional assets into outDir.
func buildSSGFromFile(root string, cfg *config.Config, fs *flag.FlagSet) core.ExitCode {
	output := firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output, config.DefaultOutput)
	outDir := joinRoot(root, output, config.DefaultOutput)
	slog.Info("building SSG site from file layout", "root", root, "output", outDir)
	site, err := ssg.LoadFromDir(root)
	if err != nil {
		return fail(err)
	}
	created, err := site.Build(outDir)
	if err != nil {
		return fail(err)
	}
	// Run detected plugins (Tailwind is the first; others follow the same pattern).
	for _, p := range plugin.Registry {
		if p.Detect(root) {
			slog.Info("plugin detected", "plugin", p.Name())
			if err := p.Build(root, outDir); err != nil {
				slog.Warn("plugin build failed", "plugin", p.Name(), "err", err)
			} else {
				slog.Info("plugin built", "plugin", p.Name())
			}
		}
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
	output := firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output, config.DefaultOutput)
	outDir := joinRoot(root, output, config.DefaultOutput)
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

// buildManuscript renders the project's content/ directory with mdbind.
// Settings come from krewire.yaml in the project root, overridden by flags.
// In a hybrid project (ssg pages present) the book suppresses its generated
// root TOC so the ssg landing page owns "/". Include/exclude path globs
// resolve flag > krewire.yaml `build:` > mdbind defaults (README skipped).
func buildManuscript(root string, cfg *config.Config, fs *flag.FlagSet, hybrid bool) core.ExitCode {
	title := firstNonEmpty(flagValue(fs, "title"), cfg.Title, moduleName(root))
	input := firstNonEmpty(flagValue(fs, "input"), cfg.Input)
	if input == "" {
		input = config.DefaultInput
		if !hasDir(root, input) && hasDir(root, "manuscript") {
			input = "manuscript"
		}
	}
	include := globList(fs, "include", cfg.Build.Include)
	exclude := globList(fs, "exclude", cfg.Build.Exclude)
	bcfg := book.Config{
		Input:      joinRoot(root, input, config.DefaultInput),
		Output:     joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput),
		Title:      title,
		Author:     firstNonEmpty(flagValue(fs, "author"), cfg.Author),
		BasePath:   firstNonEmpty(flagValue(fs, "base"), cfg.Base, "/"),
		Version:    cfg.Version,
		NavLinks:   navFromConfig(cfg.Nav),
		FooterText: cfg.Footer,
		Theme:      bookThemeFrom(fs, cfg.Theme),
		MountPath:  cfg.Book.Mount,
		NoRootTOC:  hybrid && cfg.Book.TOC == nil,
		Include:    include,
		Exclude:    exclude,
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

// globList resolves a repeatable/comma-separated glob flag over its
// krewire.yaml fallback. Unset everywhere returns nil (mdbind defaults);
// an explicitly empty value disables that filter side.
func globList(fs *flag.FlagSet, name string, fallback []string) []string {
	raw := flagValue(fs, name)
	if raw == "" {
		return fallback
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
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

// bookThemeFrom resolves the mdbind theme from the same flag/config as
// themeFrom but returns the book's local Theme type so mdbind can stay
// framework-free. The two themes stay in sync so a project can depend on
// both framework and mdbind and keep a single krewire.yaml.
func bookThemeFrom(fs *flag.FlagSet, cfg *config.Theme) *book.Theme {
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
	t := &book.Theme{Default: mode}
	if cfg != nil {
		t.Light = bookPalette(cfg.Light, book.DefaultLightPalette)
		t.Dark = bookPalette(cfg.Dark, book.DefaultDarkPalette)
	}
	return t
}

func bookPalette(p config.Palette, defaults book.Palette) book.Palette {
	out := defaults
	if p.Base1 != "" {
		out.Base1 = book.Color(p.Base1)
	}
	if p.Base1Content != "" {
		out.Base1Content = book.Color(p.Base1Content)
	}
	if p.Base2 != "" {
		out.Base2 = book.Color(p.Base2)
	}
	if p.Base2Content != "" {
		out.Base2Content = book.Color(p.Base2Content)
	}
	if p.Base3 != "" {
		out.Base3 = book.Color(p.Base3)
	}
	if p.Base3Content != "" {
		out.Base3Content = book.Color(p.Base3Content)
	}
	if p.Primary != "" {
		out.Primary = book.Color(p.Primary)
	}
	if p.PrimaryContent != "" {
		out.PrimaryContent = book.Color(p.PrimaryContent)
	}
	if p.Secondary != "" {
		out.Secondary = book.Color(p.Secondary)
	}
	if p.SecondaryContent != "" {
		out.SecondaryContent = book.Color(p.SecondaryContent)
	}
	if p.Accent != "" {
		out.Accent = book.Color(p.Accent)
	}
	if p.AccentContent != "" {
		out.AccentContent = book.Color(p.AccentContent)
	}
	if p.Ghost != "" {
		out.Ghost = book.Color(p.Ghost)
	}
	if p.GhostContent != "" {
		out.GhostContent = book.Color(p.GhostContent)
	}
	if p.Neutral != "" {
		out.Neutral = book.Color(p.Neutral)
	}
	if p.NeutralContent != "" {
		out.NeutralContent = book.Color(p.NeutralContent)
	}
	if p.Success != "" {
		out.Success = book.Color(p.Success)
	}
	if p.SuccessContent != "" {
		out.SuccessContent = book.Color(p.SuccessContent)
	}
	if p.Info != "" {
		out.Info = book.Color(p.Info)
	}
	if p.InfoContent != "" {
		out.InfoContent = book.Color(p.InfoContent)
	}
	if p.Warning != "" {
		out.Warning = book.Color(p.Warning)
	}
	if p.WarningContent != "" {
		out.WarningContent = book.Color(p.WarningContent)
	}
	if p.Error != "" {
		out.Error = book.Color(p.Error)
	}
	if p.ErrorContent != "" {
		out.ErrorContent = book.Color(p.ErrorContent)
	}
	return out
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
// content/, manuscript/) is sufficient (KWF-DF3PL FRK-FLS-001/002, KWL-K1N2Q).
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
		for _, marker := range []string{"content", "manuscript"} {
			if info, err := os.Stat(filepath.Join(cur, marker)); err == nil && info.IsDir() {
				return cur, nil
			}
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
