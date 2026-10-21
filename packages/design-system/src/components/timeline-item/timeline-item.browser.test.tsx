import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { Timeline, TimelineItem } from './timeline-item';

afterEach(cleanup);

test('timeline items are list items with a machine-readable time', () => {
  const { getByRole, getAllByRole } = render(
    <Timeline label="History">
      <TimelineItem title="Activated" dateTime="2026-01-15T10:02:00Z" timeLabel="15 Jan" />
    </Timeline>,
  );
  expect(getByRole('list', { name: 'History' })).toBeTruthy();
  expect(getAllByRole('listitem')).toHaveLength(1);
  expect(getByRole('time').getAttribute('datetime')).toBe('2026-01-15T10:02:00Z');
});
