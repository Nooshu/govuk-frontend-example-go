# Contributing

Thanks for helping maintain this **GDS-compliant frontend** template. This guide is for **human developers**. Coding agents should follow [`AGENTS.md`](AGENTS.md) and the playbooks under [`docs/`](docs/README.md).

## Before you start

1. Read [`docs/project-purpose.md`](docs/project-purpose.md) and [`docs/onboarding.md`](docs/onboarding.md).
2. Confirm the stack in [`docs/tech-stack.md`](docs/tech-stack.md) (Go; Node only pins Frontend, compiles Sass, and runs shared baseline tests).
3. Prefer official guidance listed in [`docs/guidance-sources.md`](docs/guidance-sources.md).
4. Sync shared playbooks from the agnostic template when needed: [`docs/syncing-from-template.md`](docs/syncing-from-template.md).
5. Priorities: frontend web performance → frontend security → reduced maintenance → accessibility → inclusive design.

## Non-negotiables (short)

- GOV.UK Frontend only for UI — **no** React/Vue/Angular/Svelte (etc.).
- Generate component HTML with the **native Go renderers** in `internal/govuk`, which track Frontend macros. Do not shell out to Node to render, and do not copy-paste HTML from releases.
- Official fixtures for **100% HTML parity** of **Go** output vs every fixture `html`; never edit fixture `html` to pass tests. A Nunjucks-only check is not enough.
- **100%** code coverage (functions, branches, statements) when application code exists.
- Before upgrading Frontend, read https://github.com/alphagov/govuk-frontend/releases/latest.
- HTTP responses use the shared [`baseline/`](baseline/). Compress with Brotli; Gzip is only the fallback when the client does not advertise `br`.
- Compile CSS via Sass (`styles/` → Frontend `@use` → `govuk-overrides.scss` last). Never use `!important` in service CSS. See [`docs/styles.md`](docs/styles.md).
- Document every change for **humans and agents** ([docs/documentation-structure.md](docs/documentation-structure.md)).
- Follow the **latest** Go best practices in [docs/tech-stack.md](docs/tech-stack.md). Shared Node tooling stays ESM.

Full list: [`AGENTS.md`](AGENTS.md).

## Consistency tooling

```sh
npm install
npm run build:styles    # Sass → dist/stylesheets/application.css
npm start               # example service at http://127.0.0.1:3000
npm test                # baseline, Sass pipeline, fixtures, and the example service; 100% coverage
npm run verify          # docs + build:styles + lint:go + tests
npm run sync:template   # shared paths from Nooshu/govuk-frontend-example
```

See [`docs/syncing-from-template.md`](docs/syncing-from-template.md). Dotfiles and lint setup match the language-agnostic template. Go adds `go test`, `npm run lint:go` (`go vet`, `go fix -diff`, `staticcheck`), and 100% coverage on `./internal/...`. Node remains for the Frontend pin, Sass, and the shared baseline tests.

### Dotfiles (do not bypass)

| File                                        | Role                                                   |
| ------------------------------------------- | ------------------------------------------------------ |
| `.editorconfig`                             | Indentation, charset, newlines across editors          |
| `.gitignore` / `.gitattributes`             | Ignore hygiene + line endings                          |
| `.nvmrc` / `.npmrc`                         | Node version + npm behaviour for Frontend/docs tooling |
| `.prettierrc` / `.prettierignore`           | Shared formatting                                      |
| `.markdownlint-cli2.jsonc`                  | Markdown consistency                                   |
| `.vscode/settings.json` / `extensions.json` | Shared editor defaults                                 |
| `.cursor/rules/`                            | Agent consistency rules                                |
| `.github/`                                  | PR template, Dependabot                                |

## Pull requests

- Keep changes focused; update `/docs` (and `AGENTS.md` links) when behaviour or process changes — dual audience, same PR.
- Follow Go’s latest best practices; do not introduce outdated stack idioms. Prefer the standard library.
- Split finished work into focused commits with comprehensive messages (see [`AGENTS.md`](AGENTS.md)).
- Use the PR template checklist.
- For Frontend bumps: follow [`docs/upgrading-govuk-frontend.md`](docs/upgrading-govuk-frontend.md).

## Licence and security

- [MIT License](LICENSE)
- [SECURITY.md](SECURITY.md) — private vulnerability reporting; Frontend upgrade and encoding expectations

## Documentation for both audiences

See [`docs/documentation-structure.md`](docs/documentation-structure.md). Put detail in `/docs`; keep `AGENTS.md` as the agent index.
