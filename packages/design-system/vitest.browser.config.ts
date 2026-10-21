import { storybookTest } from '@storybook/addon-vitest/vitest-plugin';
import { playwright } from '@vitest/browser-playwright';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// Runs in the pinned Playwright image in CI (A-23); locally it needs `npx playwright install chromium`.
// A new object per project: Vitest mutates the instances it expands.
const browser = () => ({
  enabled: true,
  headless: true,
  provider: playwright(),
  instances: [{ browser: 'chromium' as const }],
});

const storyProject = (theme: 'light' | 'dark') => ({
  plugins: [react(), storybookTest({ configDir: '.storybook' })],
  define: { 'import.meta.env.VITE_STORY_THEME': JSON.stringify(theme) },
  test: { name: `stories-${theme}`, browser: browser() },
});

export default defineConfig({
  test: {
    projects: [
      storyProject('light'),
      storyProject('dark'),
      {
        plugins: [react()],
        test: { name: 'components', include: ['src/**/*.browser.test.tsx'], browser: browser() },
      },
    ],
  },
});
