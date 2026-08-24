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

	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/libs/core"
)

// RegisterRun registers flags for the run command.
func RegisterRun(fs *flag.FlagSet) {
	fs.String("addr", "", "listen address for the app (default :8080)")
	registerRuntimeFlags(fs)
}

// RunRun builds and runs a fullstack app in production mode: it compiles the
// project into a temp binary, executes it passing the listen address through
// APP_ADDR, and forwards SIGINT/SIGTERM so the app's graceful shutdown runs
// inside the app. Site and book projects are rejected with guidance to use
// build + serve.
// RunRun builds and runs a fullstack app in production mode: it compiles the
// project into a temp binary, executes it passing the listen address through
// APP_ADDR, and forwards SIGINT/SIGTERM so the app's graceful shutdown runs
// inside the app. Site and book projects are rejected with guidance to use
// build + serve.
func RunRun(fs *flag.FlagSet) core.ExitCode {
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
	case shape.KindSite, shape.KindBook:
		fmt.Fprintf(os.Stderr, "kiw run: this is a %s project — build and serve it instead: 'kiw build' then 'kiw serve'\n", res.Kind)
		return core.ExitCodeUsage
	default:
		fmt.Fprintln(os.Stderr, "kiw run: no app found — add a main.go building a web.App, or run 'kiw new <project>' first")
		return core.ExitCodeUsage
	}
}

// runApp builds the app to a temp binary and runs it, streaming output and
// forwarding signals.
func runApp(rt *runtimeEnv, fs *flag.FlagSet) core.ExitCode {
	if rt.cfg.IsSSG() || hasDir(rt.root, "manuscript") {
		if code := exportSiteAssets(rt.root, fs); code != core.ExitCodeSuccess {
			return code
		}
	}

	dir, err := os.MkdirTemp("", "krewire-run-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "app")
	addr := firstNonEmpty(flagValue(fs, "addr"), ":8080")

	slog.Info("building app", "bin", bin)
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = rt.root
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fail(fmt.Errorf("build app: %w", err))
	}

	slog.Info("running app", "addr", addr)
	cmd := exec.Command(bin)
	cmd.Dir = rt.root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), addr, rt.env, rt.debug)
	if err := cmd.Start(); err != nil {
		return fail(err)
	}

	return waitChild(cmd)
}

// runCLI builds and runs a CLI project in production mode: it compiles the
// root main.go into a temp binary and executes it, forwarding SIGINT/SIGTERM
// so the CLI's own exit path resolves the process code. No listen address is
// injected for CLI projects.
func runCLI(rt *runtimeEnv, fs *flag.FlagSet) core.ExitCode {
	dir, err := os.MkdirTemp("", "krewire-run-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "app")

	slog.Info("building CLI", "bin", bin)
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = rt.root
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fail(fmt.Errorf("build cli: %w", err))
	}

	cmd := exec.Command(bin, fs.Args()...)
	cmd.Dir = rt.root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), "", rt.env, rt.debug)
	if err := cmd.Start(); err != nil {
		return fail(err)
	}

	return waitChild(cmd)
}

// waitChild forwards SIGINT/SIGTERM to the child process and resolves the
// exit code from the child's status.
func waitChild(cmd *exec.Cmd) core.ExitCode {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	for {
		select {
		case err := <-done:
			if err == nil {
				return core.ExitCodeSuccess
			}
			return childExitCode(cmd, err)
		case sig := <-sigCh:
			slog.Info("forwarding signal to app", "signal", sig)
			if err := cmd.Process.Signal(sig); err != nil {
				slog.Warn("signal forward failed", "error", err)
			}
		}
	}
}

// childExitCode resolves a completed child command to a canonical exit code,
// preferring the child's own exit status when available.
func childExitCode(cmd *exec.Cmd, err error) core.ExitCode {
	if cmd.Process != nil && cmd.ProcessState != nil {
		return core.ExitCodeFromInt(cmd.ProcessState.ExitCode())
	}
	if err != nil {
		return core.ExitCodeFailure
	}
	return core.ExitCodeSuccess
}

// exportSiteAssets refreshes the site/book export so embedded assets are
// current before the app starts (RND-RUN-005). No-op when no site is declared.
func exportSiteAssets(root string, fs *flag.FlagSet) core.ExitCode {
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	switch {
	case cfg.IsSSG():
		return buildSSGFromConfig(root, cfg, fs)
	case hasDir(root, "manuscript"):
		return buildManuscript(root, cfg, fs)
	default:
		return core.ExitCodeSuccess
	}
}
