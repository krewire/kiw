# Specification — Krewire × Node.js Ecosystem Integration

| Field       | Value                                      |
| ----------- | ------------------------------------------ |
| SpecID      | KWN-NDJ5S                                  |
| Title       | Krewire × Node.js Ecosystem Integration    |
| Status      | Draft                                      |
| Date        | 2026-08-30                                 |
| Author      | Krewire Contributors                        |
| Domain      | Ecosystem Integration                     |
| Scope       | ARCH                                        |

## Scope

This spec targets the **Krewire ecosystem** (Workspace → Module → Domain → Service → Unit). It defines how the Krewire Go framework and `kiw` devtool integrate with the **Node.js ecosystem** — the dominant JavaScript runtime for web services — without either ecosystem shadowing or replacing the other.

Containment: `Workspace ⊃ Module ⊃ Domain ⊃ Service ⊃ Unit`. The integration operates at the **Module** level (a Node.js project is a JavaScript/TypeScript module; a Krewire Go service is a Go module) and the **Service** level (Krewire `service`/`worker` kinds running alongside Node.js backends and frontends).

---

## 1. Context

### 1.1 The Node.js Ecosystem Landscape (2026)

Node.js remains the dominant runtime for JavaScript/web services. As of 2026:
- **Node.js 26.x** is the Current LTS; **Node.js 24.x** is Active LTS.
- The ecosystem is split into two broad categories of frameworks:

**Fullstack meta-frameworks** (frontend + backend in one framework):
- **SvelteKit** — Svelte's fullstack framework. ~82K GitHub stars, ~500K npm downloads/week. 93% developer satisfaction in 2026. 20-40% smaller bundles than React-based alternatives. Low learning curve. Vercel-backed (Rich Harris employed by Vercel). Adapters for Vercel, Netlify, Cloudflare, Node.js. Growing rapidly as a Next.js alternative.
- **Nuxt** — Vue's meta-framework. Nitro engine targets Node, serverless, edge. Strong in enterprise dashboards and content-heavy Vue apps.
- **Remix** — React-based, Shopify-acquired. Web standards, progressive enhancement, form-first. React Router v7 merger.
- **Astro** — Static-first, zero-JS-by-default. Excellent for blogs, docs, marketing. Content-heavy sites.
- **SolidStart** — Solid.js meta-framework. Fine-grained reactivity, matching or beating Svelte in update performance.

**Backend-only frameworks** (server-side logic):
- **NestJS** — Enterprise TypeScript backend. Decorator syntax, modules, controllers, providers, dependency injection. First-party integrations: TypeORM, Mongoose, Prisma, JWT, Passport, Swagger. NestJS 11 released 2026. Used by enterprise teams (O2 Czech Republic, others). Dominant in the enterprise Node.js backend segment.
- **Express** — Minimalist, ubiquitous. Still the most-installed Node.js web framework.
- **Fastify** — High-performance, schema-driven. Growing rapidly.
- **Hono** — Edge-native, lightweight. Bun/Cloudflare Workers focused.
- **Elysia** — Bun-optimized, TypeScript-first.
- **tRPC** — End-to-end type safety. TypeScript-only.

### 1.2 Krewire's Unique Differentiator in Node.js

Krewire enters the Node.js ecosystem not as another JavaScript framework, but as a **Go-native alternative** at multiple layers:

| Krewire Capability | What It Replaces/Complements in Node.js | Language |
|--------------------|-----------------------------------------|----------|
| `framework/runtime` (Go→WASM VDOM) | SvelteKit/Next.js/Vue/React frontend | Go compiled to WASM |
| `framework/web` + `ssg` | Node.js HTTP server + SSR templates | Go |
| `framework/service` | NestJS/Express/Fastify microservices | Go |
| `framework/worker` | Node.js background job libraries (Bull, Agenda) | Go |
| `framework/infra` | Terraform, Pulumi, CloudFormation | Go |
| `kiw` CLI | npm scripts, Makefile, CI config | Go |
| Guild template | AI agent configuration (AGENTS.md, opencode.json) | Go tool |

Krewire's **WASM runtime** (`KWF-T4X9P`) is the most significant differentiator: it compiles Go to WebAssembly, providing a VDOM-based frontend with progressive hydration — **without any JavaScript framework**. This means a Krewire project can ship a complete fullstack application where both the frontend and backend are written in Go, running in the browser and on the server respectively.

