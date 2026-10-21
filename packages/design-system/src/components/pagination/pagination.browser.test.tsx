import { cleanup, render } from '@testing-library/react';
import { useState } from 'react';
import { afterEach, expect, test } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Pagination } from './pagination';

afterEach(cleanup);

function Harness({ pages }: { pages: number }) {
  const [page, setPage] = useState(1);
  return (
    <>
      <p>Page {page}</p>
      <Pagination
        label="Results pages"
        previousLabel="Previous"
        nextLabel="Next"
        hasPrevious={page > 1}
        hasNext={page < pages}
        onPrevious={() => setPage(page - 1)}
        onNext={() => setPage(page + 1)}
      />
    </>
  );
}

test('the navigation is labelled and pages with Enter and Space', async () => {
  const { getByRole, getByText } = render(<Harness pages={3} />);
  expect(getByRole('navigation', { name: 'Results pages' })).toBeTruthy();
  expect((getByRole('button', { name: 'Previous' }) as HTMLButtonElement).disabled).toBe(true);

  getByRole('button', { name: 'Next' }).focus();
  await userEvent.keyboard('{Enter}');
  await userEvent.keyboard(' ');
  expect(getByText('Page 3')).toBeTruthy();
});

test('focus moves to the other button when the pressed one becomes disabled', async () => {
  const { getByRole } = render(<Harness pages={2} />);
  const previous = getByRole('button', { name: 'Previous' });
  const next = getByRole('button', { name: 'Next' });

  next.focus();
  await userEvent.keyboard('{Enter}');
  expect((next as HTMLButtonElement).disabled).toBe(true);
  expect(document.activeElement).toBe(previous);

  await userEvent.keyboard('{Enter}');
  expect((previous as HTMLButtonElement).disabled).toBe(true);
  expect(document.activeElement).toBe(next);
});
