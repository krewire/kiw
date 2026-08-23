package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/krewire/kiw/internal/gomod"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/libs/core"
)

func RegisterWs(fs *flag.FlagSet) {}

func RunWs(fs *flag.FlagSet) core.ExitCode {
	sub := fs.Arg(0)
	args := fs.Args()
	switch sub {
	case "info":
		return runWsInfo(fs)
	case "list", "ls":
		return runWsList(fs)
	case "add":
		return runWsAdd(fs, args[1:])
	case "remove", "rm":
		return runWsRemove(fs, args[1:])
	case "sync":
		return runWsSync(fs)
	case "exec":
		return runWsExec(fs, args[1:])
	case "", "help", "-h", "--help":
		return wsHelp()
	default:
		fmt.Fprintf(os.Stderr, "kiw: unknown ws sub-command %q\n", sub)
		return wsHelp()
	}
}

func wsHelp() core.ExitCode {
	fmt.Fprint(os.Stdout, `kiw ws — workspace for monorepo, multi-repo, multi-project, microservices

Usage:
  kiw ws info                    show workspace type and members
  kiw ws list                   list projects with kind and module
  kiw ws add <path>             add module to go.work (multi-repo)
  kiw ws remove <path>          remove module from go.work
  kiw ws sync                   run go work sync
  kiw ws exec <cmd> -- [args]   run command in each workspace member

Examples:
  kiw ws info
  kiw ws list
  kiw ws add ./services/payment
  kiw ws exec -- go test ./...
  kiw ws exec -- go vet ./...

Workspace types:
  monorepo  — single go.mod, no go.work (e.g., single service)
  multi-repo — go.work with use directives (hub with framework/libs/kiw)
  multi-project — go.work members have different kinds (app, service, worker)
  microservices — multiple kind=service members
`)
	return core.ExitCodeUsage
}

func findWorkspaceRoot(dir string) (string, string, []string) {
	// returns root, type, members
	cur, err := filepath.Abs(dir)
	if err != nil {
		return dir, "unknown", nil
	}
	for {
		if _, err := os.Stat(filepath.Join(cur, "go.work")); err == nil {
			members := parseGoWork(filepath.Join(cur, "go.work"))
			return cur, "multi-repo (go.work)", members
		}
		if _, err := os.Stat(filepath.Join(cur, "go.mod")); err == nil {
			parent := filepath.Dir(cur)
			if _, err := os.Stat(filepath.Join(parent, "go.work")); err == nil {
				members := parseGoWork(filepath.Join(parent, "go.work"))
				return parent, "multi-repo (go.work)", members
			}
			return cur, "monorepo (go.mod)", []string{cur}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return dir, "unknown", nil
		}
		cur = parent
	}
}

func parseGoWork(path string) []string {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var members []string
	lines := strings.Split(string(b), "\n")
	inUse := false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "use (" {
			inUse = true
			continue
		}
		if inUse && t == ")" {
			inUse = false
			continue
		}
		if inUse {
			t = strings.Trim(t, `"`)
			if t != "" && !strings.HasPrefix(t, "//") {
				members = append(members, t)
			}
		} else if strings.HasPrefix(t, "use ") {
			v := strings.TrimSpace(strings.TrimPrefix(t, "use"))
			v = strings.Trim(v, `"`)
			if v != "" {
				members = append(members, v)
			}
		}
	}
	return members
}

func runWsInfo(fs *flag.FlagSet) core.ExitCode {
	dir, _ := os.Getwd()
	root, typ, members := findWorkspaceRoot(dir)
	fmt.Fprintf(os.Stdout, "workspace: %s\n", root)
	fmt.Fprintf(os.Stdout, "type: %s\n", typ)
	if len(members) == 0 {
		fmt.Fprintln(os.Stdout, "members: (none)")
	} else {
		fmt.Fprintln(os.Stdout, "members:")
		for _, m := range members {
			fmt.Fprintf(os.Stdout, "  - %s\n", m)
		}
	}
	goWork := filepath.Join(root, "go.work")
	if _, err := os.Stat(goWork); err == nil {
		fmt.Fprintf(os.Stdout, "go.work: %s\n", goWork)
	}
	return core.ExitCodeSuccess
}

