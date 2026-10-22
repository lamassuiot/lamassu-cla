import '@lamassu-cla/design-system/tokens.css';
import '@lamassu-cla/design-system/base.css';
import { applyTheme } from '@lamassu-cla/design-system';
import { cleanup, render } from '@testing-library/react';
import axe from 'axe-core';
import { afterEach, expect, test, vi } from 'vitest';
import { page } from 'vitest/browser';
import { createMemoryRouter, type RouteObject } from 'react-router';
import { RouterProvider } from 'react-router/dom';
import { copy } from './copy/en';
import { portalRoutes, routes } from './routes';

// F06-T5 (AC-06-5): every screen, in both themes, at a wide window, at 200% zoom of a 1280 px
// window (640 px), and at the 320 px reflow width.
afterEach(async () => {
  cleanup();
  vi.restoreAllMocks();
  await page.viewport(1280, 900);
});

const failingPage: RouteObject = {
  index: true,
  Component: () => {
    throw new Error('screen failed');
  },
};

const screens: [string, () => RouteObject[], string, string][] = [
  ['landing', () => routes, '/', copy.landing.title],
  ['not found', () => routes, '/no-such-page', copy.notFound.title],
  ['error', () => portalRoutes([failingPage]), '/', copy.error.title],
];

const cases = screens.flatMap((screen) =>
  (['light', 'dark'] as const).flatMap((theme) =>
    [1280, 640, 320].map((width) => [...screen, theme, width] as const),
  ),
);

test.each(cases)(
  '%s screen (%s) in the %s theme at %s px',
  async (_n, tree, path, title, theme, width) => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    applyTheme(theme);
    await page.viewport(width, 900);
    const router = createMemoryRouter(tree(), { initialEntries: [path] });
    const { findByRole } = render(<RouterProvider router={router} />);
    await findByRole('heading', { level: 1, name: title });

    const { violations } = await axe.run(document, {
      runOnly: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'],
    });
    expect(violations.map((v) => `${v.id}: ${v.nodes.map((n) => n.target).join(', ')}`)).toEqual(
      [],
    );

    const { scrollWidth, clientWidth } = document.documentElement;
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth);
  },
);
