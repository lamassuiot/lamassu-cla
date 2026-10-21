import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { Button } from '../button/button';
import { ConfirmationDialog, type ConfirmationDialogProps } from './confirmation-dialog';

type DemoProps = Pick<
  ConfirmationDialogProps,
  'title' | 'description' | 'confirmLabel' | 'cancelLabel' | 'destructive'
>;

function Demo(props: DemoProps) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button onClick={() => setOpen(true)}>Open confirmation</Button>
      <ConfirmationDialog
        {...props}
        open={open}
        onConfirm={() => setOpen(false)}
        onCancel={() => setOpen(false)}
      />
    </>
  );
}

const meta = {
  component: Demo,
  args: {
    title: 'Publish version 2.0?',
    description: 'Published versions cannot be changed. Corrections need a new version.',
    confirmLabel: 'Publish',
    cancelLabel: 'Cancel',
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Open confirmation' }));
  },
} satisfies Meta<typeof Demo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Destructive: Story = {
  args: {
    title: 'Revoke this authorization?',
    description: 'Contributors covered only by this authorization will no longer be covered.',
    confirmLabel: 'Revoke',
    destructive: true,
  },
};
