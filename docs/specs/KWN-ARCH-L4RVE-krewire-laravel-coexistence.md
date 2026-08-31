# Specification — Krewire × Laravel Coexistence & Integration

| Field       | Value                                      |
| ----------- | ------------------------------------------ |
| SpecID      | KWN-L4RVE                                  |
| Title       | Krewire × Laravel Coexistence & Integration |
| Status      | Draft                                      |
| Date        | 2026-08-30                                 |
| Author      | Krewire Contributors                        |
| Domain      | Ecosystem Integration                     |
| Scope       | ARCH                                        |

## Scope

This spec targets the **Krewire ecosystem** (Workspace → Module → Domain → Service → Unit). It defines how the Krewire Go framework and `kiw` devtool coexist with, and integrate into, the **Laravel PHP ecosystem** without either ecosystem shadowing or replacing the other.

Containment: `Workspace ⊃ Module ⊃ Domain ⊃ Service ⊃ Unit`. The integration operates at the **Module** level (a Laravel project is a PHP module; a Krewire Go service is a Go module) and the **Service** level (Krewire `service`/`worker` kinds running alongside Laravel `app`).

---

## 1. Context

Krewire is an Indonesia-based open-source community that ships a unified Go framework (`github.com/krewire/framework`) plus a single devtool (`kiw`) covering every web-service workload: CLI, app, site, book, worker, service, and infra. Laravel is the dominant PHP framework in Indonesia and worldwide, known for rapid web development, Eloquent ORM, and a vast ecosystem of packages.

Both ecosystems serve Indonesian developers who build digital products. Krewire's vision is to make software development nearly free and accessible; Laravel's strength is developer productivity for web applications. Rather than competing, these ecosystems can **complement** each other: Laravel handles web routing, ORM, authentication, and blade templating at PHP speed, while Krewire provides Go-compiled performance for microservices, workers, static site generation, WASM frontends, and infrastructure automation.

The Krewire ecosystem already has a mechanism to install into any project directory — the **Guild template** (`github.com/krewire/guild`), invoked via `kiw guild install`. This installs AI agent configuration (AGENTS.md, opencode.json, .agents/) without touching the host project's source code. This spec extends that concept into full functional integration.

### 1.1 Why Coexistence Matters

Indonesian developers overwhelmingly use PHP/Laravel for web development (64M+ MSMEs, ~60% of GDP). Asking them to abandon Laravel for a pure-Go stack creates adoption friction and risks the ecosystem "shadowing" rather than "extending" existing tooling. Coexistence removes that friction: developers keep Laravel where it excels and add Krewire where Go is better.

### 1.2 Precedent: Go + Laravel Integration

The `govel` package (`mpge/govel`) demonstrates that Go binaries called from Laravel via stdin/stdout JSON or gRPC is a proven, production-viable pattern. Krewire can adopt and extend this pattern with its own tooling (`kiw build`, `kiw run`, `kiw worker`).

---

## 2. Problem Statement

### 2.1 The Adoption Barrier

Krewire's current onboarding assumes a greenfield Go project: `kiw new <name>` creates a Go module from scratch. There is no documented or supported path to add Krewire capabilities to an existing Laravel codebase. Developers who want to use Krewire alongside Laravel must manually configure cross-repo tooling, manage two independent build systems, and figure out communication patterns themselves.

### 2.2 The Shadowing Risk

If Krewire positions itself as a Laravel replacement, it risks:
- Alienating the massive Laravel community in Indonesia
- Forcing unnecessary rewrites of proven Laravel applications
- Creating a fragmented developer ecosystem where PHP and Go compete instead of collaborate

### 2.3 The Integration Gap

Today, a Laravel project and a Krewire project in the same repository would:
- Have no shared build/deploy pipeline
- No cross-service discovery
- No unified development experience
- No documented coexistence patterns

This gap must be closed for Krewire to fulfill its mission of making software development accessible to Indonesian developers — regardless of which language they start with.

---

## 3. Goals

