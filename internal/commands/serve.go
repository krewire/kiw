package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/libs/core"
	"github.com/krewire/mdbind/book"
)

// RegisterServe registers flags for the serve command.
func RegisterServe(fs *flag.FlagSet) {
	fs.String("input", "", "manuscript directory (default manuscript)")
	fs.String("addr", "", "listen address (default :8080)")
	registerRuntimeFlags(fs)
}

// RunServe starts the current project locally for every kind
// (KWN-6K41E RND-SRV-001): an app is compiled and listened on, a cli
// executes with argument passthrough, and site/book projects preview their
// static output over HTTP. It shares the runtime bootstrap with run and dev
// (RND-SRV-002).
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

	switch res.Kind {
	case shape.KindApp:
		return runApp(rt, fs)
	case shape.KindCLI:
		return runCLI(rt, fs)
	case shape.KindSite:
		if rt.cfg.IsSSG() {
			return serveSSG(fs, rt.cfg)
		}
		fmt.Fprintln(os.Stderr, "kiw serve: site detected without an ssg: config")
		return core.ExitCodeUsage
	case shape.KindBook:
		return serveBook(rt.root, fs)
	default:
		fmt.Fprintln(os.Stderr, "kiw serve: no project found — add a root main.go, an `ssg:` key in krewire.yaml, or a manuscript/ directory")
		return core.ExitCodeUsage
	}
}

// serveBook previews a manuscript/ directory with the mdbind web router.
func serveBook(root string, fs *flag.FlagSet) core.ExitCode {
	input := joinRoot(root, flagValue(fs, "input"), "manuscript")
	addr := firstNonEmpty(flagValue(fs, "addr"), ":8080")
	b, err := book.Load(input, "", "")
	if err != nil {
		return fail(err)
	}
	slog.Info("serving book", "addr", addr)
	if err := book.Serve(b, addr); err != nil {
		return fail(err)
	}
	return core.ExitCodeSuccess
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
