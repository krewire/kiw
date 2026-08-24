package commands

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/libs/core"
)

// RegisterDeploy registers flags for the deploy command.
func RegisterDeploy(fs *flag.FlagSet) {
	fs.String("target", "binary", "deploy target: binary or gh-pages")
}

// RunDeploy validates the project then stages a deployable artifact in dist/:
// the compiled app binary (binary target) and/or the exported site/ (gh-pages
// target). It never publishes; it only stages artifacts (RND-DEP-004).
func RunDeploy(fs *flag.FlagSet) core.ExitCode {
	target := flagValue(fs, "target")
	switch target {
	case "binary", "gh-pages":
	default:
		fmt.Fprintf(os.Stderr, "kiw deploy: unknown target %q — supported: binary, gh-pages\n", target)
		return core.ExitCodeUsage
	}

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

	if code := RunTest(nil); code != core.ExitCodeSuccess {
		return code
	}

	dist := joinRoot(root, "", "dist")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		return fail(err)
	}
	slog.Info("staging deploy artifacts", "dist", dist)

	if res.Kind == shape.KindApp && target == "binary" {
		if code := buildBinary(root, dist); code != core.ExitCodeSuccess {
			return code
		}
	}
	if cfg.IsSSG() || hasDir(root, "manuscript") || target == "gh-pages" {
		if code := exportSiteAssets(root, fs); code != core.ExitCodeSuccess {
			return code
		}
		siteOut := joinRoot(root, flagValue(fs, "output"), firstNonEmpty(cfg.Output, "site"))
		if err := copyDir(siteOut, filepath.Join(dist, "site")); err != nil {
			return fail(err)
		}
	}

	fmt.Println("deployed " + dist)
	return core.ExitCodeSuccess
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
