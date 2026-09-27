import policy from './policy.mjs';

const COOKIE_NAME = /^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/;
const COOKIE_VALUE = /^[\x21\x23-\x2B\x2D-\x3A\x3C-\x5B\x5D-\x7E]*$/;
const SAME_SITE = new Set(['Lax', 'Strict', 'None']);

export function buildSetCookie(name, value, options = {}) {
  if (typeof name !== 'string' || !COOKIE_NAME.test(name)) {
    throw new TypeError('invalid cookie name');
  }
  if (typeof value !== 'string' || !COOKIE_VALUE.test(value)) {
    throw new TypeError('invalid cookie value');
  }
  if (!options || typeof options !== 'object' || Array.isArray(options)) {
    throw new TypeError('cookie options must be an object');
  }

  const sameSite = options.sameSite ?? policy.cookie.sameSite;
  if (!SAME_SITE.has(sameSite)) throw new TypeError('invalid SameSite');

  const secure = options.secure ?? policy.cookie.secure;
  const httpOnly = options.httpOnly ?? policy.cookie.httpOnly;
  if (typeof secure !== 'boolean' || typeof httpOnly !== 'boolean') {
    throw new TypeError('secure and httpOnly must be booleans');
  }
  if (sameSite === 'None' && !secure) {
    throw new TypeError('SameSite=None requires Secure');
  }

  const path = options.path ?? policy.cookie.path;
  if (typeof path !== 'string' || !path.startsWith('/')) {
    throw new TypeError('cookie Path must start with /');
  }

  const hostPrefixed = name.startsWith('__Host-');
  if (options.hostPrefix === true && !hostPrefixed) {
    throw new TypeError('hostPrefix requires a __Host- cookie name');
  }
  if (hostPrefixed && !secure) throw new TypeError('__Host- cookies require Secure');
  if (hostPrefixed && path !== '/') throw new TypeError('__Host- cookies require Path=/');
  if (hostPrefixed && options.domain !== undefined) {
    throw new TypeError('__Host- cookies must not set Domain');
  }
  if (options.domain !== undefined && !/^[A-Za-z0-9.-]+$/.test(options.domain)) {
    throw new TypeError('invalid Domain');
  }

  let maxAge;
  if (options.maxAge !== undefined) {
    if (!Number.isInteger(options.maxAge) || options.maxAge < 0) {
      throw new TypeError('maxAge must be a non-negative integer');
    }
    maxAge = options.maxAge;
  }

  const parts = [`${name}=${value}`];
  if (maxAge !== undefined) parts.push(`Max-Age=${maxAge}`);
  if (options.domain !== undefined) parts.push(`Domain=${options.domain}`);
  parts.push(`Path=${path}`);
  if (secure) parts.push('Secure');
  if (httpOnly) parts.push('HttpOnly');
  parts.push(`SameSite=${sameSite}`);
  return parts.join('; ');
}
