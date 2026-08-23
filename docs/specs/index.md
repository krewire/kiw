# Specifications Index — Krewire Devtool

This directory holds the formal specifications for the krewire devtool.

Ordered by **impact-to-effort** (high impact, low effort first) and **dependency chain** (foundations first).

| SpecID    | Title                                      | Status | Depends On |
| --------- | ------------------------------------------ | ------ | ---------- |
| [KWN-Z0VFC](./KWN-DEVTOOL-Z0VFC-krewire-devtool.md)      | Krewire Devtool — Initial Specification | Draft | — |
| [KWN-6K41E](./KWN-RUN-6K41E-krewire-run-dev-deploy.md) | krewire run/dev/deploy              | Draft | KWN-Z0VFC |
| [KWN-RD3WS](./KWN-SCAFFOLD-RD3WS-project-scaffolding.md) | Project Scaffolding                  | Draft | KWN-Z0VFC |
| [KWN-7QM2X](./KWN-INIT-7QM2X-init-project-variants.md) | Init Project Variants                    | Draft | KWN-RD3WS |
| [KWN-1QGI2](./KWN-BUILD-1QGI2-project-building.md)   | Project Building                      | Draft | KWN-Z0VFC |
| [KWN-BNKJC](./KWN-INFO-BNKJC-project-information.md) | Project Information                  | Draft | KWN-Z0VFC |
| [KWN-P0FWA](./KWN-TEST-P0FWA-project-validation.md)  | Project Validation                   | Draft | KWN-Z0VFC |
| [KWN-JB7PW](./KWN-INFO-JB7PW-version-reporting.md)   | Version Reporting                     | Draft | KWN-Z0VFC |
| [KWN-MZ4LE](./KWN-GUILD-MZ4LE-guild-install-command.md) | krewire guild install Command        | Draft | KWN-Z0VFC |
| [KWN-Q7X4M](./KWN-CONF-Q7X4M-config-directory-and-dotenv.md) | Config Directory & Dotenv (provider → config/ → .env → krewire.yaml) | Draft | KWL-K4T7W, KWL-2X1QZ, KWN-RD3WS |

## Conventions

- Each specification is stored as a single Markdown file named `{ProjectId}-{Scope}-{SpecID}-{slug}.md`.
- SpecIDs are unique, random 5-character alphanumeric codes (e.g., `KWN-Z0VFC`).
- New specifications must be added to this index when created.
- Ordering: impact-to-effort (high impact, low effort first), then dependency chain (foundations first).