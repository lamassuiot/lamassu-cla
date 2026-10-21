import type { Meta, StoryObj } from '@storybook/react-vite';
import { Navigation } from './navigation';

const meta = {
  component: Navigation,
  args: {
    label: 'Main',
    items: [
      { href: '/dashboard', label: 'Dashboard', current: true },
      { href: '/agreements', label: 'Agreements' },
      { href: '/organizations', label: 'Organizations' },
      { href: 'https://example.com/docs', label: 'Documentation' },
    ],
  },
} satisfies Meta<typeof Navigation>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
