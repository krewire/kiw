# Specification — kiw Script Runner — Go as Scripting Language

| Field   | Value |
|---------|-------|
| SpecID  | KWN-SCRIPT-9F3KQ |
| Title   | kiw Script Runner — `kiw run <task>` and `kiw run path/to/file.go` |
| Status  | Draft |
| Date    | 2026-08-27 |
| Author  | Krewire Contributors |
| Domain  | Devtool — Project Commands |
| Reviewers | Krewire Maintainers |
| Deciders | Krewire Maintainers |
| Stakeholders | CLI users, Worker operators, AI Agents |

## 1. Summary

Extend `kiw run` — today only `kiw run`/`kiw run -- <args>` for full `app`/`cli` projects (`KWN-RUN-6K41E`) — to also run **single Go files** (`kiw run path/to/file.go -- args`) and **named tasks** (`kiw run <task>` where `<task>` is `scripts.<name>` in `krewire.yaml`). Go's `go run` with build cache already enables scripting; `kiw` becomes the project-aware runtime that injects `krewire.yaml` + `.env` + `KIW_ENV/KIW_DEBUG`, forwards signals, and preserves backward compatibility.

## 2. Background & Context

`kiw` (module `github.com/krewire/kiw`, binary `kiw`) is the single CLI for all 8 kinds (`internal/docs/project-vision.md`, `libs/core/workload.go:Kind` — `app/cli/site/book/worker/service/infra/kernel`). `kiw run` (`kiw/internal/commands/run.go:34`) currently compiles the whole module (`go build -o /tmp/... .`) and execs the binary; `kiw dev` watches and restarts. `krewire.yaml` is the only config (`libs/config`, `kiw/internal/config/conf.go:DefaultOutput .krewire/build`).

Developers coming from `composer run` / `npm run` / `python file.py` expect two ergonomics Go lacks natively:
* **File-as-script:** `go run tools/migrate.go` works but ignores project env (`krewire.yaml` → `.env` → `KIW_ENV` per `KWL-K4T7W` / `KWN-CONF-Q7X4M`) and signal forwarding that `kiw run` already provides for apps.
* **Task alias:** `npm run seed` maps to `scripts.seed` in `package.json`; Krewire has no equivalent in `krewire.yaml`, so teams fall back to `make`/`just` outside the single-CLI promise (`KWF-M8K2Q`).

Both are **devtool concerns**, not new workloads — they enrich `cli`/`kernel` without new `framework/*` imports, staying `core` stdlib-only and `kern` imperative.

## 3. Problem Statement

* **Who:** Developers and workers writing one-off Go scripts (migrations, seeds, linters, generators) inside a Krewire project.
* **Pain:** `go run ./tools/seed.go` bypasses `krewire.yaml`/`.env` precedence, `KIW_ENV`/`KIW_DEBUG` injection (`run.go:childEnviron`), and unified exit codes (`core.ExitCodeSuccess/Failure/Usage`). `make` is introduced as a second runner, breaking *one CLI* (`KWF-M8K2Q` principle). No `kiw run --help` documents file or task mode; existing `kiw/docs/specs/KWN-RUN-6K41E:RND-RUN-001..005` do not cover it.
* **If not solved:** Teams keep two toolchains (`go` + `make`), scripts diverge per project, onboarding cost rises — directly contradicting Krewire's near-zero-cost mission for 64M MSMEs.

## 4. Goals & Non-Goals

### Goals
* G1 — `kiw run path/to/file.go [-- args]` executes a single Go file with project env inherited (Must, testable).
* G2 — `kiw run <task>` executes `scripts.<task>` from `krewire.yaml` with `sh -c` (Unix) / `cmd /C` (Windows), streaming stdio and propagating exit code (Must).
* G3 — Backward compatible: bare `kiw run` still builds `app`/`cli` projects exactly as `KWN-RUN-6K41E:RND-RUN-001` (Must).
* G4 — Spec traceability `KWL-TEST-P8M4L`: every Must requirement has `// Tests for KWN-SCRIPT-9F3KQ` and `// Spec: KWN-SCRIPT-9F3KQ KWN-SCR-### Scope: Module` (Must).

