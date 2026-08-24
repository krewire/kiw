# Specification — Project Validation

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | KWN-P0FWA                                   |
| Title       | Project Validation                         |
| Status      | Draft                                       |
| Date        | 2026-08-18                                  |
| Author      | Krewire Contributors                         |
| Domain      | Devtools — Validation                      |

## 1. Context

`krewire test` validates the current project by running its test suite. It
locates the enclosing Go module, streams the suite's output through, and
resolves the result to a canonical exit code.

## 2. Problem Statement

Running a project's tests requires remembering exact tool invocations and
interpreting raw tool output. There is no single, ecosystem-consistent way to
say "validate this project". The devtool normalizes that experience: one
command, one exit-code contract, structured start diagnostics.

## 3. Goals

- G1 — Validate the current project with a single command.
- G2 — Stream the test suite output to the user unchanged.
- G3 — Map the outcome to canonical exit codes (0 success, 1 failure).
- G4 — Report clearly when not inside a Go module.

## 4. Non-Goals

- NG1 — Coverage reporting, benchmarks, or code-quality analysis in this phase.
- NG2 — Remote or CI orchestration from the devtool.
- NG3 — Multi-module repository-wide validation in this phase.

## 5. Requirements

| ID         | Requirement                                                       | Priority |
| ---------- | ----------------------------------------------------------------- | -------- |
| RND-TS-001 | Resolve the module from the nearest `go.mod` walking upward.      | Must     |
| RND-TS-002 | Outside a Go module, print guidance and return the usage exit code (2). | Must |
| RND-TS-003 | Run the module's tests via `go test ./...`.                       | Must     |
| RND-TS-004 | Stream the command's stdout and stderr through to the user.       | Must     |
| RND-TS-005 | Resolve a failed suite to the failure exit code (1).              | Must     |
| RND-TS-006 | Resolve a passed suite to the success exit code (0).              | Must     |
| RND-TS-007 | Log the start of validation as a structured record.               | Must     |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Fidelity.** No output is re-encoded or buffered heavily.
- NFR3 — **Portability.** Linux, macOS, and Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass.

## 7. Success Criteria

- S1 — A passing scaffolded project exits 0; a deliberately failing test exits 1.
- S2 — Outside a module, `krewire test` exits 2 without invoking anything.
- S3 — Test output is identical to running `go test ./...` directly.

## 8. Related Specifications

| SpecID    | Title                                           |
| --------- | ----------------------------------------------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md)      | Krewire Devtool — Initial Specification |
| [KWN-RD3WS](./KWN-SCAFFOLD-RD3WS-project-scaffolding.md) | Project Scaffolding                |
| [KWF-KAKQL](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-KAKQL-cli-errors-diagnostics.md) | CLI Errors & Diagnostics |
| [KWL-W0J2X](https://github.com/krewire/libs/blob/main/docs/specs/KWL-CORE-W0J2X-errors-exit-codes.md) | Core Errors & Exit Codes |

## 9. References

- [KWF-5XJFC](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-5XJFC-cli-application-model.md) — CLI Application Model.
- [kiw](https://github.com/krewire/kiw) — the validating devtool.