### 1.3 Why Integrate Rather Than Replace

The Node.js ecosystem has an enormous installed base: millions of developers, thousands of npm packages, and extensive TypeScript tooling. Asking these developers to abandon everything for a pure-Go stack is neither practical nor aligned with Krewire's mission. Coexistence allows:

- Node.js developers to adopt Krewire for performance-critical services, workers, and infrastructure without learning Go for everything
- Krewire developers to leverage the rich Node.js ecosystem (npm packages, TypeScript tooling, TypeScript-based frontends) where Go isn't the right fit
- Gradual migration: start with Node.js, add Krewire services as performance needs grow, eventually move more to Go

---

## 2. Problem Statement

### 2.1 The JavaScript Tooling Tax

Node.js developers face a "JavaScript toolchain tax": package managers (npm, yarn, pnpm), bundlers (Webpack, Vite, Turbopack), state management libraries, CSS-in-JS, TypeScript configuration, linting, testing frameworks — each with its own config, version conflicts, and update cycles. Krewire's Go-native approach eliminates this tax for the parts of the stack where Go excels.

### 2.2 The Frontend Fragmentation

The Node.js frontend landscape is fragmented across React/Next.js, Vue/Nuxt, SvelteKit, SolidStart, Astro, and more. Each requires a separate skill set, separate tooling, and separate mental model. Krewire's WASM runtime offers a **unified Go-based frontend** that works across all deployment targets without a JavaScript framework — but it must coexist with, not confront, the existing ecosystem.

### 2.3 The Enterprise Backend Gap

Node.js backend frameworks range from minimalist (Express) to enterprise (NestJS). Teams that need high-throughput, concurrent, or CPU-intensive backend services often hit Node.js's single-threaded limitations. Krewire's Go services fill this gap — providing the same backend capabilities (API, workers, message queues) with Go's concurrency model and performance.

### 2.4 No Documented Integration Path

Today, there is no documented or supported way to add Krewire capabilities to a Node.js project. A developer using SvelteKit or NestJS who wants to use Krewire for performance-critical tasks must manually configure cross-runtime tooling, manage two independent build systems, and figure out communication patterns themselves.

---

## 3. Goals

| ID        | Goal                                                                                             | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| G1        | **Non-intrusive AI agent install.** `kiw guild install` must work inside any Node.js project directory, adding Krewire's AI agent config without modifying any JS/TS file. | Must |
| G2        | **Side-by-side Go service.** `kiw new` inside a Node.js project creates a Go module subdirectory that coexists with `package.json`, `node_modules/`, and framework-specific directories (`src/`, `pages/`, `app/`). | Must |
| G3        | **Build and run integration.** `kiw build` and `kiw run` must work for Krewire Go projects inside a Node.js repository, compiling Go binaries that are addressable from Node.js. | Must |
| G4        | **WASM frontend coexistence.** Krewire's `framework/runtime` (Go→WASM) must be installable as a frontend replacement or co-exist alongside a Node.js meta-framework (SvelteKit, Nuxt, etc.) for specific routes or mount points. | Should |
| G5        | **HTTP/gRPC communication.** Krewire Go services must communicate with Node.js backends via standard protocols (HTTP/JSON, gRPC, message queues). | Must |
| G6        | **`kiw info` recognizes Node.js projects.** When run inside a Node.js project, `kiw info` reports the project type and lists co-located Krewire projects. | Should |
| G7        | **Unified deploy.** `kiw deploy --plan` generates infrastructure that includes both Node.js and Go services. | Could |
| G8        | **Shared observability.** Krewire Go service logs and Node.js logs are distinguishable but support distributed tracing across the boundary. | Could |

---

## 4. Non-Goals

| ID        | Non-Goal                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------ |
| NG1       | Replacing Node.js's JavaScript runtime or npm ecosystem. Krewire does not compile or execute JS/TS. |
| NG2       | Providing a Krewire equivalent of Next.js's App Router, Nuxt's Nitro, or NestJS's module system. Krewire uses `krewire.yaml` and Go, not JS/TS decorators. |
| NG3       | Migrating existing Node.js applications to Krewire. This spec enables coexistence, not migration.    |
| NG4       | Modifying Node.js framework internals. Krewire integrates at the service/binary/boundary level.    |
| NG5       | Supporting `npm install krewire` as a Node.js package. Krewire distributes as Go modules and binaries, not npm packages (except the `krewire/krewire-laravel` PHP wrapper pattern — not applicable here). |
| NG6       | Providing a Krewire plugin for SvelteKit, Nuxt, NestJS, or any specific Node.js framework in this phase. Framework-specific plugins are future work. |
| NG7       | Replacing TypeScript tooling. Krewire does not provide TypeScript compilers, type checkers, or linting for JS/TS. |

