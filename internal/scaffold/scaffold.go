// Package scaffold generates new Krewire projects.
package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	modFramework = "github.com/krewire/framework"
	modLibs      = "github.com/krewire/libs"
)

// Sentinels returned by the scaffold functions.
var (
	// ErrInvalidName is returned when the project name contains invalid
	// characters or path separators.
	ErrInvalidName = errors.New("invalid project name: use letters, digits, dashes, dots, or underscores")
	// ErrProjectExists is returned when the target directory exists and is
	// not empty.
	ErrProjectExists = errors.New("refusing to scaffold: directory already exists and is not empty")
	// ErrNotEmpty is returned when a variant target must be empty, such as a
	// --template clone target.
	ErrNotEmpty = errors.New("refusing to clone: target directory is not empty")
	// ErrConflict is returned when equipping a variant would overwrite a
	// file outside the kernel set (go.mod, krewire.yaml, .gitignore, main.go).
	ErrConflict = errors.New("refusing to equip: file already exists")
)

var namePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// Variant is the project shape equipped by Equip.
type Variant string

const (
	// VariantApp is a fullstack monolith with the entry point at the root
	// main.go (canonical layout KWF-CCI0N).
	VariantApp Variant = "app"
	// VariantStatic is a declarative static site configured solely through
	// the `ssg:` key in krewire.yaml (KWF-PT8OD). No ssg.yaml is produced.
	VariantStatic Variant = "static"
	// VariantBook is a manuscript book assembled by mdbind (KWM-FX9H2).
	VariantBook Variant = "book"
	// VariantCLI is a command-line application on framework/tui + libs/core,
	// mirroring the kiw devtool layout.
	VariantCLI Variant = "cli"
	// VariantTemplate clones a remote starter repository into the target.
	VariantTemplate Variant = "template"
)

// Options configures kernel generation.
type Options struct {
	// Name is the project name and directory name. Required.
	Name string
	// Dir is the parent directory to create the project in. Defaults to the
	// current working directory.
	Dir string
	// Module is the Go module path. Defaults to the project name.
	Module string
}

// file is one output file of a scaffold.
type file struct {
	name string
	body string
}

// kernelBody is the stdlib-only main.go written by `kiw new`. init
// upgrades it in place for the app variant and removes it for site/book
// variants.
const kernelBody = `package main

import "fmt"

func main() {
	fmt.Println("hello, krewire")
}
`

// New generates a new minimal Krewire project kernel — go.mod, krewire.yaml,
// main.go, and .gitignore — and returns the created paths relative to Dir.
func New(opts Options) ([]string, error) {
	if opts.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}
	if !namePattern.MatchString(opts.Name) {
		return nil, ErrInvalidName
	}
	if opts.Dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		opts.Dir = wd
	}

	module := opts.Module
	if module == "" {
		module = opts.Name
	}

	target := filepath.Join(opts.Dir, opts.Name)
	empty, err := targetEmpty(target)
	if err != nil {
		return nil, err
	}
	if !empty {
		return nil, ErrProjectExists
	}

	var created []string
	for _, f := range kernel(opts.Name, module) {
		if err := writeFile(filepath.Join(target, f.name), []byte(f.body)); err != nil {
			return nil, err
		}
		created = append(created, filepath.Join(opts.Name, f.name))
	}
	return created, nil
}

// kernel returns the minimal kernel file set.
func kernel(name, module string) []file {
	return []file{
		{goModFile, fmt.Sprintf("module %s\n\ngo 1.22\n", module)},
		{krewireYaml, fmt.Sprintf("project:\n  name: %s\n", name)},
		{mainGo, kernelBody},
		{gitignoreFile, gitignoreTemplate(name)},
	}
}

// EquipOptions configures Equip. Dir is the project root to shape in place.
type EquipOptions struct {
	// Dir is the project root directory to equip. Required.
	Dir string
	// Name is the resolved project name, provided by the commands layer:
	// the module base for the app variant, the target directory base for
	// static and book. Required.
	Name string
	// Variant selects the shape to equip.
	Variant Variant
	// Module is the existing module path (from go.mod). Required by the app
	// variant to rewrite go.mod with pinned requires.
	Module string
	// Title is used by the static and book variants as the site title.
	Title string
	// FrameworkVersion pins the framework require. Defaults to latest.
	FrameworkVersion string
	// LibsVersion pins the libs require. Defaults to v0.1.0.
	LibsVersion string
	// TemplateURL is the git URL cloned by the template variant.
	TemplateURL string
}

// Equip shapes an existing project root into the requested variant and
// returns the created paths, relative to Dir. The kernel files (go.mod,
// krewire.yaml, .gitignore, main.go) may be upgraded in place; any other
// pre-existing file is refused.
func Equip(opts EquipOptions) ([]string, error) {
	if opts.Dir == "" {
		return nil, fmt.Errorf("equip: directory is required")
	}
	switch opts.Variant {
	case VariantApp:
		return equipApp(opts)
	case VariantStatic:
		return equipStatic(opts)
	case VariantBook:
		return equipBook(opts)
	case VariantCLI:
		return equipCLI(opts)
	case VariantTemplate:
		return equipTemplate(opts)
	default:
		return nil, fmt.Errorf("unknown variant %q", opts.Variant)
	}
}

const (
	goModFile     = "go.mod"
	krewireYaml   = "krewire.yaml"
	gitignoreFile = ".gitignore"
	mainGo        = "main.go"
)

// kernelUpgrade names the files init may overwrite in place.
var kernelUpgrade = map[string]bool{
	goModFile:     true,
	krewireYaml:   true,
	gitignoreFile: true,
	mainGo:        true,
}

// writeVariant writes files into Dir, refusing to overwrite anything outside
// the kernel set, and returns the created paths relative to Dir.
func writeVariant(dir string, files []file) ([]string, error) {
	var created []string
	for _, f := range files {
		path := filepath.Join(dir, f.name)
		if _, err := os.Stat(path); err == nil && !kernelUpgrade[f.name] {
			return nil, fmt.Errorf("%w: %s", ErrConflict, f.name)
		}
		if err := writeFile(path, []byte(f.body)); err != nil {
			return nil, err
		}
		created = append(created, f.name)
	}
	return created, nil
}

// removeKernelMain removes the kernel's placeholder main.go (written by
// `kiw new`) when equipping a site or book variant, so shape detection is
// pinned by project.kind rather than a stray app marker. It is a no-op when
// the file is absent or was modified by the user.
func removeKernelMain(dir string) error {
	path := filepath.Join(dir, mainGo)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(string(data)) == strings.TrimSpace(kernelBody) {
		return os.Remove(path)
	}
	return nil
}

func gitignoreTemplate(name string) string {
	return fmt.Sprintf(`# Compiled binary
/%s

# Krewire build outputs
site/
.krewire/

# Editor and OS noise
.idea/
.vscode/
*.swp
.DS_Store
`, name)
}

// targetEmpty reports whether target is an absent path or an empty directory.
func targetEmpty(target string) (bool, error) {
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, fmt.Errorf("cannot scaffold: %s exists and is not a directory", target)
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func writeFile(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
