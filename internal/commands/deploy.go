package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/libs/core"
)

// pagesBranches lists the branches probed (in order) for site publishing
// when --branch is unset.
var pagesBranches = []string{"gh-deploy", "gh-pages"}

// RegisterDeploy registers flags for the deploy command.
func RegisterDeploy(fs *flag.FlagSet) {
	fs.String("target", "binary", "deploy target: binary, gh-pages, or infra")
	fs.String("remote", "origin", "git remote to publish the pages branch to")
	fs.String("branch", "", "pages branch to publish (default: existing gh-deploy or gh-pages, else gh-pages)")
	fs.Bool("dry-run", false, "stage dist/ but skip publishing and tests")
	fs.String("message", "", "commit message for the pages publication")
	fs.Bool("plan", false, "infra: show plan without applying (KWF-B7N3D)")
	fs.Bool("auto-approve", false, "infra: skip confirmation prompt")
	fs.Bool("destroy", false, "infra: tear down infrastructure")
	fs.String("env", "", "infra: target environment (local, production, testing)")
}

// RunDeploy validates the project then stages a deployable artifact in
// dist/: the compiled app binary (binary target) and/or the exported site.
// The gh-pages target additionally publishes the built site to the project's
// pages branch on the configured remote (--dry-run skips both tests and
// publishing). The infra target uses framework/infra for plan/apply (KWF-B7N3D).
func RunDeploy(fs *flag.FlagSet) core.ExitCode {
	target := flagValue(fs, "target")
	switch target {
	case "binary", "gh-pages", "infra":
	default:
		fmt.Fprintf(os.Stderr, "kiw deploy: unknown target %q — supported: binary, gh-pages, infra\n", target)
		return core.ExitCodeUsage
	}
	dryRun := flagBool(fs, "dry-run")

	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: "+err.Error())
		return core.ExitCodeUsage
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}
	res, err := shape.Detect(root, cfg.Kind())
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: "+err.Error())
		return core.ExitCodeUsage
	}
	if res.Kind == shape.KindNone {
		fmt.Fprintln(os.Stderr, "kiw deploy: no project found — run 'kiw new <project>' first")
		return core.ExitCodeUsage
	}

	if target == "infra" {
		return runInfraDeploy(root, cfg, fs, dryRun)
	}

	if !dryRun && hasFile(root, "go.mod") {
		if code := RunTest(nil); code != core.ExitCodeSuccess {
			return code
		}
	}

	dist := joinRoot(root, "", ".krewire/dist")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		return fail(err)
	}
	slog.Info("staging deploy artifacts", "dist", dist)

	if res.Kind == shape.KindApp && target == "binary" {
		if code := buildBinary(root, dist); code != core.ExitCodeSuccess {
			return code
		}
	}
	siteStaged := ""
	if cfg.IsSSG() || hasDir(root, "content") || hasDir(root, "manuscript") || target == "gh-pages" {
		if code := exportSiteAssets(root, fs); code != core.ExitCodeSuccess {
			return code
		}
		siteOut := joinRoot(root, firstNonEmpty(flagValue(fs, "output"), flagValue(fs, "o"), cfg.Output), config.DefaultOutput)
		if err := copyDir(siteOut, filepath.Join(dist, "site")); err != nil {
			return fail(err)
		}
		siteStaged = filepath.Join(dist, "site")
	}

	if target == "gh-pages" && !dryRun && siteStaged != "" {
		return publishPages(root, siteStaged, fs)
	}

	fmt.Println("deployed " + dist)
	return core.ExitCodeSuccess
}

// runInfraDeploy handles infra target deployment (KWF-B7N3D FRK-INFRA-050/051).
func runInfraDeploy(root string, cfg *config.Config, fs *flag.FlagSet, dryRun bool) core.ExitCode {
	plan := flagBool(fs, "plan")
	destroy := flagBool(fs, "destroy")
	autoApprove := flagBool(fs, "auto-approve")
	env := firstNonEmpty(flagValue(fs, "env"), cfg.Env, "local")

	slog.Info("infra deploy", "env", env, "plan", plan, "destroy", destroy)

	// For now, infra deploy is a placeholder that shows what would happen.
	// Full implementation requires provider-specific SDK integration.
	fmt.Printf("infra deploy: env=%s plan=%t destroy=%t auto-approve=%t\n", env, plan, destroy, autoApprove)
	if plan {
		fmt.Println("  plan: would show resource diff (not yet implemented)")
	}
	if destroy {
		fmt.Println("  destroy: would tear down infrastructure (not yet implemented)")
	}
	if dryRun {
		fmt.Println("  dry-run: no changes made")
		return core.ExitCodeSuccess
	}
	if !autoApprove && !plan {
		fmt.Println("  use --auto-approve to apply changes, or --plan to preview")
	}
	return core.ExitCodeSuccess
}

