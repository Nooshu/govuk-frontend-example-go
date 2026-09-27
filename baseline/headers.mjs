import { createHash, randomBytes } from 'node:crypto';

import { buildPreloadLinkHeader } from './performance.mjs';
import policy from './policy.mjs';

const DOCUMENT_KINDS = new Set(['document', 'sensitive-document']);
const DOWNLOAD_KINDS = new Set(['download', 'sensitive-download']);

const REFERRER_POLICIES = new Set([
  'no-referrer',
  'same-origin',
  'origin',
  'strict-origin',
  'origin-when-cross-origin',
  'strict-origin-when-cross-origin',
]);

const CORP_VALUES = new Set(['same-origin', 'same-site', 'cross-origin']);
const COEP_VALUES = new Set(['require-corp', 'credentialless', 'unsafe-none']);
const NONCE_PATTERN = /^[A-Za-z0-9+/_-]{22,}={0,2}$/;
const HASH_PATTERN = /^sha256-[A-Za-z0-9+/]+=*$/;

export function createNonce() {
  return randomBytes(policy.nonce.bytes).toString('base64');
}

export function strongEtag(body) {
  const bytes = etagBytes(body);
  const digest = createHash('sha256').update(bytes).digest('base64url');
  return `"${digest}"`;
}

export function buildContentSecurityPolicy(options = {}) {
  const extras = options.cspSources ?? {};
  if (typeof extras !== 'object' || Array.isArray(extras)) {
    throw new TypeError('cspSources must be an object');
  }

  const known = new Set(Object.keys(policy.csp.directives));
  for (const name of Object.keys(extras)) {
    if (!known.has(name)) throw new TypeError(`unknown CSP directive: ${name}`);
  }

  const parts = [];
  for (const [name, tokens] of Object.entries(policy.csp.directives)) {
    const extra = extras[name] ?? [];
    if (!Array.isArray(extra)) throw new TypeError(`${name} sources must be an array`);
    if (tokens.length === 0) {
      if (extra.length > 0) throw new TypeError(`${name} does not take sources`);
      parts.push(name);
      continue;
    }

    const merged = [...tokens];
    for (const source of extra) {
      merged.push(assertCspSource(name, source, options.allowUnsafeCspSources === true));
    }
    if (name === 'script-src') appendScriptSources(merged, options);
    parts.push(`${name} ${merged.join(' ')}`);
  }
  return parts.join('; ');
}

export function buildResponseHeaders(options) {
  if (!options || typeof options !== 'object') {
    throw new TypeError('options object is required');
  }
  const { kind } = options;
  if (!Object.hasOwn(policy.cacheControl, kind)) {
    throw new TypeError(`unknown response kind: ${String(kind)}`);
  }
  if (typeof options.secureTransport !== 'boolean') {
    throw new TypeError('secureTransport boolean is required');
  }
  if (options.hstsPreload === true && !options.secureTransport) {
    throw new TypeError('HSTS preload requires secureTransport');
  }
  if (options.setsCookie !== undefined && typeof options.setsCookie !== 'boolean') {
    throw new TypeError('setsCookie must be a boolean');
  }

  const document = DOCUMENT_KINDS.has(kind);
  const download = DOWNLOAD_KINDS.has(kind);
  rejectMisplacedOptions(options, document, download);
  if (options.setsCookie === true && !document) {
    throw new TypeError('Set-Cookie belongs on HTML documents, not on this response kind');
  }
  if (download && (options.filename === undefined || options.filename === '')) {
    throw new TypeError('download filename is required');
  }

  const headers = {};
  const contentType = resolveContentType(kind, options.contentType);
  if (contentType !== undefined) headers['Content-Type'] = contentType;
  headers['Cache-Control'] = cacheControlFor(kind, options.setsCookie === true);
  if (download) headers['Content-Disposition'] = contentDisposition(options.filename);

  Object.assign(headers, policy.headers.all);
  applyReferrer(headers, options.referrerPolicy);
  applyCorp(headers, options.crossOriginResourcePolicy);

  if (document) applyDocumentHeaders(headers, options);

  if (options.secureTransport) {
    headers['Strict-Transport-Security'] = hstsValue(options.hstsPreload === true);
  }

  headers.Vary = varyHeader(options.vary);
  if (options.preload !== undefined) headers.Link = buildPreloadLinkHeader(options.preload);

  return { kind, headers, remove: [...policy.remove] };
}

