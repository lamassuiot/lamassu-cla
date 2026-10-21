import { act, cleanup, render } from '@testing-library/react';
import { afterEach, expect, test, vi } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Button } from '../button/button';
import { LiveRegionProvider } from '../live-region/live-region';
import { ToastProvider, type ToastStatus, useToast } from './toast';

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

function Trigger({ message, status }: { message: string; status?: ToastStatus }) {
  const show = useToast();
  return <Button onClick={() => show({ message, ...(status && { status }) })}>{message}</Button>;
}

function renderToasts(triggers: { message: string; status?: ToastStatus }[]) {
  return render(
    <LiveRegionProvider>
      <ToastProvider labels={{ region: 'Notifications', dismiss: 'Dismiss' }}>
        {triggers.map((trigger) => (
          <Trigger key={trigger.message} {...trigger} />
        ))}
      </ToastProvider>
    </LiveRegionProvider>,
  );
}

test('showing a toast announces it and does not move focus', async () => {
  const { getByRole } = renderToasts([{ message: 'Saved', status: 'success' }]);
  const trigger = getByRole('button', { name: 'Saved' });
  await userEvent.click(trigger);

  const region = getByRole('region', { name: 'Notifications' });
  expect(region.textContent).toContain('Saved');
  expect(getByRole('status').textContent?.trim()).toBe('Saved');
  expect(document.activeElement).toBe(trigger);
});

test('a danger toast is announced assertively', async () => {
  const { getByRole } = renderToasts([{ message: 'Signing failed', status: 'danger' }]);
  await userEvent.click(getByRole('button', { name: 'Signing failed' }));
  expect(getByRole('alert').textContent?.trim()).toBe('Signing failed');
});

test('toasts do not disappear on a timer', async () => {
  const { getByRole } = renderToasts([{ message: 'Saved' }]);
  // Fake timers before showing the toast, so any timer it schedules can be advanced.
  vi.useFakeTimers({ toFake: ['setTimeout', 'setInterval'] });
  await userEvent.click(getByRole('button', { name: 'Saved' }));
  await act(() => vi.advanceTimersByTimeAsync(10 * 60 * 1000));
  expect(getByRole('region', { name: 'Notifications' }).textContent).toContain('Saved');
});

test('each dismiss button is described by its message', async () => {
  const { getByRole, getAllByRole } = renderToasts([{ message: 'First' }, { message: 'Second' }]);
  await userEvent.click(getByRole('button', { name: 'First' }));
  await userEvent.click(getByRole('button', { name: 'Second' }));
  const dismissButtons = getAllByRole('button', { name: 'Dismiss' });
  expect(
    dismissButtons.map(
      (button) =>
        document.getElementById(button.getAttribute('aria-describedby') ?? '')?.textContent,
    ),
  ).toEqual(['First', 'Second']);
});

test('dismissing with the keyboard keeps focus in the toast region', async () => {
  const { getByRole, getAllByRole } = renderToasts([{ message: 'First' }, { message: 'Second' }]);
  await userEvent.click(getByRole('button', { name: 'First' }));
  await userEvent.click(getByRole('button', { name: 'Second' }));
  const region = getByRole('region', { name: 'Notifications' });

  getAllByRole('button', { name: 'Dismiss' })[0]?.focus();
  await userEvent.keyboard('{Enter}');
  const remaining = getAllByRole('button', { name: 'Dismiss' });
  expect(remaining).toHaveLength(1);
  expect(document.activeElement).toBe(remaining[0]);

  await userEvent.keyboard('{Enter}');
  expect(region.textContent).not.toContain('Second');
  expect(document.activeElement).toBe(region);
});
