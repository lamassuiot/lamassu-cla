import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test, vi } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Button } from './button';

afterEach(cleanup);

// Real key presses through Playwright (AC-05-3); simulated events misreport focus.
test('is reachable with Tab and activated with Enter and Space', async () => {
  const onClick = vi.fn();
  const { getByRole } = render(<Button onClick={onClick}>Sign the agreement</Button>);
  const button = getByRole('button', { name: 'Sign the agreement' });

  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(button);
  expect(button.matches(':focus-visible')).toBe(true);

  await userEvent.keyboard('{Enter}');
  await userEvent.keyboard(' ');
  expect(onClick).toHaveBeenCalledTimes(2);
});

test('is skipped by Tab and not activated when disabled', async () => {
  const onClick = vi.fn();
  const { getByRole } = render(
    <>
      <Button disabled onClick={onClick}>
        Disabled
      </Button>
      <Button>Next</Button>
    </>,
  );
  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('button', { name: 'Next' }));
  expect(onClick).not.toHaveBeenCalled();
});

test('defaults to type="button" so it never submits a form by accident', () => {
  const { getByRole } = render(<Button>Save</Button>);
  expect(getByRole('button').getAttribute('type')).toBe('button');
});
