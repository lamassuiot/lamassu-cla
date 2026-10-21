import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { Button } from '../button/button';
import { Drawer, type DrawerProps } from './drawer';

function Demo({ side }: Pick<DrawerProps, 'side'>) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button onClick={() => setOpen(true)}>Open menu</Button>
      <Drawer
        open={open}
        onClose={() => setOpen(false)}
        title="Menu"
        closeLabel="Close menu"
        {...(side && { side })}
      >
        <nav aria-label="Main">
          <ul>
            <li>
              <a href="/agreements">Agreements</a>
            </li>
            <li>
              <a href="/organizations">Organizations</a>
            </li>
          </ul>
        </nav>
      </Drawer>
    </>
  );
}

const meta = {
  component: Demo,
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'Open menu' }));
  },
} satisfies Meta<typeof Demo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const End: Story = {};
export const Start: Story = { args: { side: 'start' } };
