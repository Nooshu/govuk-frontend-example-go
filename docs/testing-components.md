# Testing components

How we keep **100% HTML parity** with GOV.UK Frontend — now and after upgrades.

Authoritative upstream: https://frontend.design-system.service.gov.uk/testing-your-html/

## What must be compared

| Compare                                              | Required?              | Purpose                                                                               |
| ---------------------------------------------------- | ---------------------- | ------------------------------------------------------------------------------------- |
| **Go `govuk.Render` HTML → official fixture `html`** | **Yes — primary gate** | Proves this line’s native Go ports match the pinned Frontend release                  |
| **Nunjucks macro HTML → stored fixture `html`**      | Optional — secondary   | Proves fixtures are not stale relative to the pinned Frontend macros (freshness only) |

**Do not** ship a setup that only compares Nunjucks macros to fixture HTML. That never exercises the Go renderer. This line must run the **parity suite** against **Go** output for **every** fixture in every shipped component’s `fixtures.json` from the pinned GOV.UK Frontend release.

**This line (Go):** generate component HTML with native ports in `internal/govuk` that track Frontend macros / `template.njk`. Do **not** shell out to Node/Nunjucks at request time. Parity always compares **Go `Render` output** to official fixtures. A Nunjucks-only suite, if run, stays on Node with the pinned package and never replaces the Go gate.

## Why fixtures exist

Official `fixtures.json` files (one set per component per Frontend release) are the contract for **extensive 100% HTML parity testing** of whatever backend generates frontend markup. Sync them from the same `govuk-frontend` version as CSS/JS. Set them up from day one so every shipped component proves byte-for-byte equality — and so upgrades catch drift automatically.

Do **not** treat copy-pasted HTML from Design System examples or release notes as the source of truth; macros + fixtures are. Do **not** treat “Nunjucks still matches fixtures” as proof the backend is correct.

## Coverage gate (100% code)

Application and library code under test must maintain **100%** coverage of:

- **functions**
- **branches**
- **statements**

Wire Go coverage (`go test` + `scripts/check-go-coverage.mjs`) so local verify and CI **fail** below 100% statements on `./internal/...`. Do **not** normalise fixture HTML or skip parity cases to inflate coverage. Exclude only generated/vendor assets and explicit, documented exceptions (if any) in [tech-stack.md](tech-stack.md).

## Layers

| Layer                          | What it proves                                   | How                                                                                                                            |
| ------------------------------ | ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| **Parity suite (primary)**     | **Go HTML** matches fixture `html`               | Map fixture `options` → **`govuk.Render`** → ordinal string equality vs fixture `html` — cover **all** fixtures extensively    |
| **Nunjucks suite (secondary)** | Stored fixtures still match Frontend macros      | Optional Node scripts via `govuk-frontend` Nunjucks; compare to fixture `html` — freshness only; does not replace the Go suite |
| **HTTP smoke** (if applicable) | Preview/fixture surfaces wired                   | Hit preview and raw-fixture endpoints in a Testing env                                                                         |
| **Structural**                 | Nav lists components; no demos on index          | Assert homepage entries and absence of embedded demos                                                                          |
| **Fixture sync**               | App fixtures ≡ copies used by the Nunjucks suite | Byte-identical file check                                                                                                      |

## Hard rules

1. **Never** edit fixture `html` to make tests pass.
2. **Never** normalise or pretty-print HTML before compare.
3. **Never** mix Frontend versions between CSS/JS and fixtures.
4. Prefer in-process parity for the bulk of cases (fast, no HTTP); aim to exercise **every** fixture for each shipped component through **`govuk.Render`**.
5. After upgrade, the **Go parity suite** must be green (and any optional Nunjucks freshness check) — see [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md). Always read https://github.com/alphagov/govuk-frontend/releases/latest first.
6. **Never** treat a green Nunjucks suite alone as release criteria for this Go line.

## Encoding

Parity depends on **Nunjucks `escape`** (`&#39;` for `'`, real newlines in textarea values). Use the shared helper in [creating-components.md](creating-components.md) when not calling Nunjucks directly — the backend must emit the same encoding the fixtures expect.

## Commands

Document in [tech-stack.md](tech-stack.md):

- **Go parity tests** (`internal/govuk` — Go output vs **all** fixtures) — mandatory
- coverage report enforcing **100%** statements on `./internal/...`
- optional Nunjucks fixture verification (Node + `govuk-frontend`) — freshness only
- a single “verify” gate for local + CI that includes the Go parity suite

`npm test` runs the Node baseline/Sass gates, builds styles, then `npm run test:go` (race + coverage). `npm run verify` also runs docs and `lint:go`.

## Preview as human parity browser

Previews load fixtures with [`govuk.LoadFixtures`](../internal/govuk/fixtures.go) and render with [`govuk.Render`](../internal/govuk/render.go) — the **same path** as the parity suite — so attribute key order and JSON number spelling match. Do not feed fixture options through `map[string]any` for parity checks: that re-sorts keys and widens integers to float64, which falsely reports mismatches.

Previews render **one selected fixture**, show **Go-rendered HTML**, and display a parity success/fail banner. Details: [preview-server.md](preview-server.md).

## Adding tests for a new component

Follow steps 8–9 in [creating-components.md](creating-components.md). Clone the closest sibling’s **Go parity** tests; cover every fixture name. Add an optional Nunjucks freshness check only if you maintain one.
