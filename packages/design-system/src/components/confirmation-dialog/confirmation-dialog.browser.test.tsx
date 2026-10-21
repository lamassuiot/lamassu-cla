import { cleanup, render } from '@testing-library/react';
import { useState } from 'react';
import { afterEach, expect, test, vi } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Button } from '../button/button';
import { ConfirmationDialog } from './confirmation-dialog';

afterEach(cleanup);

function Harness({ onConfirm, onCancel }: { onConfirm: () => void; onCancel: () => void }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <Button onClick={() => setOpen(true)}>Revoke</Button>
      <ConfirmationDialog
        open={open}
        destructive
        title="Revoke this authorization?"
        description="Covered contributors lose coverage."
        confirmLabel="Revoke authorization"
        cancelLabel="Keep authorization"
        onConfirm={() => {
          onConfirm();
          setOpen(false);
        }}
        onCancel={() => {
          onCancel();
          setOpen(false);
        }}
      />
    </>
  );
}

async function openConfirmation() {
  const onConfirm = vi.fn();
  const onCancel = vi.fn();
  const view = render(<Harness onConfirm={onConfirm} onCancel={onCancel} />);
  const opener = view.getByRole('button', { name: 'Revoke' });
  await userEvent.click(opener);
  return { ...view, opener, onConfirm, onCancel };
}

test('focus starts on cancel, so Enter does not confirm', async () => {
  const { getByRole, queryByRole, opener, onConfirm, onCancel } = await openConfirmation();
  expect(getByRole('dialog', { name: 'Revoke this authorization?' })).toBeTruthy();
  expect(document.activeElement).toBe(getByRole('button', { name: 'Keep authorization' }));

  await userEvent.keyboard('{Enter}');
  expect(onCancel).toHaveBeenCalledTimes(1);
  expect(onConfirm).not.toHaveBeenCalled();
  expect(queryByRole('dialog')).toBeNull();
  expect(document.activeElement).toBe(opener);
});

test('Escape cancels and does not confirm', async () => {
  const { queryByRole, opener, onConfirm, onCancel } = await openConfirmation();
  await userEvent.keyboard('{Escape}');

  expect(onCancel).toHaveBeenCalledTimes(1);
  expect(onConfirm).not.toHaveBeenCalled();
  expect(queryByRole('dialog')).toBeNull();
  expect(document.activeElement).toBe(opener);
});

test('only the explicit confirming action confirms', async () => {
  const { getByRole, opener, onConfirm, onCancel } = await openConfirmation();
  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('button', { name: 'Revoke authorization' }));
  await userEvent.keyboard('{Enter}');

  expect(onConfirm).toHaveBeenCalledTimes(1);
  expect(onCancel).not.toHaveBeenCalled();
  expect(document.activeElement).toBe(opener);
});
