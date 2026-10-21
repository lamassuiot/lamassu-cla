import type { Meta, StoryObj } from '@storybook/react-vite';
import { StepIndicator } from './step-indicator';

const meta = {
  component: StepIndicator,
  args: {
    label: 'Signing progress',
    steps: ['Review', 'Sign', 'Verify', 'Done'],
    current: 2,
    completedLabel: 'completed',
  },
} satisfies Meta<typeof StepIndicator>;

export default meta;
type Story = StoryObj<typeof meta>;

export const InProgress: Story = {};
export const FirstStep: Story = { args: { current: 0 } };
