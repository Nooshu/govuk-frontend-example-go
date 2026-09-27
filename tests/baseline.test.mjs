import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import {
  applyResponseHeaders,
  appendVary,
  assertBaselinePolicy,
  baselinePolicy,
  buildContentSecurityPolicy,
  buildPreloadLinkHeader,
  buildResponseHeaders,
  buildSetCookie,
  clearSiteDataHeader,
  createNonce,
  scriptSourceHash,
  strongEtag,
} from '../baseline/index.mjs';

const httpsDocument = { kind: 'document', secureTransport: true };

function clonePolicy(mutate) {
  const copy = structuredClone(baselinePolicy);
  mutate(copy);
  return copy;
}

describe('baseline policy', () => {
  it('accepts the shipped contract', () => {
    assert.equal(assertBaselinePolicy(baselinePolicy), undefined);
    assert.equal(baselinePolicy.version, 1);
    assert.equal(baselinePolicy.performance.budgets.thirdPartyRequests, 0);
    assert.equal(baselinePolicy.performance.budgets.cssCompressedBytes, 32768);
    assert.equal(baselinePolicy.performance.budgets.jsCompressedBytes, 32768);
    assert.equal(baselinePolicy.performance.compression.standard, 'br');
    assert.equal(baselinePolicy.performance.compression.fallback, 'gzip');
    assert.equal(baselinePolicy.performance.httpMinimum, 'h2');
    assert.equal(baselinePolicy.performance.httpPreferred, 'h3');
  });

  it('hashes the GOV.UK Frontend js-enabled snippet to the published digest', () => {
    assert.equal(
      scriptSourceHash(baselinePolicy.jsEnabledSnippet),
      baselinePolicy.jsEnabledScriptHash,
    );
    assert.equal(
      baselinePolicy.jsEnabledScriptHash,
      'sha256-GUQ5ad8JK5KmEWmROf3LZd9ge94daqNvd8xy9YS1iDw=',
    );
  });

  it('rejects a non-string script source', () => {
    assert.throws(() => scriptSourceHash(undefined), /script source must be a string/);
  });

  it('rejects a policy version other than 1', () => {
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => (policy.version = 2))),
      /version must be 1/,
    );
  });

  it('rejects values that are not policy objects', () => {
    assert.throws(() => assertBaselinePolicy(null), /expected an object/);
    assert.throws(() => assertBaselinePolicy('policy'), /expected an object/);
  });

  it('rejects a snippet that does not match its hash', () => {
    const copy = clonePolicy((policy) => {
      policy.jsEnabledSnippet = `${policy.jsEnabledSnippet} `;
    });
    assert.throws(() => assertBaselinePolicy(copy), /js-enabled snippet hash/);
  });

  it('rejects a hash that drifts from the Frontend docs', () => {
    const copy = clonePolicy((policy) => {
      policy.jsEnabledSnippet = 'var stayed = true;';
      policy.jsEnabledScriptHash = scriptSourceHash('var stayed = true;');
    });
    assert.throws(() => assertBaselinePolicy(copy), /drifted from GOV.UK Frontend docs/);
  });

  it('rejects a non-string snippet', () => {
    const copy = clonePolicy((policy) => {
      policy.jsEnabledSnippet = 12;
    });
    assert.throws(() => assertBaselinePolicy(copy), /js-enabled snippet hash/);
  });

  it('rejects a weakened header, removal list, or cache policy', () => {
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => delete policy.headers)),
      /X-Content-Type-Options/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.remove = policy.remove.filter((name) => name !== 'Server');
          }),
        ),
      /must remove Server/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => delete policy.cacheControl)),
      /cache document/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.cacheControl.document = 'public, max-age=60';
          }),
        ),
      /cache document/,
    );
  });

  it('rejects a weakened content security policy', () => {
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => delete policy.csp)),
      /csp default-src/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.csp.directives['default-src'] = ['*'];
          }),
        ),
      /csp default-src/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.csp.directives['upgrade-insecure-requests'] = ["'self'"];
          }),
        ),
      /upgrade-insecure-requests/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            delete policy.csp.directives['upgrade-insecure-requests'];
          }),
        ),
      /upgrade-insecure-requests/,
    );
  });

  it('rejects weakened HSTS, permissions, compression, vitals, nonce, or cookies', () => {
    assert.throws(() => assertBaselinePolicy(clonePolicy((policy) => delete policy.hsts)), /hsts/);
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.hsts.includeSubDomains = false;
          }),
        ),
      /hsts/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => (policy.permissionsPolicy = ['camera']))),
      /permissions-policy/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => (policy.permissionsPolicy = null))),
      /permissions-policy/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => delete policy.remove)),
      /must remove Server/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => delete policy.cookie)),
      /cookie defaults/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.permissionsPolicy = [...policy.permissionsPolicy, 1, 'camera', 'Nope'];
          }),
        ),
      /permission camera/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => delete policy.performance)),
      /compression/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.performance.compression = { standard: 'gzip', fallback: 'gzip' };
          }),
        ),
      /compression/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.performance.budgets.thirdPartyRequests = 2;
          }),
        ),
      /third-party budget/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.performance.coreWebVitals.inpMs = 500;
          }),
        ),
      /core web vitals/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.performance.compression = { standard: 'br', fallback: 'identity' };
          }),
        ),
      /compression/,
    );
    assert.throws(
      () =>
        assertBaselinePolicy(
          clonePolicy((policy) => {
            policy.performance.coreWebVitals.cls = 1;
          }),
        ),
      /core web vitals/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => delete policy.nonce)),
      /nonce/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => (policy.nonce.minLength = 8))),
      /nonce/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => (policy.cookie.secure = false))),
      /cookie defaults/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => (policy.cookie.sameSite = 'None'))),
      /cookie defaults/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => (policy.cookie.httpOnly = false))),
      /cookie defaults/,
    );
    assert.throws(
      () => assertBaselinePolicy(clonePolicy((policy) => (policy.cookie.path = '/app'))),
      /cookie defaults/,
    );
  });
});

