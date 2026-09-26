package commands

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/libs/core"
)

func TestWorkerCommand_KindMismatch(t *testing.T) {
	dir := t.TempDir()
	yaml := "project:\n  name: myapp\n  kind: app\n"
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	_ = os.Chdir(dir)

	fs := flag.NewFlagSet("worker", flag.ContinueOnError)
	RegisterWorker(fs)

	code := RunWorker(fs)
	if code != core.ExitCodeUsage {
		t.Fatalf("expected exit code %v, got %v", core.ExitCodeUsage, code)
	}
}

func TestWorkerCommand_DLQListEmpty(t *testing.T) {
	dir := t.TempDir()
	yaml := "project:\n  name: myworker\n  kind: worker\n"
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	defer os.Chdir(orig)
	_ = os.Chdir(dir)

	fs := flag.NewFlagSet("worker", flag.ContinueOnError)
	RegisterWorker(fs)
	_ = fs.Set("dlq", "list")
	_ = fs.Set("queue", "file")

	code := RunWorker(fs)
	if code != core.ExitCodeSuccess {
		t.Fatalf("expected exit code success, got %v", code)
	}
}
