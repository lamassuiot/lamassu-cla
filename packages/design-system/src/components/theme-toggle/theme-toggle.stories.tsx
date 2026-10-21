import type { Meta, StoryObj } from '@storybook/react-vite';
import { ThemeToggle } from './theme-toggle';

const meta = {
  component: ThemeToggle,
  args: { legend: 'Theme', labels: { light: 'Light', dark: 'Dark', system: 'Match system' } },
} satisfies Meta<typeof ThemeToggle>;

export default meta;

export const Default: StoryObj<typeof meta> = {};
