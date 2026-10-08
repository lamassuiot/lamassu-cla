import { colorPrimitives } from './primitives.ts';
import type { Gradient, Theme } from './semantic.ts';
import { gradients, semanticColors, shadows } from './semantic.ts';
import { borderWidths, fontFamilies, motion, radii, spacing, typography } from './scales.ts';

export const tokenPrefix = '--cla-';

// `bg.canvas` in group `color` becomes `--cla-color-bg-canvas`.
export function cssVar(name: string, group?: string): string {
  const path = group ? `${group}.${name}` : name;
  return `${tokenPrefix}${path.replaceAll('.', '-')}`;
}

const colorRef = (primitive: string) => `var(${cssVar(primitive, 'color')})`;

function gradientValue({ angle, from, to }: Gradient): string {
  return `linear-gradient(${angle}, ${colorRef(from)}, ${colorRef(to)})`;
}

function kebab(key: string): string {
  return key.replace(/[A-Z]/g, (c) => `-${c.toLowerCase()}`);
}

function themeIndependent(): string[] {
  const lines: string[] = [];
  for (const [name, value] of Object.entries(colorPrimitives)) {
    lines.push(`${cssVar(name, 'color')}: ${value};`);
  }
  const scales = { ...fontFamilies, ...spacing, ...radii, ...borderWidths, ...motion };
  for (const [name, value] of Object.entries(scales)) {
    lines.push(`${cssVar(name)}: ${value};`);
  }
  for (const [name, style] of Object.entries(typography)) {
    for (const [key, value] of Object.entries(style)) {
      lines.push(`${cssVar(name)}-${kebab(key)}: ${value};`);
    }
  }
  return lines;
}

function themed(theme: Theme): string[] {
  const lines = [`color-scheme: ${theme};`];
  for (const [name, value] of Object.entries(semanticColors)) {
    lines.push(`${cssVar(name, 'color')}: ${colorRef(value[theme])};`);
  }
  for (const [name, value] of Object.entries(gradients)) {
    lines.push(`${cssVar(name)}: ${gradientValue(value[theme])};`);
  }
  for (const [name, value] of Object.entries(shadows)) {
    lines.push(`${cssVar(name)}: ${value[theme]};`);
  }
  return lines;
}

function rule(selector: string, lines: string[], indent = ''): string {
  const body = lines.map((line) => `${indent}  ${line}`).join('\n');
  return `${indent}${selector} {\n${body}\n${indent}}`;
}

export function renderTokensCss(): string {
  return `${[
    '/* Generated from src/tokens by `npm run tokens -w @lamassu-cla/design-system`. Do not edit. */',
    rule(':root', themeIndependent()),
    rule(':root,\n[data-theme="light"]', themed('light')),
    rule('[data-theme="dark"]', themed('dark')),
    `@media (prefers-color-scheme: dark) {\n${rule(':root:not([data-theme])', themed('dark'), '  ')}\n}`,
  ].join('\n\n')}\n`;
}
