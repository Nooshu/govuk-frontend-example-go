# GOV.UK Frontend example (Go)

> [!IMPORTANT]
> You are free to fork this repository and use it for your own purposes, and to modify and maintain it as you see fit.

> [!WARNING]
> 🚨 **Example repository only**
>
> This repository was created as a demonstration and will not be actively maintained or supported. It is not an official UK government project and is not endorsed, maintained, or supported by any UK government department, the Government Digital Service (GDS), or the GOV.UK Design System team.
>
> I will not be providing ongoing maintenance, updates, security fixes, or technical support.
>
> Use this code at your own risk. You are responsible for reviewing, testing, securing, maintaining, and ensuring the suitability of the code before using it in any service or production environment. I accept no responsibility or liability for any loss, damage, security issue, service failure, or other consequence resulting from its use.
>
> This repository is released under the MIT Licence. See the [LICENSE](LICENSE) file for the full licence terms.

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
