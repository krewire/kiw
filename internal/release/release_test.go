package release

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/krewire/libs/core"
)

func TestBump(t *testing.T) {
	v := core.MustParseVersion("0.3.1")
	if got := Bump(v, BumpPatch); got.String() != "0.3.2" {
		t.Errorf("Bump patch = %s, want 0.3.2", got)
	}
	if got := Bump(v, BumpMinor); got.String() != "0.4.0" {
		t.Errorf("Bump minor = %s, want 0.4.0", got)
	}
	if got := Bump(v, BumpMajor); got.String() != "1.0.0" {
		t.Errorf("Bump major = %s, want 1.0.0", got)
	}
}

func TestPlanReleasingLibsPropagatesToAllDependents(t *testing.T) {
	edits, err := Plan([]core.ModuleName{core.ModuleLibs}, BumpPatch)
	if err != nil {
		t.Fatal(err)
	}
	// own bump + framework, mdbind, guild, ship, kiw = 6 edits
	if len(edits) != 6 {
		t.Fatalf("got %d edits, want 6: %+v", len(edits), edits)
	}
	nv := Bump(mustCur(core.ModuleLibs), BumpPatch)
	for _, e := range edits {
		if e.Module == core.ModuleLibs {
			if e.To != versionDecl(core.ModuleLibs, nv.String()) {
				t.Errorf("libs own edit To = %q, want %q", e.To, versionDecl(core.ModuleLibs, nv.String()))
			}
		} else {
			want := reqDecl(core.ModuleLibs, nv.String())
			if e.To != want {
				t.Errorf("dependent %s To = %q, want %q", e.Module, e.To, want)
			}
		}
	}
}

func TestPlanReleasingFrameworkTouchesOnlyKiw(t *testing.T) {
	edits, err := Plan([]core.ModuleName{core.ModuleFramework}, BumpPatch)
	if err != nil {
		t.Fatal(err)
	}
	if len(edits) != 2 {
		t.Fatalf("got %d edits, want 2 (framework + kiw): %+v", len(edits), edits)
	}
}

func TestPlanAllModules(t *testing.T) {
	var all []core.ModuleName
	for _, m := range Modules {
		all = append(all, m.Name)
	}
	edits, err := Plan(all, BumpPatch)
	if err != nil {
		t.Fatal(err)
	}
	// 6 own bumps + 9 dependent edges (libs:5, framework:1, mdbind:1, guild:1, ship:1) = 15
	if len(edits) != 15 {
		t.Fatalf("got %d edits, want 15: %+v", len(edits), edits)
	}
}

func TestApplyWritesAndValidates(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "v.go")
	orig := `var Version = core.MustParseVersion("0.1.0")`
	if err := os.WriteFile(p, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}
	edits := []Edit{{Module: core.ModuleGuild, File: "v.go", From: `var Version = core.MustParseVersion("0.1.0")`, To: `var Version = core.MustParseVersion("0.2.0")`}}

	if _, err := Apply(edits, dir, true); err != nil {
		t.Fatalf("dry-run apply error: %v", err)
	}
	if data, _ := os.ReadFile(p); string(data) != orig {
		t.Errorf("dry-run modified the file: %q", string(data))
	}
	if _, err := Apply(edits, dir, false); err != nil {
		t.Fatalf("apply error: %v", err)
	}
	if data, _ := os.ReadFile(p); string(data) != `var Version = core.MustParseVersion("0.2.0")` {
		t.Errorf("file not updated: %q", string(data))
	}
}

func TestApplyRejectsAmbiguousMatch(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "v.go")
	_ = os.WriteFile(p, []byte(`x: core.MustParseVersion("0.1.0")
y: core.MustParseVersion("0.1.0")`), 0o644)
	edits := []Edit{{Module: core.ModuleGuild, File: "v.go", From: `core.MustParseVersion("0.1.0")`, To: `core.MustParseVersion("0.2.0")`}}
	if _, err := Apply(edits, dir, false); err == nil {
		t.Error("expected ambiguous-match error")
	}
}

func mustCur(n core.ModuleName) core.Version {
	v, err := CurrentVersion(n)
	if err != nil {
		panic(err)
	}
	return v
}
