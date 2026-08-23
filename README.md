# Krewire

**Krewire** is the single CLI entry point for the entire Krewire ecosystem — `github.com/krewire/krewire`. It drives all eight project kinds (`app`, `cli`, `site`, `book`, `worker`, `service`, `infra`, `kernel`) from one binary, one config file (`krewire.yaml`), and one workflow. The binary is named **`kiw`** for fast typing (`krewire` remains the module/repo name).

The devtool dogfoods the unified framework's `cli` package, so the tool that manages the ecosystem is itself built on the stack it manages.

> Unified vision: [`KWF-M8K2Q`](../framework/docs/specs/KWF-ARCH-M8K2Q-unified-framework-vision.md)

## Commands

| Command | Purpose | Project Kinds |
|---------|---------|---------------|
| `kiw new <name>` | Scaffold a minimal kernel | any new project |
| `kiw init` | Equip a kernel in place (default: fullstack `app`) | kernel |
| `kiw init --static` | Equip a declarative static site (`ssg:`) | kernel |
| `kiw init --book` | Equip a manuscript book (`book`) | kernel |
| `kiw init --cli` | Equip a CLI app (`cli`) | kernel |
| `kiw init --worker` | Equip a worker (`worker`) | kernel — *planned* |
| `kiw init --service` | Equip a microservice (`service`) | kernel — *planned* |
| `kiw init --infra` | Equip cloud infra (`infra`) | kernel — *planned* |
| `kiw init --template <git-url>` | Clone a starter | empty dir |
| `kiw build` | Build (binary / `site/` / infra plan) | all |
| `kiw serve` | Serve the built site | `site`, `book` |
| `kiw run [args]` | Build & run the binary | `app`, `cli`, `worker`, `service` |
| `kiw dev` | Rebuild + auto-restart (incl. WASM) | `app`, `cli`, `worker`, `service` |
| `kiw worker` | Run background workers | `worker` — *planned* |
| `kiw deploy` | Provision infra + deploy (`--plan`/`--preview`/`--destroy`) | all deployable — *planned* |
| `kiw dashboard` | Local dev dashboard (services, logs, traces, infra) | `worker`, `service`, `infra` — *planned* |
| `kiw generate` | Code generation (OpenAPI, config, etc.) | all — *planned* |
| `kiw test` | Run `go test ./...` (spawn Go toolchain) | all |
| `kiw vet` | Run `go vet ./...` (spawn Go toolchain) | all |
| `kiw fmt` | Check/format with `gofmt -l/-w` or `go fmt ./...` (spawn Go toolchain) | all |
| `kiw info` | Environment + detected kind | all |
| `kiw version` | CLI + framework versions | all |
| `kiw guild install` | Install the Guild AI template | any project |
| `kiw help <cmd>` / `kiw <cmd> help` / `kiw <cmd> --help` | Show help for a command | all |

`build`/`run`/`dev` dispatch on the detected kind; `deploy` owns the path from artifact to running infrastructure (AWS + Kubernetes first via `framework/infra`).

## Getting Started

### Prerequisites

- Go 1.22+

### Build

```bash
go build -o kiw ./cmd/kiw
# legacy path still works:
# go build -o kiw ./cmd/krewire
```

### Use

```bash
./kiw version
./kiw info
./kiw new myapp && cd myapp
../kiw build
../kiw serve   # for site/book
../kiw run     # for app/cli/worker/service
```

Scaffold variants:

```bash
./kiw new hello
cd hello && go run . hello   # CLI example before equipping
```

## Design

- **Dogfooding** — `cmd/kiw` (compat: `cmd/krewire`) is built on `github.com/krewire/framework/tui` with ecosystem exit codes (0/1/2) and `term` output.
- **Single config** — all kinds use `krewire.yaml` only; no `ssg.yaml`.
- **Kind dispatch** — `kiw info` prints the detected kind; `kiw build` picks the pipeline (SSG vs. book vs. binary vs. infra plan).

## Specifications

- `KWN-DEVTOOL-Z0VFC` — Krewire devtool
- `KWN-INIT-7QM2X` — Init project variants
- `KWN-BUILD-1QGI2` — Project building
- `KWN-RUN-6K41E` — Run / dev / deploy

Full index: [`krewire/docs/specs/index.md`](docs/specs/index.md).

## License

MIT — see [LICENSE](LICENSE).
