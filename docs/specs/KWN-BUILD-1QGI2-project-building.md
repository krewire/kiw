# Specification — Project Building

| Field       | Value                                       |
| ----------- | ------------------------------------------- |
| SpecID      | KWN-1QGI2                                   |
| Title       | Project Building                           |
| Status      | Draft                                       |
| Date        | 2026-08-18                                  |
| Author      | Krewire Contributors                         |
| Domain      | Devtool                                     |

## 1. Context

The `krewire build` command turns the current Krewire project into a static
website. It recognizes two declarative project shapes: an SSG section in
`krewire.yaml` (under the `ssg:` key) built by the web/ssg static site
generator, or a `manuscript/` directory rendered by the mdbind site builder.
Projects no longer ship their own site generators — all build logic lives in
`krewire`. All configuration lives in a single `krewire.yaml` file.

## 2. Problem Statement

Every Krewire project with a website re-invents how to produce its output: some
call mdbind's CLI, some run their own `go run ./cmd/...`, and the invocation
varies from project to project. There is no single, discoverable way to build
a Krewire project's site, and no shared convention for where the manuscript or
the declarative config lives.

## 3. Goals

- G1 — One command (`krewire build`) builds the current project's website.
- G2 — Support the two standard project shapes: `ssg.yaml` and `manuscript/`.
- G3 — Resolve paths against the project root, not the working directory.
- G4 — Delegate to the framework's ssg package or mdbind instead of reimplementing them.

## 4. Non-Goals

- NG1 — Building Go binaries (`go build ./...`); `build` targets websites only.
- NG2 — Deployment, preview servers, or asset pipelines beyond what the builder produces.
- NG3 — Discovering exotic project layouts beyond the two documented shapes.

## 5. Requirements

| ID          | Requirement                                                            | Priority |
| ----------- | ----------------------------------------------------------------------- | -------- |
| RND-BLD-001 | Provide `krewire build`, operating at the nearest Go module root.       | Must     |
| RND-BLD-002 | When `krewire.yaml` contains an `ssg:` key (non-null), build it with the web/ssg generator (`ssg.BuildFromConfig`). SSG config fields (layouts, components, pages, assets, description) live under `ssg:`. | Must |
| RND-BLD-003 | When the project contains a `manuscript/` directory (and `krewire.yaml` has no `ssg:` key), build it with the mdbind site builder (`book.Build`). | Must |
| RND-BLD-013 | `krewire.yaml` is the single config file for all project shapes. Top-level fields (`title`, `author`, `base`, `output`, `theme`, `nav`, `footer`) are shared. SSG-specific configuration lives under the `ssg:` key. | Must |
| RND-BLD-004 | Default the output to `site`, resolved relative to the project root.   | Must |
| RND-BLD-005 | Support `--output`, `--base`, `--title`, `--author`, and `--theme` flags. | Must |
| RND-BLD-006 | Default the site title to the project name (last segment of the module path). | Must |
| RND-BLD-007 | Print each created path on success, like `mdbind build`.               | Must |
| RND-BLD-008 | Exit with a usage error when neither project shape is present.          | Must |
| RND-BLD-009 | Read manuscript settings from `krewire.yaml` in the project root: title, author, base, input, output, nav, footer, and theme palette overrides. | Must |
| RND-BLD-010 | CLI flags override krewire.yaml settings; unset settings fall back to defaults. | Must |
| RND-BLD-011 | Provide `krewire serve` to preview the current project's website over HTTP (both ssg and manuscript shapes). | Must |
| RND-BLD-012 | Provide `krewire init` to scaffold a sample manuscript + `krewire.yaml` into the project root. | Must |

## 6. Non-Functional Requirements

- NFR1 — **Memory safety.** The `unsafe` package must not be used.
- NFR2 — **Portability.** Linux, macOS, and Windows.
- NFR3 — **Quality gates.** `gofmt`, `go vet ./...`, and `go test ./...` must pass.
- NFR4 — **Determinism.** Identical input must yield identical output.

## 7. Success Criteria

- S1 — In a project with a `manuscript/`, `krewire build` produces `site/` with mdbind output.
- S2 — In a project with only `cmd/site`, `krewire build` runs it with the output directory.
- S3 — Flags override defaults and relative paths resolve against the project root.
- S4 — A `krewire.yaml` configures title, nav, footer, base, and palette without a project-specific cmd.
- S5 — Running `krewire build` outside a Krewire project fails with a clear usage message.

## 8. Related Specifications

| SpecID    | Title                                          |
| --------- | ---------------------------------------------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md) | Krewire Devtool — Initial Specification |
| [KWM-FX9H2](https://github.com/krewire/mdbind/blob/main/docs/specs/KWM-BUILDER-FX9H2-mdbind-site-builder.md) | mdbind Site Builder |

## 9. References

- [KWF-CMBZJ](https://github.com/krewire/framework/blob/main/docs/specs/KWF-META-CMBZJ-krewire-meta-framework.md) — Krewire Framework initial specification.
