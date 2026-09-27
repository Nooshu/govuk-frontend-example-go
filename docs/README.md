# Documentation

All lasting project detail lives here. Keep root [`AGENTS.md`](../AGENTS.md) slim and linked.

This folder is written for **two audiences**. Same facts; different entry points.

| Audience             | Start here                                                                                                                 | Style                                                   |
| -------------------- | -------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| **Human developers** | [onboarding.md](onboarding.md), [CONTRIBUTING.md](../CONTRIBUTING.md)                                                      | Narrative, repo map, how to run/verify, troubleshooting |
| **AI coding agents** | [`AGENTS.md`](../AGENTS.md), [`.cursor/skills/gds-compliant-frontend/`](../.cursor/skills/gds-compliant-frontend/SKILL.md) | Dense playbooks, checklists, non-negotiables            |

How we keep docs dual-purpose: [documentation-structure.md](documentation-structure.md).

## For human developers

| Doc                                                      | Purpose                                                      |
| -------------------------------------------------------- | ------------------------------------------------------------ |
| [project-purpose.md](project-purpose.md)                 | What this template is for                                    |
| [why-go.md](why-go.md)                                   | Why this specialised line uses Go; advantages and trade-offs |
| [architecture.md](architecture.md)                       | Packages, request lifecycle, HTML layers                     |
| [go-conventions.md](go-conventions.md)                   | Comments, layout, lint, testing norms for Go                 |
| [onboarding.md](onboarding.md)                           | Repo map, run modes, components vs patterns, troubleshooting |
| [priorities.md](priorities.md)                           | Ordered priorities                                           |
| [frontend-performance.md](frontend-performance.md)       | Caching, compression, asset placement, budgets               |
| [frontend-security.md](frontend-security.md)             | OWASP response headers, CSP, cookies                         |
| [tech-stack.md](tech-stack.md)                           | Go + Frontend (Nunjucks as reference); Node tooling pins     |
| [npm-security.md](npm-security.md)                       | Lockfile, ignore-scripts, audit, Dependabot for npm tooling  |
| [example-service.md](example-service.md)                 | Rod licence example, `npm start`, fixture previews           |
| [syncing-from-template.md](syncing-from-template.md)     | Pull shared docs/dotfiles from govuk-frontend-example        |
| [guidance-sources.md](guidance-sources.md)               | Official GDS / Service Manual / Frontend URLs                |
| [documentation-structure.md](documentation-structure.md) | Dual-audience docs + language practice rules                 |
| [CONTRIBUTING.md](../CONTRIBUTING.md)                    | How to contribute, local checks, PR expectations             |

## For AI agents

| Doc                                                                                           | Purpose                                                 |
| --------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| [`AGENTS.md`](../AGENTS.md)                                                                   | Slim playbook — always start here in agent sessions     |
| [`.cursor/skills/gds-compliant-frontend/`](../.cursor/skills/gds-compliant-frontend/SKILL.md) | Project skill (template shape, guidance URLs, coverage) |
| [`.cursor/rules/`](../.cursor/rules/)                                                         | Always-on consistency rules                             |
| Playbooks below                                                                               | Upgrade, create component/pattern, testing, UI rules    |

Agents should still open human-oriented docs when onboarding a teammate or explaining “why”.

## GOV.UK Frontend maintenance

| Doc                                                        | Purpose                                                                                               |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md) | Update plan — read [latest release](https://github.com/alphagov/govuk-frontend/releases/latest) first |
| [govuk-frontend-roadmap.md](govuk-frontend-roadmap.md)     | Upstream roadmap; what not to invent                                                                  |
| [testing-components.md](testing-components.md)             | Fixture parity, 100% coverage, Nunjucks suite, CI                                                     |
| [govuk-components.md](govuk-components.md)                 | Component architecture                                                                                |
| [creating-components.md](creating-components.md)           | Ship a new component                                                                                  |
| [creating-patterns.md](creating-patterns.md)               | Ship a Design System pattern / page                                                                   |
| [preview-server.md](preview-server.md)                     | Local preview / parity browser                                                                        |

## Assessment & standards

| Doc                                                                | Purpose                              |
| ------------------------------------------------------------------ | ------------------------------------ |
| [guidance-sources.md](guidance-sources.md)                         | Primary official URLs                |
| [service-assessment-readiness.md](service-assessment-readiness.md) | Service Standard-shaped UI checklist |
| [service-standard.md](service-standard.md)                         | 14 points (local summary)            |
| [technology-code-of-practice.md](technology-code-of-practice.md)   | TCoP (local summary)                 |
| [authoritative-references.md](authoritative-references.md)         | Link table                           |

## UI rules

| Doc                                                | Purpose                                |
| -------------------------------------------------- | -------------------------------------- |
| [page-shell.md](page-shell.md)                     | Page template, landmarks               |
| [layout-chrome.md](layout-chrome.md)               | Skip link, header, service nav, footer |
| [accessibility.md](accessibility.md)               | WCAG, focus, progressive enhancement   |
| [content-and-forms.md](content-and-forms.md)       | Content design, forms, errors          |
| [design-tokens.md](design-tokens.md)               | Colour, typography, spacing            |
| [styles.md](styles.md)                             | Sass pipeline, cascade, no !important  |
| [frontend-performance.md](frontend-performance.md) | Caching, compression, budgets          |
| [frontend-security.md](frontend-security.md)       | OWASP headers, CSP, cookies            |

## Per-component docs

Add `docs/govuk-<kebab-name>.md` when each component ships. Until then: [govuk-components.md](govuk-components.md) and the [Design System components](https://design-system.service.gov.uk/components/).
