# GOV.UK components — architecture

## Source of truth

[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/) **Nunjucks** macros and fixtures (Node package `govuk-frontend`). This repo re-implements the HTML contract in the **chosen wrapper language’s** idiomatic component / templating model so product pages never hand-write component markup. See [tech-stack.md](tech-stack.md).

## Architecture (intended)

```text
Page / pattern
  → library API (one entry per component)
    → options model (aligned with Nunjucks macro options)
      → renderer → exact HTML string
        → Frontend CSS/JS in the page shell
```

Supporting pieces (names/paths idiomatic for the wrapper language):

- **Shared HTML helpers** — Nunjucks-compatible escape + attribute serialization.
- **Fixture loader** — cached `fixtures.json` for Previews / Fixtures.
- **Options mapper** — fixture `options` → model (including edge cases).
- **Layout chrome** — shared skip link / header / footer / service nav / pattern back link.
- **Nunjucks suite (Node)** — proves stored fixtures still match Frontend macros.

See [layout-chrome.md](layout-chrome.md), [creating-components.md](creating-components.md), [testing-components.md](testing-components.md).

## Components vs patterns

|                | Components                  | Patterns               |
| -------------- | --------------------------- | ---------------------- |
| Design System  | `/components/`              | `/patterns/` and Pages |
| Implementation | Library wrappers + fixtures | Composed pages         |
| Parity suite   | Required                    | Not applicable         |

## Expected component set

Ship wrappers for Design System components that Frontend provides fixtures for, including (non-exhaustive): accordion, back link, breadcrumbs, button, character count, checkboxes, cookie banner, date input, details, error message, error summary, exit this page, fieldset, file upload, generic header, footer, header, inset text, notification banner, pagination, panel, password input, phase banner, radios, select, service navigation, skip link, summary list, table, tabs, tag, task list, text input, textarea, warning text.

Per-component deep dives: add `docs/govuk-<kebab-name>.md` as each ships. Until then use the [Design System component pages](https://design-system.service.gov.uk/components/).

## Previews

Local preview (when `NODE_ENV` is not `production`):

1. Service start page (`/` or `/cy`) links to **Preview GOV.UK components** → `/components`.
2. `/components` is the **component preview homepage**: one link per shipped component — **no live demos on that index**.
3. `/components/:name` is that component’s own page. It lists fixtures and renders the selection via **Go** `internal/govuk` `Render` (same API the service and parity suite use), with a parity banner vs official fixture `html`.
4. `/components/:name/fixture` returns the raw fixture HTML fragment for automation.

Footer and About also link to the catalogue when demos are enabled.

## Do not

- Custom CSS for Design System appearance.
- Unofficial step-nav or invented `govuk-*` chrome.
- Embed demos on the index.
- Rebuild skip link / header / footer models per request when shared chrome helpers exist.
- Force another ecosystem’s folder layout once a language is chosen — follow that language’s best practices.
