# npm dependency security

This Go line uses **Node only as tooling**: pin `govuk-frontend`, compile Sass, run shared baseline/docs tests. Request-time HTML is native Go. Even so, the npm tree is a supply-chain surface — treat the lockfile and install policy as security controls.

## Policy (non-negotiable)

| Control                         | How this repo enforces it                                                                                    |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| Exact direct versions           | `save-exact=true` in [`.npmrc`](../.npmrc); pinned entries in `package.json`                                 |
| Committed lockfile              | `package-lock.json` (lockfileVersion 3) always committed                                                     |
| Deterministic CI install        | `npm ci` in GitHub Actions (never bare `npm install` in CI)                                                  |
| No dependency lifecycle scripts | `ignore-scripts=true` in `.npmrc`                                                                            |
| Known CVE gate                  | `scripts/audit-npm.mjs` via `npm run audit:npm` — fails on high/critical, with the reviewed exceptions below |
| Registry signatures             | `npm audit signatures` in `audit:npm`                                                                        |
| Lockfile shape                  | `scripts/check-npm-lockfile.mjs` — HTTPS `registry.npmjs.org` + sha512                                       |
| Engine pin                      | Node ≥22 / npm ≥10 with `engine-strict=true`                                                                 |
| Automated bumps                 | Dependabot weekly for npm (+ cooldown); review lockfile diffs carefully                                      |

## Commands

```sh
npm ci                 # preferred install (matches CI); respects .npmrc
npm run audit:npm      # high/critical audit + signatures + lockfile lint
npm run verify         # includes audit:npm
```

Local first-time setup can use `npm ci` when the lockfile is present (same as CI). Use `npm install` only when intentionally changing dependencies, then commit both `package.json` and `package-lock.json`.

## Accepted advisories

`npm audit --audit-level=high` cannot skip one finding. [`scripts/audit-npm.mjs`](../scripts/audit-npm.mjs) still fails the build for every other high or critical advisory. An entry below is temporary: delete it when `npm audit` no longer reports that id.

| Advisory                                                                 | Package        | Why it is accepted                                                                                                                                                                                        |
| ------------------------------------------------------------------------ | -------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [GHSA-vfj7-8cjw-p6xm](https://github.com/advisories/GHSA-vfj7-8cjw-p6xm) | `braces` 3.0.3 | No release newer than 3.0.3 is on the registry. `markdownlint-cli2` pulls it in for globbing only. The Go service does not load it. `npm audit fix --force` would downgrade `markdownlint-cli2` to 0.0.4. |

## Minimal tree

Keep direct dependencies tiny:

| Package             | Why it exists                                |
| ------------------- | -------------------------------------------- |
| `govuk-frontend`    | Pinned Frontend assets, fixtures, Sass entry |
| `sass`              | Compile `styles/` → `dist/stylesheets/`      |
| `prettier`          | Docs/format gate                             |
| `markdownlint-cli2` | Docs lint gate                               |

Do **not** add frameworks, bundlers, or “helpful” npm libraries for request-time HTML. Prefer Go’s standard library for the service.

## Adding or upgrading a package

1. Prefer the standard library / existing Go code first.
2. If npm is required, add an **exact** version (`save-exact`).
3. Run `npm install <name>@<version>`, then `npm run audit:npm`.
4. Read the **lockfile** diff (transitive deps), not only `package.json`.
5. Confirm installs still work with `ignore-scripts=true` (no silent native rebuild requirement).
6. Update [tech-stack.md](tech-stack.md) if the dependency is a lasting stack choice.
7. For `govuk-frontend`, follow [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md) — never a casual bump.

## What Dependabot is for

Weekly npm PRs catch patches and advisories. Frontend upgrades still need a human following the Frontend upgrade playbook. Grouped PRs cover docs tooling; leave `govuk-frontend` ungrouped so it stays an intentional change.

## Related

- [tech-stack.md](tech-stack.md) — stack pins
- [frontend-security.md](frontend-security.md) — HTTP response security (separate from npm)
- [CONTRIBUTING.md](../CONTRIBUTING.md) — local verify
- [npm lockfile + ci guidance](https://docs.npmjs.com/cli/v10/commands/npm-ci)
