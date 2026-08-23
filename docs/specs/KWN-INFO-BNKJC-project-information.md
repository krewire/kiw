# Specification — Project Information

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | KWN-BNKJC                                   |
| Title       | Project Information                        |
| Status      | Draft                                       |
| Date        | 2026-08-18                                  |
| Author      | Krewire Contributors                         |
| Domain      | Devtools — Inspection                      |

## 1. Context

`krewire info` prints environment and project information in the spirit of
`artisan info`: a deterministic snapshot of the tooling a project sits on —
the krewire CLI version, the Krewire Framework and Krewire Libraries versions, the
Go toolchain — alongside the project's own identity.

## 2. Problem Statement

Developers repeatedly lose sight of which stack — Go version, framework
version, libraries version, module path — a given project depends on. Tools
either omit this context or bury it in verbose telemetry. Teams need one
command that states, plainly and deterministically, what a project is and what
it runs on.

## 3. Goals

- G1 — Print environment facts: CLI, framework, libraries, Go, GOOS/GOARCH.
- G2 — Print project facts: directory, module path, whether it is built on Krewire.
- G3 — Keep output deterministic and readable by humans and shells.
- G4 — Work both inside and outside a Go module.

## 4. Non-Goals

- NG1 — Continuous monitoring or live metrics.
- NG2 — Remote/project telemetry or analytics.
- NG3 — Re-printing framework feature catalogs or specification counts.

## 5. Requirements

### 5.1 Environment Reporting

| ID         | Requirement                                                       | Priority |
| ---------- | ----------------------------------------------------------------- | -------- |
| RND-IF-001 | Report the krewire CLI version.                                    | Must     |
| RND-IF-002 | Report the Krewire Framework version from build metadata.          | Must     |
| RND-IF-003 | Report the Krewire Libraries version from build metadata.          | Must     |
| RND-IF-004 | Report the Go runtime version, GOOS, and GOARCH.                  | Must     |

### 5.2 Project Reporting

| ID         | Requirement                                                       | Priority |
| ---------- | ----------------------------------------------------------------- | -------- |
| RND-IF-005 | Report the current working directory.                             | Must     |
| RND-IF-006 | Resolve the module path from the nearest `go.mod` walking upward. | Must     |
| RND-IF-007 | Report whether the project is built on Krewire.                    | Must     |
| RND-IF-008 | Outside a Go module, still report environment facts (no failure). | Must     |
| RND-IF-009 | Order output deterministically; never append unstable counters.   | Must     |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Determinism.** Identical environments produce byte-identical output.
- NFR3 — **Portability.** Linux, macOS, and Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass.

## 7. Success Criteria

- S1 — `krewire info` inside a Krewire project reports "Built on Krewire: yes" and the correct module path.
- S2 — `krewire info` outside a module exits 0 and prints environment facts only.
- S3 — The command emits no specification counts or other extraneous noise.
- S4 — Output parsing by simple line-based tools is reliable.

## 8. Related Specifications

| SpecID    | Title                                           |
| --------- | ----------------------------------------------- |
| [KWN-JB7PW](./KWN-INFO-JB7PW-version-reporting.md)   | Version Reporting                  |
| [KWF-NPFSE](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-NPFSE-cli-output-formatting.md) | CLI Output & Formatting |
| [KWF-5XJFC](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-5XJFC-cli-application-model.md) | CLI Application Model  |
| [KWL-W0J2X](https://github.com/krewire/libs/blob/main/docs/specs/KWL-CORE-W0J2X-errors-exit-codes.md) | Core Errors & Exit Codes |

## 9. References

- [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md) — Krewire Devtool initial specification.
- [KWF-CMBZJ](https://github.com/krewire/framework/blob/main/docs/specs/KWF-META-CMBZJ-krewire-meta-framework.md) — Krewire Framework initial specification.