import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { Button } from '../button/button';
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState } from './state-panel';

afterEach(cleanup);

test('loading is a polite, busy status region with its title', () => {
  const { getByRole } = render(<LoadingState title="Loading agreements" />);
  const status = getByRole('status');
  expect(status.getAttribute('aria-busy')).toBe('true');
  expect(status.textContent).toContain('Loading agreements');
});

test.each([
  ['empty', EmptyState],
  ['error', ErrorState],
  ['permission denied', PermissionDeniedState],
] as const)('%s state is a named region with a heading, description, and action', (_, State) => {
  const { getByRole } = render(
    <State title="Title text" description="Description text" action={<Button>Recover</Button>} />,
  );
  const region = getByRole('region', { name: 'Title text' });
  expect(getByRole('heading', { level: 2, name: 'Title text' })).toBeTruthy();
  expect(region.textContent).toContain('Description text');
  expect(getByRole('button', { name: 'Recover' })).toBeTruthy();
});

test('the heading level follows the surrounding structure', () => {
  const { getByRole } = render(<ErrorState title="Nested" headingLevel={3} />);
  expect(getByRole('heading', { level: 3, name: 'Nested' })).toBeTruthy();
});

test('no description element is rendered when none is given', () => {
  const { container } = render(<EmptyState title="Only a title" />);
  expect(container.querySelectorAll('p')).toHaveLength(0);
});
