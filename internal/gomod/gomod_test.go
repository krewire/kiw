package gomod

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMod(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReadParsesModuleAndRequireBlock(t *testing.T) {
	dir := t.TempDir()
	p := writeMod(t, dir, `module github.com/acme/demo

go 1.22

require (
	github.com/krewire/framework v0.0.0-20260818-abcdef
	github.com/krewire/libs v0.1.0
)
`)

	m, err := Read(p)
	if err != nil {
		t.Fatal(err)
	}
	if m.Path != "github.com/acme/demo" {
		t.Errorf("Path = %q, want github.com/acme/demo", m.Path)
	}
	if got := m.Requires["github.com/krewire/framework"]; got != "v0.0.0-20260818-abcdef" {
		t.Errorf("framework version = %q", got)
	}
	if got := m.Requires["github.com/krewire/libs"]; got != "v0.1.0" {
		t.Errorf("libs version = %q", got)
	}
	if !m.UsesKrewire() {
		t.Error("UsesKrewire() = false, want true")
	}
}

func TestReadParsesSingleLineRequire(t *testing.T) {
	dir := t.TempDir()
	p := writeMod(t, dir, `module demo

go 1.22

require github.com/krewire/libs v0.1.0
`)

	m, err := Read(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Requires["github.com/krewire/libs"]; got != "v0.1.0" {
		t.Errorf("libs version = %q, want v0.1.0", got)
	}
}

func TestReadRejectsMissingModule(t *testing.T) {
	dir := t.TempDir()
	p := writeMod(t, dir, "go 1.22\n")
	if _, err := Read(p); err == nil {
		t.Error("Read() error = nil, want a module directive error")
	}
}

func TestFindWalksUpDirectories(t *testing.T) {
	dir := t.TempDir()
	writeMod(t, dir, "module github.com/acme/demo\n\ngo 1.22\n")
	inner := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}

	m, err := Find(inner)
	if err != nil {
		t.Fatal(err)
	}
	if m.Path != "github.com/acme/demo" {
		t.Errorf("Path = %q", m.Path)
	}
}

func TestFindMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := Find(dir); err == nil {
		t.Error("Find() error = nil, want an error")
	}
}
