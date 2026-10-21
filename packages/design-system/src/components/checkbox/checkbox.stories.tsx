import type { Meta, StoryObj } from '@storybook/react-vite';
import { Checkbox } from './checkbox';

const meta = {
  component: Checkbox,
  args: { label: 'I have read the agreement' },
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Checked: Story = { args: { defaultChecked: true } };
export const WithHint: Story = {
  args: { hint: 'You can download a copy before signing.' },
};
export const Disabled: Story = { args: { disabled: true } };
