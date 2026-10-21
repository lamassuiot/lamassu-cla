import type { Meta, StoryObj } from '@storybook/react-vite';
import { fn } from 'storybook/test';
import { Pagination } from './pagination';

const meta = {
  component: Pagination,
  args: {
    label: 'Agreement pages',
    previousLabel: 'Previous page',
    nextLabel: 'Next page',
    hasPrevious: true,
    hasNext: true,
    onPrevious: fn(),
    onNext: fn(),
  },
} satisfies Meta<typeof Pagination>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Middle: Story = {};
export const FirstPage: Story = { args: { hasPrevious: false } };
export const LastPage: Story = { args: { hasNext: false } };
