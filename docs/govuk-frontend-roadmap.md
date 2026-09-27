# GOV.UK Frontend roadmap (local summary)

Do **not** invent components or chrome that are still upstream. Prefer waiting for an official Frontend release and fixtures.

## Authoritative sources

- **Latest release (always read before upgrading):** https://github.com/alphagov/govuk-frontend/releases/latest
- [All releases](https://github.com/alphagov/govuk-frontend/releases)
- [Design System roadmap](https://design-system.service.gov.uk/community/roadmap/)
- [Upcoming components and patterns](https://design-system.service.gov.uk/community/upcoming-components-patterns/)
- [Design System cycle board](https://github.com/orgs/alphagov/projects/53)

## Watching upstream (do not invent)

Before adding a component that is not yet in this repo:

1. Read https://github.com/alphagov/govuk-frontend/releases/latest and confirm the component shipped with fixtures.
2. Check roadmap / upcoming (links above).
3. Follow [creating-components.md](creating-components.md) only after fixtures exist.
4. Upgrade the pin first if needed — [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md). Prefer Nunjucks macros over pasting HTML from the release.

## Snapshot (refresh on each upgrade)

| Status                               | Notes                                                                                                                                                                                                               |
| ------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Latest upstream (as of Sep 2026)** | [v6.5.1](https://github.com/alphagov/govuk-frontend/releases/latest) — includes Feedback and Language navigation as Trial components from 6.5.0; 6.5.1 fixes Service navigation + right-aligned Language navigation |
| **Already covered here**             | _TBD until first pin_                                                                                                                                                                                               |
| **Do not hand-build**                | Dark mode chrome, AI patterns, or unreleased components; do not copy-paste HTML instead of macros                                                                                                                   |

## After upgrade

Update the “Already covered here” row with the exact `govuk-frontend` version and call out newly shipped wrappers.
