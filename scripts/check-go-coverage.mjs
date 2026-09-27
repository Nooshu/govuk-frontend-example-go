#!/usr/bin/env node
/**
 * Fails unless go tool cover reports 100.0% statements for the aggregated
 * coverage.out from `go test ./internal/...`. cmd/server is excluded from the
 * gate (process entry only).
 */
import { spawnSync } from 'node:child_process';

const cover = spawnSync('go', ['tool', 'cover', '-func=coverage.out'], {
  encoding: 'utf8',
});
if (cover.status !== 0) {
  process.stderr.write(cover.stderr || cover.stdout || 'go tool cover failed\n');
  process.exit(cover.status ?? 1);
}

const report = cover.stdout;
process.stdout.write(report);

const total = [...report.matchAll(/^total:\s+\(statements\)\s+([\d.]+)%$/gm)].at(-1);
if (!total) {
  console.error('check-go-coverage: could not find total statement coverage');
  process.exit(1);
}
const percent = Number(total[1]);
if (percent < 100) {
  console.error(`check-go-coverage: statement coverage is ${percent}%, want 100%`);
  process.exit(1);
}
