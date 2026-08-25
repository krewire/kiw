# Specification — krewire run / dev / deploy

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | KWN-6K41E                                   |
| Title       | krewire run/dev/deploy — Fullstack Project Commands |
| Status      | Draft                                       |
| Date        | 2026-08-19                                  |
| Author      | Krewire Contributors                         |
| Domain      | Devtool — Project Commands                 |

## 1. Context

The krewire CLI currently builds sites (`rvn build`) and previews them
(`rvn serve`, `rvn init`). With the fullstack monolith (KWF-C4087) apps own an
explicit `main.go`, pages render server-side (KWF-0F2EB), and APIs run on the
same listener (KWF-230KF). The CLI must now drive those apps too.

This spec adds `krewire run` (production serve of an app), `krewire dev`
(development with automatic rebuild/restart), and project-shape detection so
one binary serves three kinds of projects:

| Kind  | Marker                                   | Command path                       |
| ----- | ---------------------------------------- | ---------------------------------- |
| App   | `main.go`/`cmd/*` building a `web.App`  | `krewire run`, `krewire dev`         |
| Site  | `krewire.yaml` + `ssg:`                   | `krewire build`, `krewire serve`     |
| Book  | `manuscript/`                            | `krewire build`, `krewire serve`     |

## 2. Problem Statement

Today there is no way to run a fullstack app from the CLI: `serve` previews a
built static site, and there is no rebuild-on-change workflow for the app's
Go + assets. Developers fall back to manual `go run .` and `go build` loops,
losing the "one command, everything works" promise. Production deployment
likewise has no canonical step: build the binary, export static pages, run assets.

`run`, `dev`, and shape detection close that gap — they make the monolith
first-class without touching the existing static book/site commands.

## 3. Goals

- G1 — Detect a project's kind by its marker files with a stable precedence (explicit > app > site > book).
- G2 — `krewire run` builds and serves an app in production mode (graceful lifecycle via FRK-SRV-021).
- G3 — `krewire dev` rebuilds Go and re-exports static pages on change, restarting with minimal downtime.
- G4 — Preserve `build`/`serve`/`init` semantics for site and book kinds exactly.
- G5 — All project kinds exit through `core` exit codes (0 success / 1 failure / 2 usage).
- G6 — Keep dependencies: the CLI never links the app's backend packages at compile time; it shells out to `go run`/`go build`.
- G7 — `krewire serve` is the universal local-start verb for every kind,
  sharing one runtime bootstrap (`krewire.yaml` → `.env` → env/debug) with
  `run` and `dev`. `run` remains the explicit compile-execute verb with
  argument passthrough. Lineage: dynamic-framework CLIs start projects with
  `serve`/`server` (php artisan serve, rails server, hugo server, jekyll
  serve); compiled-language CLIs keep `run` for one-shot execution of a built
  artifact with args (go run, cargo run, dotnet run).

## 4. Non-Goals

- NG1 — Live in-browser hot reload (HMR); process restart is sufficient in this phase.
- NG2 — A production process manager, container images, or orchestrator; `deploy` remains documented guidance, not a platform.
- NG3 — Changing the site/book build pipeline (KWN-1QGI2).
- NG4 — TLS termination, reverse-proxy, or multi-service config.

## 5. Requirements

### 5.1 Project Shape Detection

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| RND-SHD-001 | Detect **app** when a `main.go` at the project root (or a `cmd/` directory containing packages with `func main`) exists and builds as package `main`. | Must |
| RND-SHD-002 | Detect **site** when `krewire.yaml` declares an `ssg:` key (or `ssg.yaml` exists). | Must |
| RND-SHD-003 | Detect **book** when `manuscript/` exists.                        | Must |
| RND-SHD-004 | Precedence is explicit cfg `project.kind` > app > site > book; ambiguous inputs show a diagnostic and exit `core.ExitCodeUsage`. | Should |
| RND-SHD-005 | Expose detection as an internal API (`internal/shape`) so future commands reuse it; document it in `krewire info`. | Must |

### 5.2 krewire run

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| RND-RUN-001 | `krewire run` compiles the app (`go build` into a temp binary) and executes it with `APP_ADDR`/`--addr` passed through as the listen address. | Must |
| RND-RUN-002 | `krewire run` forwards signals (SIGINT/SIGTERM) to the child so graceful shutdown (FRK-SRV-021) runs in the app, not the CLI. | Must |
| RND-RUN-003 | For site/book kinds, `krewire run` is rejected with a usage error pointing at `build` + `serve`. | Should |
| RND-RUN-004 | Child stdout/stderr stream through; the CLI exits with the child's exit code. | Must |
| RND-RUN-005 | `krewire run` first builds the site/book style export if present (`public/`/`site/`) so embedded assets are current. | Should |

