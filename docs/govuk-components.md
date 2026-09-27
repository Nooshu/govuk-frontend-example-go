# GOV.UK components — architecture

## Source of truth

[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/) **Nunjucks** macros and fixtures (Node package `govuk-frontend`) are the upstream behaviour reference. This **Go** line re-implements that HTML contract in `internal/govuk` so product pages never hand-write component markup. See [tech-stack.md](tech-stack.md) and [architecture.md](architecture.md).

## Architecture (this line)

```text
Page / pattern (html/template + pages.Renderer)
  → render.Renderer (govukrender in production)
    → govuk.Params (ordered options)
      → govuk.Render → exact HTML string
        → Frontend CSS/JS in the page shell
```

Supporting pieces:

- **Shared HTML helpers** — Nunjucks-compatible escape + attribute serialisation (`internal/govuk`, `internal/htmlutil`).
- **Fixture loader** — cached `fixtures.json` for previews and the parity suite.
- **Ordered Params** — fixture `options` decoded with `jsontext` so key order and number spelling match.
- **Layout chrome** — shared skip link / header / footer / service nav / pattern back link.
- **Optional Nunjucks suite (Node)** — proves stored fixtures still match Frontend macros (freshness only).

See [layout-chrome.md](layout-chrome.md), [creating-components.md](creating-components.md), [testing-components.md](testing-components.md).

## Components vs patterns

|                | Components                        | Patterns               |
| -------------- | --------------------------------- | ---------------------- |
| Design System  | `/components/`                    | `/patterns/` and Pages |
| Implementation | `internal/govuk` ports + fixtures | Composed pages         |
| Parity suite   | Required (Go vs every fixture)    | Not applicable         |

## Expected component set

Ship Go ports for Design System components that Frontend provides fixtures for, including (non-exhaustive): accordion, back link, breadcrumbs, button, character count, checkboxes, cookie banner, date input, details, error message, error summary, exit this page, fieldset, file upload, generic header, footer, header, inset text, notification banner, pagination, panel, password input, phase banner, radios, select, service navigation, skip link, summary list, table, tabs, tag, task list, text input, textarea, warning text.

Per-component deep dives: add `docs/govuk-<kebab-name>.md` as each ships. Until then use the [Design System component pages](https://design-system.service.gov.uk/components/).
