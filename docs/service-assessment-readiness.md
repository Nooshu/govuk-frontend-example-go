# Service assessment readiness

Use this when building or reviewing pages, components, or patterns so work stays aligned with a **GOV.UK Service Standard** assessment — especially points **4** (simple to use), **5** (everyone can use the service), **11** (tools), **13** (common components and patterns), and **14** (reliable).

**Using the Design System or this repo does not automatically make a service accessible or assessment-ready.**

Still needed outside this repo: user research (including disabled users), accessibility audit, accessibility statement by public beta, assisted digital planning, privacy/security evidence, performance KPIs, open source policy, and operational evidence.

See:

- [Design System accessibility](https://design-system.service.gov.uk/accessibility/)
- [Making your service accessible](https://www.gov.uk/service-manual/helping-people-to-use-your-service/making-your-service-accessible-an-introduction)
- [service-standard.md](service-standard.md)
- [authoritative-references.md](authoritative-references.md)

## Non-negotiables for assessment-shaped UI

1. **Reuse common components and patterns** ([point 13](https://www.gov.uk/service-manual/service-standard/point-13-use-common-standards-components-patterns)) — library components and pattern demos; never invent one-off `govuk-`\* chrome.
2. **Look and behave like GOV.UK** ([point 4](https://www.gov.uk/service-manual/service-standard/point-4-make-the-service-simple-to-use)) — page template, pinned Frontend CSS/JS, sentence case, one primary action, one thing per page where possible.
3. **WCAG 2.2 AA baseline** ([point 5](https://www.gov.uk/service-manual/service-standard/point-5-make-sure-everyone-can-use-your-service)) — skip link, one `h1`, visible focus, keyboard-only paths, Error summary + field errors, `novalidate`.
4. **Progressive enhancement** — body class snippet; `initAll()`; core tasks without Frontend JS ([browser support](https://frontend.design-system.service.gov.uk/browser-support/)).
5. **Exact Frontend HTML for components** — fixture parity ([testing your HTML](https://frontend.design-system.service.gov.uk/testing-your-html/)); sanitise any `html` options.
6. **Reliable downtime UX** ([point 14](https://www.gov.uk/service-manual/service-standard)) — Page not found / Problem with the service / Service unavailable patterns; cookies page + cookie banner by public beta.
7. **No ad-hoc custom CSS** — use the Sass pipeline and `govuk-overrides.scss` cascade ([styles.md](styles.md)); never `!important` in service CSS.

## Quick page review

Mirror the checklist in `[AGENTS.md](../AGENTS.md)`: shell, components only, back link xor breadcrumbs, form errors, focus styles, trusted HTML, pattern guidance.

## Outside this repo

User research evidence, assisted digital, privacy reviews, penetration testing, performance KPIs from production traffic, open source policy, published accessibility statement, and an accessibility audit before public beta.

The response-header and cache baseline in [frontend-security.md](frontend-security.md) and [frontend-performance.md](frontend-performance.md) is the starting point for points 9 and 14. It does not replace an assessment.
