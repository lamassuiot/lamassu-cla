import { EmptyState } from '@lamassu-cla/design-system';
import { Link } from 'react-router';
import { copy } from '../copy/en';
import { usePageTitle } from '../shell/use-page-title';

export function NotFoundScreen() {
  usePageTitle(copy.notFound.title);
  return (
    <EmptyState
      headingLevel={1}
      title={copy.notFound.title}
      description={copy.notFound.description}
      action={<Link to="/">{copy.notFound.home}</Link>}
    />
  );
}
