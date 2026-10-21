import type { Meta, StoryObj } from '@storybook/react-vite';
import { FormField } from '../form-field/form-field';
import { Input } from './input';

const meta = {
  component: Input,
  decorators: [
    (Story, { args }) => (
      <FormField label="Email address">
        {(control) => <Story args={{ ...args, ...control }} />}
      </FormField>
    ),
  ],
  args: { type: 'email', defaultValue: 'ada@example.org' },
} satisfies Meta<typeof Input>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Empty: Story = { args: { defaultValue: '' } };
export const Disabled: Story = { args: { disabled: true } };
