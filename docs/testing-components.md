# Testing components

How we keep **100% HTML parity** with GOV.UK Frontend — now and after upgrades.

Authoritative upstream: https://frontend.design-system.service.gov.uk/testing-your-html/

## What must be compared

| Compare                                              | Required?              | Purpose                                                                           |
| ---------------------------------------------------- | ---------------------- | --------------------------------------------------------------------------------- |
| **Backend / library HTML → official fixture `html`** | **Yes — primary gate** | Proves the wrapper language’s interpretation of the component matches the release |
| **Nunjucks macro HTML → stored fixture `html`**      | Yes — secondary        | Proves fixtures are not stale relative to the pinned Frontend macros              |

**Do not** ship a setup that only compares Nunjucks macros to fixture HTML. That never exercises the backend language’s renderer. Language lines and services built from this template must run the **parity suite** against backend output for **every** fixture in every shipped component’s `fixtures.json` from the pinned GOV.UK Frontend release.

**Stacks:** Prefer generating HTML via **Nunjucks macros** from `govuk-frontend` when the language can call them. If the backend wraps or re-implements macros, parity still compares **that backend output** to official fixtures. The Nunjucks suite stays on **Node** with the pinned package.

## Why fixtures exist

Official `fixtures.json` files (one set per component per Frontend release) are the contract for **extensive 100% HTML parity testing** of whatever backend generates frontend markup. Sync them from the same `govuk-frontend` version as CSS/JS. Set them up from day one so every shipped component proves byte-for-byte equality — and so upgrades catch drift automatically.

Do **not** treat copy-pasted HTML from Design System examples or release notes as the source of truth; macros + fixtures are. Do **not** treat “Nunjucks still matches fixtures” as proof the backend is correct.

## Coverage gate (100% code)

Application and library code under test must maintain **100%** coverage of:

- **functions**
- **branches**
- **statements**

Wire the language’s coverage tool so local verify and CI **fail** below 100% on all three. Do **not** normalise fixture HTML or skip parity cases to inflate coverage. Exclude only generated/vendor assets and explicit, documented exceptions (if any) in [tech-stack.md](tech-stack.md).

## Layers

| Layer                          | What it proves                                    | How                                                                                                                              |
| ------------------------------ | ------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| **Parity suite (primary)**     | **Backend / library HTML** matches fixture `html` | Map fixture `options` → **backend render** → ordinal string equality vs fixture `html` — cover **all** fixtures extensively      |
| **Nunjucks suite (secondary)** | Stored fixtures still match Frontend macros       | Node scripts render via `govuk-frontend` Nunjucks; compare to fixture `html` — freshness only; does not replace the parity suite |
| **HTTP smoke** (if applicable) | Preview/fixture surfaces wired                    | Hit preview and raw-fixture endpoints in a Testing env                                                                           |
| **Structural**                 | Nav lists components; no demos on index           | Assert homepage entries and absence of embedded demos                                                                            |
| **Fixture sync**               | App fixtures ≡ copies used by the Nunjucks suite  | Byte-identical file check                                                                                                        |

## Hard rules

1. **Never** edit fixture `html` to make tests pass.
2. **Never** normalise or pretty-print HTML before compare.
3. **Never** mix Frontend versions between CSS/JS and fixtures.
4. Prefer in-process parity for the bulk of cases (fast, no HTTP); aim to exercise **every** fixture for each shipped component through the **backend** renderer.
5. After upgrade, both the **parity suite** and the **Nunjucks suite** must be green — see [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md). Always read https://github.com/alphagov/govuk-frontend/releases/latest first.
6. **Never** treat a green Nunjucks suite alone as release criteria for a language line.

## Encoding

Parity depends on **Nunjucks `escape`** (`&#39;` for `'`, real newlines in textarea values). Use the shared helper in [creating-components.md](creating-components.md) when not calling Nunjucks directly — the backend must emit the same encoding the fixtures expect.

## Commands

Document in [tech-stack.md](tech-stack.md):

- **library / parity tests** (backend output vs **all** fixtures) — mandatory
- coverage report enforcing **100%** functions, branches, statements
- Nunjucks fixture verification (Node + `govuk-frontend`) — freshness
- a single “verify” gate for local + CI that includes the parity suite

Today, before a wrapper language exists, the shared baseline and Sass pipeline are covered by Node’s test runner (`npm test`). When a wrapper lands, its parity suite becomes part of `verify`.

## Preview as human parity browser

Previews render **one selected fixture**, show **library/backend HTML** beside official fixture `html`, and display a parity success/fail banner. Details: [preview-server.md](preview-server.md).

## Adding tests for a new component

Follow steps 8–9 in [creating-components.md](creating-components.md). Clone the closest sibling’s **backend parity** tests and Nunjucks render script; cover every fixture name in both layers.
