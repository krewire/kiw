# Philosophy — Krewire Devtool

## Philosophy

**One binary, every workload.** `kiw` is the only entry point (binary `kiw`, module `github.com/krewire/krewire`). New projects start as `kernel` (`go.mod`+`krewire.yaml`+`main.go`+`.gitignore`) and are equipped to a kind via `kiw init` (`--static`/`--book`/`--cli`/`--worker`/`--service`/`--infra`).

**Principles:**

- **Unified dispatch.** Same `kiw` drives `app`/`cli`/`worker`/`service` (`run`/`dev`), `site`/`book` (`build`/`serve`), and `infra` (`deploy`).
- **Spec-driven commands.** Every command has a `KWN-*` spec in `krewire/internal`; requirements declare `Scope` (`KWL-ARCH-J2K9Q`).
- **Modular at every Scope (SRP/SoC).** Industry **Single Responsibility Principle** + **Separation of Concerns** + **High Cohesion/Low Coupling** (Parnas, Unix "Do one thing well"): one `Command` = one `internal/commands` file, one `Func` = one behavior; never stack many unrelated commands in one file. Applies down to `Func`.
- **Local-first.** `bin/kiw` (workspace-built, compat `bin/krewire`) is used to test local framework changes before release.


## Contribution

- Read `internal/docs/project-vision.md` and `internal/docs/specs/krewire/index.md` (or central `internal/docs/specs/index.md`) before changing behavior.
- Add/update tests matching project patterns; keep suite green.
- Update `README.md` / `docs/` and specs when public behavior changes; follow ecosystem spec conventions.
