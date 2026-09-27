# Creating a new GOV.UK component

Use this playbook when implementing a Design System **component** as a library wrapper. For **patterns**, use [creating-patterns.md](creating-patterns.md) instead.

Official fixture guidance: [https://frontend.design-system.service.gov.uk/testing-your-html/#using-the-html-test-files](https://frontend.design-system.service.gov.uk/testing-your-html/#using-the-html-test-files)

Related docs: [govuk-components.md](govuk-components.md), [testing-components.md](testing-components.md), [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md), [preview-server.md](preview-server.md).

**Stack note:** Structure the _wrapper_ using the **chosen language’s best practices** ([tech-stack.md](tech-stack.md)). Prefer calling **GOV.UK Frontend Nunjucks macros** for component HTML rather than copy-pasting HTML from each release. Official `fixtures.json` enables extensive **100% parity** testing of backend output.

## Goals (non-negotiable)

1. **GOV.UK Frontend is the source of truth** — official CSS/JS and `fixtures.json` from the **same** pinned `govuk-frontend` Node package.
2. **Options mirror Nunjucks macros** — names/shapes align with `macro-options.json` / fixture `options`.
3. **Exact HTML parity (backend vs fixtures)** — the **backend language’s** output equals each fixture’s `html` byte-for-byte (ordinal equality). Cover every fixture. A Nunjucks-vs-fixture check is freshness only; it does not replace backend parity. See [testing-components.md](testing-components.md).
4. **Never hand-write component markup in pages** — pages invoke the library API.
5. **No ad-hoc custom CSS** — Sass pipeline + `govuk-overrides.scss` only; no `!important` ([styles.md](styles.md)).
6. **Register in navigation** — every shipped component appears on `/components` (the preview homepage) with a preview link. The service start page links to that catalogue when demos are enabled.

## Which sibling to copy (once examples exist)

| If the component is…                        | Start from                                                  |
| ------------------------------------------- | ----------------------------------------------------------- |
| Single element / simple attributes          | Back link or Button                                         |
| List of items                               | Breadcrumbs or Accordion                                    |
| Form control that **composes** other macros | Character count (read every nested Nunjucks `template.njk`) |

## Target layout (conceptual)

Exact filenames and folders follow [tech-stack.md](tech-stack.md). Conceptually each component needs:

```text
component unit (name idiomatic for the wrapper language)
  options / model types
  public API entry
  renderer                         # builds exact HTML string
  options mapper                   # fixtures.json options → model
  fixtures.json                    # from govuk-frontend (do not invent html)

preview surface                    # Dev/Testing only — selected fixture
raw fixture endpoint               # fragment only; Dev/Testing only

parity + structural tests          # wrapper language
tests/govuk-fixtures/…             # Nunjucks verification (Node)
  <kebab-name>.fixtures.json
  render-<kebab-name>-fixtures.mjs
```

## Recommended work order

1. Copy official fixtures beside the component and into `tests/govuk-fixtures/` (or equivalent) for the Nunjucks suite.
2. Dump every fixture’s `html` + `options`; read Nunjucks `template.njk` (+ imports). Note attribute order and `{%-` / `-%}` whitespace stripping.
3. Implement models + mapper + renderer together — do not free-hand the component body in a way that drifts from the template.
4. Clone preview/fixture surfaces, parity tests, and the Node Nunjucks render script from the closest sibling.
5. Register homepage/nav entry (nav only — no live demos on the index).
6. Run parity tests → fix renderer/mapper only → Nunjucks suite → docs.
7. Visual check via preview ([preview-server.md](preview-server.md)).

## Step checklist

### 1. Obtain official fixtures

1. Confirm Design System “when to use / when not to use”.
2. From pinned `govuk-frontend` (npm/Node), copy `dist/govuk/components/<kebab-name>/fixtures.json`.
3. Skim `template.njk` and `macro-options.json`.
4. Keep a byte-identical copy for the Nunjucks suite.
5. **Never edit fixture** `html`**.** Refresh only when upgrading Frontend.

### 2. Models (Nunjucks-aligned)

- Root model mirrors macro options (`id`, `classes`, `attributes`, …) using types idiomatic for the wrapper language.
- Use a text/html pair: prefer `text` (encoded); `html` only for **trusted** markup; if both set, `html` wins; sanitise untrusted input.
- Map JSON carefully (`headingLevel` → idiomatic name in the language).
- Handle edge cases: falsy array entries, **tri-state booleans** (unset vs `false`), nested `label` / `hint` / `errorMessage` / `formGroup` / i18n maps.
- JSON numbers that Nunjucks stringifies (`maxlength`, `rows`, …) must round-trip as strings in HTML attributes where fixtures do.

### 3. Renderer

Prefer a dedicated `render(model) → string` (or the language’s equivalent) invoked by the public API.

**Read Nunjucks** `template.njk` **as the contract.** Attribute order and newlines follow the template (including `{%-` stripping).

**Encoding — match Nunjucks** `escape`, not typical framework encoders:

| Character | Entity        |
| --------- | ------------- |
| `&`       | `&`           |
| `<`       | `<`           |
| `>`       | `>`           |
| `"`       | `"`           |
| `'`       | `'`           |
| Newlines  | leave as `\n` |

Use one shared escape + attributes helper for every renderer.

**Composed macros:** If `template.njk` calls other components, implement the **composed output** in this renderer (inline helpers OK). Refactor to shared helpers later without breaking parity.

Other rules:

- Match indentation/newlines from fixture `html`.
- Preserve attribute and i18n map key order (JSON object order).
- Optional attributes: omit when falsy.
- Hard-code exact SVG / static fragments from the template when present.

### 4. Public API

Thin wrapper: accept the options model, return rendered HTML in whatever form is idiomatic (string, safe HTML type, component result). Pages call the library — they do not paste markup.

### 5. Options mapper

Map fixture `options` → model without losing edge cases: nested text/html, attribute dictionaries, number→string, tri-state booleans, omit empty error wrappers.

### 6. Fixtures endpoint (raw HTML)

- Expose a Dev/Testing-only way to fetch one fixture as a **fragment** (no layout).
- Route naming is stack-specific; document it in [tech-stack.md](tech-stack.md).

### 7. Previews surface

- List fixture names; **render only the selected** fixture.
- Show parity banner (library HTML vs official `html`).
- Back link to home — never use fixture hrefs for chrome.
- Syntax highlighting assets on Previews only — not global layout.
- Development / Testing only.

### 8. Tests

1. Structural/smoke: index lists preview; no embedded demo on index.
2. Parity (**primary**): one case per fixture name; in-process mapper + **backend** renderer; ordinal string equality to fixture `html`; cache fixtures. This is the wrapper language’s interpretation under test.
3. Optional HTTP smoke for fixture + preview surfaces when the stack has an HTTP app.
4. Use the wrapper language’s normal test isolation patterns.

### 9. Nunjucks fixture verification (Node)

- Depend on the same `govuk-frontend` pin via npm/Node.
- Script renders each fixture through Frontend’s Nunjucks macros and compares to stored `html` (trim trailing newline only if needed).
- Purpose: catch **stale fixtures** only. Library / backend parity tests catch **renderer drift**.
- A green Nunjucks suite without a green backend parity suite is **not** done.
- Typical shape: `tests/govuk-fixtures/render-<kebab-name>-fixtures.mjs` — document the exact runner in [tech-stack.md](tech-stack.md).

### 10. Navigation + docs

- Add nav item (display name, one-line description, preview path, Design System URL).
- Add `docs/govuk-<kebab-name>.md` and link from [README.md](README.md), [govuk-components.md](govuk-components.md), [testing-components.md](testing-components.md), [upgrading-govuk-frontend.md](upgrading-govuk-frontend.md).
- Add an agent-usage section (when to use, model table, minimal example, do/don’t).

### 11. Visual verify

Use the preview server; hard-refresh after rebuilds. Confirm the component is listed and default fixture shows parity success.

## Common pitfalls

| Symptom                        | Likely cause                                         |
| ------------------------------ | ---------------------------------------------------- |
| Whitespace-only parity failure | Indentation ≠ Nunjucks; `{%-` strips spaces/newlines |
| `'` or `                       |                                                      |
| `                              | Framework encoder instead of Nunjucks `escape`       |
| Attribute order differs        | Code property order ≠ `template.njk` order           |
| Numbers wrong/missing          | JSON numbers not mapped like Nunjucks                |
| Tri-state bool wrong           | Coerced missing → `false`                            |
| i18n key order wrong           | Plural map order not preserved                       |
| Composed control fails         | Nested macros not implemented                        |
| IDs off by one                 | Skipped falsy items; must preserve positions         |
| Fixture route is a full page   | Must return fragment only                            |
| Preview back goes to `/home`   | Used fixture href for chrome                         |
| Editing fixtures “fixes” tests | Update renderer instead                              |

## Do / don’t

**Do:** start from fixtures + Nunjucks `template.njk` + closest sibling; keep one coherent unit per component; prove **backend vs fixture** parity before calling done; use idiomatic types/modules/tests for the wrapper language; keep a Node Nunjucks suite for fixture freshness.

**Don’t:** hand-paste `govuk-`\* into pages; ship without nav/preview; embed demos on index; add ad-hoc CSS or `!important`; normalise HTML in tests; invent fixture HTML; nest incompatible components; pretend Frontend is not Node/Nunjucks upstream; ship the prebuilt minified Frontend CSS instead of the Sass pipeline; claim parity from Nunjucks-only checks without backend vs fixture tests.
