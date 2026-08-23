package scaffold

import "fmt"

func internalAppTemplate(module string) string {
	return fmt.Sprintf(`// Package app composes the application: it wires every
// dependency through the app container and returns the ready
// *web.App. cmd only calls this and runs.
package app

import (
	"io/fs"

	rvapp "github.com/krewire/framework/app"
	rvweb "github.com/krewire/framework/web"
	"%s"
	"%s"
	"%s"
	"%s"
	"%s"
	"%s"
)

type Config struct {
	Meta *config.Metadata
	Cfg  config.Config
}

// webProvider assembles the SSR side of the app: theme, layout, pages, and
// embedded static assets.
type webProvider struct{}

// Register binds the web app singleton.
func (p *webProvider) Register(c *rvapp.Container) error {
	return rvapp.Singleton[*rvweb.App](c, func() *rvweb.App { return rvweb.NewApp() })
}

// Boot configures the app once the container can resolve it.
func (p *webProvider) Boot(c *rvapp.Container) error {
	a, err := rvapp.Resolve[*rvweb.App](c)
	if err != nil {
		return err
	}
	a.Theme(theme.New())
	a.Layout(layouts.Shell())

	sub, err := fs.Sub(assets.Files, "public")
	if err != nil {
		return err
	}
	a.Static("/static", sub)
	pages.Register(a)
	return nil
}

// New assembles the full application and returns the ready *web.App
func New(meta *config.Metadata, appCfg config.Config) (*rvweb.App, error) {
	container, err := rvapp.NewApp(
		&configProvider{meta: meta, cfg: appCfg},
		&webProvider{},
		&http.Provider{},
	).Build()
	if err != nil {
		return nil, err
	}
	return rvapp.Resolve[*rvweb.App](container)
}

type configProvider struct {
	meta *config.Metadata
	cfg  config.Config
}

// Register binds the loaded config into the container.
func (p *configProvider) Register(c *rvapp.Container) error {
	return rvapp.Singleton[*Config](c, func() *Config {
		return &Config{Meta: p.meta, Cfg: p.cfg}
	})
}

// Boot does nothing for config.
func (p *configProvider) Boot(c *rvapp.Container) error { return nil }
`, module+"/internal/config", module+"/internal/http", module+"/assets", module+"/web/layouts", module+"/web/pages", module+"/web/theme")
}

func internalConfigTemplate() string {
	return "// Package config defines the typed application configuration\n" +
		"// krewire.yaml: metadata and project-level config (metadata only)\n" +
		"// cfg.yaml: key-value runtime configuration (key-value pairs)\n" +
		"package config\n\n" +
		"import (\n" +
		"\t\"fmt\"\n" +
		"\t\"os\"\n\n" +
		"\trconfig \"github.com/krewire/libs/config\"\n" +
		"\trcfg \"github.com/krewire/libs/cfg\"\n" +
		"\t\"github.com/krewire/libs/validate\"\n" +
		")\n\n" +
		"// Metadata is the project metadata from krewire.yaml\n" +
		"type Metadata struct {\n" +
		"\tAddr  string `yaml:\"addr\" env:\"APP_ADDR\" validate:\"required\"`\n" +
		"\tTitle string `yaml:\"title\" env:\"TITLE\" validate:\"required\"`\n" +
		"\tKind  string `yaml:\"kind\" validate:\"required\"`\n" +
		"\tName  string `yaml:\"name\" validate:\"required\"`\n" +
		"\tVersion string `yaml:\"version\"`\n" +
		"}\n\n" +
		"// Config is the key-value runtime configuration from cfg.yaml\n" +
		"type Config rcfg.Config\n\n" +
		"// LoadMetadata reads krewire.yaml from path, overlays the environment, and returns a\n" +
		"// validated Metadata.\n" +
		"func LoadMetadata(path string) (*Metadata, error) {\n" +
		"\tcfg := &Metadata{Addr: \":8080\", Title: \"Krewire Monolith\", Kind: \"app\", Version: \"0.1.0\"}\n" +
		"\tif err := rconfig.Load(path, cfg); err != nil {\n" +
		"\t\treturn nil, err\n" +
		"\t}\n" +
		"\tif err := rconfig.Override(cfg, os.LookupEnv); err != nil {\n" +
		"\t\treturn nil, err\n" +
		"\t}\n" +
		"\tif err := validate.Struct(cfg); err != nil {\n" +
		"\t\treturn nil, fmt.Errorf(\"config: %w\", err)\n" +
		"\t}\n" +
		"\treturn cfg, nil\n" +
		"}\n\n" +
		"// LoadConfig loads key-value configuration from cfg.yaml\n" +
		"func LoadConfig(path string) (Config, error) {\n" +
		"\treturn rcfg.Load(path)\n" +
		"}\n\n" +
		"// ConfigWithDefaults loads cfg.yaml and applies defaults\n" +
		"func ConfigWithDefaults(path string, defaults Config) (Config, error) {\n" +
		"\tcfg, err := rcfg.Load(path)\n" +
		"\tif err != nil {\n" +
		"\t\treturn nil, err\n" +
		"\t}\n" +
		"\treturn cfg.WithDefaults(defaults), nil\n" +
		"}\n\n" +
		"`"
}

