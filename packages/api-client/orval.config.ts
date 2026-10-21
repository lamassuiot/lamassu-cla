import { defineConfig } from 'orval';

// Generated from the committed contract; `npm run generate:check` fails on drift (AC-02-7, A-22).
export default defineConfig({
  lamassu: {
    input: { target: '../../specs/api/openapi.yaml' },
    output: {
      target: './src/generated/client.ts',
      client: 'fetch',
      mode: 'single',
      clean: true,
      override: {
        mutator: { path: './src/fetcher.ts', name: 'apiFetch' },
      },
    },
  },
});
