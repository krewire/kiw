# Specification — Multi-Runtime Build Orchestration

| Field       | Value                                        |
| ----------- | -------------------------------------------- |
| SpecID      | KWN-MLTB                                     |
| Title       | Multi-Runtime Build Orchestration            |
| Status      | Draft                                        |
| Date        | 2026-08-30                                   |
| Author      | Krewire Contributors                          |
| Domain      | Devtool                                      |
| Scope       | BUILD                                        |

## 1. Context

The Krewire `kiw build` command currently handles two project shapes: SSG sites (`ssg:` key in `krewire.yaml` or `pages/` directory) and mdbind books (`content/` directory). Both produce static websites into `.krewire/build`.

When Krewire coexists with other runtimes — specifically **Node.js** (`package.json`, `npm`/`pnpm`/`yarn`) and **PHP/Laravel** (`composer.json`) — developers need a unified build command that can orchestrate builds across all runtimes in the correct order. A typical multi-runtime project might require:

- `npm run build` for a SvelteKit/Next.js/Nuxt frontend (outputs to `dist/` or `build/`)
- `composer install` for Laravel dependencies
- `kiw build` for Krewire static sites, Go binaries, or WASM modules

Without coordination, developers must manually run these in sequence, remember the correct order, and manage cross-runtime output directories.

## 2. Problem Statement

### 2.1 Build Order Dependencies

A Krewire Go service may depend on a Node.js frontend being built first (e.g., the Go service serves the built static assets). A Laravel app may need its Composer dependencies installed before the Go worker can connect to the same database. Today, `kiw build` only builds the Krewire project shape and has no awareness of other runtime build steps.

### 2.2 No Unified Build Command

Developers in a Laravel + Krewire or Node.js + Krewire project must run:
```bash
composer install       # or npm install
npm run build          # if Node.js frontend
kiw build              # Krewire site/binary
```
There is no single command that understands the full dependency graph.

### 2.3 Output Directory Conflicts

Different runtimes write to different output directories (`dist/`, `build/`, `public/`, `.krewire/build/`, `public/build/`). `kiw build` must know where other runtimes place their output to avoid conflicts and enable proper asset serving.

---

## 3. Goals

| ID  | Goal                                                                                              | Priority |
| --- | ------------------------------------------------------------------------------------------------- | -------- |
| G1  | **Unified build command.** `kiw build` orchestrates builds for all detected runtimes in the project. | Must     |
| G2  | **Declarative build dependencies.** `krewire.yaml` declares `build.depends_on` listing other runtime build commands that must complete before `kiw build`. | Must     |
| G3  | **Independent build coexistence.** `composer install`, `npm install`, and `go build` execute independently without interfering with each other's dependencies or output directories. | Must     |
| G4  | **Runtime detection.** `kiw build` detects Node.js (`package.json`) and PHP (`composer.json`) projects in the repository. | Should   |
| G5  | **Output coordination.** Build artifacts from all runtimes are discoverable for `kiw deploy` and `kiw serve`. | Should   |

---

## 4. Non-Goals

| ID  | Non-Goal                                                                                          |
| --- | ------------------------------------------------------------------------------------------------- |
| NG1 | Replacing `npm`, `pnpm`, `yarn`, or `composer`. Krewire invokes them, does not reimplement them. |
| NG2 | Building Go binaries for non-Krewire projects. Only Krewire Go projects (`krewire.yaml` present). |
| NG3 | Managing CI/CD pipelines. This spec covers local dev and `kiw build` only.                       |
| NG4 | Building frontend frameworks directly. `kiw build` invokes `npm run build` which uses the framework's own builder. |

---

## 5. Requirements

### 5.1 Build Dependencies Configuration (Must)

| ID           | Requirement                                                                                             | Priority |
| ------------ | ------------------------------------------------------------------------------------------------------- | -------- |
| FRK-BLD-001  | `krewire.yaml` may define a `build.depends_on` field as a list of command strings (e.g., `["npm run build", "composer install --no-dev"]`). | Must     |
| FRK-BLD-002  | Each command in `build.depends_on` is executed **in order** before `kiw build` runs the Krewire build. | Must     |
| FRK-BLD-003  | Commands in `build.depends_on` run in the **project root directory** so that `npm`, `pnpm`, `yarn`, `composer` resolve their config files correctly. | Must     |
| FRK-BLD-004  | If any command in `build.depends_on` exits non-zero, `kiw build` aborts and returns the exit code.    | Must     |
| FRK-BLD-005  | `build.depends_on` commands inherit the current environment; `.env` variables are available.           | Should   |

### 5.2 Unified Build Orchestration (Must)

| ID           | Requirement                                                                                             | Priority |
| ------------ | ------------------------------------------------------------------------------------------------------- | -------- |
| FRK-BLD-010  | `kiw build` executes `build.depends_on` commands sequentially, then runs the standard Krewire build (SSG, book, or binary per `KWN-BUILD-1QGI2`). | Must     |
| FRK-BLD-011  | A new `--skip-deps` flag skips `build.depends_on` execution, running only the Krewire build.          | Should   |
| FRK-BLD-012  | A new `--deps-only` flag runs only `build.depends_on` commands, skipping the Krewire build.            | Could    |

### 5.3 Runtime Detection (Should)

| ID           | Requirement                                                                                             | Priority |
| ------------ | ------------------------------------------------------------------------------------------------------- | -------- |
| FRK-BLD-020  | If `package.json` exists and has a `scripts.build` field, `kiw info` reports it and suggests adding to `build.depends_on`. | Should   |
| FRK-BLD-021  | If `composer.json` exists, `kiw info` reports it and suggests `composer install` in `build.depends_on`. | Should   |
| FRK-BLD-022  | Detected runtime build commands are **not auto-executed**; they must be explicitly declared in `build.depends_on`. | Must     |

