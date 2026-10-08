// Writes src/styles/tokens.css from the token definitions; `--check` fails on drift instead.
import { readFileSync, writeFileSync } from 'node:fs';
import { renderTokensCss } from '../src/tokens/css.ts';

const target = new URL('../src/styles/tokens.css', import.meta.url);
const expected = renderTokensCss();

if (process.argv.includes('--check')) {
  let actual = '';
  try {
    actual = readFileSync(target, 'utf8');
  } catch {
    // A missing file is reported as drift below.
  }
  if (actual !== expected) {
    console.error(
      'tokens.css is out of date. Run `npm run tokens -w @lamassu-cla/design-system` and commit it.',
    );
    process.exit(1);
  }
} else {
  writeFileSync(target, expected);
}
