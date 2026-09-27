#!/usr/bin/env node
/**
 * Fails if package-lock.json looks unsafe for this repo's npm policy:
 * - lockfileVersion must be 3+
 * - every package entry with a tarball must use https://registry.npmjs.org/
 * - every such entry must carry a sha512 integrity hash
 * - no git:/github:/http:/file: resolved URLs for dependencies
 *
 * This is a small, dependency-free gate so we do not grow the npm attack surface
 * just to lint the lockfile.
 */
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const lockPath = join(root, 'package-lock.json');

const ALLOWED_HOST = 'registry.npmjs.org';

/** @returns {never} */
function fail(message) {
  console.error(`check-npm-lockfile: ${message}`);
  process.exit(1);
}

const lock = JSON.parse(readFileSync(lockPath, 'utf8'));

if (!Number.isInteger(lock.lockfileVersion) || lock.lockfileVersion < 3) {
  fail(`lockfileVersion must be >= 3 (got ${lock.lockfileVersion})`);
}

const packages = lock.packages;
if (!packages || typeof packages !== 'object') {
  fail('missing packages map (expected lockfileVersion 2/3 shape)');
}

let checked = 0;
for (const [name, entry] of Object.entries(packages)) {
  if (!entry || typeof entry !== 'object') continue;
  // Root package has no resolved/integrity.
  if (name === '') continue;

  const resolved = entry.resolved;
  const integrity = entry.integrity;

  if (resolved == null && integrity == null) {
    // Link / workspace style entries without a tarball — not used here, but ignore safely.
    continue;
  }

  if (typeof resolved !== 'string' || resolved.length === 0) {
    fail(`${name || '(root)'}: missing resolved URL`);
  }

  let url;
  try {
    url = new URL(resolved);
  } catch {
    fail(`${name}: resolved is not a URL: ${resolved}`);
  }

  if (url.protocol !== 'https:') {
    fail(`${name}: resolved must be https (got ${url.protocol} for ${resolved})`);
  }
  if (url.hostname !== ALLOWED_HOST) {
    fail(`${name}: resolved host must be ${ALLOWED_HOST} (got ${url.hostname})`);
  }
  if (typeof integrity !== 'string' || !integrity.startsWith('sha512-')) {
    fail(`${name}: missing sha512 integrity`);
  }
  checked += 1;
}

if (checked === 0) {
  fail('no package tarballs were checked');
}

console.log(
  `check-npm-lockfile: ok (${checked} packages, registry=${ALLOWED_HOST}, lockfileVersion=${lock.lockfileVersion})`,
);
