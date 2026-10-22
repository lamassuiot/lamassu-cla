import type { RouteObject } from 'react-router';
import { ErrorScreen } from './screens/error-screen';
import { PortalShell } from './shell/portal-shell';

// Pages render inside the shell; the error screen is bundled with the shell so it works when a
// page's chunk fails to load (AC-06-4).
export function portalRoutes(pages: RouteObject[]): RouteObject[] {
  return [
    { element: <PortalShell />, children: [{ errorElement: <ErrorScreen />, children: pages }] },
  ];
}

// Screens load on demand (route-level code splitting).
export const routes = portalRoutes([
  {
    index: true,
    lazy: async () => ({
      Component: (await import('./screens/landing-screen')).LandingScreen,
    }),
  },
  {
    path: '*',
    lazy: async () => ({
      Component: (await import('./screens/not-found-screen')).NotFoundScreen,
    }),
  },
]);
