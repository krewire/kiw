package scaffold

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCreatesKernel(t *testing.T) {
	parent := t.TempDir()
	created, err := New(Options{Name: "demo", Dir: parent})
	if err != nil {
		t.Fatal(err)
	}
	// Kernel type creates 4 files: go.mod, krewire.yaml, main.go, .gitignore.
	if len(created) != 4 {
		t.Fatalf("created %d files, want 4: %v", len(created), created)
	}

	assertFileContains(t, filepath.Join(parent, "demo", "go.mod"), "module demo")
	assertFileContains(t, filepath.Join(parent, "demo", "go.mod"), "go 1.22")
	assertFileNotContains(t, filepath.Join(parent, "demo", "go.mod"), "github.com/krewire/framework")
	assertFileContains(t, filepath.Join(parent, "demo", "krewire.yaml"), "name: demo")
	assertFileContains(t, filepath.Join(parent, "demo", "main.go"), "package main")
	assertFileNotContains(t, filepath.Join(parent, "demo", "main.go"), "github.com/krewire/framework")
	assertFileContains(t, filepath.Join(parent, "demo", ".gitignore"), "/demo")
}

func TestNewModuleOverride(t *testing.T) {
	parent := t.TempDir()
	if _, err := New(Options{Name: "demo", Dir: parent, Module: "github.com/acme/demo"}); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(parent, "demo", "go.mod"), "module github.com/acme/demo")
}

func TestNewRefusesNonEmptyDirectory(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := New(Options{Name: "demo", Dir: parent}); !errors.Is(err, ErrProjectExists) {
		t.Errorf("New() error = %v, want ErrProjectExists", err)
	}
}

func TestNewInvalidName(t *testing.T) {
	parent := t.TempDir()
	if _, err := New(Options{Name: "a/b", Dir: parent}); !errors.Is(err, ErrInvalidName) {
		t.Errorf("New() error = %v, want ErrInvalidName", err)
	}
}

// newKernel creates a kernel named "demo" inside parent and returns its path.
func newKernel(t *testing.T, parent string) string {
	t.Helper()
	if _, err := New(Options{Name: "demo", Dir: parent, Module: "example.com/demo"}); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(parent, "demo")
}

func TestEquipApp(t *testing.T) {
	dir := newKernel(t, t.TempDir())
	created, err := Equip(EquipOptions{
		Dir:     dir,
		Variant: VariantApp,
		Name:    "demo",
		Module:  "example.com/demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "go.mod"), "module example.com/demo")
	assertFileContains(t, filepath.Join(dir, "main.go"), "config.LoadMetadata")
	assertFileNotContains(t, filepath.Join(dir, "main.go"), "cfg.yaml")
	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "kind: app")
	assertFileContains(t, filepath.Join(dir, "internal/app/app.go"), "func New(meta *config.Metadata) (*rvweb.App, error)")
	for _, path := range []string{
		"internal/config/config.go",
		"internal/http/http.go",
		"web/layouts/shell.go",
		"web/pages/pages.go",
		"web/theme/theme.go",
		"assets/embed.go",
		"assets/public/app.css",
		"assets/public/app.js",
		"README.md",
		".gitignore",
	} {
		found := false
		for _, c := range created {
			if c == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("created files missing %s: %v", path, created)
		}
	}
}

func TestEquipAppRefusesNonKernelOverwrite(t *testing.T) {
	dir := newKernel(t, t.TempDir())
	if err := os.MkdirAll(filepath.Join(dir, "internal/app"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal/app/app.go"), []byte("user code"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Equip(EquipOptions{Dir: dir, Variant: VariantApp, Name: "demo", Module: "example.com/demo"})
	if !errors.Is(err, ErrConflict) {
		t.Errorf("Equip() error = %v, want ErrConflict", err)
	}
}

func TestEquipAppMissingModule(t *testing.T) {
	dir := t.TempDir()
	if _, err := Equip(EquipOptions{Dir: dir, Variant: VariantApp}); err == nil {
		t.Error("Equip() = nil error, want module error")
	}
}

func TestEquipCLI(t *testing.T) {
	dir := newKernel(t, t.TempDir())
	created, err := Equip(EquipOptions{
		Dir:     dir,
		Variant: VariantCLI,
		Name:    "demo",
		Module:  "example.com/demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "go.mod"), "module example.com/demo")
	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "kind: cli")
	assertFileContains(t, filepath.Join(dir, "main.go"), "tui.NewApp")
	assertFileContains(t, filepath.Join(dir, "main.go"), "internal/commands")
	assertFileContains(t, filepath.Join(dir, "internal/commands/commands.go"), "func Hello(_ *flag.FlagSet) core.ExitCode")
	assertFileContains(t, filepath.Join(dir, "internal/commands/commands.go"), "hello, demo")
	for _, path := range []string{"README.md", ".gitignore"} {
		found := false
		for _, c := range created {
			if c == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("created files missing %s: %v", path, created)
		}
	}
	if len(created) != 7 {
		t.Fatalf("equip cli created %d files, want 7: %v", len(created), created)
	}
}

func TestEquipCLIMissingModule(t *testing.T) {
	dir := t.TempDir()
	if _, err := Equip(EquipOptions{Dir: dir, Variant: VariantCLI, Name: "demo"}); err == nil {
		t.Error("Equip() = nil error, want module error")
	}
}

func TestEquipStatic(t *testing.T) {
	dir := newKernel(t, t.TempDir())
	created, err := Equip(EquipOptions{Dir: dir, Variant: VariantStatic, Name: "demo", Title: "My Site"})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "kind: site")
	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "ssg:")
	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "title: My Site")
	// No ssg.yaml is produced.
	if _, err := os.Stat(filepath.Join(dir, "ssg.yaml")); !os.IsNotExist(err) {
		t.Error("ssg.yaml should not exist")
	}
	// The kernel main.go is removed so shape is pinned by project.kind.
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Error("kernel main.go should be removed for a static project")
	}
	if len(created) == 0 {
		t.Fatal("equip static created no files")
	}
}