func internalHttpTemplate() string {
	return "// Package http contains the application's HTTP layer (FRK-STR-003): handlers\n" +
		"// and routes for the JSON API.\n" +
		"package http\n\n" +
		"import (\n" +
		"\t\"net/http\"\n\n" +
		"\trvapp \"github.com/krewire/framework/app\"\n" +
		"\trvweb \"github.com/krewire/framework/web\"\n" +
		"\t\"github.com/krewire/libs/validate\"\n" +
		")\n\n" +
		"// Handler serves the JSON API endpoints.\n" +
		"type Handler struct{}\n\n" +
		"// NewHandler builds the API handler.\n" +
		"func NewHandler() *Handler { return &Handler{} }\n\n" +
		"// Status reports service health.\n" +
		"func (h *Handler) Status(w http.ResponseWriter, _ *http.Request, _ rvweb.Params) {\n" +
		"\ttype statusResp struct {\n" +
		"\t\tStatus  string `json:\"status\"`\n" +
		"\t\tVersion string `json:\"version\"`\n" +
		"\t}\n" +
		"\trvweb.JSON(w, http.StatusOK, statusResp{Status: \"ok\", Version: \"0.6.0\"})\n" +
		"}\n\n" +
		"// Contact validates and stores a contact request.\n" +
		"func (h *Handler) Contact(w http.ResponseWriter, r *http.Request, _ rvweb.Params) {\n" +
		"\ttype contactReq struct {\n" +
		"\t\tEmail string `json:\"email\" validate:\"required,email\"`\n" +
		"\t\tName  string `json:\"name\" validate:\"required,len=3\"`\n" +
		"\t}\n" +
		"\tvar req contactReq\n" +
		"\tif err := rvweb.ReadJSON(r, &req); err != nil {\n" +
		"\t\trvweb.Error(w, err)\n" +
		"\t\treturn\n" +
		"\t}\n" +
		"\tif err := validate.Struct(&req); err != nil {\n" +
		"\t\trvweb.Error(w, &rvweb.HTTPError{Status: http.StatusBadRequest, Code: \"invalid\", Message: err.Error()})\n" +
		"\t\treturn\n" +
		"\t}\n" +
		"\trvweb.JSON(w, http.StatusCreated, map[string]string{\"email\": req.Email})\n" +
		"}\n\n" +
		"// Provider registers the API handler and its routes (KWF-D1CNT).\n" +
		"type Provider struct{}\n\n" +
		"// Register binds the handler into the container.\n" +
		"func (p *Provider) Register(c *rvapp.Container) error {\n" +
		"\treturn rvapp.Singleton[*Handler](c, func() *Handler { return NewHandler() })\n" +
		"}\n\n" +
		"// Boot mounts the API routes once the web app is assembled.\n" +
		"func (p *Provider) Boot(c *rvapp.Container) error {\n" +
		"\ta, err := rvapp.Resolve[*rvweb.App](c)\n" +
		"\tif err != nil {\n" +
		"\t\treturn err\n" +
		"\t}\n" +
		"\th, err := rvapp.Resolve[*Handler](c)\n" +
		"\tif err != nil {\n" +
		"\t\treturn err\n" +
		"\t}\n" +
		"\ta.Router().Get(\"/api/status\", h.Status)\n" +
		"\ta.Router().Post(\"/api/contact\", h.Contact)\n" +
		"\treturn nil\n" +
		"}\n"
}
