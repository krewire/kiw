package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/libs/core"
	"github.com/krewire/mdbind/book"
)

// RegisterServe registers flags for the serve command.
func RegisterServe(fs *flag.FlagSet) {
	fs.String("input", "", "content directory (default content)")
	fs.String("addr", "", "listen address (default :8080)")
	registerRuntimeFlags(fs)
}

// RunServe starts the current project locally for every kind
// (KWN-6K41E RND-SRV-001): an app is compiled and listened on, a cli
// executes with argument passthrough, and site/book projects preview their
// static output over HTTP. A hybrid project (ssg + book) serves the merged
// .krewire/build statically. It shares the runtime bootstrap with run and
// dev (RND-SRV-002).
func RunServe(fs *flag.FlagSet) core.ExitCode {
	rt, code := bootRuntime(fs)
	if code != core.ExitCodeSuccess {
		return code
	}
	explicit := firstNonEmpty(flagValue(fs, "kind"), rt.cfg.Kind())
	res, err := shape.Detect(rt.root, explicit)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: "+err.Error())
		return core.ExitCodeUsage
	}

	hasSSGShape := hasDir(rt.root, "pages") || rt.cfg.IsSSG()
	hasBookShape := hasDir(rt.root, "content") || hasDir(rt.root, "manuscript")

	switch res.Kind {
	case shape.KindApp:
		return runApp(rt, fs)
	case shape.KindCLI:
		return runCLI(rt, fs)
	case shape.KindSite:
		if hasSSGShape && hasBookShape {
			return serveStaticBuild(rt.root, fs)
		}
		if rt.cfg.IsSSG() {
			return serveSSG(fs, rt.cfg)
		}
		if hasDir(rt.root, "pages") {
			return serveStaticBuild(rt.root, fs)
		}
		fmt.Fprintln(os.Stderr, "kiw serve: site detected without an ssg: config")
		return core.ExitCodeUsage
	case shape.KindBook:
		if hasSSGShape && hasBookShape {
			return serveStaticBuild(rt.root, fs)
		}
		return serveBook(rt.root, fs)
	default:
		fmt.Fprintln(os.Stderr, "kiw serve: no project found — add a root main.go, an `ssg:` key in krewire.yaml, or a content/ directory")
		return core.ExitCodeUsage
	}
}

// serveBook previews the content directory with the mdbind handler. The
// input resolves like every other setting: --input flag > krewire.yaml
// `input:` > config.DefaultInput ("content"). Include/exclude globs follow
// the same precedence as kiw build so preview matches the export.
func serveBook(root string, fs *flag.FlagSet) core.ExitCode {
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	input := joinRoot(root, flagValue(fs, "input"), firstNonEmpty(cfg.Input, config.DefaultInput))
	addr := firstNonEmpty(flagValue(fs, "addr"), ":8080")
	b, err := book.LoadWithRules(
		input, "", "", "/",
		globList(fs, "include", cfg.Build.Include),
		globList(fs, "exclude", cfg.Build.Exclude),
	)
	if err != nil {
		return fail(err)
	}
	slog.Info("serving book", "addr", addr)
	if err := book.Serve(b, addr); err != nil {
		return fail(err)
	}
	return core.ExitCodeSuccess
}

// serveStaticBuild previews the built .krewire/build output over HTTP — the
// shared preview for hybrid (ssg + book) projects and file-mode sites.
func serveStaticBuild(root string, fs *flag.FlagSet) core.ExitCode {
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	outDir := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
	if info, err := os.Stat(filepath.Join(outDir, "index.html")); err != nil || info.IsDir() {
		fmt.Fprintf(os.Stderr, "kiw serve: no build output at %s — run 'kiw build' first\n", outDir)
		return core.ExitCodeUsage
	}
	addr := firstNonEmpty(flagValue(fs, "addr"), ":8080")
	slog.Info("serving static build", "dir", outDir, "addr", addr)
	if err := http.ListenAndServe(addr, extensionlessFS(outDir)); err != nil {
		return fail(err)
	}
	return core.ExitCodeSuccess
}

// extensionlessFS serves the build output with extensionless URL resolution,
// mirroring the emitted sibling .html files: /docs → docs.html and
// /docs/x → docs/x.html. Directory listings are never shown.
func extensionlessFS(root string) http.Handler {
	fs := http.Dir(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean("/" + r.URL.Path)
		if p == "/" {
			p = "/index.html"
		}
		for _, candidate := range []string{p, p + ".html"} {
			f, err := fs.Open(candidate)
			if err != nil {
				continue
			}
			st, statErr := f.Stat()
			closeErr := f.Close()
			if statErr != nil || closeErr != nil || st.IsDir() {
				continue
			}
			http.ServeFile(w, r, filepath.Join(root, filepath.FromSlash(candidate)))
			return
		}
		http.NotFound(w, r)
	})
}

// serveSSG previews the project's declarative ssg site from krewire.yaml.
func serveSSG(fs *flag.FlagSet, cfg *config.Config) core.ExitCode {
	ssgCfg := cfg.ToSSGConfig()
	addr := firstNonEmpty(flagValue(fs, "addr"), ":8080")
	slog.Info("serving site", "addr", addr)
	if err := http.ListenAndServe(addr, ssgCfg.Site().Handler()); err != nil {
		return fail(err)
	}
	return core.ExitCodeSuccess
}
