# Syncing from the language-agnostic template

This repository is the **TypeScript** specialised line of [govuk-frontend-example](https://github.com/Nooshu/govuk-frontend-example). Shared playbooks and hygiene files still live in that template; this repo adds a TypeScript wrapper stack on top.

## Remotes

| Remote     | Points at                                  | Role                                             |
| ---------- | ------------------------------------------ | ------------------------------------------------ |
| `origin`   | `Nooshu/govuk-frontend-example-typescript` | This TypeScript project                          |
| `template` | `Nooshu/govuk-frontend-example`            | Language-agnostic source of shared docs/dotfiles |

```sh
git remote -v
# origin    https://github.com/Nooshu/govuk-frontend-example-typescript.git
# template  https://github.com/Nooshu/govuk-frontend-example.git
```

## Recommended: path sync (safe for divergence)

Pulls only the paths listed in [`template-sync.paths`](../template-sync.paths) — Frontend playbooks, the shared [`baseline/`](../baseline/) performance and security contract, the Sass pipeline under [`styles/`](../styles/) and `scripts/build-styles*.mjs`, guidance, EditorConfig, Prettier, docs CI, licence/security — **without** overwriting TypeScript-specific files (`docs/tech-stack.md`, `package.json`, `README.md`, `AGENTS.md`, `src/`, …).

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

Then apply [`baseline/`](../baseline/) in this line's server (`applyResponseHeaders`, `buildSetCookie`, Brotli). Serve CSS from the compiled Sass output (`npm run build:styles`), not `govuk-frontend.min.css`. Do not keep a weaker header or cache policy in `src/`. Playbooks: [frontend-performance.md](frontend-performance.md), [frontend-security.md](frontend-security.md), [styles.md](styles.md).

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

- `docs/tech-stack.md` — TypeScript pin and tooling
- `docs/project-purpose.md` / `docs/onboarding.md` / `docs/documentation-structure.md` — may mention this line
- `docs/syncing-from-template.md` — this file
- `README.md`, `AGENTS.md`, `CONTRIBUTING.md`
- `package.json` / lockfile / `tsconfig.json` / `src/`
- Cursor skill/rules if specialised

## Humans vs agents

- **Humans:** run the script, review the diff, commit.
- **Agents:** same — never force-overwrite TypeScript stack files when syncing; follow [`AGENTS.md`](../AGENTS.md) and [tech-stack.md](tech-stack.md).
