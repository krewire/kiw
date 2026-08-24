# Specification — Config Directory & Dotenv

| Field       | Value                                          |
| ----------- | ---------------------------------------------- |
| SpecID      | KWN-Q7X4M                                      |
| Title       | Config Directory & Dotenv (provider → config/ → .env → krewire.yaml) |
| Status      | Draft                                          |
| Date        | 2026-08-23                                     |
| Author      | Krewire Contributors                            |
| Domain      | Devtools — Project Configuration               |

## 1. Context

Laravel popularized a layered configuration flow: service providers consume a
`config/` directory whose files define structure and defaults, reading
environment-specific values from `.env` via `env('KEY', 'default')`. Krewire
projects are Go; an equivalent layering gives projects the same ergonomics
while keeping `krewire.yaml` focused on devtool behavior.

## 2. Problem Statement

Today every runtime value must live either in code literals or in
`krewire.yaml`. There is no place for machine-local secrets and environment
tuning (.env), no conventional home for application defaults (config/), and
`krewire.yaml` conflates devtool behavior with application settings — making
environment promotion error-prone.

## 3. Goals

- G1 — A conventional project `config/` package holding typed, documented
  application settings as compiled Go getters.
- G2 — A `.env` file (plus committed `.env.example`) loaded into the process
  environment at boot; real environment variables always win over `.env`.
- G3 — Resolution flow: service provider → `config/` getter → `.env` /
  process env → code default; `krewire.yaml` remains authoritative only for
  internal `kiw` behavior (kind pins, dirs, ssg/book settings) and acts as
  last-resort fallback for shared keys.
- G4 — Scaffold generates the convention on `kiw init`.

## 4. Non-Goals

- NG1 — Dynamic/hot reload of config files.
- NG2 — Secrets management or encryption.
- NG3 — Per-environment yaml variants (`krewire.prod.yaml`).
- NG4 — Replacing `libs/config`'s YAML+env overlay loader used by tooling.

## 5. Requirements

| ID            | Requirement                                                                                        | Priority | Scope    |
| ------------- | -------------------------------------------------------------------------------------------------- | -------- | -------- |
| KWL-CFGV-005  | Scaffold emits `.env.example`, `.env` (gitignored), and a `config/config.go` whose getters read env with in-code defaults. | Should | Domain |
| KWL-CFGV-006  | `kiw run` / `kiw dev` load `.env` before spawning the child so providers see resolved values.        | Must     | Func     |

The library primitives behind this flow — `config.LoadDotEnv`,
`config.ParseDotEnv`, and `(config.Vars).GetOr` — live in
`github.com/krewire/libs/config` and are specified in
[KWL-2X1QZ](https://github.com/krewire/libs/blob/main/docs/specs/KWL-CONFIG-2X1QZ-configuration-loading.md)
§5.5 (rows `CFG-DOTV-001..003`, `CFG-KV-001..005`, formerly `libs/cfg`).

## 6. Non-Functional Requirements

- NFR1 — Determinism: identical `.env` content resolves identically.
- NFR2 — No `unsafe`; stdlib plus `gopkg.in/yaml.v3` only in `libs/config`.
- NFR3 — Quality gates pass in every touched repo.

## 7. Success Criteria

- S1 — A project with `APP_NAME="Demo"` in `.env` and a getter
  `Name() string { return envOr("APP_NAME", "Krewire App") }` resolves
  `"Demo"` under `kiw run`, and `"Krewire App"` when `.env` is absent.
- S2 — An exported shell variable overrides the same key in `.env`.
- S3 — `krewire.yaml` gains no application-value fields; it keeps kind/dir/
  build concerns only.

## 8. Related Specifications

| SpecID    | Title                                    |
| --------- | ---------------------------------------- |
| KWL-M1ZKS | Krewire Libraries — Initial Specification |
| KWL-K4T7W | Environments & Debug Mode                |
| KWL-2X1QZ | Configuration Loading (YAML + env overlay) |
| KWN-RD3WS | Project Scaffolding                      |
