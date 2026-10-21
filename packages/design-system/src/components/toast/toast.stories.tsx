import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '../button/button';
import { LiveRegionProvider } from '../live-region/live-region';
import { ToastProvider, type ToastStatus, useToast } from './toast';

function Demo({ status, message }: { status: ToastStatus; message: string }) {
  const show = useToast();
  return <Button onClick={() => show({ message, status })}>Show notification</Button>;
}

const meta = {
  component: Demo,
  decorators: [
    (Story) => (
      <LiveRegionProvider>
        <ToastProvider labels={{ region: 'Notifications', dismiss: 'Dismiss' }}>
          <Story />
        </ToastProvider>
      </LiveRegionProvider>
    ),
  ],
  args: { status: 'success', message: 'Agreement signed.' },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Show notification' }));
  },
} satisfies Meta<typeof Demo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Success: Story = {};
export const Info: Story = { args: { status: 'info', message: 'Verification is in progress.' } };
export const Warning: Story = {
  args: { status: 'warning', message: 'Your session expires in five minutes.' },
};
export const Danger: Story = {
  args: { status: 'danger', message: 'The signing session could not be started.' },
};