### 5.4 Output Coordination (Should)

| ID           | Requirement                                                                                             | Priority |
| ------------ | ------------------------------------------------------------------------------------------------------- | -------- |
| FRK-BLD-030  | `krewire.yaml` may declare `build.outputs` mapping runtime names to their output directories (e.g., `nodejs: "dist", laravel: "public/build", krewire: ".krewire/build"`). | Should   |
| FRK-BLD-031  | `kiw build` aggregates all declared outputs into a build manifest for `kiw deploy` and `kiw serve`.    | Should   |
| FRK-BLD-032  | When serving via `kiw serve`, the combined output is served from a unified virtual root.                | Could    |

---

## 6. Non-Functional Requirements

| ID  | Requirement               | Detail                                                                      |
| --- | ------------------------- | --------------------------------------------------------------------------- |
| NFR1 | **Memory safety.**        | The `unsafe` package must not be used.                                      |
| NFR2 | **Portability.**          | Works on Linux, macOS, and Windows.                                         |
| NFR3 | **Quality gates.**        | `gofmt -l .`, `go vet ./...`, `go test ./...` must pass for the build module. |
| NFR4 | **Determinism.**          | Identical input and dependency order must yield identical output.           |
| NFR5 | **Timeout safety.**       | Each `build.depends_on` command has a configurable timeout (default: 5 min). |

---

## 7. Success Criteria

| ID  | Criterion                                                                 | Verification                                                      |
| --- | ------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| S1  | `krewire.yaml` with `build.depends_on: ["npm run build"]` runs `npm run build` before Krewire build. | Integration test: create project with both, verify order.         |
| S2  | `composer install` in `build.depends_on` runs in project root and populates `vendor/`. | Integration test: verify `vendor/` exists after `kiw build`.      |
| S3  | Failed `build.depends_on` command aborts `kiw build` with non-zero exit.  | Test: failing command returns error, Krewire build not executed.  |
| S4  | `kiw build --skip-deps` skips `build.depends_on` and runs only Krewire build. | Test: `--skip-deps` flag behavior.                                |
| S5  | `kiw info` detects `package.json` and `composer.json` and suggests `build.depends_on` entries. | Manual test: `kiw info` output parsing.                           |

---

## 8. Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        kiw build                                │
│                                                                  │
│  1. Read krewire.yaml → parse build.depends_on                  │
│  2. For each command in build.depends_on (in order):            │
│     ├── exec in project root                                    │
│     ├── inherit environment (.env loaded)                       │
│     ├── timeout protection (default 5 min)                      │
│     └── on failure → abort, return exit code                    │
│  3. Run standard Krewire build:                                 │
│     ├── SSG build (if ssg:/pages/) → .krewire/build             │
│     ├── Book build (if content/) → .krewire/build               │
│     ├── Binary build (if service/worker/app) → ./bin/           │
│     └── WASM build (if runtime target=wasm) → .krewire/build    │
│  4. Aggregate outputs per build.outputs → build manifest        │
└─────────────────────────────────────────────────────────────────┘
```

### Example `krewire.yaml` for Laravel + Krewire + Node.js Frontend

```yaml
project:
  kind: site

ssg:
  layouts: layouts
  components: components
  pages: pages

runtime:
  namespace: go
  env_file: .env

build:
  depends_on:
    - "composer install --no-dev"
    - "npm run build"
  outputs:
    krewire: ".krewire/build"
    nodejs: "dist"
    laravel: "public/build"
```

---

## 9. Related Specifications

| SpecID                                    | Title                                          | Relationship                                                    |
| ----------------------------------------- | ---------------------------------------------- | --------------------------------------------------------------- |
| [KWN-BUILD-1QGI2](./KWN-BUILD-1QGI2-project-building.md)     | Project Building                              | Base Krewire build spec. MLTB extends it with multi-runtime orchestration. |
| [KWN-CONF-MRCN](./KWN-CONF-MRCN-multi-runtime-config.md)     | Multi-Runtime Unified Configuration           | Defines `build.depends_on` and `build.outputs` config. MLTB implements the execution. |
| [KWN-RUN-6K41E](./KWN-RUN-6K41E-krewire-run-dev-deploy.md)   | run/dev/deploy                                | `kiw run` and `kiw dev` use the same runtime detection.         |
| [KWN-ARCH-L4RVE](./KWN-ARCH-L4RVE-krewire-laravel-coexistence.md) | Krewire × Laravel Coexistence                 | Build coordination for Laravel + Krewire projects.              |
| [KWN-ARCH-NDJ5S](./KWN-ARCH-NDJ5S-nodejs-ecosystem-integration.md) | Krewire × Node.js Ecosystem Integration       | Build coordination for Node.js + Krewire projects.              |
| [KWF-COMM-XBRG](https://github.com/krewire/framework/blob/main/docs/specs/KWF-COMM-XBRG-cross-runtime-bridge.md) | Cross-Runtime Bridge Protocol                 | Build artifacts serve as communication bridge (static assets).  |

---

## 10. References

- [Node.js `npm run build`](https://docs.npmjs.com/cli/v10/commands/npm-run-script) — standard build script
- [Composer `install`](https://getcomposer.org/doc/03-cli.md#install) — PHP dependency installation
- [Go `os/exec`](https://pkg.go.dev/os/exec) — command execution
- [Krewire Project Building (`KWN-BUILD-1QGI2`)](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-BUILD-1QGI2-project-building.md) — base build spec

---

## 11. Revision History

Revision history is tracked by **git**, not in-file metadata. To view changes:

```bash
git log --oneline -- docs/specs/KWN-BUILD-MLTB-multi-runtime-build.md
```

Initial draft: 2026-08-30.