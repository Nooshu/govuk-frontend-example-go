# Project purpose

This repository is the **TypeScript** specialised line of the **GDS-compliant** government frontend template.

Language-agnostic sibling: [Nooshu/govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example). Keep shared playbooks in sync via [syncing-from-template.md](syncing-from-template.md).

## Intent

1. Meet [Service Standard](https://www.gov.uk/service-manual/service-standard) and [Technology Code of Practice](https://www.gov.uk/guidance/the-technology-code-of-practice) expectations for common components and accessibility — as far as the UI layer can.
2. Use **TypeScript on Node** for application logic and HTML generation (ESM).
3. Use **[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)** (latest pinned release) as the **only** frontend component library.
4. Do **not** introduce SPA **frontend frameworks** for GOV.UK UI.
5. Prefer **Nunjucks macros** and official **fixtures** for **100% HTML parity** testing.

## Priorities

1. Frontend web performance
2. Frontend security
3. Reduced maintenance
4. Accessibility
5. Inclusive design

See [priorities.md](priorities.md) and [tech-stack.md](tech-stack.md).

## Agent skill

[`.cursor/skills/gds-compliant-frontend/SKILL.md`](../.cursor/skills/gds-compliant-frontend/SKILL.md)
