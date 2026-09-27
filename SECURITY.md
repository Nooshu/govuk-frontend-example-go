# Security policy

## Supported versions

This repository is a **template**. Security fixes apply to the latest commit on the default branch. Consuming services should pin dependencies (including `govuk-frontend`) and track [govuk-frontend releases](https://github.com/alphagov/govuk-frontend/releases/latest) themselves.

| Version / branch      | Supported        |
| --------------------- | ---------------- |
| Latest `main`         | Yes              |
| Older commits / forks | Best effort only |

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Prefer private disclosure:

1. Use GitHub’s **[Private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability)** for this repository if enabled, or
2. Contact the repository maintainers through a private channel they publish (for example a security email in the GitHub org or profile).

Include as much detail as you can: affected paths, reproduction steps, impact, and any suggested fix.

We aim to acknowledge reports within **7 days** and to keep you informed of progress until the issue is resolved or declined with rationale.

## Scope

In scope for this template:

- Secrets or credentials accidentally documented or committed here
- Unsafe defaults in shared tooling, CI, or example configuration
- XSS or injection risk in documented HTML/encoding guidance (for example misuse of unsanitised `html` options)
- Supply-chain issues in this repo’s direct npm tooling once disclosed

Out of scope (report upstream or to the consuming service):

- Vulnerabilities in [GOV.UK Frontend](https://github.com/alphagov/govuk-frontend) itself — use that project’s security process
- Issues only present in a downstream service built from this template
- Denial of service against third-party sites or speculative attacks without a concrete defect here

## Frontend security expectations

This template prioritises **frontend security** (see [`docs/priorities.md`](docs/priorities.md)):

- Prefer GOV.UK Frontend **Nunjucks macros** and fixture-compatible encoding; do not bypass escaping
- Treat component `html` options as untrusted until sanitised; prefer `text`
- Do not add SPA/frontend frameworks that expand client attack surface for GOV.UK UI
- Keep `govuk-frontend` pinned and review https://github.com/alphagov/govuk-frontend/releases/latest before upgrading
- Apply the shared response-header baseline ([`docs/frontend-security.md`](docs/frontend-security.md), [`baseline/`](baseline/)) on every HTML and asset response

## Dependencies

Run `npm audit` (and the wrapper language’s equivalent once chosen) as part of routine maintenance. Dependabot is configured for npm and GitHub Actions where enabled.

## Licence

This project is licensed under the MIT License — see [`LICENSE`](LICENSE). GOV.UK Design System / Frontend assets and content remain subject to their own Crown copyright / OGL terms as published by GDS.