func TestEquipBook(t *testing.T) {
	dir := newKernel(t, t.TempDir())
	created, err := Equip(EquipOptions{Dir: dir, Variant: VariantBook, Name: "demo", Title: "My Book"})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "kind: book")
	assertFileContains(t, filepath.Join(dir, "krewire.yaml"), "title: My Book")
	assertFileContains(t, filepath.Join(dir, "content/docs/01-introduction.md"), "# Introduction")
	if _, err := os.Stat(filepath.Join(dir, "main.go")); !os.IsNotExist(err) {
		t.Error("kernel main.go should be removed for a book project")
	}
	if len(created) == 0 {
		t.Fatal("equip book created no files")
	}
}

func TestEquipStaticKeepsModifiedMain(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\n// user main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Equip(EquipOptions{Dir: dir, Variant: VariantStatic, Name: "demo"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err != nil {
		t.Fatalf("user-modified main.go should be kept: %v", err)
	}
}

func TestEquipUnknownVariant(t *testing.T) {
	dir := t.TempDir()
	if _, err := Equip(EquipOptions{Dir: dir, Variant: Variant("nope")}); err == nil {
		t.Error("Equip() = nil error, want unknown variant error")
	}
}

func TestEquipTemplateRequiresEmptyDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Equip(EquipOptions{Dir: dir, Variant: VariantTemplate, TemplateURL: "https://example.com/x.git"})
	if !errors.Is(err, ErrNotEmpty) {
		t.Errorf("Equip() error = %v, want ErrNotEmpty", err)
	}
}

func TestEquipTemplateMissingURL(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "empty")
	if _, err := Equip(EquipOptions{Dir: dir, Variant: VariantTemplate}); err == nil {
		t.Error("Equip() = nil error, want missing URL error")
	}
}

func TestEquipTemplateClonesLocalRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	src := t.TempDir()
	if err := gitInitCommit(src); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "clone")
	created, err := Equip(EquipOptions{Dir: dir, Variant: VariantTemplate, TemplateURL: filepath.ToSlash(src)})
	if err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, filepath.Join(dir, "hello.txt"), "from template")
	if len(created) == 0 {
		t.Fatal("clone created no files")
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), want) {
		t.Errorf("%s does not contain %q", filepath.Base(path), want)
	}
}

func assertFileNotContains(t *testing.T, path, notWant string) {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), notWant) {
		t.Errorf("%s should not contain %q", filepath.Base(path), notWant)
	}
}

// gitInitCommit creates a tiny git repository whose tracked files are listed
// by walkCreated.
func gitInitCommit(dir string) error {
	for _, f := range []struct{ name, body string }{
		{"hello.txt", "from template"},
	} {
		if err := os.WriteFile(filepath.Join(dir, f.name), []byte(f.body), 0o644); err != nil {
			return err
		}
	}
	cmd := exec.Command("git", "-C", dir, "init", "-q")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "-C", dir, "add", ".")
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "-C", dir, "-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", "init")
	return cmd.Run()
}