### Non-Goals
* NG1 — No Go interpreter/REPL (`yaegi`, `gomacro`) and no `unsafe` (security NFR from `KWN-RUN-6K41E:NFR1`).
* NG2 — No new `project.kind`; scripting is file-level, not a workload (`KWL-ARCH-J2K9Q`).
* NG3 — No shebang (`#!/usr/bin/env kiw run`) or persistent hash cache in MVP — deferred to Full phase.
* NG4 — No HMR/live reload for scripts — `kiw dev` remains for `app`; scripts use `kiw run`.

### 4.5 Assumptions & Constraints

| ID | Assumption / Constraint | Type | Validation |
|----|-------------------------|------|------------|
| A1 | Go 1.22+, `go` toolchain on PATH, `GOCACHE` enabled | Assumption | `go version` in `kiw info` |
| A2 | `krewire.yaml` is single config; `.env` overlay via `KWN-CONF-Q7X4M`/`KWL-K4T7W` | Assumption | `kiw/internal/config` existent |
| A3 | Scripts are trusted (from version-controlled `krewire.yaml`) | Assumption | Docs note |
| C1 | `libs/core` stays stdlib-only; `kiw` never imports `framework/*` | Constraint | `go vet ./...` + `arch-guard` |
| C2 | Exit codes 0/1/2 via `core.ExitCode*` (`KWL-W0J2X`) | Constraint | `core.ExitCodeFromInt` |
| C3 | No `go.work` at repo root; `go.work` only at workspace hub | Constraint | `AGENTS.md` |

### 4.6 Glossary

| Term | Definition | Source |
|------|------------|--------|
| Task alias | `scripts.<name>` in `krewire.yaml` → shell command, like `npm run` | This spec |
| File runner | `kiw run path/to/file.go` delegation to `go run` with project env | This spec |
| Runtime bootstrap | `krewire.yaml` → `.env` → `KIW_ENV/KIW_DEBUG` → child env | `KWN-RUN-6K41E:RND-SRV-002` |

## 5. Requirements

### 5.1 Functional Requirements

| ID | Requirement | Scope | Priority | RFC 2119 |
|----|-------------|-------|----------|----------|
| KWN-SCR-001 | `kiw run <path>` where `<path>` exists and ends with `.go` **MUST** execute `go run <path>` with all trailing args forwarded after `--` or directly, inheriting project env (`KIW_ENV`, `KIW_DEBUG`, `APP_ADDR` empty) and streaming stdio; exit code **MUST** be the child's (`core.ExitCodeFromInt`). | Module | Must | MUST |
| KWN-SCR-002 | `kiw run <task>` where `<task>` matches a key in `krewire.yaml` `scripts` map and is not a `.go` file **MUST** execute the command string via `sh -c` (Unix) or `cmd /C` (Windows), streaming stdio, forwarding SIGINT/SIGTERM, and returning the child's exit code; unknown task **MUST** exit `core.ExitCodeUsage` with `Available tasks: ...` hint. | Module | Must | MUST |
| KWN-SCR-003 | Bare `kiw run` (no args) **MUST** preserve existing `app`/`cli` behavior (`go build -o /tmp ... .` then exec) per `KWN-RUN-6K41E:RND-RUN-001`; `site`/`book` kinds **MUST** still be rejected with `build + serve` guidance. | Module | Must | MUST |
| KWN-SCR-004 | Precedence **MUST** be: file exists (`.go`) > scripts key > project run; ambiguity (`seed.go` file and `scripts.seed` both exist) **MUST** prefer file and emit a warning to stderr `warning: both file and task "seed" exist; running file`. | Module | Must | MUST |
| KWN-SCR-005 | `krewire.yaml` **MUST** support top-level `scripts: map[string]string` (optional, zero-value nil) validated by `libs/validate` (no `validate:"required"`); missing map **MUST** be treated as empty. | Module | Must | MUST |
| KWN-SCR-006 | `kiw run --help` and `kiw help run` **MUST** document file and task modes with examples. | Module | Should | SHOULD |
| KWN-SCR-007 | `kiw info` **SHOULD** list `scripts` keys when a `krewire.yaml` with `scripts` is detected. | Module | Should | SHOULD |

