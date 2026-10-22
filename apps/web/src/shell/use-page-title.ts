import { useLayoutEffect } from 'react';
import { documentTitle } from '../copy/en';

// Each screen sets its own title, so the title also updates on the error and not-found screens.
export function usePageTitle(pageTitle?: string) {
  useLayoutEffect(() => {
    document.title = documentTitle(pageTitle);
  }, [pageTitle]);
}
