import { PageHeader } from '@lamassu-cla/design-system';
import { copy } from '../copy/en';
import { usePageTitle } from '../shell/use-page-title';

export function LandingScreen() {
  usePageTitle(copy.landing.title);
  return <PageHeader title={copy.landing.title} />;
}
