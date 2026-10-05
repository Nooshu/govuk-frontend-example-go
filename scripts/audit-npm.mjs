#!/usr/bin/env node
/**
 * Fails the build on high or critical npm advisories, except a reviewed allowlist.
 *
 * `npm audit --audit-level=high` cannot ignore one advisory. The allowlist is
 * only for a finding that has no patched release and is not on the request path.
 * Delete an entry as soon as `npm audit` stops reporting it.
 */
import { spawnSync } from 'node:child_process';
import { pathToFileURL } from 'node:url';

/** @type {ReadonlyMap<string, string>} */
export const ALLOWED_ADVISORIES = new Map([
  [
    'GHSA-vfj7-8cjw-p6xm',
    'braces <=3.0.3 has no patched release. It is a devDependency of markdownlint-cli2 (globbing only) and is not loaded when the Go service handles a request. https://github.com/advisories/GHSA-vfj7-8cjw-p6xm',
  ],
]);

const BLOCKING = new Set(['high', 'critical']);

/**
 * @param {string} url
 * @returns {string | null}
 */
export function advisoryId(url) {
  const match = String(url).match(/GHSA-[0-9a-z]{4}-[0-9a-z]{4}-[0-9a-z]{4}/i);
  return match ? match[0] : null;
}

/**
 * @param {unknown} report npm audit --json
 * @returns {{ name: string, severity: string }[]}
 */
export function unexpectedAdvisories(report) {
  const vulnerabilities =
    report && typeof report === 'object' && 'vulnerabilities' in report
      ? report.vulnerabilities
      : null;
  if (!vulnerabilities || typeof vulnerabilities !== 'object') return [];

  /** @type {Set<string>} */
  const allowed = new Set();
  let changed = true;
  while (changed) {
    changed = false;
    for (const [name, vuln] of Object.entries(vulnerabilities)) {
      if (allowed.has(name) || !vuln || typeof vuln !== 'object') continue;
      if (!BLOCKING.has(/** @type {{ severity?: string }} */ (vuln).severity ?? '')) {
        allowed.add(name);
        changed = true;
        continue;
      }
      if (viaIsAllowed(/** @type {{ via?: unknown }} */ (vuln).via, allowed)) {
        allowed.add(name);
        changed = true;
      }
    }
  }

  /** @type {{ name: string, severity: string }[]} */
  const unexpected = [];
  for (const [name, vuln] of Object.entries(vulnerabilities)) {
    if (!vuln || typeof vuln !== 'object') continue;
    const severity = /** @type {{ severity?: string }} */ (vuln).severity ?? '';
    if (!BLOCKING.has(severity) || allowed.has(name)) continue;
    unexpected.push({ name, severity });
  }
  return unexpected;
}

/**
 * @param {unknown} via
 * @param {Set<string>} allowed
 */
function viaIsAllowed(via, allowed) {
  if (!Array.isArray(via) || via.length === 0) return false;
  return via.every((item) => {
    if (typeof item === 'string') return allowed.has(item);
    if (!item || typeof item !== 'object' || !('url' in item)) return false;
    const id = advisoryId(String(/** @type {{ url?: unknown }} */ (item).url ?? ''));
    return id != null && ALLOWED_ADVISORIES.has(id);
  });
}

/**
 * @param {unknown} report
 * @returns {string[]}
 */
export function acceptedAdvisoryIds(report) {
  const vulnerabilities =
    report && typeof report === 'object' && 'vulnerabilities' in report
      ? report.vulnerabilities
      : null;
  if (!vulnerabilities || typeof vulnerabilities !== 'object') return [];

  /** @type {Set<string>} */
  const ids = new Set();
  for (const vuln of Object.values(vulnerabilities)) {
    if (!vuln || typeof vuln !== 'object' || !('via' in vuln)) continue;
    const via = /** @type {{ via?: unknown }} */ (vuln).via;
    if (!Array.isArray(via)) continue;
    for (const item of via) {
      if (!item || typeof item !== 'object' || !('url' in item)) continue;
      const id = advisoryId(String(/** @type {{ url?: unknown }} */ (item).url ?? ''));
      if (id && ALLOWED_ADVISORIES.has(id)) ids.add(id);
    }
  }
  return [...ids];
}

function main() {
  const result = spawnSync('npm', ['audit', '--json'], { encoding: 'utf8' });
  if (result.error) {
    console.error(`audit-npm: ${result.error.message}`);
    process.exit(1);
  }

  /** @type {unknown} */
  let report;
  try {
    report = JSON.parse(result.stdout);
  } catch {
    console.error(result.stderr || result.stdout || 'audit-npm: npm audit did not return JSON');
    process.exit(1);
  }

  if (report && typeof report === 'object' && 'error' in report && report.error) {
    const summary =
      typeof report.error === 'object' && report.error && 'summary' in report.error
        ? String(report.error.summary)
        : String(report.error);
    console.error(`audit-npm: ${summary}`);
    process.exit(1);
  }

  const unexpected = unexpectedAdvisories(report);
  if (unexpected.length > 0) {
    for (const item of unexpected) {
      console.error(`audit-npm: ${item.severity} vulnerability in ${item.name}`);
    }
    console.error('audit-npm: failing. Run `npm audit` for the full report.');
    process.exit(1);
  }

  for (const id of acceptedAdvisoryIds(report)) {
    console.log(`audit-npm: accepted ${id} — ${ALLOWED_ADVISORIES.get(id)}`);
  }
  console.log('audit-npm: no unexpected high or critical advisories');
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main();
}