| ID        | Goal                                                                                             | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| G1        | **Non-intrusive AI agent install.** `kiw guild install` must work inside a Laravel project directory, adding Krewire's AI agent config without modifying any PHP code. | Must |
| G2        | **Side-by-side project scaffolding.** `kiw new` inside a Laravel project directory must create a Go module subdirectory (e.g., `go-services/`) that coexists with `composer.json`, `app/`, `config/`, etc. | Must |
| G3        | **Build and run integration.** `kiw build` and `kiw run` must work for Krewire projects inside a Laravel repository, compiling Go binaries that are addressable from PHP. | Must |
| G4        | **Communication bridge.** Define standard integration patterns (HTTP/gRPC, process bridge, message queue) for Laravel to call Krewire Go services. | Must |
| G5        | **Unified `krewire.yaml` discovery.** `kiw info` must recognize a Laravel project and report its relationship with any co-located Krewire projects. | Should |
| G6        | **Shared infrastructure.** `kiw deploy` must support deploying both Laravel (PHP) and Krewire (Go) services from a single `infra/` declaration. | Should |
| G7        | **PHP wrapper package.** A `krewire/krewire-laravel` Composer package that wraps `kiw` commands as Artisan commands and provides a Laravel-native facade. | Could |
| G8        | **Cross-stack dev experience.** `kiw dev` must watch Go sources while Laravel's dev server runs concurrently, with unified logging. | Could |

---

## 4. Non-Goals

| ID        | Non-Goal                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------ |
| NG1       | Replacing Laravel's PHP runtime with Go. Krewire does not compile or execute PHP code.            |
| NG2       | Providing a Laravel Eloquent driver in Go. Krewire does not replicate ORM functionality.          |
| NG3       | Embedding Krewire as a Laravel PHP extension or Composer package in this phase (saved for G7).    |
| NG4       | Migrating existing Laravel applications to Krewire. This spec enables coexistence, not migration.  |
| NG5       | Modifying Laravel's core architecture. Krewire integrates at the service/binary boundary.         |
| NG6       | Supporting Laravel Forge, Vapor, or Envoyer as deploy targets. Infrastructure integration is future work (G6). |

---

## 5. Requirements

### 5.1 Must Requirements

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| KWN-LVR-001 | **Guild install works in Laravel projects.** `kiw guild install <laravel-project-dir>` installs AGENTS.md, opencode.json, and .agents/ into the target directory without modifying `composer.json`, `app/`, `config/`, `routes/`, or any PHP file. | Must |
| KWN-LVR-002 | **AGENTS.md is project-agnostic.** The installed AGENTS.md must contain the existing disclaimer: "If the current project is not a Krewire project, follow the generic principles and skip Krewire-specific sections marked Krewire." | Must |
| KWN-LVR-003 | **Side-by-side scaffolding.** `kiw new <name> --dir <laravel-project-dir>` creates a Go module (go.mod, krewire.yaml, main.go, .gitignore) as a subdirectory of the Laravel project without touching Laravel files. | Must |
| KWN-LVR-004 | **No port or process conflict by default.** When both `php artisan serve` (default :8000) and `kiw run` (default :8080) run simultaneously, there must be no port conflict unless explicitly configured to the same port. | Must |
| KWN-LVR-005 | **Independent build systems.** `composer install` and `kiw build` must execute independently without interfering with each other's dependencies or output directories. | Must |
| KWN-LVR-006 | **Process bridge pattern.** `kiw run` must support spawning a Go binary that reads JSON from stdin and writes JSON to stdout, enabling Laravel's Symfony Process component to communicate with it. | Must |
| KWN-LVR-007 | **HTTP/gRPC bridge pattern.** Krewire `service` kind must support HTTP/JSON and gRPC endpoints that Laravel can call via Guzzle HTTP client or a gRPC PHP client. | Must |
| KWN-LVR-008 | **Directory isolation.** Krewire project files (go.mod, krewire.yaml, go/, internal/) must not conflict with Laravel files (composer.json, app/, config/, bootstrap/). | Must |
| KWN-LVR-009 | **Exit code compatibility.** `kiw` commands inside a Laravel project must return standard exit codes (0 success, 1 failure, 2 usage) without being affected by Laravel's own exit codes. | Must |

