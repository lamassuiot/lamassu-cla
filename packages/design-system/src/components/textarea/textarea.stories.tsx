import type { Meta, StoryObj } from '@storybook/react-vite';
import { FormField } from '../form-field/form-field';
import { Textarea } from './textarea';

const meta = {
  component: Textarea,
  decorators: [
    (Story, { args }) => (
      <FormField label="Reason">{(control) => <Story args={{ ...args, ...control }} />}</FormField>
    ),
  ],
  args: { defaultValue: 'The contributor left the organization.' },
} satisfies Meta<typeof Textarea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Empty: Story = { args: { defaultValue: '' } };
export const Disabled: Story = { args: { disabled: true } };
