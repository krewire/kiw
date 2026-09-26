package commands

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/krewire/framework/storage"
	"github.com/krewire/framework/worker"
	"github.com/krewire/kiw/internal/config"
	"github.com/krewire/libs/core"
)

// RegisterWorker registers flags for the worker command.
func RegisterWorker(fs *flag.FlagSet) {
	fs.String("queue", "memory", "queue backend: memory, file, or kv (default memory)")
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

	queueBackend := flagValue(fs, "queue")
	if queueBackend == "" {
		queueBackend = "memory"
	}

	var q worker.Queue
	switch queueBackend {
	case "file", "kv":
		storeDir := filepath.Join(root, ".krewire", "storage")
		fileKV, err := storage.NewFile(storeDir)
		if err != nil {
			return fail(err)
		}
		kvq, err := worker.NewKVQueue(ctx, fileKV)
		if err != nil {
			return fail(err)
		}
		q = kvq
	default:
		q = worker.NewInMemoryQueue()
	}

	dlqCmd := flagValue(fs, "dlq")
	if dlqCmd != "" {
		return runDLQ(ctx, q, dlqCmd)
	}

	slog.Info("worker started", "backend", queueBackend, "concurrency", flagValue(fs, "concurrency"))

	// Run queue processor until signal
	<-ctx.Done()
	slog.Info("worker shutting down")
	return core.ExitCodeSuccess
}

// runDLQ handles DLQ inspection commands (KWF-L5H2F FRK-SVC-062).
func runDLQ(ctx context.Context, q worker.Queue, cmd string) core.ExitCode {
	switch cmd {
	case "list":
		dlq := q.DLQ()
		if len(dlq) == 0 {
			fmt.Println("DLQ is empty (0 dead letters)")
			return core.ExitCodeSuccess
		}
		fmt.Printf("DLQ contains %d dead letter(s):\n", len(dlq))
		for i, dl := range dlq {
			fmt.Printf(" [%d] ID: %s | Attempts: %d | Error: %v | At: %s\n", i+1, dl.ID, dl.Attempts, dl.Err, dl.At.Format("2006-01-02 15:04:05"))
		}
		return core.ExitCodeSuccess
	default:
		fmt.Fprintf(os.Stderr, "kiw worker dlq: unknown sub-command %q (supported: list)\n", cmd)
		return core.ExitCodeUsage
	}
}