### 5.2 Should Requirements

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| KWN-LVR-010 | **`kiw info` detects Laravel.** When run inside a Laravel project directory (detected by `composer.json`), `kiw info` reports the project as a Laravel codebase and lists any co-located Krewire projects. | Should |
| KWN-LVR-011 | **Unified `.gitignore`.** When `kiw new` creates a Go subdirectory inside a Laravel project, the Go `.gitignore` must not overwrite Laravel's `.gitignore`. Both files must coexist. | Should |
| KWN-LVR-012 | **Shared `.env` awareness.** Krewire's `kiw run` must respect Go environment variables without interfering with Laravel's `.env` variables. | Should |
| KWN-LVR-013 | **Queue integration.** A Krewire Go worker (`kiw worker`) must be dispatchable from Laravel's queue system via Redis, RabbitMQ, or a message queue bridge. | Should |
| KWN-LVR-014 | **Shared logging.** Krewire Go service logs and Laravel logs must be distinguishable (different prefixes/namespaces) but support aggregation in a shared log output. | Should |

### 5.3 Could Requirements

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| KWN-LVR-015 | **`krewire/krewire-laravel` Composer package.** A PHP package providing Artisan commands (`php artisan krewire:build`, `php artisan krewire:run`, `php artisan krewire:guild-install`) and a `Krewire` facade that wraps `kiw` CLI calls. | Could |
| KWN-LVR-016 | **`kiw dev` concurrency.** `kiw dev` watches Go sources and restarts the Go binary while Laravel's `php artisan serve` runs concurrently, with unified log streaming. | Could |
| KWN-LVR-017 | **`kiw deploy` multi-service.** `kiw deploy --plan` generates an infrastructure plan that includes both the Laravel app and Krewire Go services. | Could |
| KWN-LVR-018 | **WASM frontend for Laravel.** Krewire's `framework/runtime` (Go→WASM) provides the interactive frontend for a Laravel backend, replacing Blade JS components with WASM-powered UI. | Could |
| KWN-LVR-019 | **Shared service registry.** Krewire's `framework/service` registry discovers both Laravel APIs and Go microservices for unified service mesh routing. | Could |

---

## 6. Non-Functional Requirements

| ID | Requirement | Detail |
| -- | ----------- | ------ |
| NFR1 | **Runtime isolation.** Go binaries and PHP-FPM processes must run as separate OS processes. No shared memory, no shared runtime. Communication only through defined boundaries (HTTP, gRPC, stdin/stdout, message queues). |
| NFR2 | **Dependency isolation.** `go.mod` and `composer.json` are independent. No shared dependency resolution. `go build` must not touch `vendor/`; Composer must not touch `go.sum`. |
| NFR3 | **Memory safety.** The `unsafe` package must not be used in any Krewire code that integrates with Laravel. |
| NFR4 | **Deterministic builds.** `kiw build` produces reproducible Go binaries regardless of the Laravel project's state. |
| NFR5 | **Portability.** Integration must work on Linux, macOS, and Windows. |
| NFR6 | **Quality gates.** `gofmt -l .`, `go vet ./...`, `go test ./...` must pass for the Krewire components. Laravel's own `composer lint` / PHPStan must pass independently. |
| NFR7 | **Idempotence.** Re-running `kiw guild install` on a Laravel project must not modify existing managed files unless `--force` is passed. |
| NFR8 | **Security.** Go binaries spawned by Laravel must validate all input from PHP and enforce the same timeout/security boundaries as any external process call. |
| NFR9 | **Performance.** Go worker startup via process bridge must not add more than 100ms overhead per invocation (excluding Go compilation time). gRPC persistent connections have zero startup overhead. |
| NFR10 | **Observability.** Laravel and Krewire logs must include distinguishable source markers. Trace IDs must propagate across the PHP→Go boundary when available. |

---

## 7. Success Criteria

