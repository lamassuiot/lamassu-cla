import { AppShell, LoadingState, ThemeToggle } from '@lamassu-cla/design-system';
import { useLayoutEffect, useRef } from 'react';
import { Outlet, useLocation, useNavigate, useNavigation } from 'react-router';
import { copy } from '../copy/en';
import { shouldRoute } from './should-route';

const items = [{ href: '/', label: copy.nav.home }];

export function PortalShell() {
  const location = useLocation();
  const navigate = useNavigate();
  const navigation = useNavigation();
  const firstRender = useRef(true);

  // AC-06-2: after each navigation, focus starts at the new page's heading.
  // biome-ignore lint/correctness/useExhaustiveDependencies: runs once per location change.
  useLayoutEffect(() => {
    if (firstRender.current) {
      firstRender.current = false;
      return;
    }
    const heading = document.querySelector<HTMLElement>('main h1');
    if (heading) {
      heading.tabIndex = -1;
      heading.focus();
    }
  }, [location.key]);

  return (
    <AppShell
      labels={copy.shell}
      brand={copy.brand}
      actions={<ThemeToggle legend={copy.theme.legend} labels={copy.theme.labels} />}
      navigation={{
        label: copy.shell.navigation,
        items: items.map((item) => ({ ...item, current: item.href === location.pathname })),
        onNavigate: (event, href) => {
          if (shouldRoute(event, href)) {
            event.preventDefault();
            navigate(href);
          }
        },
      }}
    >
      {navigation.state === 'loading' ? <LoadingState title={copy.loading.title} /> : <Outlet />}
    </AppShell>
  );
}
