import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { Button } from '../button/button';
import { FormField } from '../form-field/form-field';
import { Input } from '../input/input';
import { Dialog } from './dialog';

function Demo({ title, description }: { title: string; description?: string }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button onClick={() => setOpen(true)}>Open dialog</Button>
      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        title={title}
        {...(description && { description })}
        actions={
          <>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button onClick={() => setOpen(false)}>Save</Button>
          </>
        }
      >
        <FormField label="Organization name">
          {(control) => <Input {...control} defaultValue="Example Org" />}
        </FormField>
      </Dialog>
    </>
  );
}

const meta = {
  component: Demo,
  args: {
    title: 'Edit organization',
    description: 'The new name is shown on future agreements.',
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Open dialog' }));
  },
} satisfies Meta<typeof Demo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const WithoutDescription: Story = { args: { title: 'Edit organization' } };
