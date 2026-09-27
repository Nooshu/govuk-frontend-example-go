# Tech stack

**Status: Go** — this is the Go specialised line of [govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example).

Sync shared docs/dotfiles from the language-agnostic template: [syncing-from-template.md](syncing-from-template.md).

## Two layers

| Layer                          | Stack                                                                                               | Notes                                                                                                                                    |
| ------------------------------ | --------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **GOV.UK Frontend (upstream)** | **Node** package (`govuk-frontend`), **Nunjucks** macros (`template.njk`), official `fixtures.json` | Fixed by GDS. Node is for install, fixtures, Sass, and optional freshness checks — **not** for request-time HTML in this line.           |
| **This line (wrapper)**        | **Go** (≥1.25), module `github.com/Nooshu/govuk-frontend-example-go`                                | Server-side HTML generated **natively in Go**. Tracks Frontend macros/`template.njk` and proves **backend ≡ every fixture**. No SPA UIs. |

## Prefer the Go standard library

Build on packages that ship with Go before inventing helpers or pulling frameworks:

| Concern                         | Prefer                                                                                             | Avoid / notes                                                                                       |
| ------------------------------- | -------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| HTTP server and routing         | `net/http` (Go 1.22+ method/path `ServeMux`)                                                       | Gin, Echo, Chi — extra surface without benefit for this example                                     |
| Request cookies                 | `http.Request.Cookie` / `Cookies`                                                                  | Hand-rolled `Cookie` header parsers                                                                 |
| Set-Cookie serialization        | `http.Cookie` + `Valid` + `String`, after baseline policy checks                                   | Hand-built `Set-Cookie` strings                                                                     |
| Request body size limits        | `http.MaxBytesReader` (+ `http.MaxBytesError` for 413)                                             | `io.LimitReader` alone for request bodies                                                           |
| Form fields (urlencoded)        | `net/url.ParseQuery` (or `Request.ParseForm` when the body is still on the request)                |                                                                                                     |
| Multipart uploads               | `mime/multipart` — this line keeps only the filename, so it does **not** call `ParseMultipartForm` | Storing unvalidated upload bytes in memory                                                          |
| Page document shell             | `html/template` + `embed`                                                                          | Hand-pasted full page HTML; React/Vue/etc. for UI                                                   |
| Component HTML                  | Idiomatic Go ports in `internal/govuk` (`strings.Builder` / helpers) matching fixtures             | Shelling out to Node/Nunjucks; incomplete third-party wrappers that skip fixture parity             |
| Text escaping (Nunjucks parity) | Small local escaper (`&quot;`, `&#39;`, `\` → `&#92;`) in `internal/govuk` / `internal/htmlutil`   | Relying only on `html.EscapeString` / `html/template` for fixture text — they do not match Nunjucks |
| JSON (fixtures, policy)         | `encoding/json`                                                                                    |                                                                                                     |
| Gzip                            | `compress/gzip`                                                                                    |                                                                                                     |
| Brotli                          | [`github.com/andybalholm/brotli`](https://github.com/andybalholm/brotli) (no stdlib Brotli)        | Reimplementing Brotli                                                                               |
| Sessions / tokens / ETags       | `crypto/rand`, `crypto/sha256`; in-memory `session.Store`                                          |                                                                                                     |
| CSRF compare                    | `crypto/subtle.ConstantTimeCompare`                                                                | Plain `==` on tokens                                                                                |
| Multipart uploads (HTTP API)    | `mime/multipart` via `net/http`                                                                    |                                                                                                     |
| Logging                         | `log/slog`                                                                                         | Ad-hoc `fmt.Println` in handlers                                                                    |
| Tests                           | `testing`, `net/http/httptest`; optional `github.com/google/go-cmp` for HTML diffs                 |                                                                                                     |
| Collections                     | `maps`, `slices`, `strings`                                                                        |                                                                                                     |

### Evaluated and not adopted

| Library / approach                                                                          | Why not here                                                                                                                                                                      |
| ------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Gin / Echo / Chi                                                                            | `net/http` ServeMux is enough; frameworks add surface without helping Frontend parity                                                                                             |
| [`unrolled/secure`](https://github.com/unrolled/secure) and similar header middleware       | This line must emit the shared [`baseline/policy.json`](../baseline/policy.json) contract (CSP hash, cache kinds). A generic middleware would drift from the Node baseline oracle |
| [`alexedwards/scs`](https://github.com/alexedwards/scs), Gorilla sessions                   | Fine for a production service; this example keeps a tiny in-memory `session.Store` so personal answers stay off disk. Swap the store interface when you need Redis/SQL            |
| Gorilla CSRF / nosurf                                                                       | Double-submit cookie + `crypto/subtle` is enough for the example. Prefer a maintained CSRF package when you add cross-site cookie complexity                                      |
| [`github.com/0xnu/govuk-frontend-go`](https://pkg.go.dev/github.com/0xnu/govuk-frontend-go) | Incomplete vs Frontend macros; Gin-centric; does not prove official `fixtures.json` byte-parity. Do not replace `internal/govuk`                                                  |
| [`a-h/templ`](https://github.com/a-h/templ), gomponents, Jet, Pongo2, Quicktemplate         | Fine for **ordinary** Go HTML pages. They do **not** reproduce Nunjucks whitespace, attribute order, or escape rules (`&#39;`, `\`), so they cannot own GOV.UK component HTML     |
| Calling Nunjucks from Go at request time                                                    | Forbidden here: request-time HTML must be native Go. Node stays for the pin, Sass, and fixture freshness only                                                                     |

### HTML rendering (two layers)

This line splits HTML into two jobs. Different tools fit each:

| Layer                          | What it is                                                        | Choice here                                    | Notes                                                                                                                       |
| ------------------------------ | ----------------------------------------------------------------- | ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| **Page shell / service pages** | Layout, forms journey, catalogue chrome around components         | **`html/template` + `embed`** (stdlib)         | Auto-escapes; embeds `.gohtml` files; calls `component` for GOV.UK blocks. Do not invent a second full document template    |
| **GOV.UK components**          | Every `govuk-*` block; must match official `fixtures.json` `html` | **`internal/govuk` ports** (`strings.Builder`) | Source of truth is Frontend `template.njk` + fixtures. Generic template engines fight attribute order and Nunjucks escaping |

**stdlib `html/template`:** use it for pages. It is the maintained Go default for HTML. Do **not** use it (or `html.EscapeString` alone) to implement component fixtures — Nunjucks escapes `'` as `&#39;` and `\` as `&#92;`, and `{%-` strips whitespace; Go’s escaper does not.

**Popular Go HTML libraries (templ, gomponents, etc.):** good ergonomics for greenfield apps. They are the wrong primary tool for **component** HTML in this template because the parity gate is byte-for-byte equality with GDS fixtures. Optional later for service-only pages only if you keep `internal/govuk` as the component API.

**GOV.UK-specific Go wrappers:** none currently prove full fixture parity for the pinned Frontend release. Keep maintaining `internal/govuk`; re-evaluate a wrapper only if it documents and tests **every** fixture `html` for the same pin.

**Third-party rule:** only add a module when the standard library lacks the feature (today: Brotli). Prefer well-known, actively maintained packages; pin versions in `go.mod`; document the choice in this file.

## Go conventions

**Every** feature and code change must follow **current Go** best practices for the pinned major version:

- Module path matches the public repo; `go 1.25` (or newer) in `go.mod`
- Standard layout: `cmd/server`, `internal/…` for non-exportable packages
- Exported identifiers documented; package comments on every package
- Table-driven tests; **100%** function / branch / statement coverage for application packages (see [testing-components.md](testing-components.md))
- `go test ./…`, `go vet ./…`, `go fmt` / `gofmt` before verify
- Component options use the same names as Frontend macros so fixture options decode cleanly

Shared Node tooling (Sass pipeline, `baseline/` JS tests, docs scripts) stays ESM / Node 22+. Dual-audience documentation: [documentation-structure.md](documentation-structure.md).

## Consistency tooling

```sh
npm install
npm run build:styles   # Sass → dist/stylesheets/application.css
npm start              # build:styles, then go run ./cmd/server — http://127.0.0.1:3000
npm test               # baseline, Sass, then go test ./… (fixture parity + service); 100% coverage
npm run verify:docs    # Prettier + markdownlint
npm run verify         # docs + build:styles + go vet + tests
npm run sync:template  # pull shared paths from language-agnostic template
```

See [CONTRIBUTING.md](../CONTRIBUTING.md).

## Shared baseline

[`baseline/`](../baseline/) is synced from the language-agnostic template. This line **implements the same policy in Go** (`internal/baseline`), reading [`baseline/policy.json`](../baseline/policy.json). It does not call the Node helpers at request time.

| Piece                                             | How this line uses it                                                                                   |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| [`baseline/policy.json`](../baseline/policy.json) | OWASP header values, CSP (including the Frontend `js-enabled` hash), cache kinds, Brotli budgets        |
| [`baseline/*.mjs`](../baseline/)                  | Shared contract + Node test gate (`npm run test:baseline`); Go mirrors behaviour in `internal/baseline` |
| [`styles/`](../styles/)                           | Sass entry compiling Frontend via `@use`, then `govuk-overrides.scss` ([styles.md](styles.md))          |

Local `npm start` is plain HTTP, so responses omit HSTS and the session cookie is `rod_session` without `Secure`. An `https:` request URL, or `X-Forwarded-Proto: https`, sends HSTS and `__Host-session`. Public HTML that sets a cookie uses `private, no-cache`. Pages that show the application use `sensitive-document` (`no-store`). The fingerprinted compiled stylesheet, Frontend script, `initAll()` module, and hashed fonts use `public, max-age=31536000, immutable`. Unhashed asset URLs use `no-cache`. Do not serve `govuk-frontend.min.css` as the long-term CSS source.

The server compresses with Brotli when the client advertises `br`, and Gzip otherwise (`compress/gzip` + `andybalholm/brotli`). `Vary: Accept-Encoding` comes from the baseline.

Details: [frontend-performance.md](frontend-performance.md), [frontend-security.md](frontend-security.md).

## Version pin

| Item                              | Value                                                                                                                                                                                |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Implementation language           | Go 1.25+                                                                                                                                                                             |
| Templating / component approach   | Native Go HTML (`internal/govuk`, `html/template` page shell); tracks Frontend macros; **no** Node render at request time                                                            |
| `govuk-frontend` (Node)           | **6.5.1** — [v6.5.1](https://github.com/alphagov/govuk-frontend/releases/tag/v6.5.1) (reviewed against [latest release](https://github.com/alphagov/govuk-frontend/releases/latest)) |
| Sass pipeline                     | `styles/application.scss` → `npm run build:styles` → `dist/stylesheets/application.css` ([styles.md](styles.md))                                                                     |
| Compression                       | Brotli via `andybalholm/brotli`; Gzip via `compress/gzip`                                                                                                                            |
| Backend parity (primary)          | `go test` — Go `Render` ≡ every official `fixtures.json` `html` (including hidden) ([testing-components.md](testing-components.md))                                                  |
| Nunjucks freshness (secondary)    | Optional; fixtures and macros come from the same pinned package — never a substitute for backend parity                                                                              |
| Page template reference           | https://design-system.service.gov.uk/styles/page-template/                                                                                                                           |
| Fixture testing guide             | https://frontend.design-system.service.gov.uk/testing-your-html/                                                                                                                     |
| Example service                   | [example-service.md](example-service.md) — `npm start`                                                                                                                               |
| Response baseline                 | [`baseline/policy.json`](../baseline/policy.json) via `internal/baseline` — [frontend-performance.md](frontend-performance.md), [frontend-security.md](frontend-security.md)         |
| Upgrade / test / preview commands | `npm run build:styles`, `npm start`, `npm test`, `npm run verify`; Frontend upgrade per [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md)                                   |

## Hard constraints (always)

See [`AGENTS.md`](../AGENTS.md): Frontend fixture parity, no SPA UI frameworks, Sass pipeline (no `!important` in service CSS), baseline headers, Brotli-first compression, 100% coverage, dual-audience docs.
