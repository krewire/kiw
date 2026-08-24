# Specification — Version Reporting

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | KWN-JB7PW                                   |
| Title       | Version Reporting                          |
| Status      | Draft                                       |
| Date        | 2026-08-18                                  |
| Author      | Krewire Contributors                         |
| Domain      | Devtools — Reporting                       |

## 1. Context

`krewire version` reports what the tool is and what it was built against: the
krewire CLI version, the Krewire Framework version, and the Krewire Libraries
version, read from build metadata.

## 2. Problem Statement

Users and CI pipelines cannot tell which stack version a tool was built with,
leading to misdiagnosis and mismatched troubleshooting. Version output is often
inconsistent — missing prefixes, invented numbers, or none at all. The devtool
provides one deterministic, honest version statement.

## 3. Goals

- G1 — Report the krewire CLI version.
- G2 — Report the exact framework and libraries versions from build metadata.
- G3 — Normalize version formatting (single leading `v`, `dev` fallback).
- G4 — Use stable display names ("Krewire", "Krewire Framework") in output.
- G5 — In workspace builds (`go.work`), where build metadata records
  `(devel)`, fall back to each module's declared version constant so the
  report stays meaningful for developers and scaffolded projects.
- G6 — Mark workspace-resolved versions as non-release (`(dev)`) so operators
  can weigh production upgrade decisions against real provenance.

## 4. Non-Goals

- NG1 — Online version checks or update suggestions.
- NG2 — Listing every transitive dependency's version.
- NG3 — Output formats beyond the initial human-readable line set.

## 5. Requirements

| ID         | Requirement                                                       | Priority |
| ---------- | ----------------------------------------------------------------- | -------- |
| RND-VS-001 | Print the CLI version as `Krewire v<version>`.                     | Must     |
| RND-VS-002 | Print the framework version from build metadata as `Krewire Framework v<version>`. | Must |
| RND-VS-003 | Print the libraries version from build metadata.                  | Must     |
| RND-VS-004 | Collapse unknown module versions to `dev` rather than guessing.   | Must     |
| RND-VS-005 | Emit exactly one leading `v` on every printed version.            | Must     |
| RND-VS-006 | Write the report to stdout only; keep diagnostics off it.         | Must     |
| RND-VS-007 | Exit 0 on success, with no further side effects.                  | Must     |
| RND-VS-008 | When the metadata version is `(devel)` or empty, resolve the version from the module's declared constant (`framework.Version`, `libs/core.CurrentVersion`); never print `(devel)`. | Must |
| RND-VS-009 | Qualify source-resolved versions with a `(dev)` marker (e.g. `v0.5.1 (dev)`); released tags print unqualified. | Must |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Determinism.** Identical builds produce identical output.
- NFR3 — **Portability.** Linux, macOS, and Windows.
- NFR4 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass.

## 7. Success Criteria

- S1 — `krewire version` output matches `krewire info`'s framework/libs lines.
- S2 — Versions are never prefixed twice and never empty.
- S3 — The command performs no network access and alters nothing.

## 8. Related Specifications

| SpecID    | Title                                           |
| --------- | ----------------------------------------------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md)      | Krewire Devtool — Initial Specification |
| [KWN-BNKJC](./KWN-INFO-BNKJC-project-information.md) | Project Information                |
| [KWF-NPFSE](https://github.com/krewire/framework/blob/main/docs/specs/KWF-CLI-NPFSE-cli-output-formatting.md) | CLI Output & Formatting |
| [KWL-W0J2X](https://github.com/krewire/libs/blob/main/docs/specs/KWL-CORE-W0J2X-errors-exit-codes.md) | Core Errors & Exit Codes |

## 9. References

- [KWF-CMBZJ](https://github.com/krewire/framework/blob/main/docs/specs/KWF-META-CMBZJ-krewire-meta-framework.md) — Krewire Framework initial specification.
- [kiw](https://github.com/krewire/kiw) — the reporting devtool.