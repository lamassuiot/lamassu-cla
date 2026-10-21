import type { Meta, StoryObj } from '@storybook/react-vite';
import { StatusBadge } from './status-badge';

const meta = {
  component: StatusBadge,
  args: { state: 'ACTIVE', label: 'Active' },
} satisfies Meta<typeof StatusBadge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Success: Story = {};
export const Warning: Story = { args: { state: 'PENDING_SIGNATURE', label: 'Pending signature' } };
export const Danger: Story = { args: { state: 'REVOKED', label: 'Revoked' } };
export const Info: Story = { args: { state: 'DRAFT', label: 'Draft' } };
export const Neutral: Story = { args: { state: 'SUPERSEDED', label: 'Superseded' } };
export const UnknownState: Story = { args: { state: 'NOT_A_STATE', label: 'Unknown' } };
export const InformationalNotice: StoryObj<typeof StatusBadge> = {
  args: { status: 'info', label: 'Notice' },
  render: (args) => <StatusBadge {...args} />,
};
