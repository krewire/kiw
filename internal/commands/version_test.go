// Tests for KWN-JB7PW
package commands

import (
	"strings"
	"testing"
)

// Spec: KWN-JB7PW RND-VS-004 Scope: Package
func TestKWN_VS_004_HumanVersion_CollapsesUnknownToDev(t *testing.T) {
	cases := map[string]string{
		"":              "dev",
		"(devel)":       "dev",
		"v(devel)":      "dev",
		"devel":         "dev",
		"0.5.1":         "v0.5.1",
		"v0.5.1":        "v0.5.1",
		"0.2.0-alpha.1": "v0.2.0-alpha.1",
	}
	for in, want := range cases {
		if got := humanVersion(in); got != want {
			t.Errorf("humanVersion(%q) = %q, want %q", in, got, want)
		}
	}
}

// Spec: KWN-JB7PW RND-VS-005 Scope: Package
func TestKWN_VS_005_HumanVersion_SingleLeadingV(t *testing.T) {
	for _, in := range []string{"v0.5.1", "0.5.1", "v(devel)", "(devel)", "devel", ""} {
		got := humanVersion(in)
		if got == "dev" {
			continue
		}
		if len(got) < 2 || got[0] != 'v' || got[1] == 'v' {
			t.Errorf("humanVersion(%q) = %q, want exactly one leading v or dev", in, got)
		}
	}
}

// Spec: KWN-JB7PW RND-VS-009 Scope: Package
func TestKWN_VS_009_QualifiedVersion_MarksWorkspaceBuilds(t *testing.T) {
	for _, path := range []string{ModFramework, ModLibs} {
		got := qualifiedVersion(path)
		if strings.Contains(got, "devel") {
			t.Errorf("qualifiedVersion(%q) = %q, must never print devel", path, got)
		}
		if !strings.HasSuffix(got, "(dev)") {
			t.Errorf("qualifiedVersion(%q) = %q, want a (dev) marker in workspace builds", path, got)
		}
	}
}

// Spec: KWN-JB7PW S1 Scope: Package
func TestKWN_JB7PW_S1_ResolveVersions_MatchInfoPaths(t *testing.T) {
	fw, libs := resolveVersions()
	if fw == "" || libs == "" {
		t.Fatalf("resolveVersions = (%q, %q), want usable versions", fw, libs)
	}
	if fw != moduleVersion(ModFramework) || libs != moduleVersion(ModLibs) {
		t.Errorf("resolveVersions (%q, %q) disagrees with moduleVersion paths (%q, %q)",
			fw, libs, moduleVersion(ModFramework), moduleVersion(ModLibs))
	}
}
