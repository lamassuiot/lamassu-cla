import type { MouseEvent } from 'react';

// Plain left clicks on same-origin paths are routed in the app; modified clicks keep the default.
export function shouldRoute(event: MouseEvent<HTMLAnchorElement>, href: string): boolean {
  return (
    !event.defaultPrevented &&
    event.button === 0 &&
    !(event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) &&
    href.startsWith('/') &&
    !href.startsWith('//')
  );
}
