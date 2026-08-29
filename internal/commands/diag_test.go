// Tests for KWL-P8W2N
package commands

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/krewire/libs/core"
	"github.com/krewire/libs/vein"
)

// Spec: KWL-P8W2N KWL-DIAGV-007 Scope: Domain
func TestKWL_DIAGV_007_BootRuntime_InstallsEnvAppropriateLogger(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":       "module demo\n\ngo 1.26\n",
		"krewire.yaml": "title: Demo\nenv: production\ndebug: true\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chdir(t, dir)

	prev := slog.Default()
	t.Cleanup(func() { slog.SetDefault(prev) })

	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	RegisterServe(fs)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}
	rt, code := bootRuntime(fs)
	if code != vein.ExitCodeSuccess {
		t.Fatalf("bootRuntime code = %d", code)
	}
	if !rt.debug {
		t.Fatal("rt.debug = false, want true from krewire.yaml")
	}
	def := slog.Default()
	if _, ok := def.Handler().(*slog.JSONHandler); !ok {
		t.Error("production default logger should use the JSON handler")
	}
	if !def.Enabled(nil, slog.LevelDebug) {
		t.Error("debug mode should enable Debug level on the default logger")
	}
}

// Spec: KWL-P8W2N KWL-DIAGV-007 Scope: Domain
func TestKWL_DIAGV_007_Fail_PrintsAttachedStackWhenDebug(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = w
	prevDebug := debugEnabled
	debugEnabled = true
	t.Cleanup(func() { os.Stderr = oldStderr; debugEnabled = prevDebug })

	code := fail(vein.WithStack(vein.FailureError("diagnostic boom")))
	w.Close()

	out, _ := io.ReadAll(r)
	if code != vein.ExitCodeFailure {
		t.Errorf("fail code = %d, want failure", code)
	}
	logged := string(out)
	if !strings.Contains(logged, "diagnostic boom") || !strings.Contains(logged, "stack trace") {
		t.Errorf("stderr missing error or stack header:\n%s", logged)
	}
	if !strings.Contains(logged, "diag_test.go") {
		t.Errorf("stack should name this test file:\n%s", logged)
	}

	debugEnabled = false
	r2, w2, _ := os.Pipe()
	os.Stderr = w2
	fail(vein.WithStack(vein.FailureError("quiet boom")))
	w2.Close()
	quiet, _ := io.ReadAll(r2)
	if strings.Contains(string(quiet), "stack trace") {
		t.Errorf("stack printed while debug off:\n%s", quiet)
	}
}

// Spec: KWL-P8W2N KWL-ERRV-010 S4 Scope: Domain
func TestKWL_ERRV_010_Fail_PrintsTreeWithHintWithoutDebug(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = w
	prevDebug := debugEnabled
	debugEnabled = false
	t.Cleanup(func() { os.Stderr = oldStderr; debugEnabled = prevDebug })

	errd := vein.WithHint(
		vein.WithAttrs(vein.UsageError("cannot read krewire.yaml"), vein.Attr{Key: "file", Value: "krewire.yaml"}),
		"run 'kiw init' to create one",
	)
	code := fail(errd)
	w.Close()

	out, _ := io.ReadAll(r)
	tree := string(out)
	for _, want := range []string{"Error: cannot read krewire.yaml", "file=krewire.yaml", "Hint: run 'kiw init' to create one"} {
		if !strings.Contains(tree, want) {
			t.Errorf("tree missing %q:\n%s", want, tree)
		}
	}
	if strings.Contains(tree, "stack trace") {
		t.Errorf("stack leaked while debug off:\n%s", tree)
	}
	if code != vein.ExitCodeFailure {
		t.Errorf("fail code = %d, want failure", code)
	}

	// usage path keeps exit code 2 while rendering the same tree.
	r2, w2, _ := os.Pipe()
	os.Stderr = w2
	code2 := usageOrFail(vein.WithHint(vein.UsageError("bad flag combo"), "see 'kiw init --help'"))
	w2.Close()
	out2, _ := io.ReadAll(r2)
	if code2 != vein.ExitCodeUsage {
		t.Errorf("usageOrFail code = %d, want usage", code2)
	}
	if !strings.Contains(string(out2), "Hint: see 'kiw init --help'") {
		t.Errorf("usage tree missing hint:\n%s", out2)
	}
}
