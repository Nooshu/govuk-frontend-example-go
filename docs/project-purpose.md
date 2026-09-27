# Project purpose

This repository is the **Go** specialised line of the **GDS-compliant** government frontend template.

Language-agnostic sibling: [Nooshu/govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example). Keep shared playbooks in sync via [syncing-from-template.md](syncing-from-template.md).

## Intent

1. Meet [Service Standard](https://www.gov.uk/service-manual/service-standard) and [Technology Code of Practice](https://www.gov.uk/guidance/the-technology-code-of-practice) expectations for common components and accessibility — as far as the UI layer can.
2. Use **Go** for application logic and HTML generation (`net/http`, `html/template`, native component ports).
3. Use **[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)** (latest pinned release) as the **only** frontend component library.
4. Do **not** introduce SPA **frontend frameworks** for GOV.UK UI.
5. Track Frontend macros / `template.njk` in Go and prove **100% HTML parity** against official fixtures for every fixture `html`.
6. Keep **Node** as tooling only: pin `govuk-frontend`, compile Sass, run shared baseline/docs tests — never render HTML at request time.

## Why Go

See [why-go.md](why-go.md) for the advantages and trade-offs of this specialised line. Package layout: [architecture.md](architecture.md). Conventions: [go-conventions.md](go-conventions.md).

## Priorities

1. Frontend web performance
2. Frontend security
3. Reduced maintenance
4. Accessibility
5. Inclusive design

See [priorities.md](priorities.md) and [tech-stack.md](tech-stack.md).

## Agent skill

[`.cursor/skills/gds-compliant-frontend/SKILL.md`](../.cursor/skills/gds-compliant-frontend/SKILL.md)