// publishPages publishes the staged site to the project's pages branch using
// a temporary throwaway clone of remote — the user's working tree is never
// touched. When the branch already exists remotely it is rebased onto; a new
// orphan history is created otherwise.
func publishPages(root, src string, fs *flag.FlagSet) core.ExitCode {
	remoteName := flagValue(fs, "remote")
	branch := flagValue(fs, "branch")
	if branch == "" {
		b, err := detectPagesBranch(root, remoteName, fs)
		if err != nil {
			return fail(err)
		}
		branch = b
	}
	url, err := gitOutput(root, "remote", "get-url", remoteName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kiw deploy: remote %q not found in %s — add it first: git remote add origin <url>\n", remoteName, root)
		return core.ExitCodeFailure
	}

	tmp, err := os.MkdirTemp("", "kiw-ghpages-")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(tmp)

	run := func(args ...string) error {
		c := exec.Command("git", args...)
		c.Dir = tmp
		c.Stdout = os.Stderr
		c.Stderr = os.Stderr
		return c.Run()
	}
	exists := true
	if err := run("-c", "advice.detachedHead=false", "clone", "--quiet", "--depth", "1", "--branch", branch, url, "."); err != nil {
		exists = false
		if err := run("init", "--quiet", "--initial-branch", branch); err != nil {
			return fail(err)
		}
		if err := run("remote", "add", "origin", url); err != nil {
			return fail(err)
		}
	}
	clearWorktree(tmp)
	if err := copyDir(src, tmp); err != nil {
		return fail(err)
	}
	os.Remove(filepath.Join(tmp, manifestName))
	if err := run("add", "-A"); err != nil {
		return fail(err)
	}
	msg := flagValue(fs, "message")
	if msg == "" {
		msg = "deploy: publish site from kiw"
	}
	status, err := gitOutput(tmp, "status", "--porcelain")
	if err == nil && strings.TrimSpace(status) == "" {
		slog.Info("pages branch already up to date", "branch", branch)
	} else if err := commitPages(run, msg); err != nil {
		return fail(err)
	} else if !exists {
		slog.Info("created fresh pages history", "branch", branch)
	}
	if err := run("push", "origin", branch); err != nil {
		return fail(err)
	}
	slog.Info("published site", "remote", remoteName, "branch", branch)
	fmt.Printf("published %s -> %s#%s\n", src, remoteName, branch)
	return core.ExitCodeSuccess
}

// commitPages commits staged changes with the fixed Krewire Bot identity so
// publication never depends on the user's git config.
func commitPages(run func(...string) error, msg string) error {
	return run("-c", "user.name=Krewire Bot", "-c", "user.email=krewire-bot@krewire.local",
		"commit", "--quiet", "-m", msg)
}

// clearWorktree removes every tracked/untracked entry except .git so the
// next publication starts from a clean slate.
func clearWorktree(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		os.RemoveAll(filepath.Join(dir, e.Name()))
	}
}

// detectPagesBranch picks an existing remote pages branch; falls back to
// "gh-pages".
func detectPagesBranch(root, remote string, fs *flag.FlagSet) (string, error) {
	for _, b := range pagesBranches {
		out, err := gitOutput(root, "ls-remote", "--heads", remote, b)
		if err != nil {
			continue
		}
		if strings.TrimSpace(out) != "" {
			return b, nil
		}
	}
	return "gh-pages", nil
}

// gitOutput runs git in dir and returns trimmed stdout.
func gitOutput(dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.Output()
	return strings.TrimSpace(string(out)), err
}

// buildBinary compiles the project into dist/<module-name>.
func buildBinary(root, dist string) core.ExitCode {
	name := moduleName(root)
	out := filepath.Join(dist, name)
	slog.Info("building app binary", "out", out)
	cmd := exec.Command("go", "build", "-o", out, ".")
	cmd.Dir = root
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fail(err)
	}
	return core.ExitCodeSuccess
}

// copyDir copies src into dst recursively.
func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("copy %s: %w", src, err)
	}
	return copyTree(src, dst, info)
}

func copyTree(src, dst string, info os.FileInfo) error {
	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, e := range entries {
			childSrc := filepath.Join(src, e.Name())
			childDst := filepath.Join(dst, e.Name())
			childInfo, err := e.Info()
			if err != nil {
				return err
			}
			if err := copyTree(childSrc, childDst, childInfo); err != nil {
				return err
			}
		}
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}
