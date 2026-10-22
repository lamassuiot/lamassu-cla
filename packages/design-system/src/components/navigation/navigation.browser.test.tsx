import '../../styles/tokens.css';
import '../../styles/base.css';
import { cleanup, render } from '@testing-library/react';
import { afterEach, beforeEach, expect, test, vi } from 'vitest';
import { userEvent } from 'vitest/browser';
import { applyTheme } from '../../theme';
import { isSafeHref, Navigation } from './navigation';

beforeEach(() => applyTheme('light'));
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

test('the landmark is labelled and the current page is marked', () => {
  const { getByRole } = render(
    <Navigation
      label="Main"
      items={[
        { href: '/dashboard', label: 'Dashboard', current: true },
        { href: '/agreements', label: 'Agreements' },
      ]}
    />,
  );
  expect(getByRole('navigation', { name: 'Main' })).toBeTruthy();
  expect(getByRole('link', { name: 'Dashboard' }).getAttribute('aria-current')).toBe('page');
  expect(getByRole('link', { name: 'Agreements' }).hasAttribute('aria-current')).toBe(false);
});

test.each(['/agreements', '/a/b?c=1#d', 'https://example.com/docs'])('accepts %s', (href) => {
  expect(isSafeHref(href)).toBe(true);
});

test.each([
  'javascript:alert(1)',
  'JavaScript:alert(1)',
  'data:text/html,x',
  'http://example.com',
  '//example.com',
  '/\\example.com',
  'agreements',
  '',
])('rejects %s', (href) => {
  expect(isSafeHref(href)).toBe(false);
  vi.spyOn(console, 'error').mockImplementation(() => undefined);
  expect(() => render(<Navigation label="Main" items={[{ href, label: 'Bad' }]} />)).toThrow(
    /relative path or an https: URL/,
  );
});

test('keyboard focus on bg.nav uses focus.ring-on-nav, which has 3:1 contrast in light theme', async () => {
  const { getByRole } = render(
    <Navigation label="Main" items={[{ href: '/dashboard', label: 'Dashboard' }]} />,
  );
  await userEvent.keyboard('{Tab}');
  const link = getByRole('link', { name: 'Dashboard' });
  expect(document.activeElement).toBe(link);
  // orange.400 (#F78C40), not the light-theme focus.ring orange.700 (#B4500A).
  expect(getComputedStyle(link).outlineColor).toBe('rgb(247, 140, 64)');
});

test('onNavigate receives the click event and the href, and can prevent the default', async () => {
  const calls: string[] = [];
  const { getByRole } = render(
    <Navigation
      label="Main"
      items={[{ href: '/agreements', label: 'Agreements' }]}
      onNavigate={(event, href) => {
        event.preventDefault();
        calls.push(href);
      }}
    />,
  );
  const before = window.location.href;
  await userEvent.click(getByRole('link', { name: 'Agreements' }));
  expect(calls).toEqual(['/agreements']);
  expect(window.location.href).toBe(before);
});
