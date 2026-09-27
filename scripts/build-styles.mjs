import { mkdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import * as sass from 'sass';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const DEFAULT_ENTRY = path.join(ROOT, 'styles', 'application.scss');
const DEFAULT_OUT = path.join(ROOT, 'dist', 'stylesheets', 'application.css');

/**
 * Compile the application Sass entry with GOV.UK Frontend via pkg: URLs.
 * @param {{ entry?: string, outFile?: string, style?: 'expanded' | 'compressed' }} [options]
 * @returns {Promise<{ css: string, outFile: string, loadedUrls: string[] }>}
 */
export async function buildStyles(options = {}) {
  if (options === null || typeof options !== 'object' || Array.isArray(options)) {
    throw new TypeError('options must be an object');
  }

  const entry = options.entry ?? DEFAULT_ENTRY;
  const outFile = options.outFile ?? DEFAULT_OUT;
  const style = options.style ?? 'compressed';

  if (typeof entry !== 'string' || entry.length === 0) {
    throw new TypeError('entry must be a non-empty string');
  }
  if (typeof outFile !== 'string' || outFile.length === 0) {
    throw new TypeError('outFile must be a non-empty string');
  }
  if (style !== 'compressed' && style !== 'expanded') {
    throw new TypeError('style must be compressed or expanded');
  }

  const result = sass.compile(entry, {
    style,
    quietDeps: true,
    importers: [new sass.NodePackageImporter(ROOT)],
  });

  await mkdir(path.dirname(outFile), { recursive: true });
  await writeFile(outFile, result.css);

  return {
    css: result.css,
    outFile,
    loadedUrls: result.loadedUrls.map((url) => url.href),
  };
}

/**
 * CLI entry used by `npm run build:styles`.
 * @param {{ write?: (chunk: string) => void }} [io]
 */
export async function runBuildStylesCli(io = {}) {
  if (io === null || typeof io !== 'object' || Array.isArray(io)) {
    throw new TypeError('io must be an object');
  }
  const write = io.write ?? ((chunk) => process.stdout.write(chunk));
  if (typeof write !== 'function') throw new TypeError('io.write must be a function');

  const { outFile, css } = await buildStyles();
  write(`built ${outFile} (${css.length} bytes)\n`);
  return { outFile, css };
}
