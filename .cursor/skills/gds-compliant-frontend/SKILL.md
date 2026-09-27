---
name: gds-compliant-frontend
description: >-
  Builds GDS-compliant government frontends from this base template using
  standardised backend languages (TypeScript, Go, Python, etc.) with GOV.UK
  Frontend macros (prefer Nunjucks) and fixture HTML parity, no SPA/frontend
  frameworks. Use when scaffolding services, choosing stack, applying Service
  Standard or Technology Code of Practice guidance, implementing GOV.UK
  components/patterns, upgrading govuk-frontend, or verifying assessment-shaped UI.
---

# GDS-compliant frontend template

## What this project is

A **base template** for **GDS-compliant** frontends that:

- Use **standardised backend technologies** (e.g. TypeScript/Node, Go, Python) for the server and HTML generation
- Use **[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)** (latest pinned version) as the **only** frontend component library
- Prefer **Nunjucks macros** for component HTML — do **not** copy-paste HTML from each Frontend release as the long-term approach
- Wire official **test fixtures** for extensive **100% HTML parity** testing of backend-generated markup
- Do **not** use frontend frameworks (React, Vue, Angular, Svelte, Next.js client apps, etc.) for UI

Priorities (in order): frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design. See [`docs/priorities.md`](../../../docs/priorities.md).

Detail for humans: [`docs/project-purpose.md`](../../../docs/project-purpose.md), [`docs/onboarding.md`](../../../docs/onboarding.md), [`CONTRIBUTING.md`](../../../CONTRIBUTING.md). Dual-audience map: [`docs/documentation-structure.md`](../../../docs/documentation-structure.md). Playbooks: [`AGENTS.md`](../../../AGENTS.md).

## Non-negotiable stack shape

| Layer               | Choice                                                                                                               |
| ------------------- | -------------------------------------------------------------------------------------------------------------------- |
| UI                  | GOV.UK Frontend only (`govuk-*`, official JS via `initAll()`)                                                        |
| HTML generation     | Prefer **Nunjucks macros** from `govuk-frontend`; otherwise thin wrappers that stay fixture-parity with those macros |
| Frontend frameworks | **Forbidden** for UI                                                                                                 |
| Parity              | Official `fixtures.json` + ordinal HTML equality of **backend** output vs **every** fixture `html`                   |
| Upstream            | Node package + Nunjucks / `template.njk` / fixtures                                                                  |

## Authoritative guidance (search these first)

1. [Technology Code of Practice](https://www.gov.uk/guidance/the-technology-code-of-practice)
2. [Assisted digital support: an introduction](https://www.gov.uk/service-manual/helping-people-to-use-your-service/assisted-digital-support-introduction)
3. [Service Standard](https://www.gov.uk/service-manual/service-standard)
4. [Service Manual](https://www.gov.uk/service-manual)
5. [GOV.UK Design System](https://design-system.service.gov.uk/)
6. [GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)

Local index: [`docs/guidance-sources.md`](../../../docs/guidance-sources.md).

## Upgrading Frontend

**Always** read https://github.com/alphagov/govuk-frontend/releases/latest before changing the pin, then follow [`docs/upgrading-govuk-frontend.md`](../../../docs/upgrading-govuk-frontend.md). Refresh fixtures from the same version; fix renderers/macros usage — never edit fixture `html`.

## Test coverage and HTML parity

- **Code:** 100% functions, branches, statements (CI fails below).
- **HTML (primary):** TypeScript `renderComponent` output must match official fixture `html` byte-for-byte for **every** fixture on every shipped component. See [`docs/testing-components.md`](../../../docs/testing-components.md).
- **HTML (secondary):** A Nunjucks-only suite (if present) proves stored fixtures still match Frontend macros — freshness only; it does **not** replace backend vs fixture parity.
- Do not weaken either gate to satisfy the other; do not treat Nunjucks-only green as done.

## Workflow reminders

1. Confirm wrapper language in [`docs/tech-stack.md`](../../../docs/tech-stack.md) (prefer Nunjucks when viable).
2. Never hand-paste `govuk-*` component HTML; use macros / library API.
3. Upgrade only after reviewing the [latest release](https://github.com/alphagov/govuk-frontend/releases/latest).
4. New components: [`docs/creating-components.md`](../../../docs/creating-components.md). Patterns: [`docs/creating-patterns.md`](../../../docs/creating-patterns.md).
5. HTTP responses use [`baseline/`](../../../baseline/) — call `applyResponseHeaders` and `buildSetCookie`. Compress with Brotli (`br`); Gzip is only the fallback when the client does not advertise `br`. Playbooks: [`docs/frontend-performance.md`](../../../docs/frontend-performance.md), [`docs/frontend-security.md`](../../../docs/frontend-security.md). Sync `baseline/` from the template; do not fork a weaker policy.
6. Compile CSS via Sass (`styles/application.scss` → Frontend `@use` → `govuk-overrides.scss` last). Never use `!important` in service CSS. Playbook: [`docs/styles.md`](../../../docs/styles.md).
7. Document every change for **humans and agents** in the same change set ([`docs/documentation-structure.md`](../../../docs/documentation-structure.md)). Update `/docs`, and `AGENTS.md` / skill / rules when contracts change.
8. Follow the **latest** best practices for TypeScript / Node in [`docs/tech-stack.md`](../../../docs/tech-stack.md) (and current Node/ESM for shared tooling). Do not fossilise outdated patterns.
9. When a coherent piece of work is finished, split it into focused commits with comprehensive messages — do not leave a large mixed working tree.