---

## 5. Requirements

### 5.1 Must Requirements

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| KWN-NDV-001 | **Guild install works in Node.js projects.** `kiw guild install <nodejs-project-dir>` installs AGENTS.md, opencode.json, and .agents/ into the target directory without modifying `package.json`, `node_modules/`, `src/`, or any JS/TS file. | Must |
| KWN-NDV-002 | **AGENTS.md is project-agnostic.** The installed AGENTS.md must contain the existing disclaimer: "If the current project is not a Krewire project, follow the generic principles and skip Krewire-specific sections marked Krewire." | Must |
| KWN-NDV-003 | **Side-by-side scaffolding.** `kiw new <name> --dir <nodejs-project-dir>` creates a Go module (go.mod, krewire.yaml, main.go, .gitignore) as a subdirectory of the Node.js project without touching Node.js files. | Must |
| KWN-NDV-004 | **No port or process conflict by default.** When Node.js dev server (commonly :3000, :5173, :8080) and `kiw run` (default :8080) run simultaneously, there must be no port conflict unless explicitly configured to the same port. | Must |
| KWN-NDV-005 | **Independent build systems.** `npm install` / `pnpm install` and `kiw build` must execute independently without interfering with each other's dependencies or output directories. | Must |
| KWN-NDV-006 | **HTTP/gRPC bridge.** Krewire `service` kind must support HTTP/JSON and gRPC endpoints that any Node.js framework (Express, Fastify, NestJS, etc.) can call via its HTTP client or gRPC client. | Must |
| KWN-NDV-007 | **Process bridge.** `kiw run` must support spawning a Go binary that reads JSON from stdin and writes JSON to stdout, enabling Node.js's `child_process` module to communicate with it. | Must |
| KWN-NDV-008 | **Message queue integration.** Krewire `worker` kind must consume from message queues (Redis, RabbitMQ, NATS) that Node.js backends also publish to. | Must |
| KWN-NDV-009 | **Directory isolation.** Krewire project files (go.mod, krewire.yaml, go/, internal/) must not conflict with Node.js files (package.json, node_modules/, src/, pages/, app/, public/). | Must |
| KWN-NDV-010 | **Exit code compatibility.** `kiw` commands inside a Node.js project must return standard exit codes without being affected by Node.js's own exit codes. | Must |

### 5.2 Should Requirements

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| KWN-NDV-011 | **`kiw info` detects Node.js.** When run inside a Node.js project directory (detected by `package.json`), `kiw info` reports the project type and lists any co-located Krewire projects. | Should |
| KWN-NDV-012 | **Shared `.gitignore`.** When `kiw new` creates a Go subdirectory inside a Node.js project, the Go `.gitignore` must not overwrite Node.js's `.gitignore`. Both files must coexist. | Should |
| KWN-NDV-013 | **Cross-runtime tracing.** Trace IDs propagate from Node.js requests to Krewire Go services, enabling distributed tracing across the runtime boundary. | Should |
| KWN-NDV-014 | **Shared logging format.** Krewire Go service logs and Node.js logs must include distinguishable source markers (e.g., `[go]` vs `[js]`) but support aggregation in a shared log output. | Should |
| KWN-NDV-015 | **WASM mount points alongside Node.js frontend.** Krewire's `framework/runtime` can mount WASM-powered interactive components inside a Node.js-rendered page (progressive hydration), similar to how Astro islands work. | Should |
| KWN-NDV-016 | **`kiw build` as unified build.** `kiw build` can trigger both `npm run build` (for the Node.js frontend) and `go build` (for Go services) in a single command. | Should |

### 5.3 Could Requirements

