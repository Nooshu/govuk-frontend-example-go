import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { PassThrough, Readable } from 'node:stream';
import { createRequire } from 'node:module';
import { describe, it } from 'node:test';

import {
  handleRequest,
  macroNameFor,
  main,
  messageFor,
  renderComponent,
} from '../scripts/render-component.mjs';

const require = createRequire(import.meta.url);
const componentsRoot = path.join(
  path.dirname(require.resolve('govuk-frontend/package.json')),
  'dist',
  'govuk',
  'components',
);

async function fixture(component, name) {
  const raw = await readFile(path.join(componentsRoot, component, 'fixtures.json'), 'utf8');
  const found = JSON.parse(raw).fixtures.find((entry) => entry.name === name);
  assert.ok(found, `missing fixture ${component}/${name}`);
  return found;
}

function collect(stream) {
  const chunks = [];
  stream.on('data', (chunk) => chunks.push(Buffer.from(chunk)));
  return () => Buffer.concat(chunks).toString('utf8');
}

describe('render-component helper', () => {
  it('derives macro names from component directory names', () => {
    assert.equal(macroNameFor('button'), 'govukButton');
    assert.equal(macroNameFor('date-input'), 'govukDateInput');
    assert.equal(macroNameFor('character-count'), 'govukCharacterCount');
  });

  it('renders a component byte-for-byte against its official fixture', async () => {
    const entry = await fixture('button', 'default');
    assert.equal(renderComponent('button', entry.options), entry.html);
  });

  it('caches the compiled template between renders', async () => {
    const entry = await fixture('tag', 'default');
    assert.equal(renderComponent('tag', entry.options), entry.html);
    assert.equal(renderComponent('tag', entry.options), entry.html);
  });

  it('defaults missing params to an empty options object', () => {
    assert.match(renderComponent('button', undefined), /^<button/);
  });

  it('rejects names that are not component directories', () => {
    assert.throws(() => renderComponent('../secrets', {}), /Unknown GOV.UK Frontend component/);
    assert.throws(() => renderComponent(42, {}), /Unknown GOV.UK Frontend component/);
  });

  it('describes thrown values that are not Errors', () => {
    assert.equal(messageFor(new Error('boom')), 'boom');
    assert.equal(messageFor('boom'), 'boom');
  });

  it('reports render failures as an error record rather than throwing', () => {
    assert.deepEqual(handleRequest({ id: 7, component: 'nope!', params: {} }), {
      id: 7,
      error: 'Unknown GOV.UK Frontend component: nope!',
    });
    assert.deepEqual(handleRequest(null), {
      id: undefined,
      error: 'request must be an object',
    });
  });

  it('renders one request from stdin in one-shot mode', async () => {
    const stdout = new PassThrough();
    const stderr = new PassThrough();
    const readStdout = collect(stdout);
    const readStderr = collect(stderr);
    const entry = await fixture('tag', 'default');

    const code = await main(
      [],
      Readable.from([JSON.stringify({ component: 'tag', params: entry.options })]),
      stdout,
      stderr,
    );

    assert.equal(code, 0);
    assert.equal(readStdout(), entry.html);
    assert.equal(readStderr(), '');
  });

  it('exits non-zero and writes the reason to stderr in one-shot mode', async () => {
    const stdout = new PassThrough();
    const stderr = new PassThrough();
    const readStderr = collect(stderr);

    const code = await main(
      [],
      Readable.from([JSON.stringify({ component: 'Not A Component' })]),
      stdout,
      stderr,
    );

    assert.equal(code, 1);
    assert.match(readStderr(), /Unknown GOV.UK Frontend component/);
  });

  it('answers newline-delimited requests in order in stream mode', async () => {
    const stdout = new PassThrough();
    const stderr = new PassThrough();
    const readStdout = collect(stdout);
    const button = await fixture('button', 'default');
    const tag = await fixture('tag', 'default');

    const code = await main(
      ['--stream'],
      Readable.from(
        [
          JSON.stringify({ id: 1, component: 'button', params: button.options }),
          '',
          JSON.stringify({ id: 2, component: 'tag', params: tag.options }),
          JSON.stringify({ id: 3, component: 'no such component' }),
          'not json',
        ].join('\n'),
      ),
      stdout,
      stderr,
    );

    assert.equal(code, 0);
    const responses = readStdout()
      .trim()
      .split('\n')
      .map((line) => JSON.parse(line));
    assert.equal(responses.length, 4);
    assert.deepEqual(responses[0], { id: 1, html: button.html });
    assert.deepEqual(responses[1], { id: 2, html: tag.html });
    assert.match(responses[2].error, /Unknown GOV.UK Frontend component/);
    assert.equal(responses[3].id, null);
    assert.ok(responses[3].error.length > 0);
  });
});
