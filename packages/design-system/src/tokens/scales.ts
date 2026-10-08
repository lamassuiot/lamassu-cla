// Theme-independent scales (design tokens, sections 3 to 7).

// System stacks until D-12 confirms the font families; no font files are bundled.
export const fontFamilies = {
  'font.family.sans':
    "system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
  'font.family.mono': "ui-monospace, SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace",
} as const;

export interface TextStyle {
  readonly size: string;
  readonly lineHeight: string;
  readonly weight: number;
  readonly letterSpacing?: string;
  readonly transform?: 'uppercase';
}

// Sizes in rem (16 px root): the specification lists them in px.
export const typography = {
  'text.display': { size: '3rem', lineHeight: '3.5rem', weight: 700 },
  'text.h1': { size: '2.25rem', lineHeight: '2.75rem', weight: 700 },
  'text.h2': { size: '1.75rem', lineHeight: '2.25rem', weight: 600 },
  'text.h3': { size: '1.375rem', lineHeight: '1.875rem', weight: 600 },
  'text.body-lg': { size: '1.125rem', lineHeight: '1.75rem', weight: 400 },
  'text.body': { size: '1rem', lineHeight: '1.5rem', weight: 400 },
  'text.body-sm': { size: '0.875rem', lineHeight: '1.25rem', weight: 400 },
  'text.label': {
    size: '0.75rem',
    lineHeight: '1rem',
    weight: 600,
    letterSpacing: '0.08em',
    transform: 'uppercase',
  },
  'text.code': { size: '0.875rem', lineHeight: '1.25rem', weight: 400 },
} as const satisfies Record<string, TextStyle>;

export const spacing = {
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
} as const;

export const radii = {
  'radius.sm': '6px',
  'radius.md': '10px',
  'radius.lg': '16px',
  'radius.pill': '9999px',
} as const;

export const borderWidths = {
  'border.width': '1px',
  'border.width-strong': '2px',
} as const;

export const motion = {
  'motion.duration.fast': '120ms',
  'motion.duration.base': '200ms',
  'motion.duration.slow': '320ms',
  'motion.easing.standard': 'cubic-bezier(0.2, 0, 0, 1)',
  'motion.easing.exit': 'cubic-bezier(0.3, 0, 1, 1)',
} as const;

// CSS custom properties cannot be used in media queries; components import these values.
export const breakpoints = {
  'bp.sm': '640px',
  'bp.md': '768px',
  'bp.lg': '1024px',
  'bp.xl': '1280px',
} as const;
