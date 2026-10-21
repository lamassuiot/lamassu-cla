import type { Preview } from '@storybook/react-vite';
import { applyTheme } from '../src/theme.ts';
import '../src/styles/tokens.css';
import '../src/styles/base.css';

// The browser-test projects set VITE_STORY_THEME so that every story runs in both themes.
const defaultTheme = import.meta.env.VITE_STORY_THEME === 'dark' ? 'dark' : 'light';
// The narrow projects set VITE_STORY_VIEWPORT; the Vitest addon applies the viewport global.
const storyViewport = import.meta.env.VITE_STORY_VIEWPORT;

const preview: Preview = {
  globalTypes: {
    theme: {
      description: 'Color theme',
      toolbar: { title: 'Theme', items: ['light', 'dark'], dynamicTitle: true },
    },
  },
  initialGlobals: {
    theme: defaultTheme,
    ...(storyViewport && { viewport: { value: storyViewport, isRotated: false } }),
  },
  decorators: [
    (Story, { globals }) => {
      applyTheme(globals.theme === 'dark' ? 'dark' : 'light');
      return <Story />;
    },
  ],
  // AC-05-11: no horizontal page scrolling at any width; tables scroll inside their own container.
  afterEach: () => {
    const { scrollWidth, clientWidth } = document.documentElement;
    if (scrollWidth > clientWidth) {
      throw new Error(
        `The story scrolls horizontally: ${scrollWidth}px of content in ${clientWidth}px.`,
      );
    }
  },
  parameters: {
    // AC-05-11: the reflow width, and a 1280 px window at 200% zoom.
    viewport: {
      options: {
        reflow: { name: 'Reflow (320 px)', styles: { width: '320px', height: '896px' } },
        zoom200: {
          name: '1280 px at 200% zoom (640 px)',
          styles: { width: '640px', height: '896px' },
        },
      },
    },
    a11y: {
      test: 'error',
      options: {
        runOnly: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'],
      },
    },
  },
};

export default preview;
