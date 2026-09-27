# Onboarding

Human-oriented map of this repository. Coding agents should treat [`AGENTS.md`](../AGENTS.md) as the dense entry point; humans should also read [`CONTRIBUTING.md`](../CONTRIBUTING.md). How docs are split for both audiences: [documentation-structure.md](documentation-structure.md).

## What this repo is

A **base template** for **GDS-compliant** frontends on **Go**: **GOV.UK Frontend** is the only UI library; **no frontend frameworks** for UI. Exact **HTML parity** of Go-rendered output against official Frontend fixtures. See [project-purpose.md](project-purpose.md). Sync shared docs from the agnostic template: [syncing-from-template.md](syncing-from-template.md).

**Stack:** Go — [tech-stack.md](tech-stack.md). Why this language: [why-go.md](why-go.md). How packages fit: [architecture.md](architecture.md). Node pins `govuk-frontend`, compiles Sass, and runs shared baseline/docs tests.

**Official guidance:** search the URLs in [guidance-sources.md](guidance-sources.md).

**Priorities:** frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design ([priorities.md](priorities.md)).

**Documentation:** every lasting change is documented for **humans and agents** ([documentation-structure.md](documentation-structure.md)).

**HTML:** native Go ports in `internal/govuk` track Frontend macros; **Go vs fixture** parity for every fixture (`npm test`). Before Frontend upgrades, always read https://github.com/alphagov/govuk-frontend/releases/latest.

## Priorities

See [priorities.md](priorities.md). Short version: frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design.

## Components vs patterns

| Kind          | What it is                                                               | How we build it                                        | Fixture parity?                                                         |
| ------------- | ------------------------------------------------------------------------ | ------------------------------------------------------ | ----------------------------------------------------------------------- |
| **Component** | Design System building block (button, text input, …)                     | Go renderer in `internal/govuk` matching Frontend HTML | **Yes** — official `fixtures.json`                                      |
| **Pattern**   | Guidance for a journey or page composition (addresses, check answers, …) | Compose shipped components into pages                  | **No** — follow Design System guidance; no invented pattern HTML suites |

## Repo map

```text
AGENTS.md                 # Slim agent playbook
docs/                     # All documentation (this folder)
baseline/                 # Shared performance and OWASP header contract, synced from the template
styles/                   # Sass entry + govuk-overrides (compiles to dist/stylesheets/)
scripts/                  # Node build helpers (styles, template sync, coverage check)
cmd/server/               # go run entry — example HTTP server
internal/
  app/                    # Routes, session cookie, response baseline
  baseline/               # Go implementation of baseline/policy.json
  components/             # Frontend package metadata + fixtures catalogue
  config/                 # Paths, port, demos flag
  govuk/                  # Native component HTML + fixture parity tests
  govukrender/            # Renderer adapter for pages
  htmlutil/               # Small HTML/text helpers
  httpx/                  # Bodies, cookies, assets, compression
  pages/                  # Page document / layout shell
  render/                 # Component Renderer interface
  service/                # Rod fishing licence journey
  session/                # In-memory session store
```

Detail: [architecture.md](architecture.md), [example-service.md](example-service.md), and [tech-stack.md](tech-stack.md).

## Run modes

| Mode    | Command                | Purpose                                                                               |
| ------- | ---------------------- | ------------------------------------------------------------------------------------- |
| Styles  | `npm run build:styles` | Compile `styles/` → `dist/stylesheets/application.css` ([styles.md](styles.md))       |
| Preview | `npm start`            | build:styles, then example service, component catalogue, and fixture previews         |
| Test    | `npm test`             | Baseline, Sass pipeline, fixture parity, and service tests. Fails below 100% coverage |
| Lint Go | `npm run lint:go`      | `go vet`, `go fix -diff`, `staticcheck`                                               |
| Verify  | `npm run verify`       | Docs, build:styles, `lint:go`, and the full test suite (Go `-race` + 100% coverage)   |
| Upgrade | see the playbook       | Frontend bump — [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md)            |

## Testing mindset

1. **Parity checks (primary)** compare **Go `govuk.Render`** output to fixture `html` with ordinal string equality — every fixture from the pinned Frontend release (`internal/govuk`).
2. **Nunjucks freshness (secondary)** — fixtures are loaded from the same pinned `govuk-frontend` package the macros come from; a separate Nunjucks-only suite is optional. Never treat “macros still match fixtures” as a substitute for the Go parity suite.
3. **Never** edit fixture `html` to make tests pass — fix the Go renderer.
4. **Never** normalise HTML in tests.
5. A green Nunjucks-only check alone does **not** prove this line’s API is correct.

Details: [testing-components.md](testing-components.md).

## Troubleshooting

| Symptom                        | Likely cause                                               |
| ------------------------------ | ---------------------------------------------------------- |
| Parity fails on whitespace     | Renderer ≠ Nunjucks `template.njk` / `{%-` stripping       |
| Encoding differs (`'` vs `'`)  | Used framework HTML encoder instead of Nunjucks `escape`   |
| Attribute order differs        | Built attributes in map order, not template / Params order |
| Preview/fixture 404 in tests   | Demos disabled (`NODE_ENV=production`) or styles not built |
| Logo unreadable / wrong header | Frontend 5 header classes with Frontend 6+ CSS             |
| Editing fixtures “fixes” tests | Wrong fix — update renderer                                |
| `test:go` skips on stylesheet  | Run `npm run build:styles` first                           |

More pitfalls: [creating-components.md](creating-components.md).

## Consistency tooling

```sh
npm install
npm run build:styles  # Sass → dist/stylesheets/application.css
npm start
npm test              # baseline, Sass pipeline, then the Go suite
npm run lint:go       # vet + fix -diff + staticcheck
npm run verify:docs   # Prettier + markdownlint
npm run verify
```

See [CONTRIBUTING.md](../CONTRIBUTING.md). Dotfiles: `.editorconfig`, `.prettierrc.json`, `.markdownlint-cli2.jsonc`, `.nvmrc`, `.vscode/` (including Go), `staticcheck.conf`, `.cursor/rules/`, `.github/`. Conventions: [go-conventions.md](go-conventions.md).

## Next reads

1. [why-go.md](why-go.md)
2. [architecture.md](architecture.md)
3. [documentation-structure.md](documentation-structure.md)
4. [tech-stack.md](tech-stack.md)
5. [example-service.md](example-service.md)
6. [testing-components.md](testing-components.md)
7. [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md) before any Frontend bump
