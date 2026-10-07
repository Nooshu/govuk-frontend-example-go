---
name: safe-dependency-updates
description: >-
  Updates npm and language-stack dependencies safely. Use when bumping packages,
  reviewing Dependabot or Renovate PRs, changing lockfiles, running npm audit or
  language vulnerability tooling, or upgrading govuk-frontend as a dependency.
  Requires the CI pipeline to be green before starting and before completing any
  dependency update.
---

# Safe dependency updates

## Hard gate: green pipeline

Do **not** start or finish a dependency update while the default-branch CI is red.

1. **Before changing any dependency:** confirm recent runs on `main` are green (`gh run list --branch main --limit 10`, or the Actions UI). If Docs / Verify / CI (or this repo’s equivalent workflows) are failing for unrelated reasons, stop and fix or escalate first.
2. **After changing dependencies:** run this repo’s full local verify, then push or open the PR and wait until **all** required checks are green.
3. **Done means:** manifests and lockfiles committed, lasting pin changes reflected in [`docs/tech-stack.md`](../../../docs/tech-stack.md) when needed, local verify green, **and** GitHub Actions green for the update commit or PR.

Do not merge Dependabot / Renovate PRs, and do not declare a manual bump complete, while checks are failing.

## When to use

- Adding, removing, or bumping a package
- Reviewing automated dependency PRs
- Lockfile-only churn that still needs review
- Security advisories (`npm audit`, Composer audit, `govulncheck`, Cargo audit, etc.)
- Pin changes including `govuk-frontend` (then also follow the Frontend upgrade playbook)

## Principles

Moved here from project agent rules so `AGENTS.md` and language `.mdc` files stay slim:

- Keep the dependency surface small; prefer the platform / standard library first
- Prefer established, actively maintained packages over convenience libraries
- Pin critical dependencies; commit lockfiles; CI must install from locks (`npm ci`, `uv sync --frozen`, `composer install --no-dev` in production images, etc.)
- Review **lockfile** diffs, not only the manifest
- Do not delete and regenerate a lockfile to “fix” problems without understanding the change
- Automated updates are welcome; never merge without tests **and** a green pipeline
- On non-Node lines, Node/`govuk-frontend` is for install, Sass, fixtures, and shared tooling — **not** request-time HTML rendering

## GOV.UK Frontend pin

Treat `govuk-frontend` as a compatibility event, not a casual bump:

1. Read https://github.com/alphagov/govuk-frontend/releases/latest
2. Follow [`docs/upgrading-govuk-frontend.md`](../../../docs/upgrading-govuk-frontend.md)
3. Refresh fixtures from the **same** version; fix renderers / macro usage — never edit fixture `html` to pass tests
4. Still obey the green-pipeline gate above

Also check the Design System [roadmap](https://design-system.service.gov.uk/community/roadmap/), [upcoming components](https://design-system.service.gov.uk/community/upcoming-components-patterns/), and [all releases](https://github.com/alphagov/govuk-frontend/releases). Local notes: [`docs/govuk-frontend-roadmap.md`](../../../docs/govuk-frontend-roadmap.md).

## This repository

| Piece             | Detail                                                                                                                                                                |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Ecosystems        | Go modules + npm tooling; GitHub Actions via Dependabot                                                                                                               |
| Manifests / locks | `go.mod`, `go.sum`, `package.json`, `package-lock.json`                                                                                                               |
| Install           | `npm ci`; Go via modules (`go mod tidy` after Go deps change)                                                                                                         |
| Verify            | `npm run verify` (includes `npm run audit:npm` and Go tests/lint)                                                                                                     |
| Audit             | `npm run audit:npm`; `govulncheck ./...` when available                                                                                                               |
| npm policy        | [`docs/npm-security.md`](../../../docs/npm-security.md) — exact versions, `ignore-scripts`, signatures, lockfile lint                                                 |
| Notes             | Prefer the Go standard library. Only add a module when stdlib lacks the feature (today: Brotli). Do not replace `internal/govuk` with incomplete third-party wrappers |

## Workflow checklist

Copy and track:

```text
Dependency update:
- [ ] main CI green before starting
- [ ] ecosystem + packages identified (one concern per commit when practical)
- [ ] manifests + lockfiles updated with the stack’s normal tool
- [ ] lockfile / transitive diff reviewed
- [ ] security tooling run when this repo defines it
- [ ] full local verify green
- [ ] docs/tech-stack updated if the pin is a lasting stack choice
- [ ] focused commit / PR
- [ ] GitHub Actions green before complete / merge
```

## Related

- Slim agent index: [`AGENTS.md`](../../../AGENTS.md)
- Template skill: [`.cursor/skills/gds-compliant-frontend/SKILL.md`](../gds-compliant-frontend/SKILL.md)
- Frontend upgrade playbook: [`docs/upgrading-govuk-frontend.md`](../../../docs/upgrading-govuk-frontend.md)
- Stack pins: [`docs/tech-stack.md`](../../../docs/tech-stack.md)
