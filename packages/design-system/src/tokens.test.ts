import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import {
  borderWidths,
  breakpoints,
  colorPrimitives,
  cssVar,
  gradients,
  motion,
  radii,
  renderTokensCss,
  semanticColors,
  shadows,
  spacing,
  type Theme,
  themes,
  tokenPrefix,
  typography,
} from './index.ts';

// Expected values are copied from specs/ux/design-tokens.md, not derived from the code under test.
const specPrimitives = {
  'navy.950': '#0B1622',
  'navy.900': '#0F1F2E',
  'navy.800': '#162B3D',
  'navy.700': '#1D3448',
  'navy.600': '#2A4760',
  'slate.500': '#5C7A94',
  'teal.950': '#0D2A2E',
  'teal.800': '#145055',
  white: '#FFFFFF',
  'orange.400': '#F78C40',
  'orange.500': '#F2711C',
  'orange.600': '#D9600F',
  'orange.700': '#B4500A',
  'gray.50': '#F7F5F2',
  'gray.100': '#EEF1F4',
  'gray.200': '#DDE2E8',
  'gray.300': '#C9D2DC',
  'gray.400': '#B8C4D0',
  'gray.500': '#7A8896',
  'gray.700': '#4A5B6C',
  'gray.800': '#3D4F61',
};

const specSemantic: Record<string, [string, string]> = {
  'bg.canvas': ['navy.950', 'gray.50'],
  'bg.surface': ['navy.800', 'white'],
  'bg.raised': ['navy.700', 'white'],
  'bg.subtle': ['navy.900', 'gray.100'],
  'bg.nav': ['navy.900', 'teal.950'],
  'text.primary': ['white', 'navy.950'],
  'text.secondary': ['gray.400', 'gray.700'],
  'text.heading': ['white', 'teal.950'],
  'text.on-nav': ['white', 'white'],
  'text.accent': ['orange.400', 'orange.700'],
  'text.on-accent': ['navy.950', 'navy.950'],
  'border.subtle': ['navy.600', 'gray.300'],
  'border.control': ['slate.500', 'gray.500'],
  'action.primary.bg': ['orange.500', 'orange.500'],
  'action.primary.bg-hover': ['orange.400', 'orange.600'],
  'focus.ring': ['orange.400', 'orange.700'],
};

// [dark text, dark background, light text, light background]
const specStatus: Record<string, [string, string, string, string]> = {
  success: ['#4CC38A', '#10302A', '#1E7A4F', '#E6F4EC'],
  warning: ['#F2B84B', '#33280F', '#8A5A00', '#FDF3DC'],
  danger: ['#F28B82', '#3A1A1C', '#B42318', '#FDE8E6'],
  info: ['#7DB4F0', '#132C47', '#1F5FA8', '#E6EFFA'],
  neutral: ['#B8C4D0', '#1F3447', '#3D4F61', '#EAEEF2'],
};

// Section 2.4: [foreground, background, minimum ratio, dark ratio, light ratio].
const specContrast: [string, string, number, number, number][] = [
  ['text.primary', 'bg.canvas', 4.5, 18.23, 16.75],
  ['text.primary', 'bg.surface', 4.5, 14.51, 18.23],
  ['text.secondary', 'bg.canvas', 4.5, 10.28, 6.42],
  ['text.secondary', 'bg.surface', 4.5, 8.19, 6.99],
  ['text.accent', 'bg.canvas', 4.5, 7.62, 4.72],
  ['text.accent', 'bg.surface', 4.5, 6.07, 5.13],
  ['text.on-accent', 'action.primary.bg', 4.5, 6.21, 6.21],
  ['text.on-accent', 'action.primary.bg-hover', 4.5, 7.62, 4.89],
  ['text.on-nav', 'bg.nav', 4.5, 16.72, 15.13],
  ['border.control', 'bg.canvas', 3, 4.05, 3.33],
  ['border.control', 'bg.surface', 3, 3.22, 3.63],
  ['focus.ring', 'bg.canvas', 3, 7.62, 4.72],
];
const specStatusMinimum = { dark: 6.41, light: 4.69 };

