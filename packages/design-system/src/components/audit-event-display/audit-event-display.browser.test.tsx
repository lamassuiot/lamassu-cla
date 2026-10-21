import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { AuditEventDisplay } from './audit-event-display';

afterEach(cleanup);

test('renders only the given fields, in order, with the event type and time', () => {
  const { getByRole, container } = render(
    <AuditEventDisplay
      title="Agreement activated"
      eventType="agreement.activated"
      dateTime="2026-01-15T10:02:00Z"
      timeLabel="15 Jan 2026"
      fields={[
        { term: 'Actor', description: 'System' },
        { term: 'Request', description: 'req-example-0001' },
      ]}
    />,
  );
  expect(getByRole('article', { name: 'Agreement activated' })).toBeTruthy();
  expect(getByRole('heading', { level: 3 }).textContent).toBe('Agreement activated');
  expect(container.querySelector('code')?.textContent).toBe('agreement.activated');
  expect(getByRole('time').getAttribute('datetime')).toBe('2026-01-15T10:02:00Z');
  expect([...container.querySelectorAll('dt')].map((term) => term.textContent)).toEqual([
    'Actor',
    'Request',
  ]);
  expect([...container.querySelectorAll('dd')].map((term) => term.textContent)).toEqual([
    'System',
    'req-example-0001',
  ]);
});
