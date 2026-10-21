import type { Meta, StoryObj } from '@storybook/react-vite';
import { PageHeader } from '../page-header/page-header';
import { AppShell } from './app-shell';

const meta = {
  component: AppShell,
  args: {
    labels: {
      skipToContent: 'Skip to content',
      menu: 'Menu',
      menuTitle: 'Navigation',
      closeMenu: 'Close navigation',
    },
    brand: <a href="/">Lamassu CLA</a>,
    navigation: {
      label: 'Main',
      items: [
        { href: '/dashboard', label: 'Dashboard', current: true },
        { href: '/agreements', label: 'Agreements' },
        { href: '/organizations', label: 'Organizations' },
      ],
    },
    children: <PageHeader title="Dashboard" description="Your agreements and coverage." />,
  },
} satisfies Meta<typeof AppShell>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const MenuOpen: Story = {
  globals: { viewport: { value: 'reflow', isRotated: false } },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Menu' }));
  },
};