| ID | Criterion | Verification |
| -- | --------- | ------------ |
| S1 | `kiw guild install` inside a fresh Laravel project installs all managed files (AGENTS.md, opencode.json, .agents/) without modifying any PHP file. | Manual test + `diff` before/after install |
| S2 | `kiw new go-backend --dir laravel-project/` creates `go-backend/` with go.mod, krewire.yaml, main.go alongside Laravel's `composer.json` and `app/`. | Directory listing verification |
| S3 | `kiw build && kiw run` inside the Go subdirectory compiles and runs a Go service on :8080 while `php artisan serve` runs on :8000. Both respond to HTTP requests. | `curl :8000` returns Laravel response; `curl :8080` returns Go response |
| S4 | Laravel's Symfony Process component can spawn a Krewire Go binary and communicate via stdin/stdout JSON. | PHP test that calls `Process::fromShellCommandline('kiw run process')` and validates JSON response |
| S5 | `kiw info` inside a Laravel project reports the project type and any co-located Krewire projects. | `kiw info` output parsing |
| S6 | `gofmt -l .`, `go vet ./...`, `go test ./...` pass for Krewire integration code. Laravel's quality gates pass independently. | CI pipeline verification |
| S7 | Re-running `kiw guild install --force` on a Laravel project overwrites managed files; running without `--force` prompts and respects user choice. | Interactive test + `--force` test |

---

## 8. Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                    Laravel Project Directory                       │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │  Laravel (PHP)                                               │ │
│  │  ├── app/            PHP application code                    │ │
│  │  ├── config/         Laravel configuration                  │ │
│  │  ├── routes/         PHP routes                              │ │
│  │  ├── composer.json   PHP dependencies                       │ │
│  │  ├── .env            Laravel environment                    │ │
│  │  └── vendor/         Composer packages                      │ │
│  └──────────────────────┬──────────────────────────────────────┘ │
│                         │                                        │
│          ┌──────────────┼──────────────────┐                     │
│          │ HTTP/gRPC    │                  │                     │
│          │ Process      │                  │                     │
│          │ Queue        │                  │                     │
│          ▼              ▼                  ▼                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │  Go Service   │  │  Go Worker   │  │  Krewire     │           │
│  │  (kiw run)    │  │  (kiw        │  │  Static      │           │
│  │  :8080        │  │   worker)    │  │  Site        │           │
│  │  krewire.yaml │  │  :9090       │  │  (kiw build) │           │
│  └──────┬───────┘  └──────┬───────┘  │  .krewire/   │           │
│         │                 │          └──────────────┘           │
│         └────────┬────────┘                                      │
│                  │                                               │
│  ┌──────────────▼──────────────────────────────────────┐         │
│  │  Shared Infrastructure (Redis, RabbitMQ, PostgreSQL) │         │
│  └─────────────────────────────────────────────────────┘         │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │  Krewire AI Agents (installed by `kiw guild install`)        │ │
│  │  ├── AGENTS.md          Project-agnostic agent constitution │ │
│  │  ├── opencode.json      AI tool configuration               │ │
│  │  └── .agents/           20+ agents, 28 skills, commands     │ │
│  └─────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

### 8.1 Communication Patterns

| Pattern | Laravel → Krewire | Use Case | Latency |
|---------|-------------------|----------|---------|
| **Process bridge** | Symfony Process → Go binary (stdin/stdout JSON) | Image processing, data transformation, report generation | ~50-150ms per invocation |
| **HTTP/JSON** | Guzzle HTTP → Krewire service | CRUD APIs, health checks, config retrieval | ~5-20ms |
| **gRPC** | gRPC PHP client → Krewire service | High-throughput, persistent connections, streaming | ~1-5ms |
| **Message queue** | Laravel queue → Redis/RabbitMQ → Krewire worker | Background jobs, async processing, cron tasks | Depends on queue driver |

---

## 9. Integration Phases

### Phase 1: Coexistence (Immediate — no new code)

- [ ] `kiw guild install` works in Laravel project directories (G1, G2 — already implemented)
- [ ] `kiw new <name> --dir <laravel-dir>` creates side-by-side Go module (G2 — scaffold exists, needs validation)
- [ ] `kiw build` / `kiw run` work for Go subdirectory (G3 — existing behavior, needs validation)

### Phase 2: Functional Integration (Build)