### 5.2 Non-Functional Requirements

| ID | Category | Requirement |
|----|----------|-------------|
| NFR1 | Performance | File runner cold `≤ 4s`, warm (GOCACHE) `≤ 1.5s` on CI baseline; task alias overhead `≤ 50ms`. |
| NFR2 | Quality Gates | `gofmt -l .` empty, `go vet ./...` clean, `go test ./...` passes (`KWN-RUN-6K41E:NFR5`). |
| NFR3 | Stdlib-first | Only `os/exec`, `os/signal`, `path/filepath`, `runtime` plus `libs/core`/`libs/validate`; no third-party. |
| NFR4 | Portability | Linux, macOS, Windows (`sh -c` vs `cmd /C` branch, `gofmt` consistent). |
| NFR5 | Security | `unsafe` **MUST NOT** be used; scripts assumed trusted; no shell injection beyond `sh -c` of the literal `scripts` value. |
| NFR6 | Compatibility | Backward compat with `KWN-RUN-6K41E` and `krewire.yaml` without `scripts` (zero-cost when unused). |

## 6. Detailed Design / Proposal

### 6.1 Architecture

```
Workspace (go.work hub, bin/kiw)
 └─ Module: github.com/krewire/kiw
     ├─ kiw/internal/config/conf.go  — add Scripts map, inject into build
     ├─ kiw/internal/commands/run.go — branch: isGoFile(path) ? runGoFile : isScript(task) ? runScript : runApp/runCLI (existing)
     ├─ kiw/internal/commands/info.go — list scripts
     └─ libs/core, libs/validate      — unchanged (core stdlib-only)
```

Dependency direction: `kiw` → `libs` (core/validate/config) — no `framework` import, preserving `AGENTS.md` layer `libs ← framework/mdbind ← kiw ← guild`.

### 6.2 API Design

```go
// kiw/internal/config/conf.go
type Config struct {
    Scripts map[string]string `yaml:"scripts"` // KWN-SCR-005
}

// kiw/internal/commands/run.go
func isGoFileArg(root, arg string) bool // stat + HasSuffix .go
func runGoFile(rt *runtimeEnv, path string, args []string) core.ExitCode  // go run path -- args
func runScript(rt *runtimeEnv, task string) core.ExitCode                  // sh -c Scripts[task]
```

CLI surface (backward compat):
```
kiw run [task|path/to/file.go] [-- args...]
  kiw run                   # project app/cli (existing)
  kiw run tools/seed.go -- --env local
  kiw run lint              # scripts.lint from krewire.yaml
```

`krewire.yaml` example:
```yaml
project: { kind: cli, name: mytools }
scripts:
  seed: "go run ./tools/seed.go --env local"
  lint: "go vet ./..."
```

### 6.3 Alternatives Considered

| Alternative | Pros | Cons | Why rejected |
|-------------|------|------|--------------|
| A — Embed `yaegi` interpreter | True instant start, no `go` needed | Adds ~5MB, `unsafe` risk, diverges from `go run` semantics, new dep | Violates NFR3/NFR5, not stdlib |
| B — New `kiw script` command | No ambiguity with `kiw run` | Two verbs for same idea, breaks *one CLI* mental model, more docs | Higher effort, lower discoverability than overloading `run` |
| C — This proposal: overload `kiw run` with file>script>project precedence | Single verb, npm/composer familiarity, leverages `go run` cache, additive | Precedence needs warning on ambiguity | Best impact-to-effort; precedence warning mitigates |