### 5.3 krewire dev

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| RND-DEV-001 | `krewire dev` runs the app in dev mode: start child, watch the Go module files and declared asset/markup roots, rebuild + restart on change. | Must |
| RND-DEV-002 | Watch is debounced; a rebuild failure keeps the previous child running and prints a structured error (KWF-NPFSE-style). | Must |
| RND-DEV-003 | Restart is a clean stop (SIGINT) followed by start, preserving graceful shutdown semantics. | Must |
| RND-DEV-004 | `krewire dev` on a site/book kind behaves exactly like `krewire serve` (existing behavior unchanged). | Must |
| RND-DEV-005 | Watched set defaults to `*.go`, `ui/components/**`, `layouts/**`, `assets/**`, `krewire.yaml`, `ssg.yaml`; configurable via flags/env. | Should |
| RND-DEV-006 | The watched set must be extensible to client JS/TS sources (e.g. `frontend/**`) with a restart (or documented out-of-band build) once the JS/TS bridge lands (KWF-F2TQC) — no CLI rework required. | Should |

### 5.4 krewire deploy

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| RND-DEP-001 | `krewire deploy` validates the project (tests only for Go projects with `go.mod`) then produces a deployable artifact: `krewire build` output plus the compiled app binary in `.krewire/dist/`. | Should |
| RND-DEP-002 | `.krewire/dist/` contains the binary and the exported `site/`; the app is expected to serve both via KWF-C4087 embedding. | Should |
| RND-DEP-003 | Provide a `--target` flag listing supported targets (e.g. `binary`, `gh-pages`); unknown targets exit `core.ExitCodeUsage`. | Should |
| RND-DEP-004 | `--target gh-pages` publishes the staged site to the project's pages branch on `--remote` (default `origin`; branch autodetected `gh-deploy` → `gh-pages`, else `gh-pages`) via a throwaway clone — the user's working tree is untouched; commits use the fixed `Krewire Bot <krewire-bot@krewire.local>` identity; `--dry-run` skips tests and publishing. Binary staging alone never publishes. | Must |

### 5.5 Universal serve & shared runtime bootstrap

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| RND-SRV-001 | `krewire serve` dispatches on project kind exactly like `run`: app builds and listens, cli executes with argument passthrough, site/book preview static output. | Must |
| RND-SRV-002 | `serve`, `run`, and `dev` share one runtime bootstrap resolving, in order: module root → `krewire.yaml` → `.env` → env/debug precedence (KWL-K4T7W). | Must |
| RND-SRV-003 | `kiw run` keeps its distinct contract: argument passthrough after flags and child exit-code propagation (RND-RUN-004). | Must |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Dependencies.** The CLI depends only on stdlib + `libs` (`core`); it does not import `framework/web` at compile time.
- NFR3 — **Restart latency.** `dev` rebuild+restart target `≤ 3 s` on a warm cache.
- NFR4 — **Portability.** Linux, macOS, Windows.
- NFR5 — **Quality gates.** `gofmt`, `go vet ./...`, `go test ./...` pass.

## 7. Success Criteria

- S1 — In an app project, `krewire run` starts the app; SIGINT triggers graceful shutdown recorded in the app logs.
- S2 — In an app project, `krewire dev` restarts after editing a `*.go` file and after editing an asset.
- S3 — In a site/book project, `run` errors with a usage diagnostic and `serve` still works untouched.
- S4 — `krewire info` reports the detected kind and the markers used.
- S5 — `krewire deploy` stages `.krewire/dist/` with the binary + exported `site/` for a mixed app+site project.
- S6 — `kiw deploy --target gh-pages --branch gh-deploy` fast-forwards the remote pages branch with the built site; stale files disappear (clean worktree each publish) and repeated runs are idempotent.

## 8. Related Specifications

| SpecID    | Title                                           |
| --------- | ----------------------------------------------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md) | Krewire Devtool — Initial Specification |
| [KWN-1QGI2](./KWN-BUILD-1QGI2-project-building.md) | Project Building                    |
| [KWN-P0FWA](./KWN-TEST-P0FWA-project-validation.md) | Project Validation                  |
| [KWN-BNKJC](./KWN-INFO-BNKJC-project-information.md) | Project Information                 |
| [KWN-RD3WS](./KWN-SCAFFOLD-RD3WS-project-scaffolding.md) | Project Scaffolding                 |
| [KWF-C4087](https://github.com/krewire/framework/blob/main/docs/specs/KWF-APP-C4087-krewire-app-framework.md) | Krewire App Framework |
| [KWF-CCI0N](https://github.com/krewire/framework/blob/main/docs/specs/KWF-STRUCT-CCI0N-app-directory-structure.md) | App Project Directory Structure Standard |
| [KWF-F2TQC](https://github.com/krewire/framework/blob/main/docs/specs/KWF-JS-F2TQC-js-ts-framework-integration.md) | JS/TS Framework Integration |
| [KWF-0F2EB](https://github.com/krewire/framework/blob/main/docs/specs/KWF-WEB-0F2EB-server-frontend-pipeline.md) | Server & Frontend Rendering Pipeline |

## 9. References

- Go stdlib `os/exec`, `os/signal`, `syscall`.
- Krewire spec conventions (this directory).