# Specification — Project Scaffolding

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | KWN-RD3WS                                   |
| Title       | Project Scaffolding                        |
| Status      | Draft                                       |
| Date        | 2026-08-18                                  |
| Author      | Krewire Contributors                         |
| Domain      | Devtools — Scaffolding                     |

## 1. Context

`krewire new <project>` bootstraps a new Krewire project as a **minimal kernel**:
`go.mod`, `krewire.yaml`, `.gitignore`, and a framework-free `main.go`. The
kernel is framework-independent and builds with the standard library alone.
Production-shaped projects are equipped in place with `krewire init`
(KWN-7QM2X), which selects the variant: monolith, static site, book, or a
remote template.

## 2. Problem Statement

Starting a new Krewire project means creating the four kernel files by hand and
re-deriving conventions for the module name, project config file, and ignore
rules. The kernel must stay tiny and dependency-free so the bootstrap is fast
and the variant choice is deferred to `init`.

## 3. Goals

- G1 — Scaffold a minimal, buildable Krewire kernel in one command.
- G2 — Produce deterministic, idiomatic, minimal output.
- G3 — Keep the kernel free of ecosystem dependencies.
- G4 — Refuse destructive operations (never overwrite existing work).
- G5 — Defer variant shaping to `krewire init`.

## 4. Non-Goals

- NG1 — Interactive project wizards beyond minimal flags.
- NG2 — Full application templates: variants are equipped by `krewire init`, not by `new`.
- NG3 — Creating or configuring remote repositories (GitHub, CI).
- NG4 — Pinning ecosystem versions: the kernel declares no ecosystem dependencies.

## 5. Requirements

| ID         | Requirement                                                       | Priority |
| ---------- | ----------------------------------------------------------------- | -------- |
| RND-SC-001 | Read the project name from the first positional argument.         | Must     |
| RND-SC-002 | Default the module path to the project name; support `--module` override. | Must |
| RND-SC-003 | Support `--dir` to choose the parent directory (default: current). | Must   |
| RND-SC-004 | Reject names with path separators or invalid characters as usage errors. | Must |
| RND-SC-005 | Refuse to scaffold into a non-empty target directory.             | Must     |
| RND-SC-006 | Emit `go.mod` with the module path and a `go` directive only — no ecosystem requires. | Must |
| RND-SC-007 | Emit a `main.go` that compiles with the standard library only.    | Must     |
| RND-SC-008 | Emit `krewire.yaml` and `.gitignore` alongside `go.mod` and `main.go`. | Must |
| RND-SC-009 | List the created files deterministically on stdout.               | Must     |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Idempotence.** Re-running on the same target fails predictably and writes nothing.
- NFR3 — **Portability.** Linux, macOS, and Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass.

## 7. Success Criteria

- S1 — `krewire new demo` produces a project that builds and runs (`go run .`), with no external dependencies.
- S2 — Scaffolding never corrupts or overwrites pre-existing files.
- S3 — `krewire init` can equip the resulting kernel into a monolith, static site, or book without rewriting the kernel files by hand.
- S4 — Invalid names and non-empty directories resolve to exit code 2.

## 8. Related Specifications

| SpecID    | Title                                           |
| --------- | ----------------------------------------------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md)      | Krewire Devtool — Initial Specification |
| [KWN-7QM2X](./KWN-INIT-7QM2X-init-project-variants.md) | Init Project Variants |
| [KWF-PZ5JU](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-PZ5JU-cli-scaffolding.md) | CLI Scaffolding            |
| [KWF-CCI0N](https://github.com/krewire/framework/blob/main/docs/specs/KWF-STRUCT-CCI0N-app-directory-structure.md) | App Project Directory Structure Standard |
| [KWF-5XJFC](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-5XJFC-cli-application-model.md) | CLI Application Model |
| [KWL-W0J2X](https://github.com/krewire/libs/blob/main/docs/specs/KWL-CORE-W0J2X-errors-exit-codes.md) | Core Errors & Exit Codes |

## 9. References

- [KWF-FGNZ9](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-FGNZ9-cli-configuration.md) — CLI Configuration.
- [KWF-MFA0T](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-MFA0T-cli-help-usage.md) — CLI Help & Usage.
- [kiw](https://github.com/krewire/kiw) — the scaffolding devtool.