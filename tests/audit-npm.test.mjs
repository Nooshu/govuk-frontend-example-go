import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import {
  acceptedAdvisoryIds,
  advisoryId,
  ALLOWED_ADVISORIES,
  unexpectedAdvisories,
} from '../scripts/audit-npm.mjs';

const BRACES = {
  source: 1240992,
  name: 'braces',
  url: 'https://github.com/advisories/GHSA-vfj7-8cjw-p6xm',
  severity: 'high',
};

describe('unexpectedAdvisories', () => {
  it('allows the braces advisory and packages that only depend on it', () => {
    const report = {
      vulnerabilities: {
        braces: { severity: 'high', via: [BRACES] },
        micromatch: { severity: 'high', via: ['braces'] },
        'fast-glob': { severity: 'high', via: ['micromatch'] },
        globby: { severity: 'high', via: ['fast-glob', 'micromatch'] },
        'markdownlint-cli2': { severity: 'high', via: ['globby', 'micromatch'] },
        'low-only': {
          severity: 'low',
          via: [{ url: 'https://github.com/advisories/GHSA-aaaa-bbbb-cccc' }],
        },
      },
    };

    assert.deepEqual(unexpectedAdvisories(report), []);
    assert.deepEqual(acceptedAdvisoryIds(report), ['GHSA-vfj7-8cjw-p6xm']);
  });

  it('still fails on any other high advisory', () => {
    const report = {
      vulnerabilities: {
        braces: { severity: 'high', via: [BRACES] },
        sass: {
          severity: 'high',
          via: [{ url: 'https://github.com/advisories/GHSA-zzzz-yyyy-xxxx', severity: 'high' }],
        },
      },
    };

    assert.deepEqual(unexpectedAdvisories(report), [{ name: 'sass', severity: 'high' }]);
  });

  it('fails when a high finding has no cause', () => {
    assert.deepEqual(
      unexpectedAdvisories({ vulnerabilities: { sass: { severity: 'critical', via: [] } } }),
      [{ name: 'sass', severity: 'critical' }],
    );
  });

  it('ignores a report with no vulnerabilities', () => {
    assert.deepEqual(unexpectedAdvisories({}), []);
    assert.deepEqual(unexpectedAdvisories(null), []);
    assert.deepEqual(acceptedAdvisoryIds(null), []);
  });
});

describe('advisoryId', () => {
  it('reads a GHSA id from an advisory URL', () => {
    assert.equal(
      advisoryId('https://github.com/advisories/GHSA-vfj7-8cjw-p6xm'),
      'GHSA-vfj7-8cjw-p6xm',
    );
    assert.equal(advisoryId('not-an-advisory'), null);
    assert.equal(ALLOWED_ADVISORIES.has('GHSA-vfj7-8cjw-p6xm'), true);
  });
});