export function applyResponseHeaders(response, options) {
  const result = buildResponseHeaders(options);
  const access = headerAccessors(response);
  for (const name of result.remove) access.remove(name);
  for (const [name, value] of Object.entries(result.headers)) access.set(name, value);
  return result;
}

export function clearSiteDataHeader(types = ['cookies', 'storage']) {
  if (!Array.isArray(types) || types.length === 0) {
    throw new TypeError('Clear-Site-Data types must be a non-empty array');
  }
  const allowed = new Set(['cache', 'cookies', 'storage', 'executionContexts', '*']);
  return types
    .map((type) => {
      if (!allowed.has(type)) throw new TypeError(`invalid Clear-Site-Data type: ${String(type)}`);
      return `"${type}"`;
    })
    .join(', ');
}

function etagBytes(body) {
  if (typeof body === 'string') return Buffer.from(body);
  if (body instanceof Uint8Array) return body;
  throw new TypeError('body must be a string or bytes');
}

function appendScriptSources(merged, options) {
  const hashes = new Set([`'${policy.jsEnabledScriptHash}'`]);
  for (const hash of options.scriptHashes ?? []) hashes.add(validateScriptHash(hash));
  for (const hash of hashes) merged.push(hash);
  if (options.scriptNonce !== undefined) {
    merged.push(`'nonce-${validateNonce(options.scriptNonce)}'`);
  }
}

function rejectMisplacedOptions(options, document, download) {
  if (!document && options.scriptNonce !== undefined) {
    throw new TypeError('scriptNonce applies to HTML documents');
  }
  if (!document && options.cspSources !== undefined) {
    throw new TypeError('cspSources applies to HTML documents');
  }
  if (!document && options.enforceContentSecurityPolicy === false) {
    throw new TypeError('CSP report-only applies to HTML documents');
  }
  if (!document && options.permissionsAllow !== undefined) {
    throw new TypeError('permissionsAllow applies to HTML documents');
  }
  if (!document && options.crossOriginEmbedderPolicy !== undefined) {
    throw new TypeError('crossOriginEmbedderPolicy applies to HTML documents');
  }
  if (!download && options.filename !== undefined) {
    throw new TypeError('filename applies to download responses');
  }
}

function cacheControlFor(kind, setsCookie) {
  if (setsCookie && kind === 'document') return 'private, no-cache';
  return policy.cacheControl[kind];
}

function applyDocumentHeaders(headers, options) {
  Object.assign(headers, policy.headers.document);
  if (options.crossOriginEmbedderPolicy !== undefined) {
    if (!COEP_VALUES.has(options.crossOriginEmbedderPolicy)) {
      throw new TypeError('invalid Cross-Origin-Embedder-Policy');
    }
    headers['Cross-Origin-Embedder-Policy'] = options.crossOriginEmbedderPolicy;
  }
  headers['Permissions-Policy'] = permissionsPolicy(options.permissionsAllow);
  const name =
    options.enforceContentSecurityPolicy === false
      ? 'Content-Security-Policy-Report-Only'
      : 'Content-Security-Policy';
  headers[name] = buildContentSecurityPolicy(options);
}

function applyReferrer(headers, referrerPolicy) {
  if (referrerPolicy === undefined) return;
  if (!REFERRER_POLICIES.has(referrerPolicy)) throw new TypeError('invalid Referrer-Policy');
  headers['Referrer-Policy'] = referrerPolicy;
}

function applyCorp(headers, crossOriginResourcePolicy) {
  if (crossOriginResourcePolicy === undefined) return;
  if (!CORP_VALUES.has(crossOriginResourcePolicy)) {
    throw new TypeError('invalid Cross-Origin-Resource-Policy');
  }
  headers['Cross-Origin-Resource-Policy'] = crossOriginResourcePolicy;
}

function varyHeader(extra = []) {
  if (!Array.isArray(extra)) throw new TypeError('vary must be an array');
  const tokens = ['Accept-Encoding'];
  for (const token of extra) {
    if (typeof token !== 'string' || !/^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/.test(token)) {
      throw new TypeError('invalid Vary token');
    }
    tokens.push(token);
  }
  return tokens.join(', ');
}

