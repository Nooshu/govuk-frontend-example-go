# Creating a new GOV.UK pattern

Use this when the user asks for a Design System **pattern** (under `/patterns/` or **Pages** on the Design System site) — e.g. Addresses, Names, Check answers, Confirmation pages, Page not found.

Do **not** create a new low-level component unit or fixture-parity suite for patterns.

**Stack note:** Compose pages using the chosen language’s idiomatic routing and templating ([tech-stack.md](tech-stack.md)).

## What patterns are here

- **Composed pages / journeys** built only from **shipped components** and the [page template](https://design-system.service.gov.uk/styles/page-template/).
- Follow Design System “when to use / when not to use” and example structures.
- No official Frontend `fixtures.json` contract for whole patterns — correctness is guidance + component parity + accessibility rules.

## Goals

1. Reuse common components (Service Standard point 13).
2. One thing per page where the pattern expects it.
3. Back link **or** breadcrumbs — not both.
4. Forms: `novalidate`, Error summary + field errors, retained values.
5. Call out out-of-scope widgets with inset text rather than inventing unofficial chrome.
6. Link to the official pattern guidance from the demo page.

## Work order

1. Read the Design System pattern page end-to-end.
2. Confirm all required building blocks already ship as components; if not, implement the **component** first ([creating-components.md](creating-components.md)) or document the gap.
3. Add a composed page under the patterns area (routing idiomatic for the stack).
4. Use shared layout chrome ([layout-chrome.md](layout-chrome.md)); put back link in before-content via shared pattern chrome when demos need “Back to patterns”.
5. Apply [page-shell.md](page-shell.md) and [accessibility.md](accessibility.md).
6. Register the pattern in the patterns index/nav.
7. Document the demo in [govuk-components.md](govuk-components.md) (patterns section) or a dedicated `docs/govuk-pattern-<kebab>.md` if complex.
8. Visual check via preview server.

## Pages patterns (extra)

For Design System **Pages** patterns (e.g. page not found, problem with the service, service unavailable, confirmation):

- Always use the page template accessibility rules.
- Always use shipped components for panels, buttons, etc.
- Prefer these for downtime / reliability UX (Service Standard point 14) — see [service-assessment-readiness.md](service-assessment-readiness.md).

## Do / don’t

**Do:** compose the library API only; match guidance structure; keep demos boring and correct; follow the language’s page/routing conventions.

**Don’t:** invent pattern-only `govuk-`\* CSS; add fixture-parity HTML suites for patterns; hand-write header/footer/skip link; combine breadcrumbs and back link.
