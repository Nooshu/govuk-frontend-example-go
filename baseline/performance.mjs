const PRELOAD_AS = new Set(['style', 'script', 'font', 'image', 'fetch']);
const VARY_TOKEN = /^[!#$%&'*+\-.^_`|~0-9A-Za-z]+$/;

export function appendVary(existing, token) {
  if (existing !== undefined && typeof existing !== 'string') {
    throw new TypeError('existing Vary must be a string');
  }
  if (typeof token !== 'string' || !VARY_TOKEN.test(token)) {
    throw new TypeError('invalid Vary token');
  }

  const seen = new Map();
  for (const part of (existing ?? '').split(',')) {
    const trimmed = part.trim();
    if (trimmed.length === 0) continue;
    const key = trimmed.toLowerCase();
    if (!seen.has(key)) seen.set(key, trimmed);
  }
  const key = token.toLowerCase();
  if (!seen.has(key)) seen.set(key, token);
  return [...seen.values()].join(', ');
}

export function buildPreloadLinkHeader(links) {
  if (!Array.isArray(links) || links.length === 0) {
    throw new TypeError('links must be a non-empty array');
  }
  return links.map((link) => formatPreload(link)).join(', ');
}

function formatPreload(link) {
  if (!link || typeof link.href !== 'string' || !isSameOriginPath(link.href)) {
    throw new TypeError('preload href must be a same-origin path');
  }
  if (!PRELOAD_AS.has(link.as)) throw new TypeError(`unsupported preload as: ${String(link.as)}`);

  const params = [`<${link.href}>`, 'rel=preload', `as=${link.as}`];
  if (link.type !== undefined) {
    if (typeof link.type !== 'string' || !/^[\w.+-]+\/[\w.+-]+$/.test(link.type)) {
      throw new TypeError('invalid preload type');
    }
    params.push(`type="${link.type}"`);
  }
  if (link.as === 'font' || link.crossorigin === true) params.push('crossorigin');
  return params.join('; ');
}

function isSameOriginPath(href) {
  return href.startsWith('/') && !href.startsWith('//') && !/[\s"<>]/.test(href);
}
