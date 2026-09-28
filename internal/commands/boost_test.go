package commands

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/libs/core"
)

func newBoostFlagSet(args ...string) *flag.FlagSet {
	fs := flag.NewFlagSet("boost", flag.ContinueOnError)
	RegisterBoost(fs)
	fs.Parse(append([]string{"install"}, args...))
	return fs
}

func newGuildFlagSet(args ...string) *flag.FlagSet {
	fs := flag.NewFlagSet("guild", flag.ContinueOnError)
	RegisterGuild(fs)
	fs.Parse(append([]string{"install"}, args...))
	return fs
}

func TestRunBoostInstallFresh(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target), strings.NewReader(""), &out)
	if code != core.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	for _, want := range []string{"AGENTS.md", "opencode.json"} {
		if _, err := os.Stat(filepath.Join(target, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	if !strings.Contains(out.String(), "Installed Boost") {
		t.Errorf("missing next-steps banner: %q", out.String())
	}
}

func TestRunGuildAliasInstallFresh(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	code := RunGuild(newGuildFlagSet(target))
	if code != core.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	for _, want := range []string{"AGENTS.md", "opencode.json"} {
		if _, err := os.Stat(filepath.Join(target, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
}

func TestRunBoostInstallConflictDeclines(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target), strings.NewReader("n\n"), &out)
	if code != core.ExitCodeUsage {
		t.Fatalf("want usage on decline, got %d", code.Int())
	}
	if data, err := os.ReadFile(filepath.Join(target, "AGENTS.md")); err != nil || string(data) != "mine" {
		t.Errorf("decline must not overwrite: %v %q", err, data)
	}
}

func TestRunBoostInstallConflictAccepts(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target), strings.NewReader("y\n"), &out)
	if code != core.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	if data, err := os.ReadFile(filepath.Join(target, "AGENTS.md")); err != nil || !strings.Contains(string(data), "Agent Constitution") {
		t.Errorf("accept must overwrite: %v %q", err, data)
	}
}

func TestRunBoostInstallForceSkipsPrompt(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "AGENTS.md"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target, "--force"), strings.NewReader(""), &out)
	if code != core.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
}

func TestRunBoostInstallDryRunWritesNothing(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target, "--dry-run"), strings.NewReader(""), &out)
	if code != core.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run must write nothing, found %d entries", len(entries))
	}
	if !strings.Contains(out.String(), "Dry run") {
		t.Errorf("missing dry-run banner: %q", out.String())
	}
}

func TestRunBoostInstallPromptsForTarget(t *testing.T) {
	target := t.TempDir()
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(), strings.NewReader(target+"\n"), &out)
	if code != core.ExitCodeSuccess {
		t.Fatalf("want success, got %d (%s)", code.Int(), out.String())
	}
	if _, err := os.Stat(filepath.Join(target, "AGENTS.md")); err != nil {
		t.Errorf("interactive install missing AGENTS.md: %v", err)
	}
}

func TestRunBoostMissingSubcommand(t *testing.T) {
	fs := flag.NewFlagSet("boost", flag.ContinueOnError)
	RegisterBoost(fs)
	if code := RunBoost(fs); code != core.ExitCodeUsage {
		t.Fatalf("want usage, got %d", code.Int())
	}
}

func TestRunBoostUnknownSubcommand(t *testing.T) {
	fs := flag.NewFlagSet("boost", flag.ContinueOnError)
	RegisterBoost(fs)
	fs.Parse([]string{"frobnicate"})
	if code := RunBoost(fs); code != core.ExitCodeUsage {
		t.Fatalf("want usage, got %d", code.Int())
	}
}

func TestRunBoostInstallMissingTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "nope")
	var out bytes.Buffer
	code := runBoostInstall(newBoostFlagSet(target), strings.NewReader(""), &out)
	if code != core.ExitCodeUsage {
		t.Fatalf("want usage for missing target, got %d", code.Int())
	}
}

func TestIsYes(t *testing.T) {
	for _, yes := range []string{"y", "Y", "yes", "YES", " yes "} {
		if !isYes(yes) {
			t.Errorf("isYes(%q) = false, want true", yes)
		}
	}
	for _, no := range []string{"", "n", "no", "x"} {
		if isYes(no) {
			t.Errorf("isYes(%q) = true, want false", no)
		}
	}
}