func runWsList(fs *flag.FlagSet) core.ExitCode {
	dir, _ := os.Getwd()
	root, _, members := findWorkspaceRoot(dir)
	if len(members) == 0 {
		members = []string{"."}
		root, _ = filepath.Abs(dir)
	}
	fmt.Fprintf(os.Stdout, "%-30s %-12s %s\n", "PROJECT", "KIND", "MODULE")
	fmt.Fprintf(os.Stdout, "%-30s %-12s %s\n", strings.Repeat("-", 30), strings.Repeat("-", 12), strings.Repeat("-", 30))
	for _, m := range members {
		abs := m
		if !filepath.IsAbs(m) {
			abs = filepath.Join(root, m)
		}
		mod, err := gomod.Read(filepath.Join(abs, "go.mod"))
		modPath := "-"
		if err == nil {
			modPath = mod.Path
		}
		kind := "-"
		if res, err := shape.Detect(abs, ""); err == nil && res.Kind != "" {
			kind = string(res.Kind)
		}
		rel := m
		if abs == root {
			rel = "."
		}
		fmt.Fprintf(os.Stdout, "%-30s %-12s %s\n", rel, kind, modPath)
	}
	return core.ExitCodeSuccess
}

func runWsAdd(fs *flag.FlagSet, args []string) core.ExitCode {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "kiw ws add: missing <path> (e.g., kiw ws add ./services/payment)")
		return core.ExitCodeUsage
	}
	dir, _ := os.Getwd()
	root, typ, _ := findWorkspaceRoot(dir)
	if !strings.Contains(typ, "go.work") {
		fmt.Fprintln(os.Stderr, "kiw ws add: no go.work found — run 'go work init' first or use monorepo")
		return core.ExitCodeUsage
	}
	path := args[0]
	cmd := exec.Command("go", "work", "use", path)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fail(err)
	}
	fmt.Fprintf(os.Stdout, "added %s to go.work\n", path)
	return core.ExitCodeSuccess
}

func runWsRemove(fs *flag.FlagSet, args []string) core.ExitCode {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "kiw ws remove: missing <path>")
		return core.ExitCodeUsage
	}
	dir, _ := os.Getwd()
	root, typ, _ := findWorkspaceRoot(dir)
	if !strings.Contains(typ, "go.work") {
		fmt.Fprintln(os.Stderr, "kiw ws remove: no go.work found")
		return core.ExitCodeUsage
	}
	path := args[0]
	cmd := exec.Command("go", "work", "edit", "-dropuse", path)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fail(err)
	}
	fmt.Fprintf(os.Stdout, "removed %s from go.work\n", path)
	return core.ExitCodeSuccess
}

func runWsSync(fs *flag.FlagSet) core.ExitCode {
	dir, _ := os.Getwd()
	root, _, _ := findWorkspaceRoot(dir)
	cmd := exec.Command("go", "work", "sync")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fail(err)
	}
	fmt.Fprintln(os.Stdout, "go work sync done")
	return core.ExitCodeSuccess
}

func runWsExec(fs *flag.FlagSet, args []string) core.ExitCode {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "kiw ws exec: missing command (e.g., kiw ws exec -- go test ./...)")
		return core.ExitCodeUsage
	}
	// strip leading -- if present
	if args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "kiw ws exec: missing command after --")
		return core.ExitCodeUsage
	}
	dir, _ := os.Getwd()
	root, _, members := findWorkspaceRoot(dir)
	if len(members) == 0 {
		members = []string{root}
	}
	var failed []string
	for _, m := range members {
		abs := m
		if !filepath.IsAbs(m) {
			abs = filepath.Join(root, m)
		}
		if _, err := os.Stat(filepath.Join(abs, "go.mod")); err != nil {
			continue
		}
		fmt.Fprintf(os.Stdout, "\n==> %s: %s\n", m, strings.Join(args, " "))
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = abs
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "kiw ws exec: %s failed: %v\n", m, err)
			failed = append(failed, m)
		}
	}
	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "kiw ws exec: %d member(s) failed: %s\n", len(failed), strings.Join(failed, ", "))
		return core.ExitCodeFailure
	}
	return core.ExitCodeSuccess
}