function luminance(hex: string): number {
  const [r, g, b] = [1, 3, 5].map((i) => {
    const c = Number.parseInt(hex.slice(i, i + 2), 16) / 255;
    return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  }) as [number, number, number];
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x) as [number, number];
  return (hi + 0.05) / (lo + 0.05);
}

function resolve(semantic: string, theme: Theme): string {
  const primitive = semanticColors[semantic as keyof typeof semanticColors][theme];
  return colorPrimitives[primitive];
}

describe('token values (AC-04-1)', () => {
  it('defines exactly the primitive and status colors in the specification', () => {
    const expected: Record<string, string> = { ...specPrimitives };
    for (const [status, [darkText, darkBg, lightText, lightBg]] of Object.entries(specStatus)) {
      expected[`status.${status}.dark.text`] = darkText;
      expected[`status.${status}.dark.bg`] = darkBg;
      expected[`status.${status}.light.text`] = lightText;
      expected[`status.${status}.light.bg`] = lightBg;
    }
    expect(colorPrimitives).toEqual(expected);
  });

  it('defines exactly the semantic and status colors in the specification', () => {
    const expected: Record<string, Record<Theme, string>> = {};
    for (const [name, [dark, light]] of Object.entries(specSemantic)) {
      expected[name] = { dark, light };
    }
    for (const status of Object.keys(specStatus)) {
      for (const part of ['text', 'bg']) {
        expected[`status.${status}.${part}`] = {
          dark: `status.${status}.dark.${part}`,
          light: `status.${status}.light.${part}`,
        };
      }
    }
    expect(semanticColors).toEqual(expected);
  });

  it('defines the gradients and shadows', () => {
    expect(gradients).toEqual({
      'gradient.hero': {
        dark: { angle: '135deg', from: 'navy.950', to: 'teal.950' },
        light: { angle: '180deg', from: 'gray.50', to: 'gray.100' },
      },
    });
    expect(shadows).toEqual({
      'shadow.sm': { dark: '0 1px 2px rgb(0 0 0 / 0.40)', light: '0 1px 2px rgb(11 22 34 / 0.08)' },
      'shadow.md': {
        dark: '0 4px 12px rgb(0 0 0 / 0.45)',
        light: '0 4px 12px rgb(11 22 34 / 0.10)',
      },
      'shadow.lg': {
        dark: '0 12px 32px rgb(0 0 0 / 0.50)',
        light: '0 12px 32px rgb(11 22 34 / 0.14)',
      },
    });
  });

  it('defines the typography scale in rem from the specified pixel values', () => {
    const px: Record<string, [number, number, number]> = {
      'text.display': [48, 56, 700],
      'text.h1': [36, 44, 700],
      'text.h2': [28, 36, 600],
      'text.h3': [22, 30, 600],
      'text.body-lg': [18, 28, 400],
      'text.body': [16, 24, 400],
      'text.body-sm': [14, 20, 400],
      'text.label': [12, 16, 600],
      'text.code': [14, 20, 400],
    };
    expect(Object.keys(typography)).toEqual(Object.keys(px));
    for (const [name, [size, lineHeight, weight]] of Object.entries(px)) {
      expect(typography[name as keyof typeof typography]).toMatchObject({
        size: `${size / 16}rem`,
        lineHeight: `${lineHeight / 16}rem`,
        weight,
      });
    }
    expect(typography['text.label']).toMatchObject({
      letterSpacing: '0.08em',
      transform: 'uppercase',
    });
  });

  it('defines the spacing, radius, border, motion, and breakpoint scales', () => {
    expect(spacing).toEqual({
      'space.0': '0',
      'space.1': '4px',
      'space.2': '8px',
      'space.3': '12px',
      'space.4': '16px',
      'space.5': '24px',
      'space.6': '32px',
      'space.7': '48px',
      'space.8': '64px',
      'space.9': '96px',
    });
    expect(radii).toEqual({
      'radius.sm': '6px',
      'radius.md': '10px',
      'radius.lg': '16px',
      'radius.pill': '9999px',
    });
    expect(borderWidths).toEqual({ 'border.width': '1px', 'border.width-strong': '2px' });
    expect(motion).toEqual({
      'motion.duration.fast': '120ms',
      'motion.duration.base': '200ms',
      'motion.duration.slow': '320ms',
      'motion.easing.standard': 'cubic-bezier(0.2, 0, 0, 1)',
      'motion.easing.exit': 'cubic-bezier(0.3, 0, 1, 1)',
    });
    expect(breakpoints).toEqual({
      'bp.sm': '640px',
      'bp.md': '768px',
      'bp.lg': '1024px',
      'bp.xl': '1280px',
    });
  });
});

