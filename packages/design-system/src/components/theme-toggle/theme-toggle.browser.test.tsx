import { cleanup, render } from '@testing-library/react';
import { afterEach, beforeEach, expect, test } from 'vitest';
import { userEvent } from 'vitest/browser';
import { themeStorageKey } from '../../theme';
import { ThemeToggle } from './theme-toggle';

const labels = { light: 'Light', dark: 'Dark', system: 'Match system' };

beforeEach(() => {
  localStorage.removeItem(themeStorageKey);
  document.documentElement.removeAttribute('data-theme');
});
afterEach(cleanup);

test('starts from the stored preference', () => {
  localStorage.setItem(themeStorageKey, 'dark');
  const { getByRole } = render(<ThemeToggle legend="Theme" labels={labels} />);
  expect(getByRole('radio', { name: 'Dark' })).toHaveProperty('checked', true);
});

test('selecting with the keyboard applies and stores the theme; system clears both', async () => {
  const { getByRole } = render(<ThemeToggle legend="Theme" labels={labels} />);
  expect(getByRole('radio', { name: 'Match system' })).toHaveProperty('checked', true);

  await userEvent.click(getByRole('radio', { name: 'Light' }));
  await userEvent.keyboard('{ArrowDown}');
  expect(getByRole('radio', { name: 'Dark' })).toHaveProperty('checked', true);
  expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  expect(localStorage.getItem(themeStorageKey)).toBe('dark');

  await userEvent.keyboard('{ArrowDown}');
  expect(document.documentElement.hasAttribute('data-theme')).toBe(false);
  expect(localStorage.getItem(themeStorageKey)).toBeNull();
});
