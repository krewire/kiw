# Specification — kiw add/remove package@version (scalable)

| Field  | Value |
| ------ | ----- |
| SpecID | KWN-PKG-7X9K2 |
| Title  | kiw add/remove — scalable package@version (plugin → gomod → npm) |
| Status | Draft |
| Date   | 2026-08-27 |
| Author | Krewire Contributors |
| Domain | Devtool — Package Management |

## 1. Context

Framework has one plugin (`twcss`/`tailwind`) and will grow. Each plugin previously required manual `npm install` + `tailwind.config.js` + `assets/tailwind.css`. `kiw` had no uniform `add/remove` — `go get` for Go, `npm install` for JS, file creation for plugins were three different mental models, breaking the `one CLI` promise (`KWF-M8K2Q`, `AGENTS.md` Progressive pipeline P0→P6). Future plugins (e.g., `daisyui`, `autoprefixer`) would have repeated the fragmentation.

This spec makes `kiw add package@version` / `kiw remove package` the single entry for **every** package kind, with `package@latest` as latest. It reuses the existing `framework/plugin` registry so new plugins need zero devtool changes.

## 2. Problem Statement

- **Current pain:** One plugin (`twcss`) required manual `npm install` + `tailwind.config.js` + `assets/tailwind.css`; no `kiw add/remove` existed. Adding `tailwindcss@1.2.3` vs `twcss@latest` had three mental models (`npm`, `go get`, file copy) breaking the `one CLI` promise.
- **Affected consumers:** `site`/`book` authors, `framework/plugin` contributors adding the next plugin, and reviewers auditing `krewire build` plugin detection.
- **Cost of leaving unsolved:** Each new plugin repeats fragmented install docs, `twcss` onboarding stays manual, and future plugins (`daisyui`, `autoprefixer`) would repeat the same file+npm divergence.

## 3. Goals

- G1 — One syntax `package@version` (`@latest` = latest, bare `pkg` = latest) for plugins, Go modules, and npm packages.
- G2 — Zero devtool change for a new plugin: `plugin.Register(&MyPlugin{})` implementing `Installer` is enough.
- G3 — Fast failure with `did you mean?` and clear `kind` (`plugin`/`gomod`/`npm`) in `slog` and `stdout`.

## 4. Non-Goals

- NG1 — No private registry proxy in v1; `npm`/`go` use their public registries.
- NG2 — No `krewire.yaml` `plugins:` list in v1; detection stays file-based (`tailwind.config.js`) to keep `one config` minimal. Future `krewire lock` is a separate spec.
- NG3 — No version pinning beyond `package@version`; `go.mod`/`package.json` remain source of truth.

## 5. Requirements

| ID | Requirement | Scope | Priority |
|----|-------------|-------|----------|
| KWN-PKG-001 | `kiw add <pkg[@ver]> ...` and `kiw remove <pkg> ...` accept 1..N args, `ParseSpec` supports `pkg`, `pkg@1.2.3`, `pkg@latest`, `@scope/pkg@1.0.0`, `github.com/foo/bar@v1.2.3` (last `@` split, scoped `@` handled). Empty/invalid specs return `ExitCodeUsage` with `usage: kiw add <package[@version]>`. | Module | Must |
| KWN-PKG-002 | Resolver chain is `Plugin → Go → Npm` (`kiw/internal/packages.DefaultChain`). `PluginResolver` uses `plugin.FindInstaller` (Name + Aliases `twcss`/`tailwindcss` + `tailwind`); `GoResolver` triggers when name contains `/` and first segment has `.` (e.g., `github.com/...`); `NpmResolver` is fallback. | Module | Must |
| KWN-PKG-003 | Plugin contract: `framework/plugin/plugin.go` `Plugin` gains `Aliases() []string`, `Installer` (`Add(root,version)`, `Remove(root)`) and helpers `Find`/`FindInstaller` (case-insensitive). `Tailwind` implements `Aliases: [twcss, tailwindcss]`, `Add` creates `tailwind.config.js` + `assets/tailwind.css` from templates if missing then `npm install -D tailwindcss@<ver>`; `Remove` deletes configs (keeps customized `assets/tailwind.css`) + `npm uninstall`. | Module | Must |
| KWN-PKG-004 | Generic installers: `npm: pkg@ver` → `npm install [-D for plugin] <pkg>@<ver>` (`latest` when `ver==""`/`latest`); `gomod: pkg@ver` → `go get <pkg>@<ver>` (`v` prefix added if missing) / `go get <pkg>@none` fallback to `go mod edit -droprequire`. Failures return `ExitCodeFailure` with `npm/go not found` hint. | Module | Must |
| KWN-PKG-005 | UX: `kiw add` logs `slog.Info adding package package= version= kind= resolver=` and stdout `→ adding <pkg@ver> (kind) … / ✓ added`; `kiw remove` mirrors `→ removing … / ✓ removed`. Unknown command already has `did you mean` via `framework/tui`. Help groups `add`/`remove` under `PROJECT` with examples `kiw add twcss@latest`, `kiw remove twcss`. | Module | Must |
| KWN-PKG-006 | Tests: `go vet ./framework/plugin ./kiw/...` clean, `go test ./kiw/internal/commands` ok, `bin/kiw --help` lists `add`/`remove`, `bin/kiw help add` shows `package@version` usage, and a temp `myapp` `kiw add twcss@latest` creates `tailwind.config.js` + `assets/tailwind.css` then `kiw remove` cleans. | Module | Must |

## 6. Non-Functional Requirements

- NFR1 — **Stdlib-first, no new deps** for devtool (`os/exec`, `path/filepath`); `plugin` stays stdlib-only.
- NFR2 — **Scalable:** adding `daisyui` is 1 file implementing `Installer`, no `kiw` change.
- NFR3 — `gofmt`/`go vet`/`go test` green; `bin/kiw` rebuilt after `framework/plugin` change.

## 7. Success Criteria

- S1 — `kiw add twcss` ≡ `kiw add tailwindcss@latest` → creates `tailwind.config.js` + `assets/tailwind.css`; `kiw remove twcss` cleans.
- S2 — `kiw add lodash@latest` (npm) and `kiw add github.com/google/uuid@latest` (gomod) resolve to `npm`/`gomod` kind and run `npm install`/`go get` (gracefully warn if `npm`/`go` missing).
- S3 — `kiw add` with no args → `ExitCodeUsage` + usage line.

## 8. Related Specifications

| SpecID | Title |
|--------|-------|
| KWF-M8K2Q | Unified Framework Vision |
| KWF-ARCH-P7L2Q | Progressive Pipeline (P0→P6) |
| KWF-UI-T9X2K | Tailwind Support (plugin detection) |
| KWL-ARCH-J2K9Q | Ecosystem Scope Levels |
| KWN-Z0VFC | Krewire Devtool — Initial |

## 9. References

- `framework/plugin/plugin.go` — `Registry`, `FindInstaller`
- `kiw/internal/packages/{spec,resolver,npm,gomod}.go` — `ParseSpec`, `DefaultChain`
- `kiw/internal/commands/add.go`, `remove.go` — `RegisterAdd`/`RunAdd`
- `AGENTS.md` — One CLI, progressive batteries

