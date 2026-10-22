import { playwright } from '@vitest/browser-playwright';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// Portal tests that need a real browser: routing, focus, and layout (F-06). CI runs them in the
// pinned Playwright image; locally they need `npx playwright install chromium`.
export default defineConfig({
  plugins: [react()],
  test: {
    name: 'portal',
    include: ['src/**/*.browser.test.tsx'],
    browser: {
      enabled: true,
      headless: true,
      provider: playwright(),
      instances: [{ browser: 'chromium', viewport: { width: 1280, height: 900 } }],
    },
  },
});
