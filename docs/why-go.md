# Why Go for this template

This is the **Go** specialised line of the GDS-compliant GOV.UK Frontend template. Humans choosing a stack, and agents justifying design decisions, should start here. Concrete package layout and tooling live in [tech-stack.md](tech-stack.md); request flow lives in [architecture.md](architecture.md).

## What “Go line” means

| Concern                    | Choice in this repository                                                                  |
| -------------------------- | ------------------------------------------------------------------------------------------ |
| Application language       | **Go** (≥1.27)                                                                             |
| HTML for GOV.UK components | Native Go in `internal/govuk` that tracks Frontend macros / `template.njk`                 |
| HTML for page chrome       | `html/template` + `embed`                                                                  |
| HTTP                       | `net/http` ServeMux                                                                        |
| Node’s role                | Pin `govuk-frontend`, compile Sass, shared baseline/docs tests — **not** request-time HTML |
| Frontend UI frameworks     | **Forbidden** (React, Vue, Angular, Svelte, …)                                             |
| Parity gate                | Go `Render` HTML ≡ every official fixture `html`                                           |

Language-agnostic sibling (shared playbooks): [Nooshu/govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example).

## Advantages of Go for GDS-shaped frontends

### 1. One binary, predictable operations

Go compiles to a single statically linked binary (with the usual caveats for `cgo`). That fits government hosting patterns that prefer few runtime dependencies: no Node process on the request path, no JVM, no separate app server. Operators get a clear process model — listen, serve, graceful `Shutdown` — which matches how this line’s `cmd/server` is written.

### 2. Standard library that covers the HTTP surface

For this template’s needs, the standard library is enough for almost everything that touches a request:

- routing (`ServeMux` method/path patterns)
- cookies, forms, multipart boundaries
- TLS-aware cookie prefixes via ordinary `http.Cookie` handling
- HTML escaping for page shells (`html/template`)
- compression (`compress/gzip`) plus one well-known Brotli dependency
- structured logging (`log/slog`)
- testing (`testing`, `httptest`, `testing/synctest`)

Fewer frameworks means fewer places for security and cache policy to drift from the shared [`baseline/`](../baseline/) contract.

### 3. Performance that matches the priority order

This line’s first priority is **frontend web performance**. Go helps on the server side without becoming a SPA:

- low allocation cost for streaming HTML into a buffer, hashing an ETag, then compressing
- cheap goroutines for timeouts and shutdown without event-loop gymnastics
- race detector (`go test -race`) as a first-class CI gate on `./internal/...`

Fast HTML responses + Brotli-first compression + correct cache kinds are easier to keep honest when the stack is thin.

### 4. Strong defaults for security-minded services

Go’s type system, bounds checks, and a culture of small APIs help keep dangerous patterns out of the hot path. This line still does the GDS/OWASP work explicitly (CSP hash, cookie flags, CSRF compare via `crypto/subtle`), but it does **not** need a large middleware ecosystem to stay close to [`baseline/policy.json`](../baseline/policy.json).

Memory safety relative to C/C++ reduces whole classes of bugs in a long-lived public service. Explicit error returns make failure paths visible in reviews — important when a missing policy field must fail at start-up, not on the first citizen request.

### 5. Testing culture that fits fixture parity

Official Frontend fixtures demand **byte-for-byte** HTML equality. Go’s table-driven tests, subtests, and coverage tooling map cleanly onto that gate:

- one test loop over every component fixture
- 100% statement coverage on application packages (CI fails below that)
- `httptest` for handlers; `httptest.NewTestServer` + `synctest` where fake time or an in-memory network helps

The parity suite is ordinary Go tests — agents and humans run the same `npm run test:go` command.

### 6. Maintainability over years of Frontend upgrades

GOV.UK Frontend releases regularly. A specialised line must re-port macro behaviour and keep fixtures green. Go’s readability, `gofmt`/`go fix`, and `staticcheck` keep a large `internal/govuk` surface navigable. Modules + a pinned `toolchain` line make CI and laptops reproduce the same compiler.

Compared with “call Nunjucks from the server on every request”, native Go renderers:

- remove a second runtime from production
- keep escape rules and attribute order under test in one language
- make upgrades a Go code change + fixture refresh, not a hybrid debugging session

### 7. Clear module boundaries (`internal/`)

Go’s `internal/` directory enforces that other modules cannot import this template’s private packages. Inside the repo, packages stay small and purposeful (`baseline`, `httpx`, `govuk`, `service`, …). That matches how agents and humans are expected to navigate the tree — see [architecture.md](architecture.md).

### 8. Progressive enhancement stays honest

Because the server always sends complete HTML, core journeys work without Frontend’s JavaScript. Go does not push teams toward a client-rendered component tree. The Design System’s progressive-enhancement model (skip link, forms with `novalidate`, `initAll()` when JS runs) stays the default.

## Trade-offs (honest)

| Trade-off                                         | How this line handles it                                                                               |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| Component HTML must track Nunjucks macros by hand | Fixture suite is the contract; never edit fixture `html` to pass                                       |
| No JSX-like component DX                          | Page shell uses `html/template`; components use a small Params API                                     |
| Team must know Go **and** enough Node for Sass    | Documented in [onboarding.md](onboarding.md); Node is tooling-only                                     |
| Fewer “batteries included” web frameworks         | Intentional — prefer stdlib; record exceptions in [tech-stack.md](tech-stack.md)                       |
| Hot reload is not built in                        | `npm start` / `go run`; optional local watchers are fine if they do not become a production dependency |

## When Go is a strong fit

- You want server-rendered GOV.UK HTML with a **hard** fixture-parity gate
- You want production free of a Node request path
- You value stdlib HTTP, explicit errors, and high test coverage
- You can invest in maintaining `internal/govuk` across Frontend upgrades

## When to consider another specialised line

- Your organisation standardises on another language for all services and cannot run Go in production
- You need request-time Nunjucks macros from the official package (that is **not** this line’s model)
- You are building a client-heavy SPA (incompatible with this template’s non-negotiables)

## Related docs

| Doc                                                | Why read it                             |
| -------------------------------------------------- | --------------------------------------- |
| [project-purpose.md](project-purpose.md)           | Intent of the template                  |
| [architecture.md](architecture.md)                 | Packages and request flow               |
| [tech-stack.md](tech-stack.md)                     | Pins, conventions, allowed dependencies |
| [go-conventions.md](go-conventions.md)             | Comments, layout, lint, testing norms   |
| [testing-components.md](testing-components.md)     | Fixture parity and coverage             |
| [frontend-performance.md](frontend-performance.md) | Why response shape matters              |
| [frontend-security.md](frontend-security.md)       | Baseline headers and cookies            |
