# Documentation structure (humans + agents)

This template and every project built from it must keep documentation **complete enough for a new human developer to onboard** and **dense enough for an AI agent to apply contracts without guessing**. That applies to every prompt-driven change and every hand-written feature.

Canonical agent index: [`AGENTS.md`](../AGENTS.md). Doc map: [`README.md`](README.md).

## Principles

1. **One source of truth per topic** — avoid duplicating long guidance in both `AGENTS.md` and `/docs`.
2. **`AGENTS.md` stays slim** — priorities, non-negotiables, checklists, and links. Agents load this first.
3. **`/docs` holds the detail** — written so a human can follow it end-to-end without reading agent meta.
4. **Skill + rules encode habits** — [`.cursor/skills/`](../.cursor/skills/) for workflows; [`.cursor/rules/`](../.cursor/rules/) for always-on consistency.
5. **Official URLs win** — [guidance-sources.md](guidance-sources.md) before inventing policy.
6. **No silent behaviour** — if code or config changes how someone builds, runs, secures, styles, or tests the service, docs change in the same change set.
7. **Latest Go practice** — document and implement current Go best practices ([tech-stack.md](tech-stack.md), [go-conventions.md](go-conventions.md)); do not fossilise outdated patterns.

## Audience cues

| Cue in a doc                                            | Meaning                                 |
| ------------------------------------------------------- | --------------------------------------- |
| “Start here”, “Repo map”, “Troubleshooting”             | Human onboarding                        |
| “Non-negotiables”, “Playbook”, “Do / don’t”, checklists | Agent-oriented (still useful to humans) |
| “Stack note”, “Go conventions”                          | Applies to both; follow tech-stack.md   |

## Mandatory documentation for every change

When a prompt, feature, or bugfix lands code or lasting behaviour, update **both** audiences as needed:

| Change type                                          | Humans (minimum)                                                                                  | Agents (minimum)                                                                                                        |
| ---------------------------------------------------- | ------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| New or changed behaviour / API / HTTP / CSS contract | Playbook or section under `/docs`; link from [README.md](README.md) if new file                   | Link from [`AGENTS.md`](../AGENTS.md) playbook table or non-negotiables when it is a contract; skill/rules if always-on |
| New workflow (run, verify, upgrade, commit)          | [onboarding.md](onboarding.md) and/or [../CONTRIBUTING.md](../CONTRIBUTING.md)                    | `AGENTS.md` checklist / non-negotiable if agents must follow it                                                         |
| Stack or tooling choice                              | [tech-stack.md](tech-stack.md)                                                                    | Language rule + tech-stack; never invent paths before it is recorded                                                    |
| Security / performance baseline                      | [frontend-security.md](frontend-security.md) / [frontend-performance.md](frontend-performance.md) | Non-negotiable + playbook links                                                                                         |
| Styles / Sass                                        | [styles.md](styles.md)                                                                            | Non-negotiable + page review                                                                                            |
| Component or pattern                                 | Creating playbooks + component doc when shipped                                                   | Creating playbooks already linked from `AGENTS.md`                                                                      |
| Sync from agnostic template                          | [syncing-from-template.md](syncing-from-template.md)                                              | Sync playbook already linked from `AGENTS.md`                                                                           |

**Done means:** a new teammate can find how to use the change from human docs, and an agent session can find the contract from `AGENTS.md` → skill → playbook without undocumented tribal knowledge.

When editing docs:

- Prefer short sections and tables over long prose.
- Link to upstream GDS pages instead of pasting large excerpts.
- If you add an agent playbook step, add a one-line human explanation nearby or in onboarding.
- Update [`docs/README.md`](README.md) whenever you add a new doc.
- Run `npm run verify:docs` after substantive Markdown edits.

## Latest language best practices

Applies to this **Go** line **and** to the shared Node tooling synced from the agnostic template:

1. This line records **Go** in [tech-stack.md](tech-stack.md) and [go-conventions.md](go-conventions.md). Follow **current Go** layout, modules, tests, packaging, and CI norms for **all** new feature work and refactors.
2. Prefer official or widely accepted current guides for the pinned major versions (Go 1.27+, Node 22+ for tooling) over outdated tutorials.
3. Shared Node tooling (`baseline/`, Sass scripts, fixture helpers) already uses current ESM practice; keep it that way.
4. Do not adopt a “best practice” that conflicts with Frontend macros, fixture parity, the performance/security baseline, or the Sass cascade.
5. When best practices change upstream, update tech-stack notes and code in focused commits — documentation and implementation together.

## Suggested reading order

**Human (day one)**

1. [../README.md](../README.md)
2. [project-purpose.md](project-purpose.md)
3. [why-go.md](why-go.md) and [architecture.md](architecture.md)
4. [onboarding.md](onboarding.md)
5. [tech-stack.md](tech-stack.md) / [go-conventions.md](go-conventions.md)
6. [../CONTRIBUTING.md](../CONTRIBUTING.md)

**Agent (every session)**

1. [../AGENTS.md](../AGENTS.md)
2. Skill: `gds-compliant-frontend`
3. Relevant playbook under `/docs`
4. [guidance-sources.md](guidance-sources.md) when policy is unclear

## Consistency tooling

Dotfiles and shared tooling keep formatting and hygiene aligned — see [../CONTRIBUTING.md](../CONTRIBUTING.md#consistency-tooling). Document language-specific linters in [tech-stack.md](tech-stack.md).
