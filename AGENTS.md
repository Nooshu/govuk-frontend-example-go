<!-- ============================================================
  GOV.UK Design System — Agent instructions
  ============================================================
  Base: Go line of GDS-compliant frontends (Go + GOV.UK Frontend)
  Sync shared docs from Nooshu/govuk-frontend-example (see docs/syncing-from-template.md)
  Detail lives in /docs and .cursor/skills/gds-compliant-frontend
  ============================================================ -->

♛ GOV.UK

# GOV.UK Frontend example (Go)

**Go** specialised line: **Go** generates HTML natively; **[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)** (latest pinned version) is the **only** UI component library. **No frontend frameworks** (React, Vue, Angular, Svelte, etc.) for UI.

Component HTML is produced by **native Go renderers** in `internal/govuk` that track Frontend macros / `template.njk`. Do **not** shell out to Node/Nunjucks for request-time HTML. Official **test fixtures** from each Frontend release are the contract: this line’s **Go-rendered HTML** must match every fixture `html` byte-for-byte. A Nunjucks-only check is not enough.

Node is used only to pin `govuk-frontend`, compile Sass, and run shared baseline/docs tests.

Language-agnostic template (shared playbooks): https://github.com/Nooshu/govuk-frontend-example — sync with `npm run sync:template` ([docs/syncing-from-template.md](docs/syncing-from-template.md)).

