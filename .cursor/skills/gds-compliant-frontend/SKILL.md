---
name: gds-compliant-frontend
description: >-
  Builds GDS-compliant government frontends on this Go line: native Go HTML
  that tracks GOV.UK Frontend macros and matches every official fixture,
  no SPA/frontend frameworks. Node is only for the govuk-frontend pin, Sass,
  and shared baseline tests. Use when scaffolding services, applying Service
  Standard or Technology Code of Practice guidance, implementing GOV.UK
  components/patterns, upgrading govuk-frontend, or verifying assessment-shaped UI.
---

# GDS-compliant frontend (Go line)

## What this project is

The **Go** line of the GDS-compliant frontend template:

- **Go** generates HTML and serves HTTP (`net/http`, `html/template`, `internal/govuk`)
- **[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)** (pinned) is the **only** UI component library
- Component HTML is a **native Go port** of Frontend macros / `template.njk`. Do **not** shell out to Node to render, and do **not** copy-paste HTML from each release
- Official **test fixtures** are the contract: Go output must match every fixture `html` byte-for-byte
- **No** frontend frameworks (React, Vue, Angular, Svelte, Next.js client apps, etc.) for UI
- **Node** installs `govuk-frontend`, compiles Sass, and runs shared baseline/docs tests

Priorities (in order): frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design. See [`docs/priorities.md`](../../../docs/priorities.md).

Detail for humans: [`docs/project-purpose.md`](../../../docs/project-purpose.md), [`docs/why-go.md`](../../../docs/why-go.md), [`docs/architecture.md`](../../../docs/architecture.md), [`docs/onboarding.md`](../../../docs/onboarding.md), [`CONTRIBUTING.md`](../../../CONTRIBUTING.md). Dual-audience map: [`docs/documentation-structure.md`](../../../docs/documentation-structure.md). Playbooks: [`AGENTS.md`](../../../AGENTS.md). Stack: [`docs/tech-stack.md`](../../../docs/tech-stack.md). Conventions: [`docs/go-conventions.md`](../../../docs/go-conventions.md).

## Non-negotiable stack shape

| Layer               | Choice                                                                                                 |
| ------------------- | ------------------------------------------------------------------------------------------------------ |
| UI                  | GOV.UK Frontend only (`govuk-*`, official JS via `initAll()`)                                          |
| HTML generation     | Native Go in `internal/govuk` and `internal/pages`; tracks macros; fixture-parity                      |
| Frontend frameworks | **Forbidden** for UI                                                                                   |
| Parity              | Official `fixtures.json` + ordinal HTML equality of **Go** output vs **every** fixture `html`          |
| Upstream            | Node package + Nunjucks / `template.njk` / fixtures (install, Sass, freshness — not request-time HTML) |
| HTTP                | `net/http` ServeMux, [`baseline/policy.json`](../../../baseline/policy.json) via `internal/baseline`   |

## Authoritative guidance (search these first)

1. [Technology Code of Practice](https://www.gov.uk/guidance/the-technology-code-of-practice)
2. [Assisted digital support: an introduction](https://www.gov.uk/service-manual/helping-people-to-use-your-service/assisted-digital-support-introduction)
3. [Service Standard](https://www.gov.uk/service-manual/service-standard)
4. [Service Manual](https://www.gov.uk/service-manual)
5. [GOV.UK Design System](https://design-system.service.gov.uk/)
6. [GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)

Local index: [`docs/guidance-sources.md`](../../../docs/guidance-sources.md).

## Upgrading Frontend

**Pipeline gate:** do not start or finish a Frontend (or any other) dependency bump while CI is red — follow [`../safe-dependency-updates/SKILL.md`](../safe-dependency-updates/SKILL.md).

**Always** read https://github.com/alphagov/govuk-frontend/releases/latest before changing the pin, then follow [`docs/upgrading-govuk-frontend.md`](../../../docs/upgrading-govuk-frontend.md). Refresh fixtures from the same version; fix the Go renderers — never edit fixture `html`.

## Test coverage and HTML parity

- **Code:** 100% functions, branches, statements for application packages (`go test ./...`). `cmd/server` is the process entry and is excluded.
- **HTML (primary):** `internal/govuk` `Render` must match official fixture `html` byte-for-byte for **every** fixture on every shipped component. See [`docs/testing-components.md`](../../../docs/testing-components.md).
- **HTML (secondary):** A Nunjucks-only suite, if someone runs one, only proves stored fixtures still match Frontend macros. It does **not** replace Go vs fixture parity.
- Do not weaken either gate to satisfy the other.

## Workflow reminders

1. This line is **Go**. Prefer the standard library. Record exceptions in [`docs/tech-stack.md`](../../../docs/tech-stack.md). Keep the npm tree minimal and lockfile-strict ([`docs/npm-security.md`](../../../docs/npm-security.md)); use `npm ci` / `npm run audit:npm`.
2. Never hand-paste `govuk-*` component HTML; call `internal/govuk` / the page renderer.
3. Upgrade only after reviewing the [latest release](https://github.com/alphagov/govuk-frontend/releases/latest), with CI green per [`safe-dependency-updates`](../safe-dependency-updates/SKILL.md).
4. New components: [`docs/creating-components.md`](../../../docs/creating-components.md). Patterns: [`docs/creating-patterns.md`](../../../docs/creating-patterns.md).
5. HTTP responses use [`baseline/`](../../../baseline/) through `internal/baseline`. Compress with Brotli (`br`); Gzip is only the fallback when the client does not advertise `br`. Playbooks: [`docs/frontend-performance.md`](../../../docs/frontend-performance.md), [`docs/frontend-security.md`](../../../docs/frontend-security.md).
6. Compile CSS via Sass (`styles/application.scss` → Frontend `@use` → `govuk-overrides.scss` last). Never use `!important` in service CSS. Playbook: [`docs/styles.md`](../../../docs/styles.md).
7. Document every change for **humans and agents** in the same change set ([`docs/documentation-structure.md`](../../../docs/documentation-structure.md)).
8. Follow current Go best practices (`go test`, `npm run lint:go`, `gofmt`, table-driven tests, package/exported comments; see [`docs/tech-stack.md`](../../../docs/tech-stack.md) and [`docs/go-conventions.md`](../../../docs/go-conventions.md)). Shared Node tooling stays ESM.
9. When a coherent piece of work is finished, split it into focused commits with comprehensive messages.
