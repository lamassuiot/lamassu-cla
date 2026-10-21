import type { Meta, StoryObj } from '@storybook/react-vite';
import { FormField } from '../form-field/form-field';
import { Select } from './select';

const meta = {
  component: Select,
  decorators: [
    (Story, { args }) => (
      <FormField label="Agreement type">
        {(control) => <Story args={{ ...args, ...control }} />}
      </FormField>
    ),
  ],
  args: {
    defaultValue: 'icla',
    options: [
      { value: 'icla', label: 'Individual agreement' },
      { value: 'ccla', label: 'Corporate agreement' },
    ],
  },
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Disabled: Story = { args: { disabled: true } };
