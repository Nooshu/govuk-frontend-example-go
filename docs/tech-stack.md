# Tech stack

**Status: TypeScript (Node)** — this is the TypeScript specialised line of [govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example).

Sync shared docs/dotfiles from the language-agnostic template: [syncing-from-template.md](syncing-from-template.md).

## Two layers

| Layer                          | Stack                                                                                               | Notes                                                                                                |
| ------------------------------ | --------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| **GOV.UK Frontend (upstream)** | **Node** package (`govuk-frontend`), **Nunjucks** macros (`template.njk`), official `fixtures.json` | Fixed by GDS. Always name Node/Nunjucks for install, fixtures, macros, encoding, verification.       |
| **This line (wrapper)**        | **TypeScript** on **Node** (≥22), ESM (`"type": "module"`)                                          | Server-side HTML from Frontend **macros** (prefer Nunjucks). **No** React/Vue/Angular/Svelte for UI. |

## TypeScript conventions

**Every** feature and code change must follow **TypeScript / Node’s latest** best practices for the pinned major versions (not outdated tutorials):

- TypeScript **7.0.2** (`typescript` on npm). `tsc` is the native compiler.
- `strict`, plus `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, `noImplicitOverride`, `noImplicitReturns`, `noFallthroughCasesInSwitch`, `noUnusedLocals`, `noUnusedParameters`, `noUncheckedSideEffectImports`, `verbatimModuleSyntax`, `isolatedModules`, and `erasableSyntaxOnly`
- `module` and `moduleResolution` are `NodeNext`. `target` and `lib` are `ES2024`, which Node 22 runs.
- `erasableSyntaxOnly` keeps the source free of enums, runtime namespaces, and parameter properties, so the types can be stripped
- ESM only, with `import type` for types
- **TSDoc** on exported functions, classes, and types (`/** … */`, `@param`, `@returns`). That is the TypeScript equivalent of JSDoc. Comments describe the API; they do not replace the types in the signature
- Prefer calling **Nunjucks macros** from `govuk-frontend` for component HTML; thin TypeScript wrappers around options → HTML only when needed — still fixture-parity
- Tests: Node’s built-in test runner via `tsx` (`npm test`)
- Coverage gate: **100%** functions, branches, statements, and lines for application code (see [testing-components.md](testing-components.md))

Shared Node tooling (Sass pipeline, `baseline/`, docs scripts) already uses current ESM / Node 22+ practice; keep it that way. Dual-audience documentation for stack notes: [documentation-structure.md](documentation-structure.md).

## Consistency tooling

```sh
npm install
npm run build:styles   # Sass → dist/stylesheets/application.css
npm start              # build:styles, then example service — http://127.0.0.1:3000
npm test               # baseline, Sass, TypeScript vs every fixture, service tests; 100% coverage
npm run typecheck
npm run verify:docs    # Prettier + markdownlint
npm run verify         # docs + build:styles + typecheck + tests
npm run sync:template  # pull shared paths from language-agnostic template
```

See [CONTRIBUTING.md](../CONTRIBUTING.md).

## Shared baseline

[`baseline/`](../baseline/) is synced from the language-agnostic template. This line calls it. It does not keep a second header or cache policy.

| Piece                                             | How this line uses it                                                                                          |
| ------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| [`baseline/policy.json`](../baseline/policy.json) | OWASP header values, CSP directives (including the Frontend `js-enabled` hash), cache kinds, Brotli budgets    |
| [`baseline/index.mjs`](../baseline/index.mjs)     | `applyResponseHeaders` on every response, `buildSetCookie` for the session cookie, `strongEtag` on public HTML |
| [`styles/`](../styles/)                           | Sass entry compiling Frontend via `@use`, then `govuk-overrides.scss` ([styles.md](styles.md))                 |

`npm run test:baseline` is the template's 100% line, branch, and function gate for `baseline/`. `npm run test:styles` is the same gate for `scripts/build-styles.mjs`. `npm test` runs both, then `build:styles`, then the TypeScript coverage gate.

Local `npm start` is plain HTTP, so responses omit HSTS and the session cookie is `rod_session` without `Secure`. An `https:` request URL, or `X-Forwarded-Proto: https`, sends HSTS and `__Host-session`. Public HTML that sets a cookie uses `private, no-cache`. Pages that show the application use `sensitive-document` (`no-store`). The fingerprinted compiled stylesheet (`/assets/application.*.css`), Frontend script, `initAll()` module, and hashed fonts use `public, max-age=31536000, immutable`. Unhashed asset URLs use `no-cache`. Do not serve `govuk-frontend.min.css` as the long-term CSS source.

The server compresses with Brotli when the client advertises `br`, and Gzip otherwise. `Vary: Accept-Encoding` comes from the baseline.

Details: [frontend-performance.md](frontend-performance.md), [frontend-security.md](frontend-security.md).

## Version pin

| Item                              | Value                                                                                                                                                                                |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Implementation language           | TypeScript 7.0.2 on Node ≥22                                                                                                                                                         |
| Templating / component approach   | Prefer Nunjucks macros from `govuk-frontend`; TypeScript for app/library logic                                                                                                       |
| `govuk-frontend` (Node)           | **6.5.1** — [v6.5.1](https://github.com/alphagov/govuk-frontend/releases/tag/v6.5.1) (reviewed against [latest release](https://github.com/alphagov/govuk-frontend/releases/latest)) |
| Sass pipeline                     | `styles/application.scss` → `npm run build:styles` → `dist/stylesheets/application.css` ([styles.md](styles.md))                                                                     |
| Nunjucks                          | 3.2.4, with Frontend’s `trimBlocks` and `lstripBlocks`                                                                                                                               |
| Backend parity (primary)          | `npm test` — TypeScript `renderComponent` ≡ every official `fixtures.json` `html` (including hidden) ([testing-components.md](testing-components.md))                                |
| Nunjucks freshness (secondary)    | Fixtures and macros come from the same pinned `govuk-frontend` package; optional separate Nunjucks-only suite — never a substitute for backend parity                                |
| Page template reference           | https://design-system.service.gov.uk/styles/page-template/                                                                                                                           |
| Fixture testing guide             | https://frontend.design-system.service.gov.uk/testing-your-html/                                                                                                                     |
| Example service                   | [example-service.md](example-service.md) — `npm start`                                                                                                                               |
| Response baseline                 | [`baseline/`](../baseline/) via `applyResponseHeaders` — [frontend-performance.md](frontend-performance.md), [frontend-security.md](frontend-security.md)                            |
| Upgrade / test / preview commands | `npm run build:styles`, `npm start`, `npm test`, `npm run typecheck`, `npm run verify`; Frontend upgrade per [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md)              |

## Hard constraints (always)

Same as the language-agnostic template: one Frontend pin, macros over pasted HTML, **backend vs fixture** parity for every fixture (Nunjucks-only is not enough), Sass pipeline with `govuk-overrides.scss` last (never `!important` in service CSS), no SPA UI frameworks, 100% coverage for application code, review https://github.com/alphagov/govuk-frontend/releases/latest before upgrades.
See [`AGENTS.md`](../AGENTS.md), [guidance-sources.md](guidance-sources.md), [creating-components.md](creating-components.md).
