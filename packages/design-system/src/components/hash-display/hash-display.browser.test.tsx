import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test, vi } from 'vitest';
import { userEvent } from 'vitest/browser';
import { LiveRegionProvider } from '../live-region/live-region';
import { HashDisplay } from './hash-display';
import '../../styles/tokens.css';

const value = 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

function renderHash() {
  return render(
    <LiveRegionProvider>
      <HashDisplay
        label="Document SHA-256"
        value={value}
        copyLabel="Copy"
        copiedMessage="Hash copied."
        copyFailedMessage="Copying failed. Select the value to copy it manually."
      />
    </LiveRegionProvider>,
  );
}

test('shows the full value in the monospace font with an accessible name', () => {
  const { getByRole } = renderHash();
  const figure = getByRole('figure', { name: 'Document SHA-256' });
  const code = figure.querySelector('code') as HTMLElement;
  expect(code.textContent).toBe(value);
  expect(getComputedStyle(code).fontFamily).toMatch(/monospace/);
  const copy = getByRole('button', { name: 'Copy' });
  expect(document.getElementById(copy.getAttribute('aria-describedby') ?? '')?.textContent).toBe(
    'Document SHA-256',
  );
});

test('copying announces success politely', async () => {
  const writeText = vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue();
  const { getByRole } = renderHash();
  getByRole('button', { name: 'Copy' }).focus();
  await userEvent.keyboard('{Enter}');

  expect(writeText).toHaveBeenCalledWith(value);
  await expect.poll(() => getByRole('status').textContent?.trim()).toBe('Hash copied.');
});

test('a rejected copy announces the failure and the value stays selectable', async () => {
  vi.spyOn(navigator.clipboard, 'writeText').mockRejectedValue(
    new DOMException('', 'NotAllowedError'),
  );
  const { getByRole } = renderHash();
  await userEvent.click(getByRole('button', { name: 'Copy' }));

  await expect
    .poll(() => getByRole('alert').textContent?.trim())
    .toBe('Copying failed. Select the value to copy it manually.');
  const code = getByRole('figure').querySelector('code') as HTMLElement;
  expect(getComputedStyle(code).userSelect).toBe('all');
});

test('a missing clipboard API is a failure, not an error', async () => {
  vi.spyOn(navigator, 'clipboard', 'get').mockReturnValue(undefined as unknown as Clipboard);
  const { getByRole } = renderHash();
  await userEvent.click(getByRole('button', { name: 'Copy' }));

  await expect
    .poll(() => getByRole('alert').textContent?.trim())
    .toBe('Copying failed. Select the value to copy it manually.');
});
