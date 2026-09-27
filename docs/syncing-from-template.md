# Syncing from the language-agnostic template

This repository is the **Go** specialised line of [govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example). Shared playbooks and hygiene files still live in that template; this repo adds native Go HTML generation (`internal/`, `cmd/server`) on top. Node remains tooling only (Frontend pin, Sass, shared baseline tests).

## Remotes

| Remote     | Points at                          | Role                                             |
| ---------- | ---------------------------------- | ------------------------------------------------ |
| `origin`   | `Nooshu/govuk-frontend-example-go` | This Go project                                  |
| `template` | `Nooshu/govuk-frontend-example`    | Language-agnostic source of shared docs/dotfiles |

```sh
git remote -v
# origin    git@github.com:Nooshu/govuk-frontend-example-go.git
# template  https://github.com/Nooshu/govuk-frontend-example.git
```

## Recommended: path sync (safe for divergence)

Pulls only the paths listed in [`template-sync.paths`](../template-sync.paths) — Frontend playbooks, the shared [`baseline/`](../baseline/) performance and security contract, the Sass pipeline under [`styles/`](../styles/) and `scripts/build-styles*.mjs`, guidance, EditorConfig, Prettier, docs CI, licence/security — **without** overwriting Go-specific files (`docs/tech-stack.md`, `docs/why-go.md`, `docs/architecture.md`, `package.json`, `README.md`, `AGENTS.md`, `go.mod`, `internal/`, `cmd/`, …).

```sh
./scripts/sync-from-template.sh
git status
git diff --cached
git commit -m "chore: sync shared paths from template"
```

Override remote or ref if needed:

```sh
TEMPLATE_REMOTE=template TEMPLATE_REF=main ./scripts/sync-from-template.sh
```

After syncing, skim release notes if Frontend guidance changed: https://github.com/alphagov/govuk-frontend/releases/latest

Then apply [`baseline/`](../baseline/) in this line’s Go server (`internal/baseline`, Brotli via `internal/httpx`). Serve CSS from the compiled Sass output (`npm run build:styles`), not `govuk-frontend.min.css`. Do not keep a weaker header or cache policy in `internal/`. Playbooks: [frontend-performance.md](frontend-performance.md), [frontend-security.md](frontend-security.md), [styles.md](styles.md).

## Alternative: full git merge

Use when you want the complete template history merged (more conflicts in specialised files):

```sh
git fetch template
git merge template/main
# resolve conflicts in tech-stack.md, README, package.json, AGENTS.md, …
```

Prefer path sync for routine doc/hygiene updates; use merge when intentionally aligning history.

## What not to sync from template

Keep local (do not add to `template-sync.paths` without care):

- `docs/tech-stack.md` — Go pin and tooling
- `docs/why-go.md` / `docs/architecture.md` / `docs/go-conventions.md`
- `docs/testing-components.md` / `docs/creating-components.md` / `docs/govuk-components.md` / `docs/frontend-security.md` / `docs/upgrading-govuk-frontend.md` — Go-line wording
- `docs/project-purpose.md` / `docs/onboarding.md` / `docs/documentation-structure.md` — may mention this line
- `docs/syncing-from-template.md` — this file
- `README.md`, `AGENTS.md`, `CONTRIBUTING.md`
- `package.json` / lockfile (Node tooling pin for this line)
- `go.mod` / `go.sum` / `cmd/` / `internal/` / `staticcheck.conf` / `.air.toml`
- `.gitattributes` / `.gitignore` — Go-oriented ignore and text attributes
- `.npmrc` — npm supply-chain policy for this line ([npm-security.md](npm-security.md))
- Cursor skill/rules if specialised

## Humans vs agents

- **Humans:** run the script, review the diff, commit.
- **Agents:** same — never force-overwrite Go stack files when syncing; follow [`AGENTS.md`](../AGENTS.md) and [tech-stack.md](tech-stack.md).
