# Frontend security

Shared OWASP response-header baseline for this template and every language line that syncs from it. The machine-readable contract is [`baseline/policy.json`](../baseline/policy.json). This **Go** line applies it in [`internal/baseline`](../internal/baseline). The Node helpers under [`baseline/*.mjs`](../baseline/) are the shared oracle for tests; they are not used on the Go request path.

Authoritative sources:

- [OWASP HTTP Headers Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html)
- [OWASP Content Security Policy Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html)
- [OWASP Cross Site Scripting Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
- [GOV.UK Frontend: import JavaScript](https://frontend.design-system.service.gov.uk/import-javascript/) (CSP hash for the `js-enabled` snippet)

Performance cache rules live in [frontend-performance.md](frontend-performance.md). Encoding and `html` options stay in [priorities.md](priorities.md) and [creating-components.md](creating-components.md).

## Language lines

Sync the whole `baseline/` directory with this repo. Do not fork a weaker header set in the language line.

| Stack              | How to apply it                                                                                                    |
| ------------------ | ------------------------------------------------------------------------------------------------------------------ |
| **Go (this line)** | `internal/baseline` reads `policy.json` and applies headers on every response                                      |
| Node tooling       | `baseline/*.mjs` — shared test oracle (`npm run test:baseline`); optional reference `import` for other Node lines  |
| Any other language | Read `baseline/policy.json` and match `buildResponseHeaders`. Use the Node helper as the oracle when you add tests |

```sh
node --input-type=module -e "import { buildResponseHeaders } from './baseline/index.mjs'; console.log(JSON.stringify(buildResponseHeaders({ kind: 'document', secureTransport: true }), null, 2))"
```

## Response kinds

Pass `secureTransport: true` for HTTPS. Omit HSTS on plain HTTP (local preview). Preload is a separate opt-in: `hstsPreload: true` only after every name that would be pinned is HTTPS, because preload is a long-lived public commitment.

| `kind`                | When                                                       | `Cache-Control`                       |
| --------------------- | ---------------------------------------------------------- | ------------------------------------- |
| `document`            | Public HTML                                                | `no-cache`                            |
| `sensitive-document`  | Personal or authenticated HTML                             | `no-store`                            |
| `fingerprinted-asset` | CSS, JS, fonts, images whose URL changes when bytes change | `public, max-age=31536000, immutable` |
| `static-asset`        | Assets whose URL does not change                           | `no-cache`                            |
| `download`            | File download                                              | `private, no-cache`                   |
| `sensitive-download`  | Personal file download                                     | `no-store`                            |

If the HTML response sets a cookie, pass `setsCookie: true`. Public documents then use `private, no-cache`. Cookies are rejected on asset kinds so static files stay cookie-free.

```js
import { applyResponseHeaders, createNonce } from './baseline/index.mjs';

applyResponseHeaders(response, {
  kind: 'document',
  secureTransport: true,
});
```

`applyResponseHeaders` accepts a Node response (`setHeader` / `removeHeader`) or a Fetch-style `{ headers: { set, delete } }`.

## Headers on HTML documents

| Header                              | Value                                            |
| ----------------------------------- | ------------------------------------------------ |
| `Content-Type`                      | `text/html; charset=utf-8`                       |
| `Content-Security-Policy`           | Built from `policy.json` (below)                 |
| `Strict-Transport-Security`         | `max-age=63072000; includeSubDomains` on HTTPS   |
| `X-Content-Type-Options`            | `nosniff`                                        |
| `X-Frame-Options`                   | `DENY`                                           |
| `Referrer-Policy`                   | `strict-origin-when-cross-origin`                |
| `Permissions-Policy`                | Every feature in `permissionsPolicy` set to `()` |
| `Cross-Origin-Opener-Policy`        | `same-origin`                                    |
| `Cross-Origin-Embedder-Policy`      | `require-corp`                                   |
| `Cross-Origin-Resource-Policy`      | `same-origin`                                    |
| `X-Permitted-Cross-Domain-Policies` | `none`                                           |
| `X-XSS-Protection`                  | `0`                                              |
| `Vary`                              | `Accept-Encoding`                                |

Asset and download responses still send `X-Content-Type-Options`, `Referrer-Policy`, `Cross-Origin-Resource-Policy`, `X-Permitted-Cross-Domain-Policies`, `X-XSS-Protection`, and HSTS on HTTPS. They do not send CSP, `Permissions-Policy`, COOP, or COEP — those apply to documents, and repeating them on every file wastes bytes.

`Cross-Origin-Resource-Policy: same-origin` is stricter than the OWASP cheat sheet’s `same-site` example. This template self-hosts Frontend assets and does not share them with sibling origins. Pass `crossOriginResourcePolicy: 'same-site'` only when a sibling origin must embed those responses.

`Cross-Origin-Embedder-Policy: require-corp` blocks cross-origin resources that do not opt in. Pass `crossOriginEmbedderPolicy: 'credentialless'` only when a page must load a cross-origin resource that you do not control.

While adopting CSP, pass `enforceContentSecurityPolicy: false` to send `Content-Security-Policy-Report-Only` instead. Enforce the real header once violations are gone.

## Content Security Policy

The policy is `default-src 'self'` with:

- `script-src 'self'` plus the GOV.UK Frontend hash `sha256-GUQ5ad8JK5KmEWmROf3LZd9ge94daqNvd8xy9YS1iDw=`
- `script-src-attr 'none'` and `style-src-attr 'none'`
- `object-src 'none'`, `base-uri 'none'`, `frame-src 'none'`, `frame-ancestors 'none'`
- `form-action 'self'`, `upgrade-insecure-requests`
- no `'unsafe-inline'`, `'unsafe-eval'`, `*`, or `http:`

The hash matches this exact snippet and nothing else:

```html
<script>
  document.body.className +=
    ' js-enabled' + ('noModule' in HTMLScriptElement.prototype ? ' govuk-frontend-supported' : '');
</script>
```

Frontend’s [import JavaScript](https://frontend.design-system.service.gov.uk/import-javascript/) guide recommends that hash. A per-response nonce also works, but it makes the HTML unique and blocks shared caches. Prefer the hash.

Call `initAll()` from an external `<script type="module" src="…">` so it is covered by `'self'`. Do not add a second inline module. If you must, pass `scriptNonce` from `createNonce()` and set the same value on the Nunjucks `cspNonce` variable. The nonce is 16 random bytes, base64, unique per response.

Extra origins go in `cspSources` (for example `{ 'img-src': ['https://assets.example.gov.uk'] }`). The helper rejects `'unsafe-inline'`, `'unsafe-eval'`, `'unsafe-hashes'`, `*`, `http:`, and `data:` / `blob:` on script directives unless `allowUnsafeCspSources` is set. Turning that flag on means the response no longer matches this baseline.

## Cookies

```js
import { buildSetCookie } from './baseline/index.mjs';

buildSetCookie('__Host-session', token, { hostPrefix: true });
// __Host-session=…; Path=/; Secure; HttpOnly; SameSite=Lax
```

Defaults from `policy.json`: `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`. Use `SameSite=Strict` when the journey can tolerate it. `SameSite=None` requires `Secure`. `__Host-` names must be `Secure`, `Path=/`, and must not set `Domain`.

**Go line:** validate against those rules in `internal/baseline`, then serialize with `http.Cookie.Valid` / `String`. Read session cookies with `http.Request.Cookie`. Compare CSRF tokens with `crypto/subtle.ConstantTimeCompare`. See [tech-stack.md](tech-stack.md).

## Downloads and logout

Downloads send `Content-Disposition: attachment` plus `X-Content-Type-Options: nosniff`. The default type is `application/octet-stream`. Pass `filename` without a path.

On logout, send `Clear-Site-Data` from `clearSiteDataHeader()` (default `"cookies", "storage"`).

## Headers the platform must strip

`buildResponseHeaders` returns `remove`:

`Server`, `X-Powered-By`, `X-AspNet-Version`, `X-AspNetMvc-Version`, `Public-Key-Pins`, `Public-Key-Pins-Report-Only`, `Expect-CT`

Strip them in the app and again at the edge if a proxy adds `Server` after the app.

## Deliberately omitted

| Header                         | Why                                                                                                          |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------ |
| `X-DNS-Prefetch-Control: off`  | Prefetch helps first load, and this template controls its own links. Priority is performance, then security. |
| `Access-Control-Allow-Origin`  | Absent means the browser keeps the same-origin policy.                                                       |
| `Expect-CT`, `Public-Key-Pins` | OWASP says do not use them.                                                                                  |
| `Trusted Types`                | Do not require them until the pinned Frontend build is proven not to need `innerHTML` sinks.                 |
| `block-all-mixed-content`      | Use `upgrade-insecure-requests` instead.                                                                     |

## HTML

Prefer component `text` options. Treat `html` options as untrusted until sanitised. Match Nunjucks `escape` when not calling Nunjucks. Details: [creating-components.md](creating-components.md).
