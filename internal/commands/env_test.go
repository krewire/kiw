// Tests for KWL-K4T7W
package commands

import (
	"errors"
	"slices"
	"testing"

	"github.com/krewire/libs/core"
)

// Spec: KWL-K4T7W KWL-ENVV-006 Scope: Func
func TestKWL_ENVV_006_ChildEnviron_ExportsRVEnvAndDebug(t *testing.T) {
	base := []string{"PATH=/bin"}
	got := childEnviron(base, ":8080", core.EnvTesting, true)
	want := []string{"PATH=/bin", "APP_ADDR=:8080", "KIW_ENV=testing", "KIW_DEBUG=1"}
	if !slices.Equal(got, want) {
		t.Errorf("childEnviron = %v, want %v", got, want)
	}

	noAddr := childEnviron(nil, "", core.EnvProduction, false)
	wantNoAddr := []string{"KIW_ENV=production", "KIW_DEBUG=0"}
	if !slices.Equal(noAddr, wantNoAddr) {
		t.Errorf("childEnviron(no addr) = %v, want %v", noAddr, wantNoAddr)
	}
}

// Spec: KWL-K4T7W KWL-ENVV-007 Scope: Func
func TestKWL_ENVV_007_UsageOrFail_MapsUsageErrorsToExit2(t *testing.T) {
	if got := usageOrFail(core.UsageError("bad")); got != core.ExitCodeUsage {
		t.Errorf("usageOrFail(UsageError) = %d, want %d", got, core.ExitCodeUsage)
	}
	if got := usageOrFail(errors.New("boom")); got != core.ExitCodeFailure {
		t.Errorf("usageOrFail(plain error) = %d, want %d", got, core.ExitCodeFailure)
	}
}
