// Tests for KWN-JB7PW
package buildinfo

import "testing"

// Spec: KWN-JB7PW RND-VS-004 RND-VS-008 Scope: Unit
func TestKWN_VS_004_ResolveVersion_WorkspaceDevelFallsBackToKnown(t *testing.T) {
	for _, path := range []string{modFramework, modLibs} {
		got, fromSource := ResolveVersion(path)
		if got == "" || got == DevelVersion {
			t.Errorf("ResolveVersion(%q) = %q, want a concrete version for workspace builds", path, got)
		}
		if !fromSource {
			t.Errorf("ResolveVersion(%q) fromSource = false, want true in workspace builds", path)
		}
	}
}

// Spec: KWN-JB7PW RND-VS-004 Scope: Unit
func TestKWN_VS_004_ResolveVersion_UnknownPathReturnsEmpty(t *testing.T) {
	got, fromSource := ResolveVersion("example.com/unknown")
	if got != "" || fromSource {
		t.Errorf("ResolveVersion(unknown) = (%q, %v), want (\"\", false)", got, fromSource)
	}
}
