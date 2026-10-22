import { cleanup, render } from '@testing-library/react';
import { afterEach, beforeEach, expect, test } from 'vitest';
import { page, userEvent } from 'vitest/browser';
import { AppShell } from './app-shell';

const preventNavigation = (event: MouseEvent) => {
  if ((event.target as Element).closest('a[href^="/"]')) {
    event.preventDefault();
  }
};

beforeEach(() => document.addEventListener('click', preventNavigation));
afterEach(async () => {
  cleanup();
  document.removeEventListener('click', preventNavigation);
  await page.viewport(414, 896);
});

function renderShell(onNavigate?: (href: string) => void) {
  return render(
    <AppShell
      labels={{
        skipToContent: 'Skip to content',
        menu: 'Menu',
        menuTitle: 'Navigation',
        closeMenu: 'Close navigation',
      }}
      brand={<span>Lamassu CLA</span>}
      navigation={{
        label: 'Main',
        items: [
          { href: '/dashboard', label: 'Dashboard', current: true },
          { href: '/agreements', label: 'Agreements' },
        ],
        ...(onNavigate && { onNavigate: (_event, href) => onNavigate(href) }),
      }}
    >
      <h1>Dashboard</h1>
    </AppShell>,
  );
}

test('the skip link is the first tab stop and moves focus to the main content', async () => {
  const { getByRole } = renderShell();
  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('link', { name: 'Skip to content' }));
  await userEvent.keyboard('{Enter}');
  expect(document.activeElement).toBe(getByRole('main'));
  expect(getByRole('banner')).toBeTruthy();
});

test('below bp.lg the navigation is in a drawer; following a link closes it', async () => {
  await page.viewport(375, 800);
  const { getByRole, queryByRole } = renderShell();
  expect(queryByRole('navigation')).toBeNull();

  const menu = getByRole('button', { name: 'Menu' });
  await userEvent.click(menu);
  const drawer = getByRole('dialog', { name: 'Navigation' });
  expect(drawer.contains(getByRole('navigation', { name: 'Main' }))).toBe(true);

  await userEvent.click(getByRole('link', { name: 'Agreements' }));
  expect(queryByRole('dialog')).toBeNull();
  expect(document.activeElement).toBe(menu);
});

test("the caller's onNavigate runs for drawer links, and the drawer still closes", async () => {
  await page.viewport(375, 800);
  const followed: string[] = [];
  const { getByRole, queryByRole } = renderShell((href) => followed.push(href));
  await userEvent.click(getByRole('button', { name: 'Menu' }));
  await userEvent.click(getByRole('link', { name: 'Agreements' }));
  expect(followed).toEqual(['/agreements']);
  expect(queryByRole('dialog')).toBeNull();
});

test('from bp.lg the navigation is a sidebar and there is no menu button', async () => {
  await page.viewport(1280, 800);
  const { getByRole, queryByRole } = renderShell();
  expect(queryByRole('button', { name: 'Menu' })).toBeNull();
  const navigation = getByRole('navigation', { name: 'Main' });
  expect(navigation.closest('dialog')).toBeNull();
  expect(getByRole('link', { name: 'Dashboard' }).getAttribute('aria-current')).toBe('page');
});
