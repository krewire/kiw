# Specifications Index — Krewire Devtool

This directory holds the formal specifications for the krewire devtool.

Ordered by **impact-to-effort** (high impact, low effort first) and **dependency chain** (foundations first).

| SpecID    | Title                                      | Status | Depends On |
| --------- | ------------------------------------------ | ------ | ---------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md)      | Krewire Devtool — Initial Specification | Draft | — |
| [KWN-6K41E](./KWN-RUN-6K41E-krewire-run-dev-deploy.md) | krewire run/dev/deploy              | Draft | KWN-Z0VFC |
| [KWN-SCRIPT-9F3KQ](./KWN-SCRIPT-9F3KQ-kiw-script-runner.md) | kiw Script Runner — `kiw run <task>` and `kiw run file.go` | Draft | KWN-6K41E, KWN-Q7X4M, KWN-Z0VFC |
| [KWN-RD3WS](./KWN-SCAFFOLD-RD3WS-project-scaffolding.md) | Project Scaffolding                  | Draft | KWN-Z0VFC |
| [KWN-7QM2X](./KWN-INIT-7QM2X-init-project-variants.md) | Init Project Variants                    | Draft | KWN-RD3WS |
| [KWN-1QGI2](./KWN-BUILD-1QGI2-project-building.md)   | Project Building                      | Draft | KWN-Z0VFC |
| [KWN-MLTB](./KWN-BUILD-MLTB-multi-runtime-build.md) | Multi-Runtime Build Orchestration      | Draft | KWN-1QGI2, KWN-CONF-MRCN |
| [KWN-BNKJC](./KWN-INFO-BNKJC-project-information.md) | Project Information                  | Draft | KWN-Z0VFC |
| [KWN-P0FWA](./KWN-TEST-P0FWA-project-validation.md)  | Project Validation                   | Draft | KWN-Z0VFC |
| [KWN-JB7PW](./KWN-INFO-JB7PW-version-reporting.md)   | Version Reporting                     | Draft | KWN-Z0VFC |
| [KWN-MZ4LE](./KWN-GUILD-MZ4LE-guild-install-command.md) | krewire guild install Command        | Draft | KWN-Z0VFC |
| [KWN-Q3M8V](./KWN-WORKSPACE-Q3M8V-workspace-command.md) | Workspace Command — Multi-Repo & Monorepo Workflows | Draft | KWN-Z0VFC |
| [KWN-Q7X4M](./KWN-CONF-Q7X4M-config-directory-and-dotenv.md) | Config Directory & Dotenv (provider → config/ → .env → krewire.yaml) | Draft | KWL-K4T7W, KWL-2X1QZ, KWN-RD3WS |
| [KWN-CONF-MRCN](./KWN-CONF-MRCN-multi-runtime-config.md) | Multi-Runtime Unified Configuration | Draft | KWN-Q7X4M, KWN-RD3WS |
| [KWN-L4RVE](./KWN-ARCH-L4RVE-krewire-laravel-coexistence.md) | Krewire × Laravel Coexistence & Integration | Draft | KWN-GUILD-MZ4LE, KWN-RD3WS, KWN-RUN-6K41E, KWF-L5H2F |
| [KWN-NDJ5S](./KWN-ARCH-NDJ5S-nodejs-ecosystem-integration.md) | Krewire × Node.js Ecosystem Integration | Draft | KWN-GUILD-MZ4LE, KWN-RD3WS, KWN-RUN-6K41E, KWF-T4X9P, KWF-SVC-L5H2F |
| [KWN-PKG-7X9K2](./KWN-PKG-7X9K2-add-remove-package.md) | kiw add/remove — scalable package@version (plugin → gomod → npm) | Draft | — |

## Conventions

- Each specification is stored as a single Markdown file named `{ProjectId}-{Scope}-{SpecID}-{slug}.md`.
- SpecIDs are unique, random 5-character alphanumeric codes (e.g., `KWN-Z0VFC`).
- New specifications must be added to this index when created.
- Ordering: impact-to-effort (high impact, low effort first), then dependency chain (foundations first).