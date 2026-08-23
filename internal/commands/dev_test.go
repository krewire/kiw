package commands

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherDetectsGoChange(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := newWatcher(dir, 20*time.Millisecond, nil)
	defer func() { close(w.done) }()

	select {
	case <-w.Changed():
		t.Fatal("must not signal on first scan")
	case <-time.After(60 * time.Millisecond):
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-w.Changed():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected change signal")
	}
}

func TestWatcherIgnoresNonWatched(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "pkg.js"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := newWatcher(dir, 20*time.Millisecond, nil)
	defer func() { close(w.done) }()

	if err := os.WriteFile(filepath.Join(dir, "node_modules", "pkg.js"), []byte("y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-w.Changed():
		t.Fatal("node_modules must not be watched")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "out")
	if err := copyDir(src, dst); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "b" {
		t.Errorf("copied content = %q, want b", got)
	}
}
