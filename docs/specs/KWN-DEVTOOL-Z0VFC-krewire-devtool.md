# Specification — Krewire Devtool

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | KWN-Z0VFC                                   |
| Title       | Krewire Devtool — Initial Specification      |
| Status      | Draft                                       |
| Date        | 2026-08-18                                  |
| Author      | Krewire Contributors                         |
| Domain      | Devtools                                    |

## 1. Context

**Krewire** is the command-line devtool for the Krewire ecosystem. It composes
the entire ecosystem — the Krewire Framework and Krewire Libraries — from a
single CLI, and operates on any project built on top of it.

The devtool dogfoods the framework's `tui` package: the tool that manages the
ecosystem is itself built on the stack it manages, exercising the same
conventions it expects other Krewire projects to follow.

### 1.1 Current State

- Module root `github.com/krewire/kiw`; binary name `kiw`.
- Built on `github.com/krewire/framework` (dependency pinned from build metadata).
- Initial command set: `version`, `info`, `new`, `build`, `test`.
- Shared module root: Go 1.22, MIT license.

## 2. Problem Statement

Composing the Krewire ecosystem today means orchestrating several tools by hand:
building a project, inspecting its environment, running tests, and bootstrapping
new work each require bespoke commands and ad-hoc conventions. Every new project
re-derives boilerplate and version pins, so projects diverge in shape and
dependency versions. Developers lack a single entry point that speaks the
ecosystem's own language.

The devtool removes that friction: one CLI, composed from the framework it
manages, with canonical exit codes, structured diagnostics, and a stable,
incrementally growing command surface.

## 3. Goals

- G1 — Provide a single CLI entry point for the Krewire ecosystem lifecycle.
- G2 — Dogfood the framework's `tui` package for every command.
- G3 — Ship an initial command set now and grow features incrementally.
- G4 — Follow ecosystem conventions: exit codes, output separation, configuration.
- G5 — Use stable display names ("Krewire", "Krewire Framework") regardless of repository names.

## 4. Non-Goals

The following are explicitly out of scope for the initial phase:

- NG1 — Package publishing or release automation for ecosystem modules.
- NG2 — A full CI/CD orchestrator or remote status dashboard.
- NG3 — Re-implementing framework or library functionality (help, errors, terminal I/O).
- NG4 — Backwards compatibility guarantees before version `1.0.0`.

## 5. Scope — Initial Command Set

The initial command set covers the essentials: scaffolding, inspection, and
validation of Krewire projects.

### 5.1 Tool Capabilities

| ID          | Requirement                                                            | Priority |
| ----------- | ----------------------------------------------------------------------- | -------- |
| RND-MT-001  | Provide a `kiw` CLI whose commands are built on the framework `tui` package. | Must    |
| RND-MT-002  | Support `new` to scaffold a new Krewire project.                        | Must     |
| RND-MT-003  | Support `info` to report environment and project information.          | Must     |
| RND-MT-004  | Support `test` to validate the current project.                        | Must     |
| RND-MT-005  | Support `version` to report the CLI, framework, and libraries versions. | Must    |
| RND-MT-006  | All commands return canonical exit codes (0/1/2) from the ecosystem model. | Must |
| RND-MT-007  | Support `build` to build the current project's website (see KWN-1QGI2). | Must    |

### 5.2 Deliverables

- D1 — `cmd/kiw` built on `github.com/krewire/framework/tui` (RND-MT-001).
- D2 — Commands `new`, `info`, `build`, `test`, `version` (RND-MT-002..005, RND-MT-007).
- D3 — CI workflow enforcing `gofmt`, `go vet`, and `go test ./...`.
- D4 — Unit tests covering scaffolding and project detection.

## 6. Architecture Constraints

- C1 — The devtool is a Go module; `cmd/kiw` is the only entry point.
- C2 — Commands may depend on `framework` and `libs` packages, never on framework internals.
- C3 — Display names are stable: the tool is branded "Krewire", the framework "Krewire Framework", regardless of repository names.
- C4 — The `unsafe` package must not be used; exported identifiers must be documented.
- C5 — New capabilities are additive commands; breaking changes are not allowed in minor/patch releases.

## 7. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Performance.** Command startup and dispatch overhead must remain minimal.
- NFR3 — **Portability.** Linux, macOS, and Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass in CI.
- NFR5 — **Determinism.** Identical input must yield identical output.

## 8. Success Criteria

- S1 — Every ecosystem lifecycle task (scaffold, inspect, validate, report) runs through `kiw`.
- S2 — All commands resolve to canonical exit codes, never ad-hoc integers.
- S3 — A scaffolded project builds and runs without edits.
- S4 — The devtool builds only against the framework and libraries it composes.

## 9. Future Phases (Expansion)

The command surface grows incrementally, phase by phase:

| Phase | Focus                 | Planned Commands        |
| ----- | --------------------- | ----------------------- |
| P1    | Core lifecycle        | `new`, `info`, `test`, `version` |
| P2    | Build & run workflow  | `build`, `run`, `lint`  |
| P3    | Publish & docs        | `publish`, `doc`        |
| P4    | Inspect & repair      | `doctor`, `diagnose`    |
| P5    | Ecosystem-wide        | `completions`, `upgrade` |

Each phase is additive and must not break the previous phases.

## 10. Related Specifications

| SpecID    | Title                                           |
| --------- | ----------------------------------------------- |
| [KWN-RD3WS](./KWN-SCAFFOLD-RD3WS-project-scaffolding.md)  | Project Scaffolding                |
| [KWN-BNKJC](./KWN-INFO-BNKJC-project-information.md)  | Project Information                |
| [KWN-P0FWA](./KWN-TEST-P0FWA-project-validation.md)   | Project Validation                 |
| [KWN-JB7PW](./KWN-INFO-JB7PW-version-reporting.md)    | Version Reporting                  |

## 11. References

- [Framework — KWF-CMBZJ](https://github.com/krewire/framework/blob/main/docs/specs/KWF-META-CMBZJ-krewire-meta-framework.md) — Krewire Framework initial specification.
- [Libraries — KWL-M1ZKS](https://github.com/krewire/libs/blob/main/docs/specs/KWL-CORE-M1ZKS-krewire-libraries.md) — Krewire Libraries initial specification.
- [KWL-W0J2X](https://github.com/krewire/libs/blob/main/docs/specs/KWL-CORE-W0J2X-errors-exit-codes.md) — Core Errors & Exit Codes.
- Project `README.md` for building and testing instructions.