# Page shell

Stay consistent with the [GOV.UK page template](https://design-system.service.gov.uk/styles/page-template/) and related styles ([Layout](https://design-system.service.gov.uk/styles/layout/), [Typography](https://design-system.service.gov.uk/styles/typography/), [Colour](https://design-system.service.gov.uk/styles/colour/)). Prefer the app’s shared layout — see [layout-chrome.md](layout-chrome.md).

## Conceptual skeleton

```html
<html class="govuk-template" lang="en">
  <body class="govuk-template__body">
    <script>
      document.body.className +=
        ' js-enabled' +
        ('noModule' in HTMLScriptElement.prototype ? ' govuk-frontend-supported' : '');
    </script>
    <!-- Skip link → #content -->
    <!-- Header (masthead) + service navigation inside govuk-template__header -->
    <div class="govuk-width-container">
      <!-- BeforeContent: phase banner / breadcrumbs / back link (not breadcrumbs + back link) -->
      <main id="main-content" class="govuk-main-wrapper">
        <!-- Site-wide Important notification: live demo, not a real government service -->
        <div class="govuk-grid-row">
          <div class="govuk-grid-column-two-thirds">
            <!-- One h1; page content via components -->
          </div>
        </div>
      </main>
    </div>
    <!-- Footer -->
    <script type="module" src="/assets/app.[fingerprint].mjs"></script>
  </body>
</html>
```

The `js-enabled` snippet must be that exact one line. The CSP hash in [`baseline/policy.json`](../baseline/policy.json) is computed from it. Whitespace changes block the script. `app.[fingerprint].mjs` is an external module that imports and calls `initAll()` — not a second inline script. See [frontend-security.md](frontend-security.md) and [frontend-performance.md](frontend-performance.md).

## Required order

1. JS detection script immediately after body open (`js-enabled` / `govuk-frontend-supported`)
2. Skip link as first focusable element
3. GOV.UK header (+ service navigation when needed)
4. Width container wrapping main content
5. Phase banner (Example / demonstration — not a live government service)
6. `main` with `id="main-content"` (must match skip-link href)
7. Site-wide **Important** notification banner (live demo warning), then page content
8. Grid row with appropriate column width
9. Footer (OGL + Crown copyright)

## Search indexing

This example must not appear in search results. The shell sets `<meta name="robots" content="noindex, nofollow">`. HTML responses also send `X-Robots-Tag: noindex, nofollow`. `/robots.txt` returns `Disallow: /` for all user agents.

## Frontend 6+ header rules

- Masthead is **blue** with white logotype (`fill="currentcolor"`).
- Homepage link class: `govuk-header__homepage-link` (not `govuk-header__link--homepage`).
- Logo `homepageUrl` is this service’s start page (`/` in English, `/cy` in Welsh) — not the GOV.UK public homepage. Layout passes it via `pageData.HomepageURL`.
- Service name lives in `govuk-service-navigation` under the masthead (and uses the same home path).
- Copy from the pinned Frontend templates/fixtures — not outdated blog examples.

## Layout patterns

| Pattern           | Use                                                 |
| ----------------- | --------------------------------------------------- |
| Two-thirds column | Main content; optimal reading line length           |
| Full width        | Forms, tables, simple content when appropriate      |
| Grid              | 12-column responsive grid                           |
| Sidebar           | One-third for secondary nav / supplementary content |
| Centered          | Single column within max-width container            |

Spacing and tokens: [design-tokens.md](design-tokens.md).
