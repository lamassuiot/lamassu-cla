import { cleanup, render } from '@testing-library/react';
import { useState } from 'react';
import { afterEach, expect, test, vi } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Button } from '../button/button';
import { Dialog } from './dialog';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

function Harness({
  onClose,
  title = 'Edit organization',
}: {
  onClose?: () => void;
  title?: string;
}) {
  const [open, setOpen] = useState(false);
  const [mounted, setMounted] = useState(true);
  const close = () => {
    onClose?.();
    setOpen(false);
  };
  return (
    <>
      <Button onClick={() => setOpen(true)}>Open</Button>
      <Button variant="secondary">Behind</Button>
      {mounted && (
        <Dialog
          open={open}
          onClose={close}
          title={title}
          description="Shown on future agreements."
          actions={
            <>
              <Button variant="secondary" onClick={close}>
                Cancel
              </Button>
              <Button onClick={() => setMounted(false)}>Remove</Button>
            </>
          }
        >
          <input aria-label="Name" />
        </Dialog>
      )}
    </>
  );
}

const inside = (dialog: HTMLElement) =>
  document.activeElement === document.body || dialog.contains(document.activeElement);

test('opening moves focus into a named, described dialog and makes the page inert', async () => {
  const { getByRole } = render(<Harness />);
  await userEvent.click(getByRole('button', { name: 'Open' }));

  const dialog = getByRole('dialog', { name: 'Edit organization' });
  expect(dialog.matches(':modal')).toBe(true);
  expect(dialog.getAttribute('aria-describedby')).toBeTruthy();
  expect(document.getElementById(dialog.getAttribute('aria-describedby') ?? '')?.textContent).toBe(
    'Shown on future agreements.',
  );
  expect(document.activeElement).toBe(getByRole('textbox', { name: 'Name' }));

  const behind = document.body.querySelector('button:not(dialog button)') as HTMLButtonElement;
  behind.focus();
  expect(dialog.contains(document.activeElement)).toBe(true);
});

test('Tab never reaches the page behind the dialog', async () => {
  const { getByRole } = render(<Harness />);
  await userEvent.click(getByRole('button', { name: 'Open' }));
  const dialog = getByRole('dialog');
  // The native modal dialog lets focus reach the browser's own controls (AC-05-5), never the page.
  for (let i = 0; i < 6; i++) {
    await userEvent.keyboard('{Tab}');
    expect(inside(dialog)).toBe(true);
  }
  for (let i = 0; i < 6; i++) {
    await userEvent.keyboard('{Shift>}{Tab}{/Shift}');
    expect(inside(dialog)).toBe(true);
  }
});

test('Escape closes the dialog and returns focus to the opener', async () => {
  const onClose = vi.fn();
  const { getByRole, queryByRole } = render(<Harness onClose={onClose} />);
  const opener = getByRole('button', { name: 'Open' });
  await userEvent.click(opener);
  await userEvent.keyboard('{Escape}');

  expect(onClose).toHaveBeenCalledTimes(1);
  expect(queryByRole('dialog')).toBeNull();
  expect(document.activeElement).toBe(opener);
});

test('closing with an action returns focus to the opener', async () => {
  const { getByRole, queryByRole } = render(<Harness />);
  const opener = getByRole('button', { name: 'Open' });
  await userEvent.click(opener);
  await userEvent.click(getByRole('button', { name: 'Cancel' }));

  expect(queryByRole('dialog')).toBeNull();
  expect(document.activeElement).toBe(opener);
});

test('removing an open dialog returns focus to the opener', async () => {
  const { getByRole, queryByRole } = render(<Harness />);
  const opener = getByRole('button', { name: 'Open' });
  await userEvent.click(opener);
  await userEvent.click(getByRole('button', { name: 'Remove' }));

  expect(queryByRole('dialog')).toBeNull();
  expect(document.activeElement).toBe(opener);
});

test('a dialog without an accessible name fails', () => {
  // React logs the thrown render error; keep the test output readable.
  vi.spyOn(console, 'error').mockImplementation(() => undefined);
  expect(() => render(<Harness title=" " />)).toThrow(/non-empty title/);
});