| ID        | Requirement                                                                                      | Priority |
| --------- | ------------------------------------------------------------------------------------------------ | -------- |
| KWN-NDV-017 | **WASM-first frontend.** Krewire's `framework/runtime` (Go→WASM) replaces the Node.js meta-framework's frontend entirely for the application, with Node.js serving only as the API/backend. | Could |
| KWN-NDV-018 | **`kiw dev` concurrency.** `kiw dev` watches Go sources and restarts the Go binary while Node.js's dev server runs concurrently, with unified log streaming. | Could |
| KWN-NDV-019 | **Shared service registry.** Krewire's `framework/service` registry discovers both Node.js APIs and Go microservices for unified service mesh routing. | Could |
| KWN-NDV-020 | **WASM widget for Node.js.** Krewire's runtime widgets (Container, Text, Button, Input, etc.) are exposed as a standalone WASM module that Node.js projects can embed as interactive elements. | Could |
| KWN-NDV-021 | **Unified `kiw deploy` multi-runtime.** `kiw deploy --plan` generates an infrastructure plan that includes both Node.js and Go services, with appropriate container images, orchestration, and scaling. | Could |
| KWN-NDV-022 | **Krewire as Node.js package manager alternative.** `kiw add` manages Go dependencies alongside `npm`/`pnpm`/`yarn` managing JS dependencies, with a unified view. | Could |

---

## 6. Non-Functional Requirements

| ID | Requirement | Detail |
| -- | ----------- | ------ |
| NFR1 | **Runtime isolation.** Go binaries and Node.js processes run as separate OS processes. No shared memory, no shared runtime. Communication only through defined boundaries (HTTP, gRPC, stdin/stdout, message queues, WASM). |
| NFR2 | **Dependency isolation.** `go.mod` and `package.json` are independent. No shared dependency resolution. `go build` must not touch `node_modules/`; `npm install` must not touch `go.sum`. |
| NFR3 | **Memory safety.** The `unsafe` package must not be used in any Krewire code that integrates with Node.js. |
| NFR4 | **Deterministic builds.** `kiw build` produces reproducible Go binaries regardless of the Node.js project's state. |
| NFR5 | **Portability.** Integration must work on Linux, macOS, and Windows. |
| NFR6 | **Quality gates.** `gofmt -l .`, `go vet ./...`, `go test ./...` must pass for the Krewire components. Node.js's own `npm test` / TypeScript checker must pass independently. |
| NFR7 | **Idempotence.** Re-running `kiw guild install` on a Node.js project must not modify existing managed files unless `--force` is passed. |
| NFR8 | **Security.** Go binaries spawned by Node.js must validate all input from JavaScript and enforce the same timeout/security boundaries as any external process call. WASM modules must run in the browser sandbox. |
| NFR9 | **Performance.** Go worker startup via process bridge must not add more than 100ms overhead per invocation. WASM hydration must complete within 500ms (per `KWF-T4X9P`). gRPC persistent connections have zero startup overhead. |
| NFR10 | **Observability.** Node.js and Krewire logs must include distinguishable source markers. Trace IDs must propagate across the JS→Go and JS→WASM boundaries when available. |
| NFR11 | **WASM size budget.** Hello-world WASM app with one mount point is ≤ 800 KB gzipped (per `KWF-T4X9P`). |
| NFR12 | **TypeScript compatibility.** WASM module must not conflict with TypeScript's module resolution or namespace declarations. |

---

## 7. Success Criteria

| ID | Criterion | Verification |
| -- | --------- | ------------ |
| S1 | `kiw guild install` inside a fresh Node.js project (Next.js, Nuxt, NestJS, or Express) installs all managed files without modifying any JS/TS file. | Manual test + `diff` before/after install |
| S2 | `kiw new go-backend --dir nodejs-project/` creates `go-backend/` with go.mod, krewire.yaml, main.go alongside `package.json` and `node_modules/`. | Directory listing verification |
| S3 | `kiw build && kiw run` inside the Go subdirectory compiles and runs a Go service while `npm run dev` runs the Node.js dev server. Both respond to HTTP requests. | `curl :3000` returns Node.js response; `curl :8080` returns Go response |
| S4 | Node.js `child_process` can spawn a Krewire Go binary and communicate via stdin/stdout JSON. | Node.js test that calls `child_process.exec('kiw run process')` and validates JSON response |
| S5 | A Krewire Go service is callable from a Node.js backend via HTTP/gRPC. | Integration test: NestJS/Express → Go service HTTP call, validates response |
| S6 | A Krewire Go worker consumes from the same Redis/RabbitMQ queue that a Node.js backend publishes to. | Integration test: Node.js publishes job, Go worker consumes and processes |
| S7 | `kiw info` inside a Node.js project reports the project type and any co-located Krewire projects. | `kiw info` output parsing |
| S8 | `gofmt -l .`, `go vet ./...`, `go test ./...` pass for Krewire integration code. Node.js quality gates pass independently. | CI pipeline verification |
| S9 | WASM frontend mounts into a Node.js-rendered page as an island (progressive hydration), showing identical content before and after hydration. | Browser test: SSR HTML readable without JS; after WASM loads, interactive |
| S10 | Distributed trace ID propagates from Node.js request header to Krewire Go service header. | Trace ID in Node.js logs matches trace ID in Go service logs |

