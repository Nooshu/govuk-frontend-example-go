# GOV.UK Frontend example (TypeScript)

**TypeScript** specialised line of the GDS-compliant frontend template: **Node + TypeScript** generates HTML; **[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)** macros (**prefer Nunjucks**) — **no** React/Vue/Angular/Svelte for UI. Official fixtures enable **100% HTML parity** of TypeScript output vs every fixture `html`.

Language-agnostic sibling (shared playbooks): [Nooshu/govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example). Sync: [`docs/syncing-from-template.md`](docs/syncing-from-template.md).

**Stack:** see [`docs/tech-stack.md`](docs/tech-stack.md).

## Priorities

Frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design.

## Who should read what

| You are…            | Start here                                                                                                                                                     |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Human developer** | [`docs/onboarding.md`](docs/onboarding.md) → [`CONTRIBUTING.md`](CONTRIBUTING.md) → [`docs/`](docs/README.md)                                                  |
| **AI coding agent** | [`AGENTS.md`](AGENTS.md) → [`.cursor/skills/gds-compliant-frontend/`](.cursor/skills/gds-compliant-frontend/SKILL.md) → playbooks in [`docs/`](docs/README.md) |

Dual-audience map: [`docs/documentation-structure.md`](docs/documentation-structure.md).

## Quick local checks

```sh
npm install
npm run build:styles    # Sass → dist/stylesheets/application.css
npm start               # build:styles, then example pages and component demos
npm test                # baseline, Sass pipeline, fixture parity, and service tests
npm run verify          # docs + build:styles + typecheck + tests
npm run sync:template   # pull shared docs/dotfiles from the agnostic template
```

## Licence and security

- Code in this repository: [MIT License](LICENSE)
- How to report vulnerabilities: [SECURITY.md](SECURITY.md)
- GOV.UK Design System and Frontend are maintained by GDS; Crown copyright / OGL apply to GOV.UK content patterns as documented on GOV.UK.
