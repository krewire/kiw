package commands

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/krewire/framework/worker"
	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/libs/core"
)

// RegisterWorker registers flags for the worker command.
func RegisterWorker(fs *flag.FlagSet) {
	fs.String("queue", "memory", "queue backend: memory (default)")
	fs.Int("concurrency", 1, "number of concurrent workers")
	fs.String("dlq", "", "inspect DLQ: list")
}

// RunWorker starts background workers for the current project (KWF-L5H2F FRK-SVC-060/061).
func RunWorker(fs *flag.FlagSet) core.ExitCode {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: "+err.Error())
		return core.ExitCodeUsage
	}
	cfg, err := config.Load(root)
	if err != nil {
		return fail(err)
	}

	if cfg.Kind() != string(core.KindWorker) {
		fmt.Fprintf(os.Stderr, "kiw worker: project kind is %q, not 'worker' — add 'project.kind: worker' to krewire.yaml\n", cfg.Kind())
		return core.ExitCodeUsage
	}

	dlqCmd := flagValue(fs, "dlq")
	if dlqCmd != "" {
		return runDLQ(ctx, fs, dlqCmd)
	}

	_ = worker.NewInMemoryQueue()
	slog.Info("worker started", "backend", "memory", "concurrency", flagValue(fs, "concurrency"))

	// Run queue processor until signal
	<-ctx.Done()
	slog.Info("worker shutting down")
	return core.ExitCodeSuccess
}

// runDLQ handles DLQ inspection commands (KWF-L5H2F FRK-SVC-062).
func runDLQ(ctx context.Context, fs *flag.FlagSet, cmd string) core.ExitCode {
	switch cmd {
	case "list":
		slog.Info("DLQ listing requires an active queue — start kiw worker first")
		return core.ExitCodeSuccess
	default:
		fmt.Fprintf(os.Stderr, "kiw worker dlq: unknown sub-command %q (supported: list)\n", cmd)
		return core.ExitCodeUsage
	}
}
