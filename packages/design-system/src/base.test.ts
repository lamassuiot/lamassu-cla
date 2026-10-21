import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { motion } from './tokens/scales.ts';

const base = readFileSync(new URL('./styles/base.css', import.meta.url), 'utf8');
const reducedMotion = base.slice(base.indexOf('@media (prefers-reduced-motion: reduce)'));

describe('base styles (AC-04-8)', () => {
  it('has a reduced-motion rule outside the cascade layers', () => {
    expect(base).toContain('@layer base, components;');
    expect(reducedMotion.length).toBeGreaterThan(0);
    expect(reducedMotion).not.toContain('@layer');
  });

  it('limits transitions to opacity and durations to motion.duration.fast', () => {
    expect(reducedMotion).toMatch(/transition-property: opacity;/);
    const durations = Object.keys(motion).filter((name) => name.includes('duration'));
    for (const name of durations.filter((n) => n !== 'motion.duration.fast')) {
      const property = `--cla-${name.replaceAll('.', '-')}`;
      expect(reducedMotion).toContain(`${property}: var(--cla-motion-duration-fast);`);
    }
  });

  it('applies the canvas, text, font, and focus-ring tokens', () => {
    for (const token of [
      '--cla-color-bg-canvas',
      '--cla-color-text-primary',
      '--cla-font-family-sans',
      '--cla-text-body-size',
      '--cla-color-focus-ring',
    ]) {
      expect(base).toContain(`var(${token})`);
    }
    expect(base).toMatch(/:focus-visible \{[^}]*outline: var\(--cla-border-width-strong\) solid/);
  });

  it('references only tokens that tokens.css defines', () => {
    const tokens = readFileSync(new URL('./styles/tokens.css', import.meta.url), 'utf8');
    const defined = new Set([...tokens.matchAll(/(--cla-[\w-]+):/g)].map(([, name]) => name));
    const used = [...base.matchAll(/var\((--cla-[\w-]+)\)/g)].map(([, name]) => name);
    expect(used.length).toBeGreaterThan(0);
    for (const name of used) {
      expect(defined.has(name as string), name).toBe(true);
    }
  });
});
