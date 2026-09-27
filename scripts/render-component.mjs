#!/usr/bin/env node
/**
 * Render GOV.UK Frontend components with the official Nunjucks macros.
 *
 * GOV.UK Frontend ships as a Node package. Its Nunjucks macros are the source of truth for
 * component HTML, so the Go application calls them through this helper rather than keeping a
 * second copy of the markup. See `docs/testing-components.md`.
 *
 * Two modes:
 *
 * - One shot (default). Read one JSON request from stdin, write the HTML to stdout.
 *   `echo '{"component":"button","params":{"text":"Save"}}' | node scripts/render-component.mjs`
 *
 * - Stream (`--stream`). Read newline-delimited JSON requests from stdin and write
 *   newline-delimited JSON responses to stdout, in the order they arrive. The Go renderer keeps
 *   one process alive for the life of the server or test run, so 700+ fixture renders do not pay
 *   Node start-up costs.
 *
 * Request: `{"id": 1, "component": "button", "params": {…}}` (`id` is optional in one-shot mode).
 * Response: `{"id": 1, "html": "…"}` or `{"id": 1, "error": "…"}`.
 *
 * Only the renderer's own outer whitespace is trimmed. Macro output is never otherwise normalised,
 * because it is compared byte-for-byte with each release's `fixtures.json`.
 */

import { createInterface } from 'node:readline';
import { createRequire } from 'node:module';
import { dirname, join } from 'node:path';

const require = createRequire(import.meta.url);
const nunjucks = require('nunjucks');

const COMPONENT_NAME = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

const govukDist = join(dirname(require.resolve('govuk-frontend/package.json')), 'dist');

const env = new nunjucks.Environment(new nunjucks.FileSystemLoader(govukDist), {
  autoescape: true,
  trimBlocks: true,
  lstripBlocks: true,
});

const compiled = new Map();

/**
 * Nunjucks macro name for a component, matching its `macro.njk`.
 *
 * @param {string} componentName - Kebab-case component name, such as `date-input`.
 * @returns {string} The macro name, such as `govukDateInput`.
 */
export function macroNameFor(componentName) {
  const pascal = componentName
    .split('-')
    .map((part) => `${part.charAt(0).toUpperCase()}${part.slice(1)}`)
    .join('');
  return `govuk${pascal}`;
}

/**
 * Render one component by calling its official Nunjucks macro.
 *
 * @param {string} componentName - Kebab-case component name.
 * @param {object} params - Macro options.
 * @returns {string} The macro HTML, with only the renderer's outer whitespace trimmed.
 * @throws {Error} When the name is not a Frontend component directory.
 */
export function renderComponent(componentName, params) {
  if (typeof componentName !== 'string' || !COMPONENT_NAME.test(componentName)) {
    throw new Error(`Unknown GOV.UK Frontend component: ${String(componentName)}`);
  }
  let template = compiled.get(componentName);
  if (!template) {
    const macroName = macroNameFor(componentName);
    const source = `{%- from "govuk/components/${componentName}/macro.njk" import ${macroName} -%}{{- ${macroName}(params) -}}`;
    template = nunjucks.compile(source, env);
    compiled.set(componentName, template);
  }
  return template.render({ params: params ?? {} }).trim();
}

/**
 * Message for a thrown value.
 *
 * @param {unknown} error - Value from a `catch` clause.
 * @returns {string} The `Error` message, or the value as text.
 */
export function messageFor(error) {
  return error instanceof Error ? error.message : String(error);
}

/**
 * Turn one parsed request into a response record.
 *
 * @param {unknown} request - Parsed JSON request.
 * @returns {{ id: unknown, html?: string, error?: string }} The response to serialise.
 */
export function handleRequest(request) {
  const id = request && typeof request === 'object' ? request.id : undefined;
  try {
    if (!request || typeof request !== 'object') throw new Error('request must be an object');
    return { id, html: renderComponent(request.component, request.params) };
  } catch (error) {
    return { id, error: messageFor(error) };
  }
}

async function readAll(stream) {
  const chunks = [];
  for await (const chunk of stream) chunks.push(Buffer.from(chunk));
  return Buffer.concat(chunks).toString('utf8');
}

/**
 * Run the helper.
 *
 * @param {string[]} argv - Arguments after the script name.
 * @param {NodeJS.ReadableStream} stdin - Request stream.
 * @param {NodeJS.WritableStream} stdout - Response stream.
 * @param {NodeJS.WritableStream} stderr - Error stream, used in one-shot mode only.
 * @returns {Promise<number>} The process exit code.
 */
export async function main(argv, stdin, stdout, stderr) {
  if (argv.includes('--stream')) {
    const lines = createInterface({ input: stdin, crlfDelay: Infinity });
    for await (const line of lines) {
      if (line.trim().length === 0) continue;
      let response;
      try {
        response = handleRequest(JSON.parse(line));
      } catch (error) {
        response = { id: null, error: messageFor(error) };
      }
      stdout.write(`${JSON.stringify(response)}\n`);
    }
    return 0;
  }

  const response = handleRequest(JSON.parse(await readAll(stdin)));
  if (response.error !== undefined) {
    stderr.write(`${response.error}\n`);
    return 1;
  }
  stdout.write(response.html);
  return 0;
}

/* node:coverage disable */
if (process.argv[1] && import.meta.url === `file://${process.argv[1]}`) {
  process.exitCode = await main(
    process.argv.slice(2),
    process.stdin,
    process.stdout,
    process.stderr,
  );
}
/* node:coverage enable */
