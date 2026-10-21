import { defineConfig } from 'vitest/config';

// Node unit tests only; component and story tests need a browser (vitest.browser.config.ts).
export default defineConfig({
  test: {
    include: ['src/**/*.test.ts'],
  },
});
