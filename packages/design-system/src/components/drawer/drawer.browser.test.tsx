import { cleanup, render } from '@testing-library/react';
import { useState } from 'react';
import { afterEach, expect, test } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Button } from '../button/button';
import { Drawer } from './drawer';

afterEach(cleanup);

function Harness() {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button onClick={() => setOpen(true)}>Menu</Button>
      <Drawer open={open} onClose={() => setOpen(false)} title="Navigation" closeLabel="Close">
        <a href="/agreements">Agreements</a>
      </Drawer>
    </>
  );
}

async function openDrawer() {
  const view = render(<Harness />);
  const opener = view.getByRole('button', { name: 'Menu' });
  await userEvent.click(opener);
  return { ...view, opener, drawer: view.getByRole('dialog', { name: 'Navigation' }) };
}

test('opening moves focus into the drawer, and Tab never reaches the page', async () => {
  const { getByRole, drawer } = await openDrawer();
  expect(drawer.matches(':modal')).toBe(true);
  expect(document.activeElement).toBe(getByRole('button', { name: 'Close' }));
  for (let i = 0; i < 4; i++) {
    await userEvent.keyboard('{Tab}');
    expect(
      document.activeElement === document.body || drawer.contains(document.activeElement),
    ).toBe(true);
  }
});

test('Escape closes the drawer and returns focus to the opener', async () => {
  const { queryByRole, opener } = await openDrawer();
  await userEvent.keyboard('{Escape}');
  expect(queryByRole('dialog')).toBeNull();
  expect(document.activeElement).toBe(opener);
});

test('the close button closes the drawer and returns focus to the opener', async () => {
  const { queryByRole, opener } = await openDrawer();
  await userEvent.keyboard('{Enter}');
  expect(queryByRole('dialog')).toBeNull();
  expect(document.activeElement).toBe(opener);
});
