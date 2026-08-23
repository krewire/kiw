package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/krewire/kiw/internal/gomod"
	"github.com/krewire/libs/core"
)

func RunTest(fs *flag.FlagSet) core.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	mod, err := gomod.Find(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: not inside a Go module — run 'kiw new <project>' first")
		return core.ExitCodeUsage
	}
	slog.Info("running tests", "module", mod.Path)
	args := []string{"test", "./..."}
	if fs != nil && len(fs.Args()) > 0 {
		args = append([]string{"test"}, fs.Args()...)
		if !containsGoPackageArg(fs.Args()) {
			hasPkg := false
			for _, a := range fs.Args() {
				if !strings.HasPrefix(a, "-") {
					hasPkg = true
					break
				}
			}
			if !hasPkg {
				args = append(args, "./...")
			}
		}
	}
	cmd := exec.Command("go", args...) // #nosec G204 — user-invoked devtool
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return core.ExitCodeFailure
	}
	return core.ExitCodeSuccess
}

func RunVet(fs *flag.FlagSet) core.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	mod, err := gomod.Find(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: not inside a Go module — run 'kiw new <project>' first")
		return core.ExitCodeUsage
	}
	slog.Info("running vet", "module", mod.Path)
	args := []string{"vet", "./..."}
	if fs != nil && len(fs.Args()) > 0 {
		args = append([]string{"vet"}, fs.Args()...)
	}
	cmd := exec.Command("go", args...) // #nosec G204 — user-invoked devtool
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return core.ExitCodeFailure
	}
	return core.ExitCodeSuccess
}

func RegisterFmt(fs *flag.FlagSet) {
	fs.Bool("write", false, "write formatted files back to disk (default: check only with gofmt -l)")
	fs.Bool("w", false, "alias for --write")
}

func RunFmt(fs *flag.FlagSet) core.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	root := dir
	if mod, err := gomod.Find(dir); err == nil {
		if mdir := moduleDir(dir); mdir != "" {
			root = mdir
		} else {
			_ = mod
		}
	}
	write := flagValue(fs, "write") == "true" || flagValue(fs, "w") == "true"
	if fs != nil {
		for _, a := range fs.Args() {
			if a == "--write" || a == "-w" || a == "write" {
				write = true
				break
			}
		}
	}
	if write {
		slog.Info("formatting code", "root", root, "mode", "write")
		cmd := exec.Command("gofmt", "-w", ".") // #nosec G204 — user-invoked devtool
		cmd.Dir = root
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fallback := exec.Command("go", "fmt", "./...") // #nosec G204
			fallback.Dir = root
			fallback.Stdout = os.Stdout
			fallback.Stderr = os.Stderr
			if ferr := fallback.Run(); ferr != nil {
				return fail(err)
			}
			return core.ExitCodeSuccess
		}
		return core.ExitCodeSuccess
	}
	slog.Info("checking format", "root", root)
	cmd := exec.Command("gofmt", "-l", ".") // #nosec G204 — user-invoked devtool
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		if len(out) > 0 {
			fmt.Fprint(os.Stdout, string(out))
		}
		return fail(err)
	}
	if len(out) > 0 {
		fmt.Fprint(os.Stdout, string(out))
		fmt.Fprintln(os.Stderr, "kiw fmt: some files need formatting — run 'kiw fmt --write' or 'gofmt -w .'")
		return core.ExitCodeFailure
	}
	return core.ExitCodeSuccess
}

func moduleDir(dir string) string {
	cur := dir
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
}

func containsGoPackageArg(args []string) bool {
	for _, a := range args {
		if strings.Contains(a, "/") || strings.HasPrefix(a, "./") || a == "." || a == "..." {
			return true
		}
	}
	return false
}
