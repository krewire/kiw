package commands

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/kiw/internal/release"
	"github.com/krewire/kiw/internal/shape"
	"github.com/krewire/libs/core"
	"github.com/krewire/libs/term"
)

type releaseModuleSlice []string

func (s *releaseModuleSlice) String() string { return strings.Join(*s, ",") }
func (s *releaseModuleSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

var (
	releaseModules releaseModuleSlice
	releaseAll     bool
	releaseBump    string
	releaseApply   bool
	releaseNotes   bool
)

// RegisterRelease declares the flags for the release command.
func RegisterRelease(fs *flag.FlagSet) {
	fs.Var(&releaseModules, "module", "module to release; repeatable (framework, libs, kiw, mdbind, guild, ship); use with --all for every module")
	fs.BoolVar(&releaseAll, "all", false, "release every Krewire module (maintainer mode, run inside the workspace)")
	fs.StringVar(&releaseBump, "bump", "patch", "version component to increment: patch|minor|major")
	fs.BoolVar(&releaseApply, "apply", false, "write the changes (default is a dry run)")
	fs.BoolVar(&releaseNotes, "notes", false, "print release notes from git log since the last tag")
}

// RunRelease stages a release. There are two audiences:
//
//   - Maintainer mode: inside the Krewire workspace with --module/--all, it bumps
//     the released modules' versions and propagates the new minimum versions into
//     every dependent module's EcosystemRequires.
//   - Project mode: inside a Krewire project (krewire.yaml with project.kind) with
//     no --module flag, it bumps the project version in krewire.yaml and builds the
//     project, so ecosystem developers can release the apps/sites/books they build.
//
// It is a dry run unless --apply is given.
func RunRelease(fs *flag.FlagSet) core.ExitCode {
	bump, err := release.ParseBump(releaseBump)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return core.ExitCodeUsage
	}

	if releaseAll || len(releaseModules) > 0 {
		return runModuleRelease(bump)
	}
	return runProjectRelease(bump)
}

// runModuleRelease implements maintainer mode (Krewire's own modules).
func runModuleRelease(bump release.BumpKind) core.ExitCode {
	var names []core.ModuleName
	if releaseAll {
		for _, m := range release.Modules {
			names = append(names, m.Name)
		}
	} else {
		for _, s := range releaseModules {
			n := core.ModuleName(s)
			if !validReleaseModule(n) {
				fmt.Fprintf(os.Stderr, "release: unknown module %q\n", s)
				return core.ExitCodeUsage
			}
			names = append(names, n)
		}
	}

	edits, err := release.Plan(names, bump)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return core.ExitCodeFailure
	}

	var root string
	if releaseApply || releaseNotes {
		if root, err = findReleaseWorkspaceRoot(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return core.ExitCodeFailure
		}
	}

	tm := term.NewTerminal()
	label := "DRY RUN"
	if releaseApply {
		label = "APPLY"
	}
	fmt.Printf("%s: %d version.go change(s) for this module release\n", tm.Paint(label, term.ColorDefault, []term.Style{term.StyleBold}), len(edits))
	for _, e := range edits {
		fmt.Printf("  • %s\n    %s\n", tm.Paint(e.File, term.ColorCyan, nil), e.Summary)
	}

	if releaseApply {
		modified, err := release.Apply(edits, root, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return core.ExitCodeFailure
		}
		fmt.Printf("modified %d file(s)\n", len(modified))
	}

	fmt.Println("\nNext steps (run manually, or automate with a release workflow):")
	for _, n := range names {
		cur, _ := release.CurrentVersion(n)
		nv := release.Bump(cur, bump)
		m := release.ManifestFor(n)
		fmt.Printf("  cd %s && git add -A && git commit -m \"release: %s %s\" && git tag v%s\n", m.Dir, n, nv.String(), nv.String())
	}
	fmt.Println("  • remove local replace directives in kiw/go.mod before publishing (see AGENTS.md release rule)")
	fmt.Println("  • run `kiw compat` to confirm the ecosystem is still mutually compatible")

	if releaseNotes {
		printReleaseNotes(root, names)
	}
	return core.ExitCodeSuccess
}