describe('generated CSS (AC-04-2, AC-04-7)', () => {
  const committed = readFileSync(new URL('./styles/tokens.css', import.meta.url), 'utf8');
  const declarations = [...committed.matchAll(/^\s*(--[\w-]+):\s*([^;]+);/gm)].map(
    ([, name, value]) => ({ name: name as string, value: value as string }),
  );

  it('matches the token definitions', () => {
    expect(committed).toBe(renderTokensCss());
  });

  it('prefixes every custom property with --cla-', () => {
    expect(declarations.length).toBeGreaterThan(0);
    for (const { name } of declarations) {
      expect(name.startsWith(tokenPrefix)).toBe(true);
    }
  });

  it('builds semantic colors and gradients only from primitive references', () => {
    const referencing = [
      ...Object.keys(semanticColors).map((name) => cssVar(name, 'color')),
      ...Object.keys(gradients).map((name) => cssVar(name)),
    ];
    const primitiveVars = new Set(Object.keys(colorPrimitives).map((p) => cssVar(p, 'color')));
    for (const { name, value } of declarations.filter((d) => referencing.includes(d.name))) {
      expect(value, name).not.toMatch(/#[0-9a-f]{3,8}\b|rgb|hsl/i);
      const refs = [...value.matchAll(/var\((--[\w-]+)\)/g)].map(([, ref]) => ref);
      expect(refs.length, name).toBeGreaterThan(0);
      for (const ref of refs) {
        expect(primitiveVars.has(ref as string), `${name} -> ${ref}`).toBe(true);
      }
    }
  });

  it('defines every semantic token for the light, dark, and system-dark selectors', () => {
    for (const selector of [':root,\n[data-theme="light"] {', '[data-theme="dark"] {']) {
      expect(committed).toContain(selector);
    }
    expect(committed).toMatch(
      /@media \(prefers-color-scheme: dark\) \{\n {2}:root:not\(\[data-theme\]\) \{/,
    );
    const semanticNames = Object.keys(semanticColors).map((name) => cssVar(name, 'color'));
    for (const name of semanticNames) {
      expect(declarations.filter((d) => d.name === name)).toHaveLength(3);
    }
  });
});

describe('contrast (AC-04-3)', () => {
  for (const theme of themes) {
    it(`meets the WCAG thresholds and matches the specification in the ${theme} theme`, () => {
      for (const [fg, bg, minimum, darkRatio, lightRatio] of specContrast) {
        const ratio = contrast(resolve(fg, theme), resolve(bg, theme));
        expect(ratio, `${fg} on ${bg}`).toBeGreaterThanOrEqual(minimum);
        expect(Math.abs(ratio - (theme === 'dark' ? darkRatio : lightRatio))).toBeLessThan(0.01);
      }
      const statusRatios = Object.keys(specStatus).map((status) =>
        contrast(resolve(`status.${status}.text`, theme), resolve(`status.${status}.bg`, theme)),
      );
      const lowest = Math.min(...statusRatios);
      expect(lowest).toBeGreaterThanOrEqual(4.5);
      expect(Math.abs(lowest - specStatusMinimum[theme])).toBeLessThan(0.01);
    });
  }
});
