// All user-visible portal text (F-06 AC-06-6). English only until D-21; screens never hold literal text.
export const copy = {
  brand: 'Lamassu CLA',
  shell: {
    skipToContent: 'Skip to content',
    menu: 'Menu',
    menuTitle: 'Navigation',
    closeMenu: 'Close navigation',
    navigation: 'Main',
  },
  theme: {
    legend: 'Theme',
    labels: { light: 'Light', dark: 'Dark', system: 'System' },
  },
  nav: { home: 'Home' },
  loading: { title: 'Loading' },
  // The landing description is added with the approved copy (D-27).
  landing: { title: 'Lamassu CLA' },
  notFound: {
    title: 'Page not found',
    description: 'The page you requested does not exist.',
    home: 'Go to the home page',
  },
  error: {
    title: 'Something went wrong',
    description: 'This page could not be shown. Reload to try again.',
    reload: 'Reload',
  },
} as const;

// `<page title> – <brand>`; the landing page uses the brand alone.
export function documentTitle(pageTitle?: string): string {
  return pageTitle && pageTitle !== copy.brand ? `${pageTitle} – ${copy.brand}` : copy.brand;
}