---

## 8. Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Node.js Project Directory                          │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Node.js (JavaScript/TypeScript)                                 │ │
│  │  ├── src/ or app/      Application code (SvelteKit/NestJS/etc.) │ │
│  │  ├── package.json      Node.js dependencies                    │ │
│  │  ├── node_modules/     Installed packages                       │ │
│  │  ├── tsconfig.json     TypeScript configuration                │ │
│  │  └── .env              Node.js environment                     │ │
│  └──────────────────────┬─────────────────────────────────────────┘ │
│                         │                                            │
│          ┌──────────────┼──────────────────┐                         │
│          │ HTTP/gRPC    │                  │                         │
│          │ child_process│                  │                         │
│          │ Message Queue│                  │                         │
│          │ WASM embed   │                  │                         │
│          ▼              ▼                  ▼                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │  Go Service   │  │  Go Worker   │  │  Krewire     │              │
│  │  (kiw run)    │  │  (kiw        │  │  WASM        │              │
│  │  :8080        │  │   worker)    │  │  Runtime     │              │
│  │  krewire.yaml │  │  :9090       │  │  .wasm       │              │
│  └──────┬───────┘  └──────┬───────┘  │  (in browser)│              │
│         │                 │          └──────────────┘              │
│         └────────┬────────┘                                         │
│                  │                                                  │
│  ┌──────────────▼──────────────────────────────────────────────┐   │
│  │  Shared Infrastructure                                       │   │
│  │  ├── Redis / RabbitMQ / NATS (message queues)               │   │
│  │  ├── PostgreSQL / MySQL / MongoDB (databases)               │   │
│  │  ├── PostgreSQL / Redis (Krewire service registry)          │   │
│  │  └── Monitoring (Prometheus, Grafana, distributed tracing)  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Krewire AI Agents (installed by `kiw guild install`)           │ │
│  │  ├── AGENTS.md          Project-agnostic agent constitution    │ │
│  │  ├── opencode.json      AI tool configuration                  │ │
│  │  └── .agents/           20+ agents, 28 skills, commands        │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Krewire WASM Frontend (optional, alongside Node.js frontend)   │ │
│  │  ├── runtime/            Go→WASM VDOM + component model        │ │
│  │  ├── framework/ui/       Theme + scoped CSS                     │ │
│  │  ├── widgets/            Layout, input, display widgets         │ │
│  │  └── hydrate("load")    Progressive hydration mount points     │ │
│  └─────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────┘
```

### 8.1 Communication Patterns

| Pattern | Node.js → Krewire | Use Case | Latency |
|---------|-------------------|----------|---------|
| **Process bridge** | Node.js `child_process` → Go binary (stdin/stdout JSON) | Image processing, data transformation, report generation, CLI tools | ~50-150ms per invocation |
| **HTTP/JSON** | Node.js HTTP client (fetch, Axios, Guzzle) → Krewire service | CRUD APIs, health checks, config retrieval, auth validation | ~5-20ms |
| **gRPC** | gRPC Node.js client → Krewire service | High-throughput, persistent connections, streaming, real-time | ~1-5ms |
| **Message queue** | Node.js queue publisher → Redis/RabbitMQ/NATS → Krewire worker | Background jobs, async processing, cron tasks, event-driven architecture | Depends on queue driver |
| **WASM embed** | Node.js HTML → WASM module loaded in browser → interactive UI | Interactive dashboards, real-time UI, form-heavy components | ~500ms first hydration |

### 8.2 Coexistence Modes

| Mode | Description | When to Use |
|------|-------------|-------------|
| **Sidecar** | Go service runs alongside Node.js app, same repo, different ports | Teams adding Go for specific performance-critical services |
| **Worker** | Go workers consume from queues that Node.js publishes to | Offloading CPU-heavy background jobs from Node.js |
| **Island (WASM)** | Krewire WASM components mount inside Node.js-rendered pages | Replacing JS-heavy interactive components with Go→WASM |
| **Fullstack Go** | Krewire provides both frontend (WASM) and backend (Go); Node.js handles only the API gateway or reverse proxy | Teams wanting to eliminate JavaScript from the stack |
| **Hybrid** | Mix of all modes — Node.js for web layer, Go for services/workers, WASM for specific interactive components | Maximum flexibility, gradual adoption |

---

## 9. Node.js Ecosystem Landscape (Integration Map)

### 9.1 Fullstack Meta-Frameworks

| Framework | Krewire Integration Path | Notes |
|-----------|--------------------------|-------|
| **SvelteKit** | WASM islands alongside Svelte components; Go service as API backend | SvelteKit's adapter system can serve Krewire-built static assets alongside Svelte output |
| **Nuxt** | WASM mount points in Nuxt pages; Go service via Nitro server middleware | Nuxt's Nitro engine supports middleware that proxies to Krewire services |
| **Remix** | Go service as Remix route handlers' backend; WASM for interactive UI | Remix's `loader`/`action` pattern can call Krewire services |
| **Astro** | Krewire WASM as Astro island; Go service for API routes | Astro's island architecture is a natural fit for WASM progressive hydration |
| **SolidStart** | WASM components alongside Solid components; Go backend | Solid's fine-grained reactivity complements WASM's imperative DOM updates |
| **Next.js** | Same patterns as other meta-frameworks | Even if Next.js declines, integration paths remain valid for existing users |

### 9.2 Backend Frameworks

| Framework | Krewire Integration Path | Notes |
|-----------|--------------------------|-------|
| **NestJS** | Go microservice as NestJS microservice via gRPC or HTTP; Go worker on NestJS queue | NestJS's microservice architecture natively supports gRPC, Redis, MQTT transports |
| **Express** | Go service behind Express reverse proxy; Go binary via `child_process` | Simplest integration — Express routes proxy to Krewire services |
| **Fastify** | Go service via Fastify proxy; message queue bridge | Fastify's performance complements Go's performance |
| **Hono** | Go service as Hono middleware target; edge-compatible | Hono's edge runtime can serve alongside Krewire edge deployments |
| **Elysia** | Go service behind Elysia; shared schema validation | Elysia's TypeScript-first schema can align with Krewire's typed configs |
| **tRPC** | Go service exposing gRPC endpoints; tRPC client → gRPC | tRPC's end-to-end typing can extend to Go service contracts |

### 9.3 Infrastructure & Tooling

| Tool | Krewire Integration | Notes |
|------|---------------------|-------|
| **Docker** | Both Node.js and Go in same or separate containers; `kiw build` produces Go binary | Standard multi-container or multi-service Docker Compose |
| **Kubernetes** | `kiw deploy --plan` generates K8s manifests for Go services; Node.js via Helm | Krewire `infra` kind targets Kubernetes |
| **CI/CD** | `kiw build` + `npm run build` in single pipeline; `kiw test` + `npm test` | Both run in GitHub Actions, GitLab CI, etc. |
| **Vercel/Netlify/Cloudflare** | Node.js frontend on platform; Krewire Go service on separate host or edge | Krewire adapters support these platforms |
| **Redis/RabbitMQ/NATS** | Shared message queues for Node.js ↔ Go worker communication | Krewire `worker` kind supports all three |

---

## 10. Integration Phases

### Phase 1: Coexistence (Immediate — no new code)

- [ ] `kiw guild install` works in Node.js project directories (G1 — already implemented)
- [ ] `kiw new <name> --dir <nodejs-dir>` creates side-by-side Go module (G2 — scaffold exists, needs validation)
- [ ] `kiw build` / `kiw run` work for Go subdirectory (G3 — existing behavior, needs validation)
- [ ] Process bridge documented: Node.js `child_process` → Go binary (stdin/stdout JSON) (G5)

### Phase 2: Functional Integration (Build)

- [ ] HTTP/gRPC service template with Node.js call examples (G6)
- [ ] Message queue integration: Node.js publisher → Krewire worker consumer (G8)
- [ ] `kiw info` detects Node.js projects (G6)
- [ ] `.gitignore` coexistence handling (G12)
- [ ] Cross-runtime distributed tracing (G13)
- [ ] Shared logging format (G14)

### Phase 3: WASM Frontend Integration (Polish)

- [ ] WASM mount points alongside Node.js meta-framework pages (G15) — follows `KWF-T4X9P` progressive hydration model
- [ ] Krewire runtime widgets as standalone WASM module embeddable in Node.js projects (G20)
- [ ] `kiw build` unified build triggering both Node.js and Go builds (G16)
- [ ] `kiw dev` concurrent watching for both runtimes (G18)

### Phase 4: Infrastructure & Scale (Full Integration)

- [ ] Unified `kiw deploy` multi-runtime infrastructure (G7, G21)
- [ ] Shared service registry discovering both Node.js and Go services (G19)
- [ ] WASM-first frontend replacing Node.js meta-framework (G17)
- [ ] Shared observability with distributed tracing across all runtimes (G8)

---

## 11. Krewire WASM Runtime as Node.js Integration Point

Krewire's `framework/runtime` (specified in `KWF-T4X9P`) is the most distinctive integration path into the Node.js ecosystem. Unlike competing JavaScript frameworks, Krewire's WASM runtime:

### 11.1 How It Works

1. **Build**: `krewire build --target=wasm` compiles Go to WebAssembly using `GOOS=js GOARCH=wasm`
2. **Output**: `.wasm` module + JS glue (~800 KB gzipped for hello-world app)
3. **Mount**: WASM component mounts into SSR HTML via `hydrate="load"|"idle"|"visible"` markers
4. **Hydrate**: Client boot scans markers, instantiates WASM module, attaches event listeners
5. **VDOM**: Same virtual DOM used for server rendering and client reconciliation
6. **Components**: `Container`, `Text`, `Button`, `Input`, `Scaffold`, `AppBar`, `Card` — all written in Go
7. **Theme**: `framework/ui` Theme vars → CSS custom properties, shared between SSR and WASM

### 11.2 Integration Modes with Node.js Frontend

| Mode | Description | Framework Compatibility |
|------|-------------|------------------------|
| **WASM island** | Krewire WASM component mounts as an interactive island inside Node.js-rendered HTML | Astro, Next.js, Nuxt, SvelteKit, any SSR framework |
| **WASM widget** | Standalone WASM module embeddable via `<script>` tag in any HTML page | Any framework or plain HTML |
| **WASM-first** | Krewire provides the entire frontend via WASM; Node.js serves only the API | Any Node.js backend |
| **WASM + SSR hybrid** | Node.js handles SSR for content pages; WASM handles interactive sections | Best for content-heavy apps with interactive features |

### 11.3 Why This Matters for Node.js

The Node.js ecosystem's frontend is fragmented across React, Vue, Svelte, Solid, and more — each requiring separate learning curves, tooling, and mental models. Krewire's WASM runtime offers:

- **One language for frontend and backend**: Go for everything, no JS/TS framework switching
- **No virtual DOM tax in the browser**: WASM runs at near-native speed
- **Progressive hydration**: SSR content is fully readable before WASM loads
- **No npm dependency for the frontend**: No `node_modules` bloat from frontend libraries
- **Type safety**: Go's type system replaces TypeScript for frontend code
- **Small bundles**: ≤ 800 KB gzipped for hello-world (competitive with SvelteKit's ~40-60% smaller bundles)

This makes Krewire a unique offering in the Node.js ecosystem: a complete Go-native fullstack that eliminates JavaScript from the frontend while still integrating with existing Node.js backends.

---

## 12. Related Specifications

| SpecID | Title | Relationship |
|--------|-------|-------------|
| [KWF-M8K2Q](https://github.com/krewire/framework/blob/main/docs/specs/KWF-ARCH-M8K2Q-unified-framework-vision.md) | Unified Vision — one Go framework for every workload | Parent vision. Krewire's workload matrix defines the capabilities that integrate with Node.js. |
| [KWL-ARCH-J2K9Q](https://github.com/krewire/libs/blob/main/docs/specs/KWL-ARCH-J2K9Q-scope-levels.md) | Scope Levels | Defines Workspace → Module → Domain → Service → Unit containment. Integration operates at Module and Service scope. |
| [KWF-T4X9P](https://github.com/krewire/framework/blob/main/docs/specs/KWF-WASM-T4X9P-wasm-client-runtime.md) | WASM Client Runtime — Go-Native Frontend | Krewire's WASM frontend that integrates with Node.js meta-frameworks as islands or replacement. |
| [KWN-GUILD-MZ4LE](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-GUILD-MZ4LE-guild-install-command.md) | guild install Command | The install mechanism that enables Path 1 (non-intrusive AI agent config). |
| [KWN-RD3WS](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-SCAFFOLD-RD3WS-project-scaffolding.md) | Project Scaffolding | `kiw new` and `kiw init` mechanisms that enable Path 2 (side-by-side Go module). |
| [KWN-RUN-6K41E](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-RUN-6K41E-krewire-run-dev-deploy.md) | run/dev/deploy | `kiw run` and `kiw dev` mechanisms that enable Path 3 (build and run integration). |
| [KWF-SVC-L5H2F](https://github.com/krewire/framework/blob/main/docs/specs/KWF-SVC-L5H2F-microservice-patterns.md) | Microservice Patterns | Defines service/worker patterns that integrate with Node.js backends via HTTP/gRPC/queue. |
| [KWF-ARCH-P7L2Q](https://github.com/krewire/framework/blob/main/docs/specs/KWF-ARCH-P7L2Q-progressive-pipeline.md) | Progressive Pipeline | Krewire's progressive scaling: static → islands → monolith → modular monolith → headless → services → mesh. Node.js can occupy any stage. |
| [KWN-CONF-MRCN](./KWN-CONF-MRCN-multi-runtime-config.md) | Multi-Runtime Unified Configuration | Defines runtime namespace isolation, `.env` loading, and `build.depends_on` for Node.js + Krewire coexistence. |
| [KWN-BUILD-MLTB](./KWN-BUILD-MLTB-multi-runtime-build.md) | Multi-Runtime Build Orchestration | Orchestrates `npm run build` and `kiw build` in a unified build command. |
| [KWN-L4RVE](https://github.com/krewire/kiw/blob/main/docs/specs/KWN-ARCH-L4RVE-krewire-laravel-coexistence.md) | Krewire × Laravel Coexistence | Sister spec for PHP ecosystem integration. Same architectural patterns apply to Node.js. |
| `govel` (`mpge/govel`) | Go-powered task execution for Laravel | Reference implementation for Go+language runtime process/gRPC bridge pattern. |

---

## 13. References

- [Node.js](https://nodejs.org) — the dominant JavaScript runtime
- [SvelteKit](https://svelte.dev/docs/kit) — Svelte fullstack meta-framework, ~82K stars, 93% satisfaction
- [NestJS](https://nestjs.com) — enterprise TypeScript backend, dominant in Node.js enterprise segment
- [Krewire Framework](https://github.com/krewire/framework) — unified Go framework
- [Krewire Devtool (`kiw`)](https://github.com/krewire/kiw) — single CLI for all workloads
- [Krewire Guild](https://github.com/krewire/guild) — AI agent template, installable into any project
- [Krewire WASM Runtime (`KWF-T4X9P`)](https://github.com/krewire/framework/blob/main/docs/specs/KWF-WASM-T4X9P-wasm-client-runtime.md) — Go→WASM VDOM frontend with progressive hydration
- [Krewire Project Vision](https://github.com/krewire/framework/blob/main/internal/docs/project-vision.md) — workload matrix, progressive pipeline
- [Krewire × Laravel Coexistence](internal docs) — sister spec applying the same patterns to PHP
- [Govel (`mpge/govel`)](https://github.com/mpge/govel) — reference implementation for Go+runtime process/gRPC bridge
- [Stack Overflow Developer Survey 2025](https://survey.stackoverflow.co/2025/technology) — JavaScript remains most-used language; Node.js dominant runtime

---

## 14. Revision History

Revision history is tracked by **git**, not in-file metadata. To view changes:

```bash
git log --oneline -- docs/specs/KWN-ARCH-NDJ5S-nodejs-ecosystem-integration.md
```

Initial draft: 2026-08-30.
