# Guidance sources

When building or reviewing work from this template, **search and follow** official guidance from these URLs before inventing local rules. Prefer upstream pages over paraphrases in this repo.

## Primary sources (always)

| Source                                | URL                                                                                                        | Use for                                                                                                                                                                                              |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Technology Code of Practice           | https://www.gov.uk/guidance/the-technology-code-of-practice                                                | Spend controls–aligned tech criteria: user needs, accessibility, open source, open standards, cloud first, security, privacy, reuse, integration, data, purchasing, sustainability, Service Standard |
| Assisted digital support              | https://www.gov.uk/service-manual/helping-people-to-use-your-service/assisted-digital-support-introduction | Users who need help to use online services; research; what support must cover                                                                                                                        |
| Service Standard                      | https://www.gov.uk/service-manual/service-standard                                                         | The 14 points services must meet                                                                                                                                                                     |
| Service Manual                        | https://www.gov.uk/service-manual                                                                          | How to apply the Standard (accessibility, agile, design, tech, research, assessments)                                                                                                                |
| GOV.UK Design System                  | https://design-system.service.gov.uk/                                                                      | Styles, components, patterns, accessibility strategy, roadmap                                                                                                                                        |
| GOV.UK Frontend                       | https://frontend.design-system.service.gov.uk/                                                             | Install (npm/Node), Nunjucks, CSS/JS, fixtures / testing HTML, browser support, APIs                                                                                                                 |
| govuk-frontend latest release         | https://github.com/alphagov/govuk-frontend/releases/latest                                                 | **Mandatory** read before every Frontend upgrade                                                                                                                                                     |
| Frontend performance (Service Manual) | https://www.gov.uk/service-manual/technology/how-to-test-frontend-performance                              | How to benchmark and iterate on frontend speed                                                                                                                                                       |
| Frontend performance (GDS Way)        | https://gds-way.digital.cabinet-office.gov.uk/standards/optimise-frontend-perf.html                        | Priority order: asset placement, compression, fingerprinting, unused code                                                                                                                            |
| OWASP HTTP Headers                    | https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html                               | Response headers for XSS, clickjacking, MIME sniffing, HSTS, isolation                                                                                                                               |
| OWASP Content Security Policy         | https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html                    | CSP directives; prefer hashes or nonces over `unsafe-inline`                                                                                                                                         |

## Technology Code of Practice — points (titles)

From the [official TCoP page](https://www.gov.uk/guidance/the-technology-code-of-practice) (consider all points for spend controls):

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

Local short note: [technology-code-of-practice.md](technology-code-of-practice.md).

## Service Standard — points (titles)

From the [official Service Standard](https://www.gov.uk/service-manual/service-standard):

1. Understand users and their needs
2. Solve a whole problem for users
3. Provide a joined up experience across all channels
4. Make the service simple to use
5. Make sure everyone can use the service
6. Have a multidisciplinary team
7. Use agile ways of working
8. Iterate and improve frequently
9. Create a secure service which protects users’ privacy
10. Define what success looks like and publish performance data
11. Choose the right tools and technology
12. Make new source code open
13. Use and contribute to open standards, common components and patterns
14. Operate a reliable service

Local short note: [service-standard.md](service-standard.md). UI checklist: [service-assessment-readiness.md](service-assessment-readiness.md).

## Assisted digital (summary)

From [Assisted digital support: an introduction](https://www.gov.uk/service-manual/helping-people-to-use-your-service/assisted-digital-support-introduction):

- Everyone who needs the service must be able to use it; some users need help with the online part (trust, confidence, access, skills, motivation).
- Support may be phone, in person, or sometimes webchat.
- Research assisted digital needs; do not assume none exist without evidence.
- This template’s UI work does not replace assisted digital planning.

## Design System + Frontend

- Components and patterns: https://design-system.service.gov.uk/
- Technical install, Nunjucks, fixture HTML testing, browsers: https://frontend.design-system.service.gov.uk/
- **Latest Frontend release (always before upgrading):** https://github.com/alphagov/govuk-frontend/releases/latest
- Prefer Nunjucks macros over copy-pasted HTML; extensive fixture parity for backend output
- Local playbooks: [creating-components.md](creating-components.md), [testing-components.md](testing-components.md), [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md)

## How agents should use this list

1. Open or search the relevant URL for the question.
2. Apply the guidance to the **backend + GOV.UK Frontend** shape of this template (no competing UI frameworks).
3. Record wrapper-specific decisions in [tech-stack.md](tech-stack.md).
4. Keep [`AGENTS.md`](../AGENTS.md) slim — put lasting detail under `/docs`.
