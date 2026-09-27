import { createHash } from 'node:crypto';

import policy from './policy.json' with { type: 'json' };

const PUBLISHED_JS_ENABLED_HASH = 'sha256-GUQ5ad8JK5KmEWmROf3LZd9ge94daqNvd8xy9YS1iDw=';

const EXACT_HEADERS = {
  'X-Content-Type-Options': 'nosniff',
  'Referrer-Policy': 'strict-origin-when-cross-origin',
  'X-XSS-Protection': '0',
  'X-Permitted-Cross-Domain-Policies': 'none',
  'Cross-Origin-Resource-Policy': 'same-origin',
  'X-Frame-Options': 'DENY',
  'Cross-Origin-Opener-Policy': 'same-origin',
  'Cross-Origin-Embedder-Policy': 'require-corp',
};

const REQUIRED_REMOVE = [
  'Server',
  'X-Powered-By',
  'X-AspNet-Version',
  'X-AspNetMvc-Version',
  'Public-Key-Pins',
  'Public-Key-Pins-Report-Only',
  'Expect-CT',
];

const REQUIRED_CSP = {
  'default-src': "'self'",
  'script-src': "'self'",
  'script-src-attr': "'none'",
  'style-src': "'self'",
  'style-src-attr': "'none'",
  'img-src': "'self'",
  'font-src': "'self'",
  'connect-src': "'self'",
  'media-src': "'self'",
  'object-src': "'none'",
  'frame-src': "'none'",
  'frame-ancestors': "'none'",
  'base-uri': "'none'",
  'form-action': "'self'",
  'manifest-src': "'self'",
  'worker-src': "'self'",
};

const CACHE_CONTROL = {
  document: 'no-cache',
  'sensitive-document': 'no-store',
  'fingerprinted-asset': 'public, max-age=31536000, immutable',
  'static-asset': 'no-cache',
  download: 'private, no-cache',
  'sensitive-download': 'no-store',
};

export function scriptSourceHash(source) {
  if (typeof source !== 'string') {
    throw new TypeError('script source must be a string');
  }
  const digest = createHash('sha256').update(source, 'utf8').digest('base64');
  return `sha256-${digest}`;
}

export function assertBaselinePolicy(candidate) {
  if (!candidate || typeof candidate !== 'object') {
    throw new Error('baseline policy: expected an object');
  }

  const problems = [];
  if (candidate.version !== 1) problems.push('version must be 1');

  const snippet = candidate.jsEnabledSnippet;
  const hash = candidate.jsEnabledScriptHash;
  if (typeof snippet !== 'string' || scriptSourceHash(snippet) !== hash) {
    problems.push('js-enabled snippet hash');
  }
  if (hash !== PUBLISHED_JS_ENABLED_HASH) {
    problems.push('js-enabled hash drifted from GOV.UK Frontend docs');
  }

  const headerGroups = candidate.headers ?? {};
  const headers = { ...(headerGroups.all ?? {}), ...(headerGroups.document ?? {}) };
  for (const [name, value] of Object.entries(EXACT_HEADERS)) {
    if (headers[name] !== value) problems.push(`${name} must be ${value}`);
  }

  const remove = new Set(candidate.remove ?? []);
  for (const name of REQUIRED_REMOVE) {
    if (!remove.has(name)) problems.push(`must remove ${name}`);
  }

  const cacheControl = candidate.cacheControl ?? {};
  for (const [kind, value] of Object.entries(CACHE_CONTROL)) {
    if (cacheControl[kind] !== value) problems.push(`cache ${kind}`);
  }

  const csp = candidate.csp ?? {};
  const directives = csp.directives ?? {};
  for (const [name, value] of Object.entries(REQUIRED_CSP)) {
    const tokens = directives[name];
    if (!Array.isArray(tokens) || tokens.join(' ') !== value) problems.push(`csp ${name}`);
  }
  const upgrade = directives['upgrade-insecure-requests'];
  if (!Array.isArray(upgrade) || upgrade.length !== 0) {
    problems.push('csp upgrade-insecure-requests');
  }

  const hsts = candidate.hsts ?? {};
  if (hsts.maxAge !== 63072000 || hsts.includeSubDomains !== true) {
    problems.push('hsts');
  }

  const permissions = candidate.permissionsPolicy;
  if (!Array.isArray(permissions) || permissions.length < 8) {
    problems.push('permissions-policy');
  } else {
    const seen = new Set();
    for (const name of permissions) {
      if (typeof name !== 'string' || !/^[a-z0-9-]+$/.test(name) || seen.has(name)) {
        problems.push(`permission ${String(name)}`);
      }
      if (typeof name === 'string') seen.add(name);
    }
  }

  const performance = candidate.performance ?? {};
  const compression = performance.compression ?? {};
  if (compression.standard !== 'br' || compression.fallback !== 'gzip') {
    problems.push('compression');
  }
  const budgets = performance.budgets ?? {};
  if (budgets.thirdPartyRequests !== 0) {
    problems.push('third-party budget');
  }
  const vitals = performance.coreWebVitals ?? {};
  if (vitals.lcpMs !== 2500 || vitals.inpMs !== 200 || vitals.cls !== 0.1) {
    problems.push('core web vitals');
  }
  const nonce = candidate.nonce ?? {};
  if (nonce.bytes !== 16 || nonce.minLength !== 24) {
    problems.push('nonce');
  }
  const cookie = candidate.cookie ?? {};
  if (
    cookie.sameSite !== 'Lax' ||
    cookie.secure !== true ||
    cookie.httpOnly !== true ||
    cookie.path !== '/'
  ) {
    problems.push('cookie defaults');
  }

  if (problems.length > 0) {
    throw new Error(`baseline policy: ${problems.join('; ')}`);
  }
}

assertBaselinePolicy(policy);

export default policy;
