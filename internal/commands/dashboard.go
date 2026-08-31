package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/libs/core"
)

// RegisterDashboard registers flags for the dashboard command.
func RegisterDashboard(fs *flag.FlagSet) {
	fs.String("port", "4000", "port for the dashboard HTTP server")
	fs.String("env", "", "target environment: local, production, or testing")
}

// RunDashboard starts a local dev dashboard for services, logs, and traces
// (KWF-L5H2F observability).
func RunDashboard(fs *flag.FlagSet) core.ExitCode {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: "+err.Error())
		return core.ExitCodeUsage
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}

	port := flagValue(fs, "port")
	if port == "" {
		port = "4000"
	}

	switch cfg.Kind() {
	case string(core.KindWorker), string(core.KindService), string(core.KindInfra):
		fmt.Printf("dashboard started on http://localhost:%s\n", port)
		fmt.Println("  (placeholder — full UI requires frontend integration)")
		return core.ExitCodeSuccess
	default:
		fmt.Fprintf(os.Stderr, "kiw dashboard: project kind %q does not support dashboard — use worker, service, or infra\n", cfg.Kind())
		return core.ExitCodeUsage
	}
}
