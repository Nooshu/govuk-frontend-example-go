# Onboarding

Human-oriented map of this repository. Coding agents should treat [`AGENTS.md`](../AGENTS.md) as the dense entry point; humans should also read [`CONTRIBUTING.md`](../CONTRIBUTING.md). How docs are split for both audiences: [documentation-structure.md](documentation-structure.md).

## What this repo is

A **base template** for **GDS-compliant** frontends on **TypeScript / Node**: **GOV.UK Frontend** is the only UI library; **no frontend frameworks** for UI. Exact **HTML parity** against official Frontend fixtures. See [project-purpose.md](project-purpose.md). Sync shared docs from the agnostic template: [syncing-from-template.md](syncing-from-template.md).

**Stack:** TypeScript — [tech-stack.md](tech-stack.md). Prefer **Nunjucks macros** from `govuk-frontend`; keep Node scripts for fixtures and verification.

**Official guidance:** search the URLs in [guidance-sources.md](guidance-sources.md).

**Priorities:** frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design ([priorities.md](priorities.md)).

**Documentation:** every lasting change is documented for **humans and agents** ([documentation-structure.md](documentation-structure.md)).

**HTML:** prefer **Nunjucks macros**; **TypeScript vs fixture** parity for every fixture (`npm test`). Before Frontend upgrades, always read https://github.com/alphagov/govuk-frontend/releases/latest.

## Priorities

See [priorities.md](priorities.md). Short version: frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design.

## Components vs patterns

| Kind          | What it is                                                               | How we build it                                  | Fixture parity?                                                         |
| ------------- | ------------------------------------------------------------------------ | ------------------------------------------------ | ----------------------------------------------------------------------- |
| **Component** | Design System building block (button, text input, …)                     | Library wrapper that renders exact Frontend HTML | **Yes** — official `fixtures.json`                                      |
| **Pattern**   | Guidance for a journey or page composition (addresses, check answers, …) | Compose shipped components into pages            | **No** — follow Design System guidance; no invented pattern HTML suites |

## Repo map

```text
AGENTS.md                 # Slim agent playbook
docs/                     # All documentation (this folder)
baseline/                 # Shared performance and OWASP header contract, synced from the template
styles/                   # Sass entry + govuk-overrides (compiles to dist/stylesheets/)
scripts/                  # Node build helpers (styles, template sync)
src/                      # TypeScript application (ESM)
  main.ts                 # npm start
  app.ts                  # Routes and the Node HTTP server
  components/             # Nunjucks macro renderer, fixtures, catalogue
  views/                  # Page templates (layout, service, demos)
  service/                # Rod licence journey
  http/                   # Request body, cookies, assets, security headers
  pages/                  # Page document around the Frontend template
  session/                # In-memory session
```

Detail: [example-service.md](example-service.md) and [tech-stack.md](tech-stack.md).

## Run modes

| Mode    | Command                | Purpose                                                                                                       |
| ------- | ---------------------- | ------------------------------------------------------------------------------------------------------------- |
| Styles  | `npm run build:styles` | Compile `styles/` → `dist/stylesheets/application.css` ([styles.md](styles.md))                               |
| Preview | `npm start`            | build:styles, then example service, component catalogue, and fixture previews                                 |
| Test    | `npm test`             | Baseline, Sass pipeline, fixture parity, and service tests. Fails below 100% coverage                         |
| Verify  | `npm run verify`       | Docs, build:styles, `lint:go` (vet / fix / staticcheck), and the full test suite (Go `-race` + 100% coverage) |
| Upgrade | see the playbook       | Frontend bump — [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md)                                    |

## Testing mindset

1. **Parity checks (primary)** compare **TypeScript `renderComponent`** output to fixture `html` with ordinal string equality — every fixture from the pinned Frontend release (`src/components/render.test.ts`).
2. **Nunjucks freshness (secondary)** — fixtures are loaded from the same pinned `govuk-frontend` package the macros come from; a separate Nunjucks-only suite is optional. Never treat “macros still match fixtures” as a substitute for the TypeScript parity suite.
3. **Never** edit fixture `html` to make tests pass — fix the renderer.
4. **Never** normalise HTML in tests.
5. A green Nunjucks-only check alone does **not** prove this line’s API is correct.

Details: [testing-components.md](testing-components.md).

## Troubleshooting

| Symptom                        | Likely cause                                                |
| ------------------------------ | ----------------------------------------------------------- |
| Parity fails on whitespace     | Renderer ≠ Nunjucks `template.njk` / `{%-` stripping        |
| Encoding differs (`'` vs `'`)  | Used framework HTML encoder instead of Nunjucks `escape`    |
| Attribute order differs        | Built attributes in code property order, not template order |
| Preview/fixture 404 in tests   | Test host not enabling Dev/Testing routes                   |
| Logo unreadable / wrong header | Frontend 5 header classes with Frontend 6+ CSS              |
| Editing fixtures “fixes” tests | Wrong fix — update renderer                                 |

More pitfalls: [creating-components.md](creating-components.md).

## Consistency tooling

```sh
npm install
npm run build:styles  # Sass → dist/stylesheets/application.css
npm start
npm test              # baseline, Sass pipeline, then the TypeScript suite
npm run verify:docs   # Prettier + markdownlint
npm run verify
```

See [CONTRIBUTING.md](../CONTRIBUTING.md). Dotfiles: `.editorconfig`, `.prettierrc.json`, `.markdownlint-cli2.jsonc`, `.nvmrc`, `.vscode/`, `.cursor/rules/`, `.github/`.

## Next reads

1. [documentation-structure.md](documentation-structure.md)
2. [tech-stack.md](tech-stack.md)
3. [page-shell.md](page-shell.md) and [layout-chrome.md](layout-chrome.md)
4. [govuk-components.md](govuk-components.md)
5. [service-assessment-readiness.md](service-assessment-readiness.md)
6. [frontend-performance.md](frontend-performance.md) and [frontend-security.md](frontend-security.md)
7. [styles.md](styles.md)
8. [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md)
