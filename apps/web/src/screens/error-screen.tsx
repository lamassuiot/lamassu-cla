import { Button, ErrorState } from '@lamassu-cla/design-system';
import { copy } from '../copy/en';
import { usePageTitle } from '../shell/use-page-title';

// AC-06-4: the error itself is never shown; the router keeps it for the browser console.
export function ErrorScreen() {
  usePageTitle(copy.error.title);
  return (
    <ErrorState
      headingLevel={1}
      title={copy.error.title}
      description={copy.error.description}
      action={<Button onClick={() => window.location.reload()}>{copy.error.reload}</Button>}
    />
  );
}
