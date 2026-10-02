# Architecture

How this Go line is put together: packages, request flow, and where GOV.UK Frontend fits. For “why Go”, see [why-go.md](why-go.md). For pins and tooling, see [tech-stack.md](tech-stack.md).

## Big picture

```text
Browser
   │
   ▼
cmd/server          # process: listen, timeouts, graceful Shutdown
   │
   ▼
internal/app        # ServeMux, session cookie, CSRF, write path
   │
   ├─► internal/pages + html/template   # page shell (landmarks, chrome slots)
   │         │
   │         └─► render.Renderer
   │                   │
   │                   └─► internal/govukrender → internal/govuk.Render
   │                         (native ports of Frontend macros / template.njk)
   │
   ├─► internal/service                 # rod licence journey (answers, validate, save)
   ├─► internal/session                 # in-memory Store
   ├─► internal/httpx                   # bodies, cookies, assets, Brotli/Gzip
   ├─► internal/baseline                # policy.json → OWASP headers, cache kinds
   ├─► internal/components              # Frontend package metadata + fixtures list
   └─► internal/config                  # paths, PORT, demos flag, Frontend version
```

Node is **off** the request path. It installs `govuk-frontend`, compiles `styles/` → `dist/stylesheets/application.css`, and runs shared baseline/docs tests.

## Repository layout

```text
AGENTS.md                 # Slim agent entry
README.md                 # Human entry
cmd/server/               # main package only
internal/
  app/                    # HTTP application
  baseline/               # Shared policy implementation (Go)
  components/             # Catalogue + fixture loading from node_modules
  config/                 # Runtime configuration
  govuk/                  # Component HTML ports + Params + parity tests
  govukrender/            # render.Renderer adapter over govuk.Render
  htmlutil/               # Small shared HTML/text helpers
  httpx/                  # Transport helpers
  pages/                  # Document / layout composition
  render/                 # Renderer interface
  service/                # Example journey domain
  session/                # Session types + memory store
baseline/                 # Synced policy JSON + Node oracle tests
styles/                   # Sass entry + overrides
scripts/                  # Node helpers (Sass, sync, coverage check)
docs/                     # Dual-audience documentation
tests/                    # Node tests for baseline + Sass
```

## Package responsibilities

| Package                | Responsibility                                                                                                     |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `cmd/server`           | Process entry. Loads config, builds `app.App`, serves with timeouts, shuts down on signal. Excluded from coverage. |
| `internal/app`         | Routes, session open/save, CSRF, demos catalogue, writing responses through the baseline.                          |
| `internal/pages`       | GOV.UK page template structure in Go (`Document`, before-content, component helper).                               |
| `internal/govuk`       | Ordered `Params`, Nunjucks-parity escaping, every shipped component renderer, fixture parity tests.                |
| `internal/govukrender` | Thin adapter: `map[string]any` page options → `govuk.Params` → HTML.                                               |
| `internal/render`      | `Renderer` interface so tests can stub components without Frontend HTML.                                           |
| `internal/service`     | Rod fishing licence steps, validation, field builders that call the component API.                                 |
| `internal/session`     | Session values + `Store` interface + in-memory implementation.                                                     |
| `internal/httpx`       | Max body size, cookie helpers, asset map, compression wrapper.                                                     |
| `internal/baseline`    | Parse `baseline/policy.json`; apply headers, cache kinds, cookie serialisation rules.                              |
| `internal/components`  | Discover component names and load official `fixtures.json` from the pinned package.                                |
| `internal/config`      | Find repo root, Frontend version, stylesheet path, port, demos flag.                                               |
| `internal/htmlutil`    | Shared escaping / small text helpers for pages.                                                                    |

## Request lifecycle

1. **Listen** — `cmd/server` binds `0.0.0.0` on `tcp4` (so hosts such as Render detect the port) and configures read/write/idle timeouts. `HOST=127.0.0.1` limits that to loopback.
2. **Route** — `app` ServeMux matches method + path (service steps, assets, demos when enabled).
3. **Session** — open or create a session; set `rod_session` or `__Host-session` when HTTPS-shaped.
4. **Handler** — validate POST (CSRF, body limits), update answers, choose view model.
5. **Render** — page shell asks `Renderer` for each GOV.UK block; `govuk.Render` emits fixture-faithful HTML.
6. **Write** — buffer body → strong ETag → conditional 304 → baseline headers + cache kind → Brotli if `Accept-Encoding` includes `br`, else Gzip → send.

Personal answers use a **sensitive document** cache kind so intermediaries do not store them. See [frontend-performance.md](frontend-performance.md) and [frontend-security.md](frontend-security.md).

## Two HTML layers

| Layer             | Tooling                          | Contract                                   |
| ----------------- | -------------------------------- | ------------------------------------------ |
| Page shell        | `html/template` + embed          | Landmarks, title, before-content, slots    |
| GOV.UK components | `internal/govuk` strings builder | Byte-equality with official fixture `html` |

Do not use generic Go HTML libraries to own component markup — they will lose Nunjucks whitespace and escape parity. Detail: [tech-stack.md](tech-stack.md), [page-shell.md](page-shell.md), [testing-components.md](testing-components.md).

## Configuration and assets

- **Config** resolves the repository root (walks up for `go.mod` / `package.json`), reads the pinned `govuk-frontend` version, and points at `dist/stylesheets/application.css` and Frontend’s JS/assets under `node_modules`.
- **Styles** must be built before tests or start (`npm run build:styles`).
- **Demos** (catalogue, fixture previews) default on locally; `NODE_ENV=production` turns them off unless `DEMOS_ENABLED=true`.

## Testing architecture

| Suite                       | Command / location                     | Proves                                      |
| --------------------------- | -------------------------------------- | ------------------------------------------- |
| Go fixture parity (primary) | `go test` in `internal/govuk`          | Go HTML ≡ every fixture `html`              |
| Go application tests        | `npm run test:go`                      | Routes, validation, baseline, 100% coverage |
| Node baseline + Sass        | `npm run test:baseline`, `test:styles` | Shared policy oracle + Sass pipeline        |
| Docs hygiene                | `npm run verify:docs`                  | Prettier + markdownlint                     |
| Go lint                     | `npm run lint:go`                      | `go vet`, `go fix -diff`, `staticcheck`     |

## Extension points for a real service

When forking this template into a live service:

1. Keep `internal/govuk` + fixture parity as the component boundary.
2. Replace `session.Store` with Redis/SQL if you need multi-instance durability.
3. Keep implementing [`baseline/policy.json`](../baseline/policy.json) — do not invent a weaker header set.
4. Add domain packages beside `service/`; do not paste `govuk-*` HTML into templates.
5. Update dual-audience docs in the same change set ([documentation-structure.md](documentation-structure.md)).

Using this repository does **not** by itself make a service assessment-ready — see [service-assessment-readiness.md](service-assessment-readiness.md).
