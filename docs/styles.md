# Styles (Sass pipeline)

This template compiles CSS from **Sass**, using GOV.UK Frontend’s Sass API. Do **not** treat `node_modules/govuk-frontend/dist/govuk/govuk-frontend.min.css` as the long-term stylesheet source.

Authoritative upstream: [Include CSS](https://frontend.design-system.service.gov.uk/include-css/).

## Cascade

| Order | Source                                                                                     | Role                                                  |
| ----- | ------------------------------------------------------------------------------------------ | ----------------------------------------------------- |
| 1     | `pkg:govuk-frontend` via `@use` in [`styles/application.scss`](../styles/application.scss) | Full Frontend CSS (settings configurable with `with`) |
| 2     | [`styles/govuk-overrides.scss`](../styles/govuk-overrides.scss)                            | Service overrides — last in the cascade               |

Source order is cascade order. Put overrides only in `govuk-overrides.scss` so they win with **equal or higher specificity**, not with `!important`.

## Commands

```sh
npm run build:styles   # writes dist/stylesheets/application.css
npm test               # includes Sass pipeline + no-!important checks
npm run verify         # docs + build:styles + tests
```

The build uses Dart Sass with `NodePackageImporter` and `quietDeps: true` so dependency deprecation noise from Frontend does not fail the compile. See [`scripts/build-styles.mjs`](../scripts/build-styles.mjs).

Serve the compiled file with the baseline cache kind for fingerprinted assets once the URL includes a content hash ([frontend-performance.md](frontend-performance.md)).

## `!important`

**Forbidden** in `styles/` (including `govuk-overrides.scss`). Tests fail if `!important` appears outside comments.

Frontend’s own `govuk-!-…` utility classes may compile to `!important` upstream. That is GDS’s override system, not a licence to add `!important` in service CSS.

## What belongs in overrides

Prefer, in order:

1. Component / pattern options from GOV.UK Frontend macros
2. Frontend’s documented spacing / width / typography override **classes** (`govuk-!-…`) in HTML
3. Cascade rules in `govuk-overrides.scss` using specificity (compose an `app-` class with a `govuk-` class)

Do not:

- Restyle yellow focus
- Invent parallel chrome that looks like GOV.UK but is hand-rolled
- Copy-paste or `@import` the prebuilt minified Frontend CSS instead of `@use`

## Selective imports

When a service only needs some components, replace the full `@use "pkg:govuk-frontend"` with the selective `@use` list from the [Include CSS](https://frontend.design-system.service.gov.uk/include-css/) guide (base → core → objects → components → utilities → Frontend overrides). Keep `@use "govuk-overrides"` **last**.

## Assets path

`application.scss` sets `$govuk-assets-path: "/assets/"`. Serve fonts and images from that path (or change the setting and the static routes together).

## Upgrades

After bumping `govuk-frontend`, run `npm run build:styles` and fix Sass breaks before parity HTML work. Re-measure Brotli CSS size ([frontend-performance.md](frontend-performance.md)).
