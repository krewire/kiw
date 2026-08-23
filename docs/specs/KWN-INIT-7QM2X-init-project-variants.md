# Specification — Init Project Variants

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | KWN-7QM2X                                   |
| Title       | Init Project Variants                      |
| Status      | Draft                                       |
| Date        | 2026-08-20                                  |
| Author      | Krewire Contributors                         |
| Domain      | Devtools — Project Init                    |

## 1. Context

After `krewire new` produces a minimal kernel (go.mod, krewire.yaml, .gitignore,
main.go), the developer needs a single command to equip that kernel into a
production-shaped project. `krewire init` extends the project in place by
adding the canonical files for the chosen variant and aligning `krewire.yaml`.

The command set for `new`/`init` is therefore split by **role**: `new`
bootstraps the kernel; `init` equips a variant.

## 2. Problem Statement

Scaffolding full projects from `new` forced every variant into a single
command with a `--type`, and there was no way to equip an existing minimal
project with a static site, a manuscript book, or a remote template without
pasting files by hand. The project config must live exclusively in
`krewire.yaml` — a separate `ssg.yaml` is not accepted.

## 3. Goals

- G1 — Keep `krewire new <project>` minimal and framework-free.
- G2 — Equip a kernel into a **fullstack monolith** when no variant flag is given.
- G3 — Equip a **static site** with `--static`, configured solely via the `ssg:` key in `krewire.yaml`.
- G4 — Equip a **manuscript book** (mdbind) with `--book`.
- G5 — Bootstrap from a remote starter with `--template <git-url>`.
- G6 — Equip a **command-line application** built on framework/tui with `--cli`.
- G7 — Never overwrite pre-existing project files.

## 4. Non-Goals

- NG1 — No `ssg.yaml` file: the declarative SSG config lives only under the `ssg:` key in `krewire.yaml`.
- NG2 — No SSR variant: `init` does not accept an SSR mode.
- NG3 — No interactive wizard: variants are selected strictly by flags.

## 5. Requirements

| ID          | Requirement                                                       | Priority |
| ----------- | ----------------------------------------------------------------- | -------- |
| RND-INIT-001 | `krewire init` operates in the current directory (or an optional positional target). | Must |
| RND-INIT-002 | With no variant flag, equip the kernel into a fullstack monolith: entry `main.go` plus `internal/`, `web/`, `assets/`. | Must |
| RND-INIT-003 | Monolith preserves the module path from the existing `go.mod` and pins framework + libs to the versions the devtool was built with (with local `replace` when run from the krewire repo). | Must |
| RND-INIT-004 | The monolith entry point is the root `main.go` (thin entry calling `config.Load` + `app.New`), matching `krewire run` (`go build .`). | Must |
| RND-INIT-005 | `--static` writes an `ssg:` key into `krewire.yaml` (sample layout, component, page, assets, theme) with `project.kind: site`. No `ssg.yaml` is created. | Must |
| RND-INIT-006 | `--book` writes a `manuscript/` sample plus `project.kind: book` and book fields in `krewire.yaml`. | Must |
| RND-INIT-007 | `--template <git-url>` shallow-clones the URL into the target directory and enacts the cloned files as the project; the target must be absent or empty. | Must |
| RND-INIT-008 | Variant flags are mutually exclusive; more than one is a usage error. | Must |
| RND-INIT-009 | Equipping a site or book variant removes the kernel's placeholder `main.go`; the project kind is pinned via `project.kind` in `krewire.yaml`. | Must |
| RND-INIT-010 | Refuse to overwrite files beyond the kernel set (`go.mod`, `krewire.yaml`, `.gitignore`, `main.go`), which `init` upgrades in place; report the conflict as a usage error. | Must |
| RND-INIT-011 | List created files deterministically on stdout and resolve to exit code 0. | Must |
| RND-INIT-012 | `--cli` equips a command-line application on framework/tui + libs/core: entry `main.go` plus `internal/commands`, pins framework + libs versions, and sets `project.kind: cli`. | Must |
| RND-INIT-013 | `krewire run` and `krewire dev` recognize the `cli` kind and build/run the CLI binary from the root `main.go` without a listen address. | Must |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Idempotence.** Re-running on an already-equipped project fails predictably and writes nothing.
- NFR3 — **Portability.** Linux, macOS, and Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass.

## 7. Success Criteria

- S1 — `krewire new demo && cd demo && krewire init` yields a monolith where `krewire run` builds and serves.
- S2 — `krewire init --static` yields a project where `krewire build` emits a site from the `ssg:` key only.
- S3 — `krewire init --book` yields a project where `krewire build` assembles the manuscript with mdbind.
- S4 — Files beyond the kernel set are never overwritten; conflicts exit with code 2.
- S5 — `krewire init --static --book` exits with a usage error (2).
- S6 — `krewire init --cli` yields a project where `krewire run <args>` and `krewire dev` build and run the CLI binary.

## 8. Related Specifications

| SpecID    | Title                                           |
| --------- | ----------------------------------------------- |
| [KWN-RD3WS](./KWN-SCAFFOLD-RD3WS-project-scaffolding.md) | Project Scaffolding (minimal kernel) |
| [KWN-1QGI2](./KWN-BUILD-1QGI2-project-building.md) | Project Building |
| [KWN-6K41E](./KWN-RUN-6K41E-krewire-run-dev-deploy.md) | Krewire run/dev/deploy |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md) | Krewire Devtool — Initial Specification |

## 9. References

- [KWF-CCI0N](https://github.com/krewire/framework/blob/main/docs/specs/KWF-STRUCT-CCI0N-app-directory-structure.md) — App Project Directory Structure.
- [KWF-PT8OD](https://github.com/krewire/framework/blob/main/docs/specs/KWF-SSG-PT8OD-static-site-generator.md) — Static Site Generator.
- [KWM-FX9H2](https://github.com/krewire/mdbind/blob/main/docs/specs/KWM-BUILDER-FX9H2-mdbind-site-builder.md) — mdbind Site Builder.