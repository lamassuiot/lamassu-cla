import type { ColorPrimitive } from './primitives.ts';

export const themes = ['dark', 'light'] as const;
export type Theme = (typeof themes)[number];

type Themed<T> = Readonly<Record<Theme, T>>;

// Semantic colors (design tokens, sections 2.2 and 2.3).
export const semanticColors = {
  'bg.canvas': { dark: 'navy.950', light: 'gray.50' },
  'bg.surface': { dark: 'navy.800', light: 'white' },
  'bg.raised': { dark: 'navy.700', light: 'white' },
  'bg.subtle': { dark: 'navy.900', light: 'gray.100' },
  'bg.nav': { dark: 'navy.900', light: 'teal.950' },
  'text.primary': { dark: 'white', light: 'navy.950' },
  'text.secondary': { dark: 'gray.400', light: 'gray.700' },
  'text.heading': { dark: 'white', light: 'teal.950' },
  'text.on-nav': { dark: 'white', light: 'white' },
  'text.accent': { dark: 'orange.400', light: 'orange.700' },
  'text.on-accent': { dark: 'navy.950', light: 'navy.950' },
  'border.subtle': { dark: 'navy.600', light: 'gray.300' },
  'border.control': { dark: 'slate.500', light: 'gray.500' },
  'action.primary.bg': { dark: 'orange.500', light: 'orange.500' },
  'action.primary.bg-hover': { dark: 'orange.400', light: 'orange.600' },
  'focus.ring': { dark: 'orange.400', light: 'orange.700' },
  'status.success.text': { dark: 'status.success.dark.text', light: 'status.success.light.text' },
  'status.success.bg': { dark: 'status.success.dark.bg', light: 'status.success.light.bg' },
  'status.warning.text': { dark: 'status.warning.dark.text', light: 'status.warning.light.text' },
  'status.warning.bg': { dark: 'status.warning.dark.bg', light: 'status.warning.light.bg' },
  'status.danger.text': { dark: 'status.danger.dark.text', light: 'status.danger.light.text' },
  'status.danger.bg': { dark: 'status.danger.dark.bg', light: 'status.danger.light.bg' },
  'status.info.text': { dark: 'status.info.dark.text', light: 'status.info.light.text' },
  'status.info.bg': { dark: 'status.info.dark.bg', light: 'status.info.light.bg' },
  'status.neutral.text': { dark: 'status.neutral.dark.text', light: 'status.neutral.light.text' },
  'status.neutral.bg': { dark: 'status.neutral.dark.bg', light: 'status.neutral.light.bg' },
} as const satisfies Record<string, Themed<ColorPrimitive>>;

export type SemanticColor = keyof typeof semanticColors;

export interface Gradient {
  readonly angle: string;
  readonly from: ColorPrimitive;
  readonly to: ColorPrimitive;
}

// Design tokens, section 2.5.
export const gradients = {
  'gradient.hero': {
    dark: { angle: '135deg', from: 'navy.950', to: 'teal.950' },
    light: { angle: '180deg', from: 'gray.50', to: 'gray.100' },
  },
} as const satisfies Record<string, Themed<Gradient>>;

// Design tokens, section 5.
export const shadows = {
  'shadow.sm': { dark: '0 1px 2px rgb(0 0 0 / 0.40)', light: '0 1px 2px rgb(11 22 34 / 0.08)' },
  'shadow.md': { dark: '0 4px 12px rgb(0 0 0 / 0.45)', light: '0 4px 12px rgb(11 22 34 / 0.10)' },
  'shadow.lg': {
    dark: '0 12px 32px rgb(0 0 0 / 0.50)',
    light: '0 12px 32px rgb(11 22 34 / 0.14)',
  },
} as const satisfies Record<string, Themed<string>>;
