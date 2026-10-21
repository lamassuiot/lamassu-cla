import type { Meta, StoryObj } from '@storybook/react-vite';
import { Input } from '../input/input';
import { FormField } from './form-field';

const meta = {
  component: FormField,
  args: {
    label: 'Display name',
    children: (control) => <Input {...control} defaultValue="Ada Example" />,
  },
} satisfies Meta<typeof FormField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const WithHint: Story = { args: { hint: 'Shown on your public agreement record.' } };
export const WithError: Story = {
  args: {
    hint: 'Shown on your public agreement record.',
    error: 'Enter a display name.',
    children: (control) => <Input {...control} defaultValue="" />,
  },
};
export const Required: Story = { args: { label: 'Display name (required)', required: true } };
