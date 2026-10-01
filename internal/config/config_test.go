package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/framework/ui"
)

func TestLoadMissingFile(t *testing.T) {
	c, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "" || c.Base != "" || c.Input != "" || c.Output != "" || c.Theme != nil || len(c.Nav) != 0 {
		t.Errorf("expected zero config, got %+v", *c)
	}
}

func TestLoadParsesConfig(t *testing.T) {
	dir := t.TempDir()
	body := `title: Krewire Documentation
author: Krewire Contributors
base: /docs/
input: manuscript
output: site
footer: Krewire Documentation — MIT License
nav:
  - text: Krewire
    url: /
theme:
  default: dark
  light:
    primary: "#7c3aed"
    accent: "#b45309"
  dark:
    primary: "#a78bfa"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Title != "Krewire Documentation" || c.Base != "/docs/" || c.Output != "site" {
		t.Errorf("parsed config = %+v", c)
	}
	if len(c.Nav) != 1 || c.Nav[0].URL != "/" {
		t.Errorf("nav = %+v", c.Nav)
	}
	if c.Theme.Default != "dark" || string(c.Theme.Light.Primary) != "#7c3aed" || c.Theme.Dark.Accent != "" {
		t.Errorf("theme = %+v", c.Theme)
	}
}

// A project may declare its version next to the kind instead of at the top
// level. Both must produce the same effective version, otherwise the ssg half
// stamps its cache-busting query while the book half does not, and a reader
// keeps a stale stylesheet on one page family and a fresh one on the other.
func TestLoadFoldsProjectVersionIntoVersion(t *testing.T) {
	dir := t.TempDir()
	body := `project:
  name: krewire
  kind: site
  version: "v0.1.0"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Version != "v0.1.0" {
		t.Errorf("Version = %q, want v0.1.0 from project.version", c.Version)
	}
}

// An explicit top-level version is the effective one: it is the documented
// location and must win over the nested declaration.
func TestLoadPrefersTopLevelVersion(t *testing.T) {
	dir := t.TempDir()
	body := `version: "2.0.0"
project:
  kind: site
  version: "v0.1.0"
`
	if err := os.WriteFile(filepath.Join(dir, "krewire.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Version != "2.0.0" {
		t.Errorf("Version = %q, want the top-level 2.0.0", c.Version)
	}
}
func TestPaletteMerge(t *testing.T) {
	p := &Palette{Primary: "#123456", Accent: "#abcdef"}
	got := p.UI(ui.DefaultLightPalette)
	if got.Primary != "#123456" || got.Accent != "#abcdef" {
		t.Errorf("overrides not applied: %+v", got)
	}
	if got.Base1 != ui.DefaultLightPalette.Base1 {
		t.Errorf("unset tokens must fall back to defaults: %+v", got)
	}
	if (*Palette)(nil).UI(ui.DefaultLightPalette) != ui.DefaultLightPalette {
		t.Error("nil palette must return the defaults unchanged")
	}
}