describe('response headers', () => {
  it('sets the OWASP document baseline on HTTPS', () => {
    const { headers, remove } = buildResponseHeaders(httpsDocument);
    assert.equal(headers['Content-Type'], 'text/html; charset=utf-8');
    assert.equal(headers['Cache-Control'], 'no-cache');
    assert.equal(headers['X-Content-Type-Options'], 'nosniff');
    assert.equal(headers['X-Frame-Options'], 'DENY');
    assert.equal(headers['Referrer-Policy'], 'strict-origin-when-cross-origin');
    assert.equal(headers['X-XSS-Protection'], '0');
    assert.equal(headers['X-Permitted-Cross-Domain-Policies'], 'none');
    assert.equal(headers['Cross-Origin-Opener-Policy'], 'same-origin');
    assert.equal(headers['Cross-Origin-Embedder-Policy'], 'require-corp');
    assert.equal(headers['Cross-Origin-Resource-Policy'], 'same-origin');
    assert.equal(headers['Strict-Transport-Security'], 'max-age=63072000; includeSubDomains');
    assert.equal(headers.Vary, 'Accept-Encoding');
    assert.equal(remove.includes('X-Powered-By'), true);
    assert.equal(remove.includes('Server'), true);
    assert.equal(remove.includes('Expect-CT'), true);
    assert.equal(remove.includes('Public-Key-Pins'), true);

    const csp = headers['Content-Security-Policy'];
    assert.match(csp, /default-src 'self'/);
    assert.match(csp, /script-src 'self' 'sha256-GUQ5ad8JK5KmEWmROf3LZd9ge94daqNvd8xy9YS1iDw='/);
    assert.match(csp, /script-src-attr 'none'/);
    assert.match(csp, /object-src 'none'/);
    assert.match(csp, /base-uri 'none'/);
    assert.match(csp, /frame-ancestors 'none'/);
    assert.match(csp, /form-action 'self'/);
    assert.match(csp, /upgrade-insecure-requests/);
    assert.equal(csp.includes('unsafe-inline'), false);
    assert.equal(csp.includes('unsafe-eval'), false);
    assert.match(headers['Permissions-Policy'], /camera=\(\)/);
    assert.match(headers['Permissions-Policy'], /geolocation=\(\)/);
    assert.equal(headers['Content-Security-Policy-Report-Only'], undefined);
  });

  it('omits HSTS on HTTP and can opt into preload on HTTPS', () => {
    const http = buildResponseHeaders({ kind: 'document', secureTransport: false });
    assert.equal(http.headers['Strict-Transport-Security'], undefined);
    const preloaded = buildResponseHeaders({ ...httpsDocument, hstsPreload: true });
    assert.equal(
      preloaded.headers['Strict-Transport-Security'],
      'max-age=63072000; includeSubDomains; preload',
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'document', secureTransport: false, hstsPreload: true }),
      /HSTS preload requires secureTransport/,
    );
  });

  it('keeps personalised HTML out of shared caches', () => {
    const sensitive = buildResponseHeaders({
      kind: 'sensitive-document',
      secureTransport: true,
      setsCookie: true,
    });
    assert.equal(sensitive.headers['Cache-Control'], 'no-store');
    assert.match(sensitive.headers['Content-Security-Policy'], /script-src /);
    const withCookie = buildResponseHeaders({ ...httpsDocument, setsCookie: true });
    assert.equal(withCookie.headers['Cache-Control'], 'private, no-cache');
    assert.throws(
      () =>
        buildResponseHeaders({
          kind: 'fingerprinted-asset',
          secureTransport: true,
          setsCookie: true,
        }),
      /Set-Cookie belongs on HTML documents/,
    );
  });

  it('caches fingerprinted assets immutably and revalidates everything else', () => {
    const fingerprinted = buildResponseHeaders({
      kind: 'fingerprinted-asset',
      secureTransport: true,
      contentType: 'text/css',
    });
    assert.equal(fingerprinted.headers['Cache-Control'], 'public, max-age=31536000, immutable');
    assert.equal(fingerprinted.headers['Content-Type'], 'text/css; charset=utf-8');
    assert.equal(fingerprinted.headers['Content-Security-Policy'], undefined);
    assert.equal(fingerprinted.headers['X-Frame-Options'], undefined);
    assert.equal(fingerprinted.headers['X-Content-Type-Options'], 'nosniff');

    const plain = buildResponseHeaders({ kind: 'static-asset', secureTransport: true });
    assert.equal(plain.headers['Cache-Control'], 'no-cache');
    assert.equal(plain.headers['Content-Type'], undefined);
  });

  it('forces downloads to be attachments', () => {
    const file = buildResponseHeaders({
      kind: 'download',
      secureTransport: true,
      filename: 'café notes.pdf',
    });
    assert.equal(file.headers['Cache-Control'], 'private, no-cache');
    assert.equal(file.headers['Content-Type'], 'application/octet-stream');
    assert.equal(
      file.headers['Content-Disposition'],
      'attachment; filename="caf_ notes.pdf"; filename*=UTF-8\'\'caf%C3%A9%20notes.pdf',
    );
    const secret = buildResponseHeaders({
      kind: 'sensitive-download',
      secureTransport: true,
      filename: 'evidence.bin',
      contentType: 'application/octet-stream',
    });
    assert.equal(secret.headers['Cache-Control'], 'no-store');
  });

  it('accepts caller content types and referrer or isolation overrides', () => {
    const html = buildResponseHeaders({
      ...httpsDocument,
      contentType: 'text/html; charset=UTF-8',
      referrerPolicy: 'no-referrer',
      crossOriginResourcePolicy: 'same-site',
      crossOriginEmbedderPolicy: 'credentialless',
      enforceContentSecurityPolicy: false,
      vary: ['Cookie'],
    });
    assert.equal(html.headers['Content-Type'], 'text/html; charset=UTF-8');
    assert.equal(html.headers['Referrer-Policy'], 'no-referrer');
    assert.equal(html.headers['Cross-Origin-Resource-Policy'], 'same-site');
    assert.equal(html.headers['Cross-Origin-Embedder-Policy'], 'credentialless');
    assert.equal(html.headers['Content-Security-Policy'], undefined);
    assert.match(html.headers['Content-Security-Policy-Report-Only'], /default-src 'self'/);
    assert.equal(html.headers.Vary, 'Accept-Encoding, Cookie');

    const image = buildResponseHeaders({
      kind: 'static-asset',
      secureTransport: true,
      contentType: 'image/png',
    });
    assert.equal(image.headers['Content-Type'], 'image/png');
    const json = buildResponseHeaders({
      kind: 'static-asset',
      secureTransport: true,
      contentType: 'application/json',
    });
    assert.equal(json.headers['Content-Type'], 'application/json; charset=utf-8');
  });

  it('extends CSP and permissions without allowing unsafe defaults', () => {
    const nonce = createNonce();
    const csp = buildContentSecurityPolicy({
      cspSources: { 'img-src': ['https://images.example.com'] },
      scriptHashes: ['sha256-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa='],
      scriptNonce: nonce,
    });
    assert.match(csp, /img-src 'self' https:\/\/images\.example\.com/);
    assert.equal(csp.includes(`'nonce-${nonce}'`), true);
    assert.match(csp, /sha256-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa=/);
    assert.match(csp, /sha256-GUQ5ad8JK5KmEWmROf3LZd9ge94daqNvd8xy9YS1iDw=/);

    const allowed = buildResponseHeaders({
      ...httpsDocument,
      permissionsAllow: { camera: 'self', microphone: 'none', usb: 'https://kiosk.example.com' },
      preload: [{ href: '/assets/font.woff2', as: 'font', type: 'font/woff2' }],
    });
    assert.match(allowed.headers['Permissions-Policy'], /camera=\(self\)/);
    assert.match(allowed.headers['Permissions-Policy'], /microphone=\(\)/);
    assert.match(allowed.headers['Permissions-Policy'], /usb=\(https:\/\/kiosk\.example\.com\)/);
    assert.match(allowed.headers.Link, /rel=preload; as=font/);

    const unsafe = buildContentSecurityPolicy({
      cspSources: { 'script-src': ["'unsafe-inline'"] },
      allowUnsafeCspSources: true,
    });
    assert.match(unsafe, /script-src 'self' 'unsafe-inline'/);
  });

  it('rejects invalid caller input', () => {
    assert.throws(() => buildResponseHeaders(null), /options object is required/);
    assert.throws(
      () => buildResponseHeaders({ kind: 'page', secureTransport: true }),
      /unknown response kind/,
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'document' }),
      /secureTransport boolean is required/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, setsCookie: 'yes' }),
      /setsCookie must be a boolean/,
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'static-asset', secureTransport: true, scriptNonce: 'x' }),
      /scriptNonce applies to HTML documents/,
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'static-asset', secureTransport: true, cspSources: {} }),
      /cspSources applies to HTML documents/,
    );
    assert.throws(
      () =>
        buildResponseHeaders({
          kind: 'static-asset',
          secureTransport: true,
          enforceContentSecurityPolicy: false,
        }),
      /CSP report-only applies to HTML documents/,
    );
    assert.throws(
      () =>
        buildResponseHeaders({ kind: 'static-asset', secureTransport: true, permissionsAllow: {} }),
      /permissionsAllow applies to HTML documents/,
    );
    assert.throws(
      () =>
        buildResponseHeaders({
          kind: 'static-asset',
          secureTransport: true,
          crossOriginEmbedderPolicy: 'require-corp',
        }),
      /crossOriginEmbedderPolicy applies to HTML documents/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, filename: 'a.pdf' }),
      /filename applies to download responses/,
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'download', secureTransport: true }),
      /download filename is required/,
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'download', secureTransport: true, filename: '' }),
      /download filename is required/,
    );
    assert.throws(
      () =>
        buildResponseHeaders({
          kind: 'download',
          secureTransport: true,
          filename: '../secret.pdf',
        }),
      /must not include a path/,
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'download', secureTransport: true, filename: 'a\\b.pdf' }),
      /must not include a path/,
    );
    assert.throws(
      () =>
        buildResponseHeaders({ kind: 'download', secureTransport: true, filename: 'foo..bar.pdf' }),
      /must not include a path/,
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'download', secureTransport: true, filename: 1 }),
      /download filename is required/,
    );
    assert.throws(
      () => buildResponseHeaders({ kind: 'download', secureTransport: true, filename: 'a"b.pdf' }),
      /invalid download filename/,
    );
    assert.throws(
      () =>
        buildResponseHeaders({
          kind: 'download',
          secureTransport: true,
          filename: `${'a'.repeat(181)}.pdf`,
        }),
      /download filename is required/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, contentType: 'text/html\r\nX: 1' }),
      /invalid Content-Type/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, contentType: '' }),
      /invalid Content-Type/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, referrerPolicy: 'unsafe-url' }),
      /invalid Referrer-Policy/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, crossOriginResourcePolicy: 'anywhere' }),
      /invalid Cross-Origin-Resource-Policy/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, crossOriginEmbedderPolicy: 'maybe' }),
      /invalid Cross-Origin-Embedder-Policy/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, vary: 'Cookie' }),
      /vary must be an array/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, vary: ['Cookie\n'] }),
      /invalid Vary token/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, vary: [1] }),
      /invalid Vary token/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: [] }),
      /cspSources must be an object/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'img-src': 'https://cdn.example' } }),
      /img-src sources must be an array/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'child-src': ['https://cdn.example'] } }),
      /unknown CSP directive/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'upgrade-insecure-requests': ["'self'"] } }),
      /does not take sources/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'script-src': ["'unsafe-inline'"] } }),
      /not allowed on script-src/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'script-src': ['*'] } }),
      /not allowed on script-src/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'script-src': ['http://cdn.example'] } }),
      /not allowed on script-src/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'script-src': ['http:'] } }),
      /not allowed on script-src/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'script-src': ['data:'] } }),
      /not allowed on script-src/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'script-src': ['blob:'] } }),
      /not allowed on script-src/,
    );
    assert.throws(
      () =>
        buildContentSecurityPolicy({
          cspSources: { 'img-src': ['https://images.example.com/a b'] },
        }),
      /invalid CSP source/,
    );
    assert.throws(() => buildContentSecurityPolicy({ scriptNonce: 'short' }), /scriptNonce/);
    assert.throws(() => buildContentSecurityPolicy({ scriptNonce: 12 }), /scriptNonce/);
    assert.throws(
      () => buildContentSecurityPolicy({ scriptNonce: `${'a'.repeat(30)}!` }),
      /scriptNonce/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'script-src': ["'unsafe-eval'"] } }),
      /not allowed on script-src/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'script-src': ["'unsafe-hashes'"] } }),
      /not allowed on script-src/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'img-src': [''] } }),
      /invalid CSP source/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ cspSources: { 'img-src': [1] } }),
      /invalid CSP source/,
    );
    assert.throws(
      () => buildContentSecurityPolicy({ scriptHashes: ['md5-abc'] }),
      /invalid script hash/,
    );
    assert.throws(() => buildContentSecurityPolicy({ scriptHashes: [12] }), /invalid script hash/);
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, contentType: 1 }),
      /invalid Content-Type/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, permissionsAllow: [] }),
      /permissionsAllow must be an object/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, permissionsAllow: null }),
      /permissionsAllow must be an object/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, permissionsAllow: { camera: 1 } }),
      /invalid permissions allow/,
    );
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, permissionsAllow: { camera: '*' } }),
      /invalid permissions allow/,
    );
    const cleared = buildResponseHeaders({ ...httpsDocument, permissionsAllow: { camera: '()' } });
    assert.match(cleared.headers['Permissions-Policy'], /camera=\(\)/);
    assert.throws(
      () => buildResponseHeaders({ ...httpsDocument, permissionsAllow: { flashlight: 'self' } }),
      /unknown permissions policy feature/,
    );
  });

  it('applies headers through Node and Fetch-style responses', () => {
    const node = {
      set: [],
      removed: [],
      setHeader(name, value) {
        this.set.push([name, value]);
      },
      removeHeader(name) {
        this.removed.push(name);
      },
    };
    const applied = applyResponseHeaders(node, httpsDocument);
    assert.equal(applied.headers['X-Frame-Options'], 'DENY');
    assert.equal(node.removed.includes('X-Powered-By'), true);
    assert.equal(
      node.set.some(([name]) => name === 'Content-Security-Policy'),
      true,
    );

    const withoutRemove = { setHeader() {} };
    applyResponseHeaders(withoutRemove, { kind: 'static-asset', secureTransport: false });

    const fetchHeaders = {
      values: new Map(),
      set(name, value) {
        this.values.set(name, value);
      },
      delete(name) {
        this.values.delete(name);
      },
    };
    applyResponseHeaders({ headers: fetchHeaders }, httpsDocument);
    assert.equal(fetchHeaders.values.get('X-Content-Type-Options'), 'nosniff');
    assert.equal(fetchHeaders.values.has('Server'), false);

    const setOnly = { headers: { set() {} } };
    applyResponseHeaders(setOnly, { kind: 'static-asset', secureTransport: true });

    assert.throws(
      () => applyResponseHeaders({}, httpsDocument),
      /setHeader\(\) or headers\.set\(\)/,
    );
    assert.throws(
      () => applyResponseHeaders({ headers: {} }, httpsDocument),
      /setHeader\(\) or headers\.set\(\)/,
    );
    assert.throws(
      () => applyResponseHeaders(null, httpsDocument),
      /setHeader\(\) or headers\.set\(\)/,
    );
  });
});

