import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '../button/button';
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState } from './state-panel';

const meta = {
  component: EmptyState,
  args: {
    title: 'No agreements yet',
    description: 'Agreements you sign appear here.',
  },
} satisfies Meta<typeof EmptyState>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {
  args: { action: <Button>Sign an agreement</Button> },
};

export const Loading: StoryObj<typeof LoadingState> = {
  args: { title: 'Loading agreements' },
  render: (args) => <LoadingState {...args} />,
};

export const ErrorWithRetry: Story = {
  args: {
    title: 'Agreements could not be loaded',
    description: 'Try again. If the problem continues, contact the maintainers.',
    action: <Button variant="secondary">Try again</Button>,
  },
  render: (args) => <ErrorState {...args} />,
};

export const PermissionDenied: Story = {
  args: {
    title: 'You do not have access to this page',
    description: 'Ask an organization administrator to add you as an authorized signer.',
  },
  render: (args) => <PermissionDeniedState {...args} />,
};
