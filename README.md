# GOV.UK Frontend example (Go)

> [!IMPORTANT]
> You are free to fork, modify, and maintain this repository for your own use.
>
> This includes using and adapting it within your department, organisation, or project.

<!-- Separate alerts. Prettier collapses the two blank lines markdownlint MD028 would accept. -->

> [!WARNING]
>
> **🚨 Example repository only**
>
> This repository is a **demonstration only**. It will not be actively maintained or supported.
>
> **It is not an official UK government project.** It is not endorsed, maintained, or supported by any UK government department, the Government Digital Service (GDS), or the GOV.UK Design System team.
>
> **No ongoing support or updates will be provided.** This includes maintenance, dependency updates, security fixes, or technical support.
>
> **Use this code at your own risk.** You are responsible for reviewing, testing, securing, and maintaining the code, and for determining whether it is suitable for use in a service or production environment.
>
> **You are free to fork, modify, and maintain this repository for your own use.**
>
> This repository is released under the [MIT Licence](LICENSE). See the licence for the full terms.

**Go** specialised line of the GDS-compliant frontend template: **Go** generates HTML natively; **[GOV.UK Frontend](https://frontend.design-system.service.gov.uk/)** macros/`template.njk` are the behaviour reference — **no** React/Vue/Angular/Svelte for UI. Official fixtures enable **100% HTML parity** of Go output vs every fixture `html`. Node is used only to pin `govuk-frontend`, compile Sass, and run shared baseline tests — not to render HTML at request time.

Language-agnostic sibling (shared playbooks): [Nooshu/govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example). Sync: [`docs/syncing-from-template.md`](docs/syncing-from-template.md).

**Stack:** [`docs/tech-stack.md`](docs/tech-stack.md). **Why Go:** [`docs/why-go.md`](docs/why-go.md). **Architecture:** [`docs/architecture.md`](docs/architecture.md). **Deploy demo:** [`docs/deploying-on-render.md`](docs/deploying-on-render.md).

## Priorities

Frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design.

## Who should read what

| You are…            | Start here                                                                                                                                                     |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Human developer** | [`docs/onboarding.md`](docs/onboarding.md) → [`docs/why-go.md`](docs/why-go.md) → [`CONTRIBUTING.md`](CONTRIBUTING.md) → [`docs/`](docs/README.md)             |
| **AI coding agent** | [`AGENTS.md`](AGENTS.md) → [`.cursor/skills/gds-compliant-frontend/`](.cursor/skills/gds-compliant-frontend/SKILL.md) → playbooks in [`docs/`](docs/README.md) |

Dual-audience map: [`docs/documentation-structure.md`](docs/documentation-structure.md).

## Quick local checks

```sh
npm ci                  # lockfile install (scripts disabled — see docs/npm-security.md)
npm run audit:npm       # high+ audit, signatures, lockfile lint
npm run build:styles    # Sass → dist/stylesheets/application.css
npm start               # build:styles, then go run ./cmd/server
npm test                # baseline, Sass pipeline, Go fixture parity, and service tests
npm run lint:go         # go vet + go fix -diff + staticcheck
npm run verify          # docs + audit:npm + build:styles + lint:go + tests
npm run sync:template   # pull shared docs/dotfiles from the agnostic template
```

## Licence and security

- Code in this repository: [MIT License](LICENSE)
- How to report vulnerabilities: [SECURITY.md](SECURITY.md)
- GOV.UK Design System and Frontend are maintained by GDS; Crown copyright / OGL apply to GOV.UK content patterns as documented on GOV.UK.
