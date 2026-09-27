# Frontend performance

Shared performance baseline for this template and every language line that syncs from it. Cache and preload behaviour is enforced by [`baseline/`](../baseline/). Budgets and placement rules are in [`baseline/policy.json`](../baseline/policy.json).

Authoritative sources:

- [How to test frontend performance](https://www.gov.uk/service-manual/technology/how-to-test-frontend-performance) (Service Manual)
- [How to optimise frontend performance](https://gds-way.digital.cabinet-office.gov.uk/standards/optimise-frontend-perf.html) (GDS Way)
- [Include CSS](https://frontend.design-system.service.gov.uk/include-css/) and [import JavaScript](https://frontend.design-system.service.gov.uk/import-javascript/)

Security headers for the same responses: [frontend-security.md](frontend-security.md).

## Language lines

Sync `baseline/` with this repo. Node calls `buildResponseHeaders`. Other languages must use the same `kind` values and the same `Cache-Control` strings from `policy.json`.

## What GDS ranks highest

From the [GDS Way](https://gds-way.digital.cabinet-office.gov.uk/standards/optimise-frontend-perf.html), do these before lower-priority work:

| Priority | Practice                                | How this template does it                                                                      |
| -------- | --------------------------------------- | ---------------------------------------------------------------------------------------------- |
| High     | CSS in the document head, defer scripts | Page template: stylesheet in `<head>`, Frontend JS as `<script type="module">` (modules defer) |
| High     | Minify, then compress with Brotli       | Compile `styles/application.scss` to CSS, then Brotli (`br`) as the compression standard       |
| High     | Fingerprint assets and cache them       | `kind: 'fingerprinted-asset'` sends `public, max-age=31536000, immutable`                      |
| High     | Ship only the CSS and JS the page needs | Import the Frontend modules you use. No second UI framework                                    |
| High     | Optimise images                         | Width and height set; modern formats; no third-party image hosts by default                    |
| Medium   | Fewer cross-origin connections          | `performance.budgets.thirdPartyRequests` is `0`                                                |
| Low      | Small cookies, HTTP/2 or HTTP/3         | Cookie-free asset responses; HTTP/2 minimum, HTTP/3 preferred                                  |

Brotli (`Content-Encoding: br`) is the compression standard for HTML, CSS, JavaScript, and other compressible responses. The HTTP server or edge compresses; the app does not set `Content-Encoding`. When the client sends `Accept-Encoding: br`, respond with Brotli. Offer Gzip only when the client does not advertise `br`.

The helper sets `Vary: Accept-Encoding` so caches store the Brotli, Gzip, and uncompressed copies separately. `policy.json` records this as `compression.standard: "br"` and `compression.fallback: "gzip"`.

## Asset placement

1. Stylesheet `<link>` in `<head>` pointing at the **compiled** Sass output (`dist/stylesheets/application.css` after `npm run build:styles`), so the first paint uses GOV.UK Frontend CSS plus any cascade overrides. Do not link the prebuilt `govuk-frontend.min.css` as the long-term source — see [styles.md](styles.md).
2. The `js-enabled` snippet immediately after `<body>` opens. It must be the one-line string in `policy.json` so the CSP hash stays valid. It is tiny and has to run before enhanced components paint.
3. One external module before `</body>` that imports `initAll` (or `createAll` for the components on that page) and calls it. Do not add a second inline script. An external file is cached and stays inside `script-src 'self'`.

```html
<script type="module" src="/assets/app.[fingerprint].mjs"></script>
```

```js
import { initAll } from './govuk-frontend.min.js';

initAll();
```

Import only the components a page needs when that is smaller than `initAll()`. UMD bundles are larger — avoid them when the page can use modules.

## Cache kinds

| `kind`                            | `Cache-Control`                       | Why                                                                                                                                         |
| --------------------------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `document`                        | `no-cache`                            | Browsers revalidate HTML so deploys pick up new asset URLs. Pair with a strong `ETag` (`strongEtag(body)`) so unchanged HTML can return 304 |
| `sensitive-document`              | `no-store`                            | Personal pages must not be stored                                                                                                           |
| `fingerprinted-asset`             | `public, max-age=31536000, immutable` | The URL changes when the bytes change                                                                                                       |
| `static-asset`                    | `no-cache`                            | Unhashed URLs must not sit in cache across deploys                                                                                          |
| `download` / `sensitive-download` | `private, no-cache` / `no-store`      | Files are not public site chrome                                                                                                            |

`no-cache` still allows storage; it requires revalidation. That is what you want for public HTML. It is not the right header for personal data — use `sensitive-document`.

Pass `setsCookie: true` on HTML that sets a cookie so shared caches do not store it. Asset kinds reject cookies.

Fingerprint real files (`app.a1b2c3d4.mjs`). `govuk-frontend.min.js` on its own is not fingerprinted; put a hash in the URL you serve, or keep it on `static-asset` so it revalidates.

## Preload

Preload only same-origin files the page uses immediately. Fonts always get `crossorigin`, which they need for the preload to match the CSS request.

```js
import { buildPreloadLinkHeader } from './baseline/index.mjs';

buildPreloadLinkHeader([
  { href: '/assets/fonts/light.[fingerprint].woff2', as: 'font', type: 'font/woff2' },
]);
```

Or pass `preload` to `buildResponseHeaders` and it sets `Link`. Do not preload third-party hosts. Revisit font preload when you upgrade Frontend — paths change between releases.

## Budgets

`policy.json` `performance.budgets` is the starting gate:

| Budget               | Value | Notes                                                                                                              |
| -------------------- | ----- | ------------------------------------------------------------------------------------------------------------------ |
| Brotli HTML          | 50KB  | Re-measure on a real page, compressed with Brotli                                                                  |
| Brotli CSS           | 32KB  | Headroom over a full Frontend Sass compile (about 13KB Brotli when this budget was set; re-measure after upgrades) |
| Brotli JS            | 32KB  | Headroom over the full `govuk-frontend.min.js` bundle (about 10KB Brotli when this budget was set)                 |
| Third-party requests | 0     | No analytics, fonts, or widgets from other origins                                                                 |

Core Web Vitals targets in the same file are the usual “good” thresholds: LCP 2.5s, INP 200ms, CLS 0.1. The Service Manual does not publish those numbers; it tells you to benchmark, then optimise. Re-measure Brotli CSS and JS after every Frontend upgrade.

## Measure

Use more than one tool, on a low-powered device, not only a developer laptop:

- [Lighthouse](https://developer.chrome.com/docs/lighthouse/overview/)
- [PageSpeed Insights](https://pagespeed.web.dev/)
- [WebPageTest](https://www.webpagetest.org/)
- Chrome DevTools network and performance panels
- [Sitespeed.io](https://www.sitespeed.io/)

Service Manual workflow: benchmark, fix the bottleneck, measure again. GDS Way also lists SpeedCurve and Calibre for scheduled runs.

Check a repeat view as well as a first view. Repeat view is where fingerprinted immutable caching shows up.
