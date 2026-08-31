# Specification — Multi-Runtime Unified Configuration

| Field       | Value                                      |
| ----------- | ------------------------------------------ |
| SpecID      | KWN-MRCN                                   |
| Title       | Multi-Runtime Unified Configuration        |
| Status      | Draft                                      |
| Date        | 2026-08-30                                 |
| Author      | Krewire Contributors                        |
| Domain      | Devtools — Configuration                  |
| Scope       | CONF                                        |

## Scope

This spec targets the **Krewire ecosystem** at the **Module** and **Domain** scope levels. It defines how Krewire's `krewire.yaml` configuration coexists with and integrates with configuration files from other runtimes — specifically `.env` (Node.js/PHP), `composer.json` (PHP), and `package.json` (Node.js) — within a single project directory.

Containment: `Workspace ⊃ Module ⊃ Domain ⊃ Service ⊃ Unit`. The configuration spec operates at the **Module** level (a project's configuration) and the **Domain** level (configuration parsing and validation).

---

## 1. Context

### 1.1 The Configuration Fragmentation Problem

A project that integrates Krewire with Node.js and/or PHP has multiple configuration files, each with different formats, purposes, and loaders:

| File | Runtime | Format | Purpose |
|------|---------|--------|---------|
| `krewire.yaml` | Krewire (Go) | YAML | Project kind, build config, service/worker settings |
| `composer.json` | PHP (Composer) | JSON | PHP dependencies, scripts, autoload |
| `package.json` | Node.js (npm/pnpm/yarn) | JSON | JS dependencies, scripts, devDependencies |
| `.env` | Both (PHP/Node.js) | key=value | Environment variables |
| `.env.example` | Both | key=value | Environment variable documentation |

These files coexist in the same directory but are loaded by different tools (`kiw`, `composer`, `npm`, `dotenv`). Without coordination, configuration drift, port conflicts, and environment variable collisions can occur.

### 1.2 Existing Krewire Config

Krewire's configuration is defined in `krewire.yaml` with typed structs loaded via `libs/config` and validated via `libs/validate`. The `libs/core` package validates `Kind`, `Workload`, and `SpecID`. This spec extends the config model to support multi-runtime awareness without modifying the core config struct.

### 1.3 The `.env` Bridge

Both PHP (Laravel) and Node.js (via `dotenv`) use `.env` files for environment variables. Krewire Go services also read environment variables via `os.Getenv`. The challenge is ensuring that:
- Environment variables don't collide across runtimes
- Each runtime can define its own namespace of variables
- Krewire can access variables from the `.env` file without loading the entire runtime's config

---

## 2. Problem Statement

### 2.1 Configuration Collisions

When Krewire and Node.js/PHP share a project directory, configuration keys can collide:
- Both may define `PORT`, `DATABASE_URL`, `REDIS_HOST`, `CACHE_DRIVER`
- Different runtimes may interpret the same variable differently (e.g., `PORT` for a Node.js server vs a Krewire Go service)

### 2.2 Build Coordination

`npm run build` and `kiw build` may need to run in sequence or in parallel. Without a unified build configuration, developers must manually coordinate which builds run first and which depend on which.

### 2.3 Environment Synchronization

Environment variables defined in `.env` must be accessible to Go processes spawned by Node.js or PHP. The `.env` file is typically loaded by the runtime's loader, but spawned child processes may not inherit the same environment.

---

## 3. Goals

| ID        | Goal                                                                                             | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| G1        | **Runtime namespace isolation.** Each runtime's configuration variables are namespaced to prevent collisions. | Must |
| G2        | **Unified `.env` loading.** Krewire can read `.env` files directly, or load variables passed by the calling runtime. | Must |
| G3        | **Build order configuration.** `krewire.yaml` can declare build dependencies on other runtimes' build steps. | Should |
| G4        | **Port coordination.** `krewire.yaml` and other runtime configs can declare ports without conflict. | Must |
| G5        | **Config validation across runtimes.** Krewire validates that its required environment variables exist in `.env`. | Should |

---

## 4. Non-Goals

| ID        | Non-Goal                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------ |
| NG1       | Replacing `composer.json` or `package.json`. Krewire does not manage PHP or Node.js dependencies. |
| NG2       | Parsing `.env` files in a runtime other than Go. Krewire reads `.env`; it does not replace `dotenv` or Laravel's `.env` loader. |
| NG3       | Synchronizing database schemas across runtimes. That's an application-level concern. |
| NG4       | Managing CI/CD pipeline configuration. Krewire's config is for project runtime, not CI/CD. |

---

## 5. Requirements

### 5.1 Runtime Namespace (Must)

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| FRK-CONF-001 | `krewire.yaml` may define a `runtime.namespace` field (e.g., `go`, `krewire`) that prefixes all Krewire-specific environment variables. | Must |
| FRK-CONF-002 | Environment variables defined in `.env` that are consumed by Krewire must be accessible via `os.Getenv` regardless of namespace. | Must |
| FRK-CONF-003 | Port declarations in `krewire.yaml` must not conflict with ports declared in other runtime configs. `kiw` must warn on conflict. | Must |

### 5.2 Unified `.env` Loading (Must)

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| FRK-CONF-010 | Krewire can load `.env` files directly using a Go `.env` parser (stdlib-compatible). | Must |
| FRK-CONF-011 | When `kiw` spawns a Go binary as a child process, it passes the `.env` variables via `cmd.Env = append(os.Environ(), ...)` so the child inherits them. | Must |
| FRK-CONF-012 | `krewire.yaml` may declare a `runtime.env_file` field pointing to a `.env` file path (default: `.env`). | Should |
| FRK-CONF-013 | If `.env` does not exist but `.env.example` does, `kiw` warns the user to copy and fill in `.env`. | Should |

### 5.3 Build Order Configuration (Should)

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| FRK-CONF-020 | `krewire.yaml` may declare `build.depends_on` listing other runtime build commands that must complete before `kiw build`. | Should |
| FRK-CONF-021 | `kiw build` executes `build.depends_on` commands in order before building the Go project. | Should |
| FRK-CONF-022 | Build dependencies are executed in the project root directory so that `npm run build` and `composer install` resolve correctly. | Should |

### 5.4 Port Coordination (Must)

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| FRK-CONF-030 | `kiw info` reports the port declared in `krewire.yaml` alongside any detected ports from other runtimes. | Must |
| FRK-CONF-031 | `kiw run` accepts `--addr` flag to override the default port, enabling coexistence with other runtimes on the same machine. | Must |
| FRK-CONF-032 | When multiple Krewire services run in the same project, each must declare a unique port. | Must |

### 5.5 Config Validation (Should)

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| FRK-CONF-040 | `kiw vet` checks that all environment variables referenced in `krewire.yaml` exist in `.env` (or have defaults). | Should |
| FRK-CONF-041 | `krewire.yaml` validation (`libs/validate`) checks that `runtime.namespace` and `runtime.env_file` fields are valid if present. | Should |

---

## 6. Non-Functional Requirements

| ID | Requirement | Detail |
| -- | ----------- | ------ |
| NFR1 | **YAML parsing.** `krewire.yaml` parsing uses `gopkg.in/yaml.v3` or stdlib-compatible parser. No custom YAML parser. |
| NFR2 | **`.env` parsing.** `.env` parsing uses a Go stdlib-compatible parser. No third-party `.env` libraries. |
| NFR3 | **No side effects.** Reading `.env` must not modify any files. Writing `.env` is not performed by `kiw`. |
| NFR4 | **Quality gates.** `gofmt -l .`, `go vet ./...`, `go test ./...` must pass for the config module. |
| NFR5 | **Cross-platform.** `.env` and `krewire.yaml` parsing works on Linux, macOS, and Windows. |

---

## 7. Success Criteria

| ID | Criterion | Verification |
| -- | --------- | ------------ |
| S1 | `krewire.yaml` with `runtime.namespace: go` and `runtime.env_file: .env` loads correctly. | `kiw info` reports the config without errors |
| S2 | Variables from `.env` are accessible to a spawned Go binary. | Integration test: spawn Go binary, read env var, validate |
| S3 | `kiw run --addr :9090` runs on port 9090 without conflicting with Node.js `:3000`. | `curl :3000` returns Node.js; `curl :9090` returns Go |
| S4 | `build.depends_on` executes `npm run build` before `kiw build`. | Test: build order verification |
| S5 | `kiw vet` warns when a referenced env var is missing from `.env`. | Test: missing env var detection |

---

## 8. Related Specifications

| SpecID | Title | Relationship |
|--------|-------|-------------|
| [KWN-Q7X4M](./KWN-CONF-Q7X4M-config-directory-and-dotenv.md) | Config Directory & Dotenv | Existing kiw config spec. This spec extends it with multi-runtime awareness. |
| [KWF-COMM-XBRG](https://github.com/krewire/framework/blob/main/docs/specs/KWF-COMM-XBRG-cross-runtime-bridge.md) | Cross-Runtime Bridge Protocol | The bridge protocol uses environment variables defined in this spec's `.env` loading mechanism. |
| [KWN-BUILD-MLTB](./KWN-BUILD-MLTB-multi-runtime-build.md) | Multi-Runtime Build | Build order depends on the `build.depends_on` configuration defined here. |
| [KWN-RUN-6K41E](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-RUN-6K41E-krewire-run-dev-deploy.md) | run/dev/deploy | `kiw run` uses `--addr` for port coordination. |

---

## 9. References

- [Node.js `dotenv`](https://github.com/motdotla/dotenv) — `.env` loading for Node.js
- [Laravel `.env`](https://laravel.com/docs/configuration#configuration-file-format) — PHP `.env` loading
- [Go `os.Getenv`](https://pkg.go.dev/os#Getenv) — Go environment variable access
- [krewire.yaml](https://github.com/krewire/kiw) — Krewire config format

---

## 10. Revision History

Revision history is tracked by **git**, not in-file metadata.

Initial draft: 2026-08-30.