function hstsValue(preload) {
  const parts = [`max-age=${policy.hsts.maxAge}`, 'includeSubDomains'];
  if (preload) parts.push('preload');
  return parts.join('; ');
}

function permissionsPolicy(allow = {}) {
  if (!allow || typeof allow !== 'object' || Array.isArray(allow)) {
    throw new TypeError('permissionsAllow must be an object');
  }
  const known = new Set(policy.permissionsPolicy);
  for (const name of Object.keys(allow)) {
    if (!known.has(name)) throw new TypeError(`unknown permissions policy feature: ${name}`);
  }
  return policy.permissionsPolicy
    .map((name) => {
      if (!Object.hasOwn(allow, name)) return `${name}=()`;
      return `${name}=${permissionAllowList(name, allow[name])}`;
    })
    .join(', ');
}

function permissionAllowList(name, value) {
  if (value === 'none' || value === '()') return '()';
  if (value === 'self') return '(self)';
  if (typeof value === 'string' && /^https:\/\/[A-Za-z0-9.-]+(?::\d+)?$/.test(value)) {
    return `(${value})`;
  }
  throw new TypeError(`invalid permissions allow for ${name}`);
}

function resolveContentType(kind, contentType) {
  const chosen = contentType === undefined ? policy.contentTypes[kind] : contentType;
  if (chosen === undefined) return undefined;
  if (typeof chosen !== 'string' || chosen.length === 0 || /[\r\n]/.test(chosen)) {
    throw new TypeError('invalid Content-Type');
  }
  if (/charset=/i.test(chosen)) return chosen;
  if (/^text\//i.test(chosen) || /json|xml|javascript/i.test(chosen)) {
    return `${chosen}; charset=utf-8`;
  }
  return chosen;
}

function contentDisposition(filename) {
  if (typeof filename !== 'string' || filename.length === 0 || filename.length > 180) {
    throw new TypeError('download filename is required');
  }
  if (filename.includes('/') || filename.includes('\\') || filename.includes('..')) {
    throw new TypeError('download filename must not include a path');
  }
  if (/[\r\n";]/.test(filename)) throw new TypeError('invalid download filename');
  const ascii = filename.replace(/[^\x20-\x7E]/g, '_');
  return `attachment; filename="${ascii}"; filename*=UTF-8''${encodeURIComponent(filename)}`;
}

function assertCspSource(directive, source, allowUnsafe) {
  if (typeof source !== 'string' || source.length === 0 || /[\s;,]/.test(source)) {
    throw new TypeError(`invalid CSP source for ${directive}`);
  }
  const blocked =
    source === '*' ||
    source === "'unsafe-inline'" ||
    source === "'unsafe-eval'" ||
    source === "'unsafe-hashes'" ||
    source === 'http:' ||
    source.startsWith('http://') ||
    (directive.startsWith('script-') && (source === 'data:' || source === 'blob:'));
  if (blocked && !allowUnsafe) {
    throw new TypeError(`${source} is not allowed on ${directive} without allowUnsafeCspSources`);
  }
  return source;
}

function validateNonce(nonce) {
  if (
    typeof nonce !== 'string' ||
    nonce.length < policy.nonce.minLength ||
    !NONCE_PATTERN.test(nonce)
  ) {
    throw new TypeError('scriptNonce must be a cryptographically random base64 string');
  }
  return nonce;
}

function validateScriptHash(hash) {
  if (typeof hash !== 'string' || !HASH_PATTERN.test(hash)) {
    throw new TypeError('invalid script hash');
  }
  return `'${hash}'`;
}

function headerAccessors(response) {
  if (response && typeof response.setHeader === 'function') {
    return {
      set(name, value) {
        response.setHeader(name, value);
      },
      remove(name) {
        if (typeof response.removeHeader === 'function') response.removeHeader(name);
      },
    };
  }
  const headers = response && response.headers;
  if (headers && typeof headers.set === 'function') {
    return {
      set(name, value) {
        headers.set(name, value);
      },
      remove(name) {
        if (typeof headers.delete === 'function') headers.delete(name);
      },
    };
  }
  throw new TypeError('response must provide setHeader() or headers.set()');
}
