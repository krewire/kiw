# Specification — Workspace Command

| Field   | Value                                          |
| ------- | ---------------------------------------------- |
| SpecID  | KWN-Q3M8V                                       |
| Title   | Workspace Command — Multi-Repo & Monorepo Workflows |
| Status  | Draft                                           |
| Date    | 2026-08-24                                      |
| Author  | Krewire Contributors                            |
| Domain  | Devtool — Workspace Command                     |

## Context

Krewire projects range from single-module monorepos to hub workspaces that
bind five or more repositories with `go.work`, and forward to microservice
layouts where each member is its own deployable kind. The devtool already
owns every other workflow around these layouts (`kiw build`, `kiw run`,
`kiw test`), but inspecting and mutating a workspace still requires raw
`go work …` invocations, manual member bookkeeping, and per-member command
loops.

The `kiw ws` sub-command shipped to close that gap before it had a
specification. This document retroactively records the requirement set the
implementation must satisfy and be tested against (spec-first debt from the
2026-08-24 documentation sync).

## Problem Statement

Developers working across multi-repo or microservice workspaces repeat three
painful motions daily:

1. **Discovery** — which members exist, what kind is each, what module path?
   Today this means reading `go.work` by hand and running detection commands
   one directory at a time.
2. **Mutation** — adding/removing a module means remembering exact
   `go work use` / `go work edit -dropuse` syntax and running it from the
   correct root.
3. **Fan-out** — running a command (tests, vet) across every member means a
   hand-written shell loop whose failure aggregation is easy to get wrong.

Leaving this unsolved costs every workspace developer minutes per hour and
makes `kiw` incomplete as the single entry point promised by the unified
vision (`KWF-M8K2Q`): the ecosystem's most layout-heavy users fall back to
raw toolchain commands.

## Goals

- G1 — One command surface for workspace discovery, mutation and fan-out.
- G2 — Layout classification that matches how Krewire hubs actually look.
- G3 — Exit codes consistent with the ecosystem contract
  (`core.ExitCodeSuccess/Failure/Usage`).

## Requirements

| ID         | Requirement                                                                                                   | Priority | Scope |
| ---------- | ------------------------------------------------------------------------------------------------------------- | -------- | ----- |
| WS-CMD-001 | `kiw ws <sub>` dispatches to `info`, `list`/`ls`, `add`, `remove`/`rm`, `sync`, `exec`; empty arg or `help`/`-h`/`--help` prints usage with exit code Usage; an unknown sub-command reports it on stderr then prints usage with exit code Usage. | Must | Unit |
| WS-CMD-002 | Workspace detection walks upward from the current directory: a directory containing `go.work` classifies as multi-repo with members parsed from both block-form and single-line `use` directives; a bare `go.mod` root classifies as monorepo with itself as sole member; neither yields unknown. | Must | Unit |
| WS-CMD-003 | `ws info` prints the workspace root path, its classification, every member, and the `go.work` location when present. | Must | Unit |
| WS-CMD-004 | `ws list` prints one PROJECT/KIND/MODULE row per member, resolving module paths via the go.mod reader and kinds via project-shape detection. | Must | Unit |
| WS-CMD-005 | `ws add <path>` and `ws remove <path>` mutate `go.work` through `go work use` / `go work edit -dropuse` executed at the workspace root; a missing argument or a non-`go.work` layout exits Usage without invoking the toolchain. | Must | Unit |
| WS-CMD-006 | `ws sync` runs `go work sync` at the workspace root.                                                            | Must | Unit |
| WS-CMD-007 | `ws exec [--] <cmd> [args…]` runs the command inside every member that has a `go.mod`, streaming output live, collecting failed members, and exiting Failure when at least one member failed; a missing command exits Usage. | Must | Unit |
| WS-CMD-008 | The help text documents all sub-commands, runnable examples, and the four workspace types (monorepo, multi-repo, multi-project, microservices). | Should | Unit |

## Non-Goals

- NG1 — Editing `go.work.repl`/replace directives beyond add/remove/sync.
- NG2 — Remote workspace definitions; the filesystem remains the source of truth.
- NG3 — Parallel fan-out execution in v1.

## Non-Functional Requirements

- NFR1 — Sub-command handling reuses the shared flag-set dispatch so `kiw ws help` behaves like every other command group.
- NFR2 — No output buffering that would break live log streaming during `exec`.

## Related Specifications

| SpecID    | Title                                    |
| --------- | ---------------------------------------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md) | Devtool CLI foundation |
| [KWN-Q7X4M](./KWN-CONF-Q7X4M-config-directory-and-dotenv.md) | Config directory & dotenv |

## References

- Go workspaces: https://go.dev/ref/mod#workspaces
