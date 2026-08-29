// Tests for KWN-Q3M8V
package commands

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/kiw/internal/gomod"
	"github.com/krewire/libs/core"
)

func newWsFlagSet(args ...string) *flag.FlagSet {
	fs := flag.NewFlagSet("ws", flag.ContinueOnError)
	_ = fs.Parse(args)
	return fs
}

func captureStdout(t *testing.T, fn func() core.ExitCode) (string, core.ExitCode) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out), code
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Spec: KWN-Q3M8V WS-CMD-001 Scope: Unit
func TestKWN_WS_CMD_001_RunWs_DispatchesHelpAndRejectsUnknown(t *testing.T) {
	if _, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("help")) }); code != core.ExitCodeUsage {
		t.Errorf("help exit code = %v, want Usage", code)
	}
	if _, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("bogus")) }); code != core.ExitCodeUsage {
		t.Errorf("unknown sub-command exit code = %v, want Usage", code)
	}
}

// Spec: KWN-Q3M8V WS-CMD-002 Scope: Unit
func TestKWN_WS_CMD_002_FindWorkspaceRoot_ClassifiesGoWorkAndMonorepo(t *testing.T) {
	hub := t.TempDir()
	writeFile(t, filepath.Join(hub, "go.work"), "go 1.26\n\nuse (\n\t./a\n\t./b\n)\n")
	nested := filepath.Join(hub, "a", "deep")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	root, typ, members := findWorkspaceRoot(nested)
	if root != hub || !strings.Contains(typ, "go.work") {
		t.Errorf("hub detection root=%q typ=%q", root, typ)
	}
	if len(members) != 2 || members[0] != "./a" || members[1] != "./b" {
		t.Errorf("block-form members = %q", members)
	}

	solo := t.TempDir()
	writeFile(t, filepath.Join(solo, "go.mod"), "module example.com/solo\n")
	root2, typ2, members2 := findWorkspaceRoot(solo)
	if root2 != solo || !strings.Contains(typ2, "monorepo") || len(members2) != 1 || members2[0] != solo {
		t.Errorf("monorepo detection root=%q typ=%q members=%q", root2, typ2, members2)
	}
}

// Spec: KWN-Q3M8V WS-CMD-003 Scope: Unit
func TestKWN_WS_CMD_003_RunWsInfo_PrintsRootTypeAndMembers(t *testing.T) {
	solo := t.TempDir()
	writeFile(t, filepath.Join(solo, "go.mod"), "module example.com/solo\n")
	chdir(t, solo)

	out, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("info")) })
	if code != core.ExitCodeSuccess {
		t.Fatalf("info exit code = %v", code)
	}
	for _, want := range []string{"workspace:", "type:", "monorepo", "members:"} {
		if !strings.Contains(out, want) {
			t.Errorf("info output missing %q in %q", want, out)
		}
	}
}

// Spec: KWN-Q3M8V WS-CMD-004 Scope: Unit
func TestKWN_WS_CMD_004_RunWsList_RendersProjectKindModuleTable(t *testing.T) {
	hub := t.TempDir()
	writeFile(t, filepath.Join(hub, "go.mod"), "module example.com/hub\n")
	writeFile(t, filepath.Join(hub, "go.work"), "go 1.26\n\nuse (\n\t.\n)\n")

	if _, err := gomod.Read(filepath.Join(hub, "go.mod")); err != nil {
		t.Fatalf("gomod fixture unreadable: %v", err)
	}
	chdir(t, hub)

	out, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("list")) })
	if code != core.ExitCodeSuccess {
		t.Fatalf("list exit code = %v", code)
	}
	for _, want := range []string{"PROJECT", "example.com/hub"} {
		if !strings.Contains(out, want) {
			t.Errorf("list output missing %q in %q", want, out)
		}
	}
}

// Spec: KWN-Q3M8V WS-CMD-005 Scope: Unit
func TestKWN_WS_CMD_005_RunWsAddRemove_UsageWithoutArgsOrGoWork(t *testing.T) {
	chdir(t, t.TempDir())

	if _, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("add")) }); code != core.ExitCodeUsage {
		t.Errorf("add without path = %v, want Usage", code)
	}
	if _, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("remove")) }); code != core.ExitCodeUsage {
		t.Errorf("remove without path = %v, want Usage", code)
	}
	if _, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("add", "./x")) }); code != core.ExitCodeUsage {
		t.Errorf("add outside go.work layout = %v, want Usage", code)
	}
}

// Spec: KWN-Q3M8V WS-CMD-006 Scope: Unit
func TestKWN_WS_CMD_006_RunWsSync_SucceedsInTrivialWorkspace(t *testing.T) {
	hub := t.TempDir()
	writeFile(t, filepath.Join(hub, "go.mod"), "module example.com/hub\n")
	writeFile(t, filepath.Join(hub, "go.work"), "go 1.26\n\nuse (\n\t.\n)\n")
	chdir(t, hub)

	if _, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("sync")) }); code != core.ExitCodeSuccess {
		t.Fatalf("sync exit code = %v, want Success", code)
	}
}

// Spec: KWN-Q3M8V WS-CMD-007 Scope: Unit
func TestKWN_WS_CMD_007_RunWsExec_AggregatesMemberFailures(t *testing.T) {
	hub := t.TempDir()
	writeFile(t, filepath.Join(hub, "good", "go.mod"), "module example.com/good\n")
	writeFile(t, filepath.Join(hub, "bad", "go.mod"), "module example.com/bad\n")
	writeFile(t, filepath.Join(hub, "go.work"), "go 1.26\n\nuse (\n\t./good\n\t./bad\n)\n")
	chdir(t, hub)

	_, ok := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("exec", "--", "true")) })
	if ok != core.ExitCodeSuccess {
		t.Fatalf("exec true exit code = %v, want Success", ok)
	}

	_, failed := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("exec", "false")) })
	if failed != core.ExitCodeFailure {
		t.Fatalf("exec false exit code = %v, want Failure", failed)
	}

	if _, code := captureStdout(t, func() core.ExitCode { return RunWs(newWsFlagSet("exec")) }); code != core.ExitCodeUsage {
		t.Errorf("exec without command = %v, want Usage", code)
	}
}
