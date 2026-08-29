package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
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
//
// Extended by KWN-SCRIPT-9F3KQ: when the first positional argument is a Go
// file path it runs `go run <file> -- args` (KWN-SCR-001); when it matches
// `scripts.<task>` in krewire.yaml it runs the shell command (KWN-SCR-002).
// Precedence is file > script > project (KWN-SCR-004).
func RunRun(fs *flag.FlagSet) core.ExitCode {
	rt, code := bootRuntime(fs)
	if code != core.ExitCodeSuccess {
		return code
	}

	args := fs.Args()
	if len(args) > 0 {
		first := args[0]
		isFile := isGoFileArg(rt.root, first)
		_, isScript := rt.cfg.Scripts[first]
		if isFile && isScript {
			fmt.Fprintf(os.Stderr, "warning: both file and task %q exist; running file\n", first)
		}
		if isFile {
			return runGoFile(rt, first, args[1:])
		}
		if isScript {
			return runScript(rt, first, args[1:])
		}
		if strings.HasSuffix(first, ".go") {
			fmt.Fprintf(os.Stderr, "kiw run: file %q not found\n", first)
			return core.ExitCodeUsage
		}
		if len(rt.cfg.Scripts) > 0 && isTaskLike(first) {
			// Unknown task — only error if project is not a runnable app/cli,
			// otherwise treat as CLI args for backward compat (KWN-SCR-003).
			// Peek detection without fully booting: check for main.go.
			if !hasDir(rt.root, "cmd") && !hasFile(rt.root, "main.go") {
				fmt.Fprintf(os.Stderr, "kiw run: unknown task %q\nAvailable tasks: %s\n", first, formatTasks(rt.cfg.Scripts))
				return core.ExitCodeUsage
			}
		}
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
	hasPages := hasDir(root, "pages")
	hasSSG := cfg.IsSSG()
	hasBook := hasDir(root, "content") || hasDir(root, "manuscript")
	if !hasPages && !hasSSG && !hasBook {
		return core.ExitCodeSuccess
	}
	// Mirror RunBuild: SSG from file (pages/) takes precedence over ssg: config
	if hasPages {
		if code := buildSSGFromFile(root, cfg, fs); code != core.ExitCodeSuccess {
			return code
		}
	} else if hasSSG {
		if code := buildSSGFromConfig(root, cfg, fs); code != core.ExitCodeSuccess {
			return code
		}
	}
	if hasBook {
		if code := buildManuscript(root, cfg, fs, hasPages || hasSSG); code != core.ExitCodeSuccess {
			return code
		}
	}
	return core.ExitCodeSuccess
}

// Tests for KWN-SCRIPT-9F3KQ
// Spec: KWN-SCRIPT-9F3KQ KWN-SCR-001 Scope: Module
func isGoFileArg(root, arg string) bool {
	if !strings.HasSuffix(arg, ".go") {
		return false
	}
	path := arg
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, arg)
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Spec: KWN-SCRIPT-9F3KQ KWN-SCR-001 Scope: Module
func runGoFile(rt *runtimeEnv, path string, extraArgs []string) core.ExitCode {
	if len(extraArgs) > 0 && extraArgs[0] == "--" {
		extraArgs = extraArgs[1:]
	}
	args := []string{"run", path}
	args = append(args, extraArgs...)
	slog.Info("running go file", "path", path)
	cmd := exec.Command("go", args...)
	cmd.Dir = rt.root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), "", rt.env, rt.debug)
	if err := cmd.Start(); err != nil {
		return fail(err)
	}
	return waitChild(cmd)
}

// Spec: KWN-SCRIPT-9F3KQ KWN-SCR-002 Scope: Module
func runScript(rt *runtimeEnv, task string, extraArgs []string) core.ExitCode {
	script, ok := rt.cfg.Scripts[task]
	if !ok {
		fmt.Fprintf(os.Stderr, "kiw run: unknown task %q\nAvailable tasks: %s\n", task, formatTasks(rt.cfg.Scripts))
		return core.ExitCodeUsage
	}
	if len(extraArgs) > 0 && extraArgs[0] == "--" {
		extraArgs = extraArgs[1:]
	}
	if len(extraArgs) > 0 {
		script = script + " " + strings.Join(extraArgs, " ")
	}
	slog.Info("running script", "task", task, "command", script)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", script)
	} else {
		cmd = exec.Command("sh", "-c", script)
	}
	cmd.Dir = rt.root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = childEnviron(os.Environ(), "", rt.env, rt.debug)
	if err := cmd.Start(); err != nil {
		return fail(err)
	}
	return waitChild(cmd)
}

// Spec: KWN-SCRIPT-9F3KQ KWN-SCR-004 Scope: Module
func isTaskLike(s string) bool {
	if s == "" {
		return false
	}
	if strings.Contains(s, "/") || strings.Contains(s, "\\") {
		return false
	}
	if strings.HasSuffix(s, ".go") {
		return false
	}
	return true
}

func formatTasks(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
