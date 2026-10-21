import type { Preview } from '@storybook/react-vite';
import { applyTheme } from '../src/theme.ts';
import '../src/styles/tokens.css';
import '../src/styles/base.css';

// The browser-test projects set VITE_STORY_THEME so that every story runs in both themes.
const defaultTheme = import.meta.env.VITE_STORY_THEME === 'dark' ? 'dark' : 'light';

const preview: Preview = {
  globalTypes: {
    theme: {
      description: 'Color theme',
      toolbar: { title: 'Theme', items: ['light', 'dark'], dynamicTitle: true },
    },
  },
  initialGlobals: { theme: defaultTheme },
  decorators: [
    (Story, { globals }) => {
      applyTheme(globals.theme === 'dark' ? 'dark' : 'light');
      return <Story />;
    },
  ],
  parameters: {
    a11y: {
      test: 'error',
      options: {
        runOnly: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'],
      },
    },
  },
};

export default preview;
