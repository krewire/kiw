package shape

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectAppByRootMainGo(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/app\n")
	writeFile(t, dir, "main.go", `package main

func main() {}

`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindApp || got.Marker != "main.go" {
		t.Errorf("got %v (%s), want app (main.go)", got.Kind, got.Marker)
	}
}

func TestDetectAppByCmdPackage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/app\n")
	writeFile(t, dir, "cmd/server/main.go", `package main

func main() {}

`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindApp || got.Marker != "cmd/*" {
		t.Errorf("got %v (%s), want app (cmd/*)", got.Kind, got.Marker)
	}
}

func TestDetectNonMainGoIsNotApp(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/lib\n")
	writeFile(t, dir, "main.go", `package lib

func Main() {}

`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindNone {
		t.Errorf("got %v, want none", got.Kind)
	}
}

func TestDetectSite(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", "ssg:\n  title: Hi\n")
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindSite {
		t.Errorf("got %v, want site", got.Kind)
	}
}

func TestDetectSiteByConfigButAppPrecedes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "krewire.yaml", "ssg:\n")
	writeFile(t, dir, "main.go", `package main

func main() {}

`)
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindApp {
		t.Errorf("got %v, want app (precedence)", got.Kind)
	}
}

func TestDetectBook(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "manuscript/index.md", "# Hi\n")
	got, err := Detect(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindBook {
		t.Errorf("got %v, want book", got.Kind)
	}
}

func TestDetectExplicitKind(t *testing.T) {
	dir := t.TempDir()
	got, err := Detect(dir, "book")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindBook || got.Marker != "project.kind" {
		t.Errorf("got %v (%s), want book (project.kind)", got.Kind, got.Marker)
	}
}

func TestDetectExplicitCLI(t *testing.T) {
	dir := t.TempDir()
	got, err := Detect(dir, "cli")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindCLI || got.Marker != "project.kind" {
		t.Errorf("got %v (%s), want cli (project.kind)", got.Kind, got.Marker)
	}
}

func TestDetectUnknownExplicitKind(t *testing.T) {
	if _, err := Detect(t.TempDir(), "lambda"); err == nil {
		t.Error("want error for unknown explicit kind")
	}
}

func TestDetectNone(t *testing.T) {
	got, err := Detect(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindNone {
		t.Errorf("got %v, want none", got.Kind)
	}
}