**LIVE guidance** — [Design System feedback](https://design-system.service.gov.uk/community/feedback/).

Skill: [`.cursor/skills/gds-compliant-frontend/SKILL.md`](.cursor/skills/gds-compliant-frontend/SKILL.md). Purpose: [`docs/project-purpose.md`](docs/project-purpose.md). Stack: [`docs/tech-stack.md`](docs/tech-stack.md).

## Priorities (in order)

1. Frontend web performance
2. Frontend security
3. Reduced maintenance
4. Accessibility
5. Inclusive design

Details: [`docs/priorities.md`](docs/priorities.md).

## Start here

| Audience                | Doc                                                                                                                     |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| **Human developers**    | [`docs/onboarding.md`](docs/onboarding.md), [`CONTRIBUTING.md`](CONTRIBUTING.md)                                        |
| **AI agents (this file)** | Keep reading; skill: [`.cursor/skills/gds-compliant-frontend/SKILL.md`](.cursor/skills/gds-compliant-frontend/SKILL.md) |
| Dual-audience docs map  | [`docs/documentation-structure.md`](docs/documentation-structure.md), [`docs/README.md`](docs/README.md)                |
| Project purpose         | [`docs/project-purpose.md`](docs/project-purpose.md)                                                                    |
| Official guidance URLs  | [`docs/guidance-sources.md`](docs/guidance-sources.md)                                                                  |
| Stack / language (Go)   | [`docs/tech-stack.md`](docs/tech-stack.md)                                                                              |
| Sync from agnostic template | [`docs/syncing-from-template.md`](docs/syncing-from-template.md)                                                    |

**Language rule:** This line is **Go**. **Every** feature and code change must follow Go’s **latest** best practices (modules, `internal/`, tests, `go vet`, packaging) as recorded in [`docs/tech-stack.md`](docs/tech-stack.md) — without weakening the non-negotiables below. Prefer the standard library (`net/http`, `html/template`, `compress/gzip`, `crypto/*`, `log/slog`) and only well-known deps where stdlib lacks a feature (e.g. `andybalholm/brotli`). Generate HTML **natively in Go**; do not call Node to render.

**GOV.UK Frontend’s own stack:** Frontend ships as a **Node** package with **Nunjucks** macros, official `fixtures.json`, and `template.njk` sources. Use Node for install, fixtures, Sass, and optional freshness checks. Refer to Nunjucks for macro options and escape behaviour when porting.

**Guidance rule:** Prefer searching the URLs in [`docs/guidance-sources.md`](docs/guidance-sources.md) over inventing local policy.

**Documentation rule:** Every prompt, feature, and code change for this template **and** projects built from it must leave **comprehensive dual-audience documentation** (humans + AI agents) in the right place. Detail: [`docs/documentation-structure.md`](docs/documentation-structure.md).

## Non-negotiables

1. **GOV.UK Frontend macros are the HTML source of truth** — render via native Go that tracks those macros. Do **not** copy-paste component HTML from release notes or the Design System site as the long-term approach, and do **not** shell out to Node just to render HTML.
2. **Backend HTML must match every official fixture** — for each shipped component, **Go `Render` output** is compared byte-for-byte to the `html` in that release’s `fixtures.json`, for **every** fixture (including hidden ones). That is the primary parity gate. A Nunjucks-only check proves fixtures are fresh; it does **not** replace backend vs fixture comparison. No normalisation; never edit fixture `html` to pass tests. See [`docs/testing-components.md`](docs/testing-components.md).
3. **No frontend UI frameworks** — no React/Vue/Angular/Svelte (or similar) for GOV.UK UI; backend + GOV.UK Frontend only.
4. **No ad-hoc custom CSS** — ship styles through the Sass pipeline in [`styles/`](styles/) (`application.scss` → GOV.UK Frontend `@use` → [`govuk-overrides.scss`](styles/govuk-overrides.scss) last). Prefer component options and Design System patterns; do not paste or serve Frontend’s prebuilt `govuk-frontend.min.css` as the long-term source. See [`docs/styles.md`](docs/styles.md).
5. **No `!important` in service CSS** — overrides must win with cascade order and specificity only. This applies to every project using this template. Frontend’s own `govuk-!-…` utilities are upstream; do not copy that pattern into service styles.
6. **Components via library API** — never hand-paste component `govuk-*` markup into pages; call `internal/govuk` / the page `Renderer`.
7. **Patterns compose components** — Design System patterns are pages/journeys, not new low-level components, and have no fixture-parity suites.
8. **WCAG 2.2 AA baseline** — skip link, one `h1`, visible focus (never override yellow focus), keyboard paths, Error summary + field errors, `novalidate`.
9. **Progressive enhancement** — core tasks work without Frontend JS; keep `js-enabled` / `govuk-frontend-supported` and `initAll()`.
10. **Do not ship unreleased GOV.UK chrome** — wait for Frontend release + fixtures. See [`docs/govuk-frontend-roadmap.md`](docs/govuk-frontend-roadmap.md).
11. **100% code coverage** — functions, branches, and statements at **100%** for application/library code under test; CI must fail below that. Do not weaken fixture HTML equality to chase coverage. See [`docs/testing-components.md`](docs/testing-components.md).
12. **Always review the latest release notes** before upgrading — https://github.com/alphagov/govuk-frontend/releases/latest — then follow [`docs/upgrading-govuk-frontend.md`](docs/upgrading-govuk-frontend.md).
13. **Performance and security baseline** — every response uses [`baseline/`](baseline/) policy via `internal/baseline` (cache kind, OWASP headers, CSP hash for the `js-enabled` snippet). Sync that directory from the template; do not invent a weaker set. Compress with Brotli first (`andybalholm/brotli`); Gzip is the fallback. See [`docs/frontend-performance.md`](docs/frontend-performance.md) and [`docs/frontend-security.md`](docs/frontend-security.md).
14. **Split finished work into focused commits** — once a coherent piece of code or docs is complete, create **specific** commits with **comprehensive** messages (why, contract impact, how to verify). Do not leave a large mixed working tree; do not squash unrelated concerns into one commit. This applies to agents and humans using this template.
15. **Document every change for humans and agents** — no feature, prompt-driven change, or behaviour lands without dual-audience docs updated in the right place (`/docs` detail, `AGENTS.md` / skill / rules links when contracts change, onboarding or CONTRIBUTING when workflow changes). Aim for easier onboarding and maintenance. See [`docs/documentation-structure.md`](docs/documentation-structure.md).
16. **Follow the latest language best practices** — all new and changed code must match Go’s current best practices in [`docs/tech-stack.md`](docs/tech-stack.md) (not outdated tutorials). Shared Node tooling (Sass, baseline, fixtures) follows current Node/ESM practice. Never weaken Frontend, parity, security, or performance non-negotiables to chase a fad.

Using this repo does **not** make a service assessment-ready. See [`docs/service-assessment-readiness.md`](docs/service-assessment-readiness.md).

## Agent playbooks

| Task                                | Doc                                                                    |
| ----------------------------------- | ---------------------------------------------------------------------- |
| Sync from agnostic template         | [`docs/syncing-from-template.md`](docs/syncing-from-template.md)       |
| Upgrade GOV.UK Frontend             | [`docs/upgrading-govuk-frontend.md`](docs/upgrading-govuk-frontend.md) |
| Add a component                     | [`docs/creating-components.md`](docs/creating-components.md)           |
| Add a pattern                       | [`docs/creating-patterns.md`](docs/creating-patterns.md)               |
| Layout / chrome                     | [`docs/layout-chrome.md`](docs/layout-chrome.md)                       |
| Fixture / parity testing            | [`docs/testing-components.md`](docs/testing-components.md)             |
| Example service                     | [`docs/example-service.md`](docs/example-service.md)                   |
| Page shell                          | [`docs/page-shell.md`](docs/page-shell.md)                             |
| Frontend performance                | [`docs/frontend-performance.md`](docs/frontend-performance.md)         |
| Frontend security                   | [`docs/frontend-security.md`](docs/frontend-security.md)               |
| Accessibility                       | [`docs/accessibility.md`](docs/accessibility.md)                       |
| Content & forms                     | [`docs/content-and-forms.md`](docs/content-and-forms.md)               |
| Design tokens (colour, type, space) | [`docs/design-tokens.md`](docs/design-tokens.md)                       |
| Styles / Sass cascade               | [`docs/styles.md`](docs/styles.md)                                     |
| Dual-audience documentation         | [`docs/documentation-structure.md`](docs/documentation-structure.md)   |
| Guidance sources                    | [`docs/guidance-sources.md`](docs/guidance-sources.md)                 |
| Authoritative links                 | [`docs/authoritative-references.md`](docs/authoritative-references.md) |

## Quick page review

Before finishing a page change:

- [ ] Page template shell / before-content / single `h1` / title
- [ ] Library API only for GOV.UK UI blocks (not pasted HTML)
- [ ] Back link **or** breadcrumbs — not both
- [ ] Forms: `novalidate`, Error summary + messages, values retained
- [ ] Focus styles untouched; no `outline: none`
- [ ] Trusted/sanitised HTML only; prefer plain text options
- [ ] HTML responses use the baseline security headers; assets use the matching cache kind
- [ ] CSS from the Sass pipeline in `<head>` (not prebuilt `govuk-frontend.min.css`); Frontend JS is an external `type="module"`; the `js-enabled` snippet matches the pinned CSP hash
- [ ] No `!important` in service styles; overrides only via `govuk-overrides.scss` specificity
- [ ] Pattern guidance followed; out-of-scope widgets called out with inset text
- [ ] Coverage remains 100% functions / branches / statements for touched library code
- [ ] Backend parity suite green: Go `Render` HTML ≡ every fixture `html` (not only Nunjucks ≡ fixtures)
- [ ] Fixture parity still green for any touched components
- [ ] Dual-audience docs updated (humans in `/docs` or CONTRIBUTING; agents via `AGENTS.md` / skill / playbook links if contracts changed)
- [ ] Code follows Go’s latest best practices ([`docs/tech-stack.md`](docs/tech-stack.md))

## Watching upstream

**Before every Frontend upgrade:** read https://github.com/alphagov/govuk-frontend/releases/latest (currently documents releases such as [v6.5.1](https://github.com/alphagov/govuk-frontend/releases/latest)).

Also: [roadmap](https://design-system.service.gov.uk/community/roadmap/), [upcoming](https://design-system.service.gov.uk/community/upcoming-components-patterns/), [all releases](https://github.com/alphagov/govuk-frontend/releases). Local: [`docs/govuk-frontend-roadmap.md`](docs/govuk-frontend-roadmap.md).

Upgrade only via [`docs/upgrading-govuk-frontend.md`](docs/upgrading-govuk-frontend.md). Do not invent dark mode, AI patterns, or unofficial chrome.
