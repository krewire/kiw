// Package shape classifies a project directory into one of the Krewire
// project kinds so commands can dispatch to the right pipeline.
package shape

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Kind identifies a Krewire project kind.
type Kind string

const (
	// KindApp is a fullstack monolith: an explicit main.go / cmd/* building
	// a web.App. Driven by kiw run and kiw dev.
	KindApp Kind = "app"
	// KindSite is a declarative static site configured by krewire.yaml (ssg:)
	// or ssg.yaml. Driven by kiw build and kiw serve.
	KindSite Kind = "site"
	// KindBook is a manuscript/ directory rendered by mdbind. Driven by
	// kiw build and kiw serve.
	KindBook Kind = "book"
	// KindCLI is a command-line application on framework/tui + libs/core,
	// built and run from the root main.go. Driven by kiw run and kiw dev.
	KindCLI Kind = "cli"
	// KindNone means no known marker was found.
	KindNone Kind = ""
)

// Result is the outcome of a detection.
type Result struct {
	// Kind is the detected project kind.
	Kind Kind
	// Marker names the file or config key that determined the kind, e.g.
	// "main.go" or "krewire.yaml#ssg".
	Marker string
}

// String implements fmt.Stringer.
func (r Result) String() string {
	if r.Kind == KindNone {
		return "none"
	}
	return string(r.Kind)
}

// Detect classifies the project rooted at dir. explicitKind, when non-empty,
// comes from config.project.kind and wins over any marker detection.
func Detect(dir, explicitKind string) (Result, error) {
	switch explicitKind {
	case string(KindApp), string(KindSite), string(KindBook), string(KindCLI):
		return Result{Kind: Kind(explicitKind), Marker: "project.kind"}, nil
	case "":
		// fall through to marker detection.
	default:
		return Result{}, fmt.Errorf("shape: unknown project kind %q", explicitKind)
	}

	if isApp(dir) {
		return Result{Kind: KindApp, Marker: appMarker(dir)}, nil
	}
	if hasFile(dir, "ssg.yaml") || hasSSGKey(dir) {
		return Result{Kind: KindSite, Marker: "krewire.yaml#ssg"}, nil
	}
	for _, marker := range []string{"content", "manuscript"} {
		if isDir(dir, marker) {
			return Result{Kind: KindBook, Marker: marker + "/"}, nil
		}
	}
	return Result{Kind: KindNone}, nil
}

// isApp reports whether dir builds an application package main directly from
// main.go or through cmd/ packages containing func main.
func isApp(dir string) bool {
	for _, p := range append([]string{dir}, cmdDirs(dir)...) {
		if packageMain(p) {
			return true
		}
	}
	return false
}

// cmdDirs returns every immediate subdirectory of dir/cmd.
func cmdDirs(dir string) []string {
	cmd := filepath.Join(dir, "cmd")
	entries, err := os.ReadDir(cmd)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, filepath.Join(cmd, e.Name()))
		}
	}
	return out
}

// appMarker names the strongest marker that flagged an app project.
func appMarker(dir string) string {
	if packageMain(dir) {
		return "main.go"
	}
	return "cmd/*"
}

// packageMain reports whether pkgDir is a package main whose files declare
// func main, excluding test files.
func packageMain(pkgDir string) bool {
	fset := token.NewFileSet()
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false
	}
	hasMain := false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(pkgDir, e.Name()), nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		if f.Name.Name != "main" {
			return false
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				hasMain = true
			}
		}
	}
	return hasMain
}

// hasSSGKey reports whether krewire.yaml declares an `ssg:` key.
func hasSSGKey(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "krewire.yaml"))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "ssg:")
}

// hasFile reports whether file exists as a regular file under dir.
func hasFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}

// isDir reports whether sub is a directory under dir.
func isDir(dir, sub string) bool {
	info, err := os.Stat(filepath.Join(dir, sub))
	return err == nil && info.IsDir()
}