// runProjectRelease implements developer mode: release the Krewire project in
// the current directory by bumping its version in krewire.yaml and building it.
func runProjectRelease(bump release.BumpKind) core.ExitCode {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return core.ExitCodeFailure
	}
	cfg, err := config.Load(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return core.ExitCodeFailure
	}
	res, _ := shape.Detect(dir, cfg.Kind())
	if res.Kind == shape.KindNone {
		fmt.Fprintln(os.Stderr, "release: not a Krewire project (no krewire.yaml with project.kind); run inside your project, or use --module/--all inside the Krewire workspace")
		return core.ExitCodeUsage
	}

	cur := cfg.Version
	prefix := ""
	stripped := strings.TrimPrefix(cur, "v")
	if cur != stripped {
		prefix = "v"
	}
	cv := core.Version{}
	if stripped == "" {
		cv = core.MustParseVersion("0.1.0")
	} else if cv, err = core.ParseVersion(stripped); err != nil {
		fmt.Fprintf(os.Stderr, "release: invalid project version %q in krewire.yaml: %v\n", cur, err)
		return core.ExitCodeUsage
	}
	nv := release.Bump(cv, bump)
	newVer := prefix + nv.String()

	tm := term.NewTerminal()
	label := "DRY RUN"
	if releaseApply {
		label = "APPLY"
	}
	fmt.Printf("%s: release %s project %q %s -> %s\n", tm.Paint(label, term.ColorDefault, []term.Style{term.StyleBold}), res.Kind, dir, cur, newVer)
	fmt.Printf("  • %s\n    %s\n", tm.Paint("krewire.yaml", term.ColorCyan, nil), fmt.Sprintf("version %s -> %s", cur, newVer))

	if releaseApply {
		if err := bumpKrewireVersion(dir, newVer); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return core.ExitCodeFailure
		}
		fmt.Println("  • building the project with `kiw build`")
		bfs := flag.NewFlagSet("build", flag.ContinueOnError)
		RegisterBuild(bfs)
		if code := RunBuild(bfs); code != core.ExitCodeSuccess {
			return code
		}
	}

	fmt.Printf("\nNext steps (run manually, or automate with a release workflow):\n")
	fmt.Printf("  git add krewire.yaml && git commit -m \"release: %s\" && git tag %s\n", newVer, newVer)
	if releaseNotes {
		if notes := gitLogNotes(dir); notes != "" {
			fmt.Printf("\nRelease notes:\n%s\n", notes)
		}
	}
	return core.ExitCodeSuccess
}

// bumpKrewireVersion sets the top-level `version:` field in dir/krewire.yaml to
// newVer, preserving the file's other content. It adds the line when missing.
func bumpKrewireVersion(dir, newVer string) error {
	p := filepath.Join(dir, "krewire.yaml")
	data, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("release: read %s: %w", p, err)
	}
	re := regexp.MustCompile(`(?m)^(version:\s*).*$`)
	updated := re.ReplaceAllString(string(data), `$1"`+newVer+`"`)
	if updated == string(data) {
		// No version: line present; append one.
		updated = string(data)
		if !strings.HasSuffix(updated, "\n") {
			updated += "\n"
		}
		updated += fmt.Sprintf("version: \"%s\"\n", newVer)
	}
	return os.WriteFile(p, []byte(updated), 0o644)
}

func validReleaseModule(n core.ModuleName) bool {
	for _, m := range release.Modules {
		if m.Name == n {
			return true
		}
	}
	return false
}

// findReleaseWorkspaceRoot walks up from the working directory to locate go.work.
func findReleaseWorkspaceRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("release: go.work not found (run from within the Krewire workspace)")
		}
		dir = parent
	}
}

// printReleaseNotes prints Conventional-Commits-style notes per released module.
func printReleaseNotes(root string, names []core.ModuleName) {
	for _, n := range names {
		m := release.ManifestFor(n)
		if notes := gitLogNotes(filepath.Join(root, m.Dir)); notes != "" {
			fmt.Printf("\nRelease notes for %s:\n%s\n", n, notes)
		}
	}
}

func gitLogNotes(dir string) string {
	last := lastGitTag(dir)
	rev := "HEAD"
	if last != "" {
		rev = last + "..HEAD"
	}
	out, err := exec.Command("git", "-C", dir, "log", rev, "--pretty=format:- %s").Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func lastGitTag(dir string) string {
	out, err := exec.Command("git", "-C", dir, "describe", "--tags", "--abbrev=0").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
