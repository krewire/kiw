package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/libs/core"
)

// RegisterDev registers flags for the dev command.
func RegisterDev(fs *flag.FlagSet) {
	fs.String("addr", "", "listen address for the app (default :8080)")
	fs.Duration("interval", 500*time.Millisecond, "file-watch polling interval")
	registerRuntimeFlags(fs)
}

// RunDev runs the project in dev mode. For an app it rebuilds and restarts
// the child on change; for site/book projects it behaves exactly like serve.
func RunDev(fs *flag.FlagSet) core.ExitCode {
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
	case shape.KindApp, shape.KindCLI:
		return devApp(rt, fs)
	case shape.KindSite:
		if rt.cfg.IsSSG() {
			return serveSSG(fs, rt.cfg)
		}
		fmt.Fprintln(os.Stderr, "kiw dev: site detected without an ssg: config")
		return core.ExitCodeUsage
	case shape.KindBook:
		return serveBook(rt.root, fs)
	default:
		fmt.Fprintln(os.Stderr, "kiw dev: no project found — run 'kiw new <project>' first")
		return core.ExitCodeUsage
	}
}

// devApp runs the app, watching the module and declared asset/markup roots,
// rebuilding and restarting on change. A failed rebuild keeps the previous
// child running (RND-DEV-002).
func devApp(rt *runtimeEnv, fs *flag.FlagSet) core.ExitCode {
	root, cfg := rt.root, rt.cfg
	env, debug := rt.env, rt.debug
	dir, err := os.MkdirTemp("", "krewire-dev-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "app")
	addr := firstNonEmpty(flagValue(fs, "addr"), ":8080")
	interval := fs.Lookup("interval").Value
	every := 500 * time.Millisecond
	if d, err := time.ParseDuration(interval.String()); err == nil && d > 0 {
		every = d
	}

	build := func() error {
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = root
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		slog.Info("dev build")
		return cmd.Run()
	}
	if err := build(); err != nil {
		return fail(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	watcher := newWatcher(root, every, cfg)
	restart := func() { restartApp(bin, root, addr, env, debug) }

	restart()
	slog.Info("watching for changes", "root", root, "interval", every)
	for {
		select {
		case sig := <-sigCh:
			slog.Info("dev received signal, stopping app", "signal", sig)
			stopChild()
			return core.ExitCodeSuccess
		case <-watcher.Changed():
			watcher.Reset()
			slog.Info("change detected")
			if err := build(); err != nil {
				slog.Error("dev build failed; keeping previous process running", "error", err)
				continue
			}
			stopChild()
			restart()
		}
	}
}
