import type { StorybookConfig } from '@storybook/react-vite';

// Development-only preview (A-23); it is not deployed or distributed (D-25).
const config: StorybookConfig = {
  framework: '@storybook/react-vite',
  stories: ['../src/**/*.stories.tsx'],
  addons: ['@storybook/addon-a11y', '@storybook/addon-vitest'],
  core: { disableTelemetry: true, disableWhatsNewNotifications: true },
};

export default config;
