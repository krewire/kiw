# Architecture — Krewire Devtool

## Module Structure

```
krewire/
├── cmd/kiw/               # Entry — built via `go build -o kiw ./cmd/kiw` (compat: cmd/krewire), dogfoods framework/tui
├── internal/
│   ├── commands/         # build/serve/run/dev, worker, deploy (--plan/--preview/--destroy), dashboard, generate, info, version, guild, test
│   ├── scaffold/         # `kiw new` / `kiw init` templates for 8 kinds
│   ├── shape/            # Kind detection (krewire.yaml + manuscript/ + infra/ + main.go)
│   ├── siteconfig/       # ssg: / book: config handling
│   ├── buildinfo/        # version embedding
│   ├── gomod/            # go.mod helpers
│   └── version/          # version reporting
└── docs/
```

**Design decisions:**

- **Kind dispatch.** `kiw info` prints detected kind; `kiw build` picks pipeline (binary / `site/` / `manuscript/`→`site/` / infra plan) based on `project.kind` or markers.
- **Dogfooding.** `cmd/kiw` (compat: `cmd/krewire`) is built on `framework/tui` with `libs/core` exit codes and `term` output.
- **No per-project `cmd/`.** All CLI behavior lives in `krewire/krewire`; projects have no `cmd/` binaries for build/serve/run.
- **Binary name.** The CLI binary is `kiw`; the module remains `github.com/krewire/krewire`.


## Conventions

- Documentation in English, Markdown, spec-driven (`internal/docs/specs/krewire/` in `krewire/internal`).
- Quality gates: `gofmt -l .`, `go vet ./...`, `go test ./...` in each Go repo; per-kind `kiw build` / `kiw build --plan` spot-checks.
- Cross-repo testing via `go.work` workspace (`./framework`, `./libs`, etc.) at hub root; `go work sync` updates `go.work.sum`.
