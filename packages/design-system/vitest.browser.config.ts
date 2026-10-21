import { storybookTest } from '@storybook/addon-vitest/vitest-plugin';
import { playwright } from '@vitest/browser-playwright';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// Runs in the pinned Playwright image in CI (A-23); locally it needs `npx playwright install chromium`.
// A new object per project: Vitest mutates the instances it expands.
const browser = (width = 414) => ({
  enabled: true,
  headless: true,
  provider: playwright(),
  instances: [{ browser: 'chromium' as const, viewport: { width, height: 896 } }],
});

// `reflow` (320 px, AC-05-11) and `zoom200` (a 1280 px window at 200% zoom) are Storybook viewport
// options in .storybook/preview.tsx; story widths come from them, not from the browser instance.
const storyProject = (theme: 'light' | 'dark', viewport?: 'reflow' | 'zoom200') => ({
  plugins: [react(), storybookTest({ configDir: '.storybook' })],
  define: {
    'import.meta.env.VITE_STORY_THEME': JSON.stringify(theme),
    'import.meta.env.VITE_STORY_VIEWPORT': JSON.stringify(viewport ?? ''),
  },
  test: {
    name: viewport ? `stories-${theme}-${viewport}` : `stories-${theme}`,
    browser: browser(),
  },
});

// Screenshots are compared only in the pinned image (`make test-visual`); host rendering differs.
const visual = process.env.CLA_VISUAL === '1';

export default defineConfig({
  test: {
    projects: [
      storyProject('light'),
      storyProject('dark'),
      // Width does not change colors, so the narrow runs use one theme.
      storyProject('light', 'reflow'),
      storyProject('light', 'zoom200'),
      {
        plugins: [react()],
        test: { name: 'components', include: ['src/**/*.browser.test.tsx'], browser: browser() },
      },
      ...(visual
        ? [
            {
              plugins: [react()],
              test: {
                name: 'visual',
                include: ['src/**/*.visual.test.tsx'],
                browser: {
                  ...browser(1280),
                  // Rendering in the pinned image is deterministic, so any color change must fail;
                  // the default threshold (0.1) let a change to a neighboring palette step pass.
                  expect: {
                    toMatchScreenshot: {
                      comparatorName: 'pixelmatch' as const,
                      comparatorOptions: { threshold: 0 },
                    },
                  },
                },
              },
            },
          ]
        : []),
    ],
  },
});