describe('cookies, preload, and logout', () => {
  it('builds a Secure HttpOnly session cookie', () => {
    assert.equal(
      buildSetCookie('session', 'abc'),
      'session=abc; Path=/; Secure; HttpOnly; SameSite=Lax',
    );
    assert.equal(
      buildSetCookie('seen', '', { maxAge: 0, sameSite: 'Strict', path: '/start' }),
      'seen=; Max-Age=0; Path=/start; Secure; HttpOnly; SameSite=Strict',
    );
    assert.equal(
      buildSetCookie('__Host-session', 'abc', { hostPrefix: true }),
      '__Host-session=abc; Path=/; Secure; HttpOnly; SameSite=Lax',
    );
    assert.equal(
      buildSetCookie('prefs', 'a', { secure: false, httpOnly: false, domain: 'example.com' }),
      'prefs=a; Domain=example.com; Path=/; SameSite=Lax',
    );
    assert.match(
      buildSetCookie('track', 'a', { sameSite: 'None' }),
      /Secure; HttpOnly; SameSite=None/,
    );
  });

  it('rejects unsafe cookie attributes', () => {
    assert.throws(() => buildSetCookie('bad name', 'a'), /invalid cookie name/);
    assert.throws(() => buildSetCookie(1, 'a'), /invalid cookie name/);
    assert.throws(() => buildSetCookie('ok', 'a b'), /invalid cookie value/);
    assert.throws(() => buildSetCookie('ok', 1), /invalid cookie value/);
    assert.throws(() => buildSetCookie('ok', 'a', null), /cookie options must be an object/);
    assert.throws(() => buildSetCookie('ok', 'a', []), /cookie options must be an object/);
    assert.throws(() => buildSetCookie('ok', 'a', { sameSite: 'lax' }), /invalid SameSite/);
    assert.throws(() => buildSetCookie('ok', 'a', { secure: 'yes' }), /must be booleans/);
    assert.throws(() => buildSetCookie('ok', 'a', { httpOnly: 'yes' }), /must be booleans/);
    assert.throws(
      () => buildSetCookie('ok', 'a', { sameSite: 'None', secure: false }),
      /SameSite=None requires Secure/,
    );
    assert.throws(() => buildSetCookie('ok', 'a', { path: 'start' }), /Path must start/);
    assert.throws(
      () => buildSetCookie('session', 'a', { hostPrefix: true }),
      /hostPrefix requires/,
    );
    assert.throws(
      () => buildSetCookie('__Host-session', 'a', { secure: false }),
      /__Host- cookies require Secure/,
    );
    assert.throws(
      () => buildSetCookie('__Host-session', 'a', { path: '/app' }),
      /__Host- cookies require Path/,
    );
    assert.throws(
      () => buildSetCookie('__Host-session', 'a', { domain: 'example.com' }),
      /must not set Domain/,
    );
    assert.throws(() => buildSetCookie('ok', 'a', { domain: 'ex ample' }), /invalid Domain/);
    assert.throws(() => buildSetCookie('ok', 'a', { maxAge: -1 }), /maxAge/);
    assert.throws(() => buildSetCookie('ok', 'a', { maxAge: 1.5 }), /maxAge/);
  });

  it('builds preload and vary headers for same-origin assets', () => {
    assert.equal(
      buildPreloadLinkHeader([
        { href: '/assets/app.css', as: 'style', type: 'text/css' },
        { href: '/assets/app.mjs', as: 'script', crossorigin: true },
      ]),
      '</assets/app.css>; rel=preload; as=style; type="text/css", </assets/app.mjs>; rel=preload; as=script; crossorigin',
    );
    assert.match(buildPreloadLinkHeader([{ href: '/font.woff2', as: 'font' }]), /crossorigin/);
    assert.equal(appendVary(undefined, 'Accept-Encoding'), 'Accept-Encoding');
    assert.equal(
      appendVary('accept-encoding, Cookie', 'Accept-Encoding'),
      'accept-encoding, Cookie',
    );
    assert.equal(appendVary('Accept-Encoding,', 'Cookie'), 'Accept-Encoding, Cookie');
    assert.throws(() => appendVary(1, 'Cookie'), /existing Vary must be a string/);
    assert.throws(() => appendVary('', 'Bad Token'), /invalid Vary token/);
    assert.throws(() => buildPreloadLinkHeader([]), /non-empty array/);
    assert.throws(() => buildPreloadLinkHeader('nope'), /non-empty array/);
    assert.throws(() => buildPreloadLinkHeader([null]), /same-origin path/);
    assert.throws(
      () => buildPreloadLinkHeader([{ href: 'https://cdn.example/a.css', as: 'style' }]),
      /same-origin path/,
    );
    assert.throws(
      () => buildPreloadLinkHeader([{ href: '//cdn.example/a.css', as: 'style' }]),
      /same-origin path/,
    );
    assert.throws(
      () => buildPreloadLinkHeader([{ href: '/a b.css', as: 'style' }]),
      /same-origin path/,
    );
    assert.throws(
      () => buildPreloadLinkHeader([{ href: '/app.js', as: 'document' }]),
      /unsupported preload as/,
    );
    assert.throws(
      () => buildPreloadLinkHeader([{ href: '/app.js', as: 'script', type: 'javascript' }]),
      /invalid preload type/,
    );
  });

  it('builds a Clear-Site-Data header for logout', () => {
    assert.equal(clearSiteDataHeader(), '"cookies", "storage"');
    assert.equal(clearSiteDataHeader(['cache', '*']), '"cache", "*"');
    assert.throws(() => clearSiteDataHeader([]), /non-empty array/);
    assert.throws(() => clearSiteDataHeader('cookies'), /non-empty array/);
    assert.throws(() => clearSiteDataHeader(['localStorage']), /invalid Clear-Site-Data type/);
  });

  it('builds a strong validator from the response body', () => {
    const fromText = strongEtag('hello');
    const fromBytes = strongEtag(new Uint8Array(Buffer.from('hello')));
    assert.equal(fromText, fromBytes);
    assert.match(fromText, /^"[A-Za-z0-9_-]+"$/);
    assert.notEqual(strongEtag(''), fromText);
    assert.throws(() => strongEtag(12), /string or bytes/);
  });
});
