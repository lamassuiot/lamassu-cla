import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Table, type TableColumn } from './table';

afterEach(cleanup);

interface Row {
  id: string;
  name: string;
  count: number;
}

const columns: TableColumn<Row>[] = [
  { key: 'name', header: 'Name', cell: (row) => row.name, rowHeader: true },
  { key: 'count', header: 'Count', cell: (row) => row.count, numeric: true },
];
const rows: Row[] = [
  { id: '1', name: 'Alpha Example', count: 3 },
  { id: '2', name: 'Beta Example', count: 12 },
];

test('the caption names the table and its container; headers have scopes', () => {
  const { getByRole, getAllByRole } = render(
    <Table caption="Members" columns={columns} rows={rows} rowKey={(row) => row.id} />,
  );
  expect(getByRole('table', { name: 'Members' })).toBeTruthy();
  expect(getByRole('region', { name: 'Members' })).toBeTruthy();
  expect(getAllByRole('columnheader').map((cell) => cell.getAttribute('scope'))).toEqual([
    'col',
    'col',
  ]);
  expect(getAllByRole('rowheader').map((cell) => cell.textContent)).toEqual([
    'Alpha Example',
    'Beta Example',
  ]);
});

test('the container is a tab stop only while the table overflows it', async () => {
  const wide: TableColumn<Row>[] = Array.from({ length: 12 }, (_, index) => ({
    key: `c${index}`,
    header: `Column header ${index}`,
    cell: (row: Row) => row.name,
  }));
  const { getByRole } = render(
    <Table caption="Wide" columns={wide} rows={rows} rowKey={(row) => row.id} />,
  );
  const region = getByRole('region', { name: 'Wide' });
  const host = region.parentElement as HTMLElement;

  host.style.width = '20rem';
  await expect.poll(() => region.tabIndex).toBe(0);
  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(region);
  await userEvent.keyboard('{ArrowRight}');
  await expect.poll(() => region.scrollLeft).toBeGreaterThan(0);

  host.style.width = '400rem';
  await expect.poll(() => region.hasAttribute('tabindex')).toBe(false);
});

test('the container becomes a tab stop when new content overflows it', async () => {
  const props = { caption: 'Growing', rowKey: (row: Row) => row.id };
  const { getByRole, rerender } = render(<Table {...props} columns={columns} rows={rows} />);
  const region = getByRole('region', { name: 'Growing' });
  (region.parentElement as HTMLElement).style.width = '30rem';
  await expect.poll(() => region.hasAttribute('tabindex')).toBe(false);

  const wide: TableColumn<Row>[] = Array.from({ length: 12 }, (_, index) => ({
    key: `c${index}`,
    header: `Column header ${index}`,
    cell: (row: Row) => row.name,
  }));
  rerender(<Table {...props} columns={wide} rows={rows} />);
  await expect.poll(() => region.tabIndex).toBe(0);
});
