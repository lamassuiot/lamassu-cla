import '@lamassu-cla/design-system/tokens.css';
import '@lamassu-cla/design-system/base.css';
import { cleanup, render } from '@testing-library/react';
import { afterEach, beforeEach, expect, test, vi } from 'vitest';
import { userEvent } from 'vitest/browser';
import { createMemoryRouter, type RouteObject } from 'react-router';
import { RouterProvider } from 'react-router/dom';
import { copy } from './copy/en';
import { portalRoutes, routes } from './routes';

const marker = Symbol('no full page load');

beforeEach(() => {
  (window as unknown as Record<symbol, boolean>)[marker] = true;
});
afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  localStorage.clear();
});

function renderAt(path: string, routeTree: RouteObject[] = routes) {
  const router = createMemoryRouter(routeTree, { initialEntries: [path] });
  return { router, ...render(<RouterProvider router={router} />) };
}

test('each route renders inside the shell with one h1 and the skip link first', async () => {
  for (const [path, title] of [
    ['/', copy.landing.title],
    ['/no-such-page', copy.notFound.title],
  ] as const) {
    const { findByRole, getByRole, getAllByRole } = renderAt(path);
    expect(await findByRole('heading', { level: 1, name: title })).toBeTruthy();
    expect(getAllByRole('heading', { level: 1 })).toHaveLength(1);
    expect(getAllByRole('banner')[0]?.closest('main')).toBeNull();
    expect(getByRole('main')).toBeTruthy();
    await userEvent.keyboard('{Tab}');
    expect(document.activeElement).toBe(getByRole('link', { name: copy.shell.skipToContent }));
    cleanup();
  }
});

test('navigation routes without a full load, sets the title, focuses the h1, and marks the link', async () => {
  const { router, findByRole, getByRole } = renderAt('/no-such-page');
  await findByRole('heading', { level: 1, name: copy.notFound.title });
  expect(document.title).toBe(`${copy.notFound.title} – ${copy.brand}`);
  const before = window.location.href;

  await userEvent.click(getByRole('link', { name: copy.nav.home }));
  const heading = await findByRole('heading', { level: 1, name: copy.landing.title });
  expect(router.state.location.pathname).toBe('/');
  expect(window.location.href).toBe(before);
  expect((window as unknown as Record<symbol, boolean>)[marker]).toBe(true);
  expect(document.title).toBe(`${copy.landing.title} – ${copy.brand}`);
  expect(document.activeElement).toBe(heading);
  expect(getByRole('link', { name: copy.nav.home }).getAttribute('aria-current')).toBe('page');
});

test('modified clicks keep the browser default and do not route', async () => {
  const { router, findByRole, getByRole } = renderAt('/no-such-page');
  await findByRole('heading', { level: 1, name: copy.notFound.title });
  const seen: boolean[] = [];
  // Runs after React's handler; records the outcome, then stops the browser from opening a tab.
  const record = (event: MouseEvent) => {
    seen.push(event.defaultPrevented);
    event.preventDefault();
  };
  document.addEventListener('click', record);
  for (const modifier of ['ctrlKey', 'metaKey', 'shiftKey'] as const) {
    getByRole('link', { name: copy.nav.home }).dispatchEvent(
      new MouseEvent('click', { bubbles: true, cancelable: true, [modifier]: true }),
    );
  }
  document.removeEventListener('click', record);
  expect(seen).toEqual([false, false, false]);
  expect(router.state.location.pathname).toBe('/no-such-page');
});

test('an unknown path shows the not-found screen with a link to the landing page', async () => {
  const { findByRole, getByRole } = renderAt('/agreements/none');
  await findByRole('heading', { level: 1, name: copy.notFound.title });
  await userEvent.click(getByRole('link', { name: copy.notFound.home }));
  expect(await findByRole('heading', { level: 1, name: copy.landing.title })).toBeTruthy();
});

test.each([
  ['a page whose code fails to load', { lazy: () => Promise.reject(new Error('secret detail')) }],
  [
    'a page that throws while rendering',
    {
      Component: () => {
        throw new Error('secret detail');
      },
    },
  ],
])('%s shows the error screen inside the shell without details', async (_name, page) => {
  vi.spyOn(console, 'error').mockImplementation(() => undefined);
  vi.spyOn(console, 'warn').mockImplementation(() => undefined);
  const { findByRole, getByRole, getAllByRole, container } = renderAt(
    '/',
    portalRoutes([{ index: true, ...page } as RouteObject]),
  );
  expect(await findByRole('heading', { level: 1, name: copy.error.title })).toBeTruthy();
  expect(getByRole('button', { name: copy.error.reload })).toBeTruthy();
  expect(getAllByRole('banner')[0]?.closest('main')).toBeNull();
  expect(container.textContent).not.toContain('secret detail');
  expect(document.title).toBe(`${copy.error.title} – ${copy.brand}`);
});

test('the theme toggle switches the theme, and the choice survives a remount', async () => {
  const first = renderAt('/');
  await first.findByRole('heading', { level: 1, name: copy.landing.title });
  await userEvent.click(first.getByRole('radio', { name: copy.theme.labels.dark }));
  expect(document.documentElement.dataset.theme).toBe('dark');
  cleanup();

  const second = renderAt('/');
  await second.findByRole('heading', { level: 1, name: copy.landing.title });
  expect(
    (second.getByRole('radio', { name: copy.theme.labels.dark }) as HTMLInputElement).checked,
  ).toBe(true);
  expect(document.documentElement.dataset.theme).toBe('dark');
});
