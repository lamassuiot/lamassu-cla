import js from '@eslint/js';
import jsxA11y from 'eslint-plugin-jsx-a11y';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import globals from 'globals';
import tseslint from 'typescript-eslint';

export default tseslint.config(
  { ignores: ['**/dist/**', '**/coverage/**', '**/node_modules/**', '.tools/**'] },
  js.configs.recommended,
  ...tseslint.configs.strict,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: { globals: { ...globals.browser } },
  },
  {
    files: ['apps/web/**/*.{ts,tsx}'],
    ...jsxA11y.flatConfigs.strict,
  },
  {
    files: ['apps/web/**/*.{ts,tsx}'],
    plugins: { 'react-hooks': reactHooks, 'react-refresh': reactRefresh },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': 'error',
    },
  },
  {
    files: ['apps/web/**/*.tsx'],
    rules: {
      // Visual values come from the design system (specs/ux/design-tokens.md).
      'no-restricted-syntax': [
        'error',
        {
          selector: 'JSXAttribute[name.name="style"]',
          message: 'Use design-system components and tokens instead of inline styles.',
        },
      ],
    },
  },
  {
    files: ['**/*.config.{js,ts,mjs}', 'eslint.config.js'],
    languageOptions: { globals: { ...globals.node } },
  },
);
