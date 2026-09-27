# Technology Code of Practice (local summary)

Canonical: https://www.gov.uk/guidance/the-technology-code-of-practice

Full indexed guidance for this template: [guidance-sources.md](guidance-sources.md).

The Technology Code of Practice (TCoP) is used for Cabinet Office spend controls and the Local Digital Declaration. Consider each point; align with mandatory points and follow as many others as practical.

## Points (titles)

1. Define user needs
2. Make things accessible and inclusive
3. Be open and use open source
4. Make use of open standards
5. Use cloud first
6. Make things secure
7. Make privacy integral
8. Share, reuse and collaborate
9. Integrate and adapt technology
10. Make better use of data
11. Define your purchasing strategy
12. Make your technology sustainable
13. Meet the Service Standard

## Relevance to this template

- Prefer open standards and common government components ([GOV.UK Frontend](https://frontend.design-system.service.gov.uk/) / [Design System](https://design-system.service.gov.uk/)).
- Be open / use open source where appropriate (point 3); make new source open aligns with Service Standard point 12.
- Share, reuse, collaborate (point 8) — do not invent parallel UI systems or frontend frameworks for GOV.UK chrome.
- Make things secure (point 6) — apply the shared response-header baseline ([frontend-security.md](frontend-security.md)), not a one-off header set per service.
- Make privacy integral (point 7) — `Referrer-Policy`, `Permissions-Policy`, and `no-store` for personal HTML are part of that baseline.
- Meet the Service Standard (point 13) — see [service-standard.md](service-standard.md) and [service-assessment-readiness.md](service-assessment-readiness.md).
- Plan for maintenance: pinned Frontend version, fixture parity, documented upgrade path ([upgrading-govuk-frontend.md](upgrading-govuk-frontend.md)).

When the implementation language is chosen, record it in [tech-stack.md](tech-stack.md) and follow that language’s current best practices — without abandoning open standards or GOV.UK Frontend.