Decision: **C** per impact-to-effort ordering (`agent-workflow` skill).

### 6.4 System Context & Diagrams

```
User
 ├─ kiw run file.go ──► bootRuntime (krewire.yaml→.env→KIW_ENV) ──► go run file.go (childEnviron + waitChild/SIGINT)
 ├─ kiw run seed ────► bootRuntime ──► Scripts[seed] ──► sh -c "go run ./tools/seed.go" (waitChild)
 └─ kiw run (no arg) ► shape.Detect (app/cli) ──► runApp/runCLI (go build -o /tmp + exec)
```

### 6.5 Cost & Performance

| Aspect | Estimate | Notes |
|--------|----------|-------|
| Dev cost | S (1 week MVP: spec 0.5d + config 1d + file runner 1.5d + task alias 2d + docs/tests 1d) | Single engineer, ~300 LOC + 150 test LOC |
| Runtime cost | Zero when unused (nil map); task alias +50ms shell spawn; file runner = `go run` cost | `GOCACHE` warm 0.8–1.5s |
| Binary size | +~2KB (no new deps) | `go vet` size proof |

### 6.6 Security, Privacy & Compliance

* Scripts trusted (version-controlled). No `unsafe`. Shell is `sh -c` of literal YAML value — no user-input interpolation. `KIW_ENV/KIW_DEBUG` sanitized via `childEnviron`.
* Secrets via `.env` / `krewire.yaml` `secret.Ref`, never inlined in `scripts` value (guidance in docs).

### 6.7 Accessibility & Internationalization

* `tui` output English per `AGENTS.md`; error `Available tasks: ...` uses `term` formatting.

### 6.8 Observability

* `log/slog` Info for `running script` / `running go file` (like `building app` in `run.go:79`). Errors via `core.FormatTree` + `printStackIfDebug`.

## 7. Dependencies & Impact

| Dependency | Relation | Impact |
|------------|----------|--------|
| `KWN-RUN-6K41E` | Extends RND-RUN-001..005 and RND-SRV-001..003 | No breaking change; precedence added |
| `KWN-CONF-Q7X4M` / `KWL-K4T7W` | Reuses `bootRuntime` env chain | Scripts inherit same chain |
| `KWL-ARCH-J2K9Q` | Scope `Module` for all requirements | Test location `kiw/internal/commands` |
| `libs/config` | `Scripts` field lives here if shared; otherwise `kiw/internal/config` | Downstream `go.mod` bump not needed (field additive) |

Impact: `framework/mdbind/guild/libs` unchanged. `go.work` unchanged. `AGENTS.md` command matrix to be synced post-implementation.

## 8. Related Specifications

| SpecID | Title |
|--------|-------|
| [KWN-RUN-6K41E](./KWN-RUN-6K41E-krewire-run-dev-deploy.md) | krewire run/dev/deploy (parent) |
| [KWN-CONF-Q7X4M](./KWN-CONF-Q7X4M-config-directory-and-dotenv.md) | Config Directory & Dotenv |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md) | Krewire Devtool Initial Spec |
| [KWL-ARCH-J2K9Q](https://github.com/krewire/libs/blob/main/docs/specs/KWL-ARCH-J2K9Q-ecosystem-scope-levels.md) | Ecosystem Scope Levels |
| [KWL-TEST-P8M4L](https://github.com/krewire/libs/blob/main/docs/specs/KWL-TEST-P8M4L-spec-driven-testing.md) | Spec-Driven Testing |

## 9. References

* `kiw/internal/commands/run.go:34` — current `RunRun` + `runApp`/`runCLI` + `waitChild`
* `kiw/internal/config/conf.go` — `Config` struct (`Output`, `Input`, `SSG`)
* `kiw/cmd/kiw/main.go:18` — `tui.NewCommand("run", ...)`
* Go `os/exec`, `os/signal`, `syscall` (stdib) — already used in `run.go`
* `npm run` / `composer run` / `python file.py` ergonomics — user request 2026-08-27