- [ ] Process bridge documentation and examples (G4)
- [ ] HTTP/gRPC service template with Laravel call examples (G4)
- [ ] `kiw info` detects Laravel projects (G5)
- [ ] `.gitignore` coexistence handling (G11)

### Phase 3: Developer Experience (Polish)

- [ ] `krewire/krewire-laravel` Composer package (G7)
- [ ] `kiw dev` concurrent watching (G8)
- [ ] Shared logging format (G14)
- [ ] Artisan commands delegating to `kiw` (G7)

### Phase 4: Infrastructure (Scale)

- [ ] Unified `kiw deploy` for PHP + Go (G6)
- [ ] Shared service registry (G19)
- [ ] WASM frontend integration (G18)

---

## 10. Related Specifications

| SpecID | Title | Relationship |
|--------|-------|-------------|
| [KWF-M8K2Q](https://github.com/krewire/framework/blob/main/docs/specs/KWF-ARCH-M8K2Q-unified-framework-vision.md) | Unified Vision — one Go framework for every workload | Parent vision. Krewire's workload matrix defines the capabilities that integrate with Laravel. |
| [KWL-ARCH-J2K9Q](https://github.com/krewire/libs/blob/main/docs/specs/KWL-ARCH-J2K9Q-scope-levels.md) | Scope Levels | Defines Workspace → Module → Domain → Service → Unit containment. Integration operates at Module and Service scope. |
| [KWN-GUILD-MZ4LE](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-GUILD-MZ4LE-guild-install-command.md) | guild install Command | The install mechanism that enables Path 1 (non-intrusive AI agent config). |
| [KWN-RD3WS](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-SCAFFOLD-RD3WS-project-scaffolding.md) | Project Scaffolding | `kiw new` and `kiw init` mechanisms that enable Path 2 (side-by-side Go module). |
| [KWN-RUN-6K41E](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-RUN-6K41E-krewire-run-dev-deploy.md) | run/dev/deploy | `kiw run` and `kiw dev` mechanisms that enable Path 3 (build and run integration). |
| [KWG-P9ZT4](https://github.com/krewire/guild/blob/main/docs/specs/KWG-INSTALL-P9ZT4-guild-module-install.md) | Guild Module & Install Library | The underlying Go library that `kiw guild install` delegates to. |
| [KWF-L5H2F](https://github.com/krewire/framework/blob/main/docs/specs/KWF-SVC-L5H2F-microservice-patterns.md) | Microservice Patterns | Defines service/worker patterns that integrate with Laravel via HTTP/gRPC/queue. |
| [KWN-CONF-MRCN](./KWN-CONF-MRCN-multi-runtime-config.md) | Multi-Runtime Unified Configuration | Defines runtime namespace isolation, `.env` loading, and `build.depends_on` for Laravel + Krewire coexistence. |
| [KWN-BUILD-MLTB](./KWN-BUILD-MLTB-multi-runtime-build.md) | Multi-Runtime Build Orchestration | Orchestrates `composer install` and `kiw build` in a unified build command. |
| `mpge/govel` | Go-powered task execution for Laravel | Reference implementation for Go+Laravel process/gRPC bridge pattern. |

---

## 11. References

- [Krewire Framework](https://github.com/krewire/framework) — unified Go framework
- [Krewire Devtool (`kiw`)](https://github.com/krewire/kiw) — single CLI for all workloads
- [Krewire Guild](https://github.com/krewire/guild) — AI agent template, installable into any project
- [Krewire Project Vision](https://github.com/krewire/framework/blob/main/internal/docs/project-vision.md) — workload matrix, progressive pipeline, principles
- [Govel (`mpge/govel`)](https://github.com/mpge/govel) — reference implementation for Go+Laravel integration
- [Laravel Framework](https://laravel.com) — the dominant PHP web framework
- [Krewire × Laravel Integration Research](internal docs) — this specification's research baseline

---

## 12. Revision History

Revision history is tracked by **git**, not in-file metadata. To view changes:

```bash
git log --oneline -- docs/specs/KWN-ARCH-L4RVE-krewire-laravel-coexistence.md
```

Initial draft: 2026-08-30.
