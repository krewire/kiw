package commands

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/libs/core"
)

func TestRunReleaseDryRunDoesNotMutate(t *testing.T) {
	// Reset the package-level release flags.
	releaseModules = releaseModuleSlice{}
	releaseAll = false
	releaseBump = "patch"
	releaseApply = false
	releaseNotes = false

	releaseModules = append(releaseModules, string(core.ModuleLibs))

	if got := RunRelease(flag.NewFlagSet("release", flag.ContinueOnError)); got != core.ExitCodeSuccess {
		t.Errorf("RunRelease dry-run = %v, want ExitCodeSuccess", got)
	}
	// Dry run must not request a workspace root or mutate files; success implies that.
}

func TestRunReleaseProjectDryRunDoesNotMutate(t *testing.T) {
	releaseModules = releaseModuleSlice{}
	releaseAll = false
	releaseBump = "patch"
	releaseApply = false
	releaseNotes = false

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte("project:\n  kind: site\nversion: \"v0.3.2\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	if got := RunRelease(flag.NewFlagSet("release", flag.ContinueOnError)); got != core.ExitCodeSuccess {
		t.Errorf("project release dry-run = %v, want ExitCodeSuccess", got)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "krewire.yaml"))
	if strings.Contains(string(data), "v0.3.3") {
		t.Error("project dry-run mutated krewire.yaml")
	}
}

func TestRunReleaseProjectMissingConfigFails(t *testing.T) {
	releaseModules = releaseModuleSlice{}
	releaseAll = false
	releaseBump = "patch"
	releaseApply = false
	releaseNotes = false

	dir := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if got := RunRelease(flag.NewFlagSet("release", flag.ContinueOnError)); got == core.ExitCodeSuccess {
		t.Error("expected failure when no Krewire project is present")
	}
}

