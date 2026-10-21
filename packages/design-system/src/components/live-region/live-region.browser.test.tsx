import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Button } from '../button/button';
import { LiveRegionProvider, useAnnounce } from './live-region';

afterEach(cleanup);

function Trigger({ message, urgent }: { message: string; urgent?: boolean }) {
  const announce = useAnnounce();
  return (
    <Button onClick={() => announce(message, urgent ? 'assertive' : 'polite')}>{message}</Button>
  );
}

test('regions exist before any message and receive polite and assertive messages', async () => {
  const { getByRole } = render(
    <LiveRegionProvider>
      <Trigger message="Saved" />
      <Trigger message="Session expired" urgent />
    </LiveRegionProvider>,
  );
  const status = getByRole('status');
  const alert = getByRole('alert');
  expect(status.textContent).toBe('');
  expect(status.getAttribute('aria-live')).toBe('polite');
  expect(alert.getAttribute('aria-live')).toBe('assertive');

  await userEvent.click(getByRole('button', { name: 'Saved' }));
  expect(status.textContent?.trim()).toBe('Saved');
  await userEvent.click(getByRole('button', { name: 'Session expired' }));
  expect(alert.textContent?.trim()).toBe('Session expired');
});

test('a repeated message changes the region content so it is announced again', async () => {
  const { getByRole } = render(
    <LiveRegionProvider>
      <Trigger message="Copied" />
    </LiveRegionProvider>,
  );
  const button = getByRole('button', { name: 'Copied' });
  await userEvent.click(button);
  const first = getByRole('status').textContent;
  await userEvent.click(button);
  expect(getByRole('status').textContent).not.toBe(first);
  expect(getByRole('status').textContent?.trim()).toBe('Copied');
});

test('the regions are visually hidden but not removed from the accessibility tree', () => {
  const { getByRole } = render(<LiveRegionProvider>{null}</LiveRegionProvider>);
  const status = getByRole('status');
  expect(status.getBoundingClientRect().width).toBeLessThanOrEqual(1);
  expect(getComputedStyle(status).display).not.toBe('none');
  expect(getComputedStyle(status).visibility).not.toBe('hidden');
});

test('useAnnounce outside the provider fails loudly', () => {
  expect(() => render(<Trigger message="x" />)).toThrow(/LiveRegionProvider/);
});
