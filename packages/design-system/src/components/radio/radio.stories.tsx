import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { RadioGroup, type RadioGroupProps } from './radio';

type Contribution = 'individual' | 'organization';

function Demo(props: Omit<RadioGroupProps<Contribution>, 'value' | 'onChange'>) {
  const [value, setValue] = useState<Contribution | undefined>('individual');
  return <RadioGroup {...props} value={value} onChange={setValue} />;
}

const meta = {
  component: Demo,
  args: {
    legend: 'You contribute',
    name: 'contribution',
    options: [
      { value: 'individual', label: 'As an individual', hint: 'Individual agreement' },
      { value: 'organization', label: 'For an organization', hint: 'Corporate agreement' },
    ],
  },
} satisfies Meta<typeof Demo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Disabled: Story = { args: { disabled: true } };
