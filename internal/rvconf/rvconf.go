// Package rvconf reads project configuration from krewire.yaml so that `kiw`
// commands can drive any project kind without a project-specific cmd.
// A krewire.yaml with an `ssg:` key triggers SSG mode; one with a manuscript/
// directory triggers book mode. All top-level fields are shared across modes.
package rvconf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/krewire/framework/ui"
	"github.com/krewire/framework/web/ssg"
	"gopkg.in/yaml.v3"
)

// Config is the krewire.yaml schema consumed by the build command.
type Config struct {
	// Project pins the project kind when discovery is ambiguous. Empty means
	// auto-detection (app > site > book precedence).
	Project Project `yaml:"project"`
	// Env pins the target environment: "local", "production", or "testing".
	// Empty resolves to local; KIW_ENV and --env override (KWL-K4T7W).
	Env string `yaml:"env"`
	// Debug enables debug mode. Defaults to false; KIW_DEBUG and --debug
	// override.
	Debug bool `yaml:"debug"`
	// Title is the site title.
	Title string `yaml:"title"`
	// Author is the site author.
	Author string `yaml:"author"`
	// Base is the URL base the site is served under, e.g. "/docs/".
	Base string `yaml:"base"`
	// Input is the manuscript directory, relative to the project root.
	Input string `yaml:"input"`
	// Output is the output directory, relative to the project root.
	Output string `yaml:"output"`
	// Nav holds optional navbar links.
	Nav []Link `yaml:"nav"`
	// Footer is optional footer text.
	Footer string `yaml:"footer"`
	// Theme configures the light/dark theme switcher.
	Theme *Theme `yaml:"theme"`
	// SSG holds the declarative SSG config. When non-nil, kiw build uses
	// the web/ssg generator instead of mdbind. Top-level fields (Title,
	// Author, Base, Output, Theme) are injected into the SSG config.
	SSG *SSGConfig `yaml:"ssg"`
}

// IsSSG reports whether this config targets the declarative SSG mode.
func (c *Config) IsSSG() bool {
	return c != nil && c.SSG != nil
}

// Kind returns the pinned project kind from config, or "" when unset.
func (c *Config) Kind() string {
	if c == nil {
		return ""
	}
	return c.Project.Kind
}

// Project holds per-project settings in krewire.yaml.
type Project struct {
	// Kind pins the project kind: "app", "cli", "site", or "book". Empty
	// lets the CLI detect it from marker files.
	Kind string `yaml:"kind"`
	// Dirs overrides the canonical directory locations (FRK-STR-010).
	Dirs Dirs `yaml:"dirs"`
}

// Dirs holds custom directory locations for the project layout.
type Dirs struct {
	Web      string `yaml:"web"`
	Public   string `yaml:"public"`
	Internal string `yaml:"internal"`
	Cmd      string `yaml:"cmd"`
}

// SSGConfig holds SSG-specific fields that live under the `ssg:` key in
// krewire.yaml. It mirrors the fields of ssg.Config that are not shared with
// the top-level Config.
type SSGConfig struct {
	// Description is injected into every page's template data.
	Description string `yaml:"description"`
	// Layouts are the site's named layouts.
	Layouts []SSGLayoutConfig `yaml:"layouts"`
	// Components are the site's named components.
	Components []SSGComponentConfig `yaml:"components"`
	// Pages are the site's output pages.
	Pages []SSGPageConfig `yaml:"pages"`
	// Assets maps output paths to file contents.
	Assets map[string]string `yaml:"assets"`
}

// SSGLayoutConfig is a layout entry under ssg.layouts.
type SSGLayoutConfig struct {
	Name  string `yaml:"name"`
	Body  string `yaml:"body"`
	Style string `yaml:"style"`
	// UI, when set, builds the layout from a reusable framework/ui shell
	// instead of Body/Style.
	UI *ui.LayoutConfig `yaml:"ui"`
}

// SSGComponentConfig is a component entry under ssg.components.
type SSGComponentConfig struct {
	Name  string `yaml:"name"`
	Body  string `yaml:"body"`
	Style string `yaml:"style"`
}

// SSGPageConfig is a page entry under ssg.pages.
type SSGPageConfig struct {
	Path   string         `yaml:"path"`
	Title  string         `yaml:"title"`
	Layout string         `yaml:"layout"`
	Root   string         `yaml:"root"`
	Data   map[string]any `yaml:"data"`
}

// ToSSGConfig converts this Config into an ssg.Config, merging top-level
// fields (Title, Output, Theme) with the SSG-specific fields under ssg:.
// Callers must check IsSSG() before calling this method.
func (c *Config) ToSSGConfig() *ssg.Config {
	s := c.SSG
	cfg := &ssg.Config{
		Title:       c.Title,
		Description: s.Description,
		Output:      c.Output,
		Assets:      s.Assets,
	}
	if c.Theme != nil {
		cfg.Theme = &ssg.ThemeConfig{
			Default: c.Theme.Default,
		}
	}
	for _, l := range s.Layouts {
		cfg.Layouts = append(cfg.Layouts, ssg.LayoutConfig{Name: l.Name, Body: l.Body, Style: l.Style, UI: l.UI})
	}
	for _, comp := range s.Components {
		cfg.Components = append(cfg.Components, ssg.ComponentConfig{Name: comp.Name, Body: comp.Body, Style: comp.Style})
	}
	for _, p := range s.Pages {
		cfg.Pages = append(cfg.Pages, ssg.PageConfig{
			Path:   p.Path,
			Title:  p.Title,
			Layout: p.Layout,
			Root:   p.Root,
			Data:   p.Data,
		})
	}
	return cfg
}

