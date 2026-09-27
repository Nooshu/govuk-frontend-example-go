import assert from 'node:assert/strict';
import { mkdtemp, readFile, readdir, rm } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { describe, it } from 'node:test';
import { fileURLToPath } from 'node:url';

import { buildStyles, runBuildStylesCli } from '../scripts/build-styles.mjs';

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const STYLES_DIR = path.join(ROOT, 'styles');

async function listScssFiles(dir) {
  const entries = await readdir(dir, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) files.push(...(await listScssFiles(full)));
    else if (entry.isFile() && entry.name.endsWith('.scss')) files.push(full);
  }
  return files;
}

describe('Sass pipeline', () => {
  it('compiles GOV.UK Frontend via @use, not the prebuilt dist CSS', async () => {
    const temp = await mkdtemp(path.join(os.tmpdir(), 'govuk-styles-'));
    try {
      const outFile = path.join(temp, 'application.css');
      const { css, loadedUrls } = await buildStyles({ outFile, style: 'compressed' });
      const written = await readFile(outFile, 'utf8');

      assert.equal(css, written);
      assert.match(css, /\.govuk-button\b/);
      assert.match(css, /\.govuk-template\b/);
      assert.match(css, /--app-stylesheet-layer:\s*govuk-overrides/);
      assert.equal(
        css.indexOf('.govuk-button') < css.indexOf('--app-stylesheet-layer:govuk-overrides') ||
          css.indexOf('.govuk-button') < css.indexOf('--app-stylesheet-layer: govuk-overrides'),
        true,
      );

      const loaded = loadedUrls.join('\n');
      assert.match(loaded, /govuk-frontend/);
      assert.match(loaded, /govuk-overrides/);
      assert.equal(loaded.includes('govuk-frontend.min.css'), false);
    } finally {
      await rm(temp, { recursive: true, force: true });
    }
  });

  it('prints the default build path from the CLI helper', async () => {
    const lines = [];
    const { outFile, css } = await runBuildStylesCli({ write: (chunk) => lines.push(chunk) });
    assert.match(outFile, /application\.css$/);
    assert.match(css, /\.govuk-button/);
    assert.match(lines.join(''), /built .*application\.css/);

    const defaultCli = await runBuildStylesCli();
    assert.match(defaultCli.outFile, /application\.css$/);

    await assert.rejects(() => runBuildStylesCli(null), /io must be an object/);
    await assert.rejects(() => runBuildStylesCli([]), /io must be an object/);
    await assert.rejects(() => runBuildStylesCli({ write: 'nope' }), /io.write must be a function/);
  });

  it('accepts expanded output and rejects invalid options', async () => {
    const temp = await mkdtemp(path.join(os.tmpdir(), 'govuk-styles-'));
    try {
      const outFile = path.join(temp, 'expanded.css');
      const { css } = await buildStyles({ outFile, style: 'expanded' });
      assert.match(css, /\n/);
      assert.match(css, /\.govuk-button/);

      await assert.rejects(() => buildStyles(null), /options must be an object/);
      await assert.rejects(() => buildStyles([]), /options must be an object/);
      await assert.rejects(() => buildStyles({ entry: '' }), /entry must be a non-empty string/);
      await assert.rejects(() => buildStyles({ entry: 1 }), /entry must be a non-empty string/);
      await assert.rejects(
        () => buildStyles({ outFile: '' }),
        /outFile must be a non-empty string/,
      );
      await assert.rejects(() => buildStyles({ outFile: 1 }), /outFile must be a non-empty string/);
      await assert.rejects(
        () => buildStyles({ style: 'nested' }),
        /style must be compressed or expanded/,
      );
    } finally {
      await rm(temp, { recursive: true, force: true });
    }
  });

  it('keeps service Sass free of !important', async () => {
    const files = await listScssFiles(STYLES_DIR);
    assert.ok(files.some((file) => file.endsWith('application.scss')));
    assert.ok(files.some((file) => file.endsWith('govuk-overrides.scss')));

    for (const file of files) {
      const source = await readFile(file, 'utf8');
      const withoutBlockComments = source.replace(/\/\*[\s\S]*?\*\//g, '');
      const withoutLineComments = withoutBlockComments.replace(/(^|[^:])\/\/.*$/gm, '$1');
      assert.equal(
        /!important\b/i.test(withoutLineComments),
        false,
        `${path.relative(ROOT, file)} must not contain !important`,
      );
    }
  });
});
