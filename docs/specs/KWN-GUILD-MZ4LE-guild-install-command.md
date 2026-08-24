# Specification — krewire guild install Command

| Field       | Value                                      |
| ----------- | ------------------------------------------ |
| SpecID      | KWN-MZ4LE                                  |
| Title       | krewire guild install Command               |
| Status      | Draft                                      |
| Date        | 2026-08-20                                 |
| Author      | Krewire Contributors                        |
| Domain      | Devtools — Guild Installer                  |

## 1. Context

Krewire Guild is the ecosystem's reusable AI agent setup, distributed as the Go
module `github.com/krewire/guild`. The krewire CLI is the single entry point for
the whole ecosystem. This spec adds `krewire guild install` — an interactive
CLI wizard that installs the Guild template into a project, replacing the
retired `scripts/install.sh` from the Guild repository.

## 2. Problem Statement

Until now the Guild template was installed through a checkout-local shell
script (`scripts/install.sh`) with positional flags, no validation, and no
connection to the ecosystem CLI. Users of the Krewire ecosystem expect every
operation — scaffolding, building, serving, testing — from `krewire`. Install
should be no different, and it must be interactive: ask where to install and
confirm before overwriting existing managed files.

## 3. Goals

- G1 — Add a `guild` command group to the krewire CLI exposing `guild install`.
- G2 — Interactively prompt for the target directory when not supplied.
- G3 — Detect existing managed files and prompt for confirmation before
      overwriting (or honor explicit `--force`).
- G4 — Support `--dry-run` to preview without writing.
- G5 — Consume the template purely from `github.com/krewire/guild`, with no copy
      of the template inside the krewire repository.

## 4. Non-Goals

- NG1 — Installing Git hooks, cloud deploy configs, or vendor-specific setup.
- NG2 — Scaffolding an entire project (`krewire new` covers that).
- NG3 — Templating variables inside the installed files.

## 5. Requirements

| ID         | Requirement                                                            | Priority |
| ---------- | ---------------------------------------------------------------------- | -------- |
| KWN-GLD-001 | Register `guild` as a CLI command group whose sub-command is `install`.  | Must   |
| KWN-GLD-002 | `krewire guild install` with no directory arg prompts for one interactively. | Must |
| KWN-GLD-003 | Accept the target directory as a positional argument; reject missing/empty targets with exit code 2. | Must |
| KWN-GLD-004 | Support `--force` to overwrite existing managed files without prompting.  | Must |
| KWN-GLD-005 | Support `--dry-run` to report files without writing.                     | Must |
| KWN-GLD-006 | Confirm overwrite interactively when managed files exist and no `--force` is given. | Must |
| KWN-GLD-007 | Print the created files deterministically and finish with next-steps guidance. | Should |
| KWN-GLD-008 | Delegate installation to `github.com/krewire/guild` `Install` API; never copy template bytes inline. | Must |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Determinism.** Same input yields the same reported file list.
- NFR3 — **Portability.** Linux, macOS, and Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, `go test ./...` must pass.
- NFR5 — **Idempotence.** Re-running on an already-installed target fails predictably and writes nothing (unless forced).

## 7. Success Criteria

- S1 — `krewire guild install` walks a user through target selection and overwrite confirmation interactively.
- S2 — `krewire guild install ./fresh` installs into `./fresh` without prompting.
- S3 — `krewire guild install --dry-run ./fresh` reports files without writing.
- S4 — `guild install` on a target containing a managed file without `--force` stops with exit code 2... after a confirmation prompt that can decline.

## 8. Related Specifications

| SpecID    | Title                                                |
| --------- | ---------------------------------------------------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md) | Krewire Devtool — Initial Specification |
| [KWG-P9ZT4](https://github.com/krewire/guild/blob/main/docs/specs/KWG-INSTALL-P9ZT4-guild-module-install.md) | Guild Module & Install Library |
| [KWG-K2N7Q](https://github.com/krewire/guild/blob/main/docs/specs/KWG-ECO-K2N7Q-krewire-native-guild-template.md) | Krewire-Native Guild Template |

## 9. References

- [guild](https://github.com/krewire/guild) — template + install library.
- [kiw](https://github.com/krewire/kiw) — the CLI.