// Link is a named navigation target shown in the navbar.
type Link struct {
	Text string `yaml:"text"`
	URL  string `yaml:"url"`
}

// Theme configures the theme switcher and per-mode palette overrides.
type Theme struct {
	// Default is the initial mode: auto, light, or dark.
	Default string `yaml:"default"`
	// Light and Dark hold palette token overrides for each mode.
	Light Palette `yaml:"light"`
	Dark  Palette `yaml:"dark"`
}

// Palette holds color token overrides. Empty tokens fall back to the ui
// defaults.
type Palette struct {
	Base1            string `yaml:"base-1"`
	Base1Content     string `yaml:"base-1-content"`
	Base2            string `yaml:"base-2"`
	Base2Content     string `yaml:"base-2-content"`
	Base3            string `yaml:"base-3"`
	Base3Content     string `yaml:"base-3-content"`
	Primary          string `yaml:"primary"`
	PrimaryContent   string `yaml:"primary-content"`
	Secondary        string `yaml:"secondary"`
	SecondaryContent string `yaml:"secondary-content"`
	Accent           string `yaml:"accent"`
	AccentContent    string `yaml:"accent-content"`
	Ghost            string `yaml:"ghost"`
	GhostContent     string `yaml:"ghost-content"`
	Neutral          string `yaml:"neutral"`
	NeutralContent   string `yaml:"neutral-content"`
	Success          string `yaml:"success"`
	SuccessContent   string `yaml:"success-content"`
	Info             string `yaml:"info"`
	InfoContent      string `yaml:"info-content"`
	Warning          string `yaml:"warning"`
	WarningContent   string `yaml:"warning-content"`
	Error            string `yaml:"error"`
	ErrorContent     string `yaml:"error-content"`
}

// Load reads krewire.yaml from dir. A missing file yields a zero Config without
// an error.
func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, "krewire.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("rvconf: read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("rvconf: parse %s: %w", path, err)
	}
	return &c, nil
}

// UI merges the palette overrides into the given default palette.
func (p *Palette) UI(defaults ui.Palette) ui.Palette {
	if p == nil {
		return defaults
	}
	if p.Base1 != "" {
		defaults.Base1 = ui.Color(p.Base1)
	}
	if p.Base1Content != "" {
		defaults.Base1Content = ui.Color(p.Base1Content)
	}
	if p.Base2 != "" {
		defaults.Base2 = ui.Color(p.Base2)
	}
	if p.Base2Content != "" {
		defaults.Base2Content = ui.Color(p.Base2Content)
	}
	if p.Base3 != "" {
		defaults.Base3 = ui.Color(p.Base3)
	}
	if p.Base3Content != "" {
		defaults.Base3Content = ui.Color(p.Base3Content)
	}
	if p.Primary != "" {
		defaults.Primary = ui.Color(p.Primary)
	}
	if p.PrimaryContent != "" {
		defaults.PrimaryContent = ui.Color(p.PrimaryContent)
	}
	if p.Secondary != "" {
		defaults.Secondary = ui.Color(p.Secondary)
	}
	if p.SecondaryContent != "" {
		defaults.SecondaryContent = ui.Color(p.SecondaryContent)
	}
	if p.Accent != "" {
		defaults.Accent = ui.Color(p.Accent)
	}
	if p.AccentContent != "" {
		defaults.AccentContent = ui.Color(p.AccentContent)
	}
	if p.Ghost != "" {
		defaults.Ghost = ui.Color(p.Ghost)
	}
	if p.GhostContent != "" {
		defaults.GhostContent = ui.Color(p.GhostContent)
	}
	if p.Neutral != "" {
		defaults.Neutral = ui.Color(p.Neutral)
	}
	if p.NeutralContent != "" {
		defaults.NeutralContent = ui.Color(p.NeutralContent)
	}
	if p.Success != "" {
		defaults.Success = ui.Color(p.Success)
	}
	if p.SuccessContent != "" {
		defaults.SuccessContent = ui.Color(p.SuccessContent)
	}
	if p.Info != "" {
		defaults.Info = ui.Color(p.Info)
	}
	if p.InfoContent != "" {
		defaults.InfoContent = ui.Color(p.InfoContent)
	}
	if p.Warning != "" {
		defaults.Warning = ui.Color(p.Warning)
	}
	if p.WarningContent != "" {
		defaults.WarningContent = ui.Color(p.WarningContent)
	}
	if p.Error != "" {
		defaults.Error = ui.Color(p.Error)
	}
	if p.ErrorContent != "" {
		defaults.ErrorContent = ui.Color(p.ErrorContent)
	}
	return defaults
}
