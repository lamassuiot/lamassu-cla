import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { AgreementStatusDisplay } from './agreement-status-display';

afterEach(cleanup);

const levels = ['Electronic signature', 'Advanced electronic signature', 'Level supplied by data'];

test.each(levels)('renders only the supplied signature level: %s', (description) => {
  const { getByRole, container } = render(
    <AgreementStatusDisplay
      title="Individual CLA"
      state="ACTIVE"
      stateLabel="Active"
      signatureLevel={{ term: 'Signature level', description }}
      details={[{ term: 'Version', description: '2.0' }]}
    />,
  );
  expect(getByRole('region', { name: 'Individual CLA' })).toBeTruthy();
  const level = [...container.querySelectorAll('dt')].find(
    (term) => term.textContent === 'Signature level',
  );
  expect(level?.nextElementSibling?.textContent).toBe(description);
  // ADR-0004: no qualified claim unless the supplied level says so.
  expect(container.textContent).not.toMatch(/qualif/i);
});

test('the state is shown as a labelled badge with its status color', () => {
  const { getByText } = render(
    <AgreementStatusDisplay
      title="Corporate CLA"
      state="REVOKED"
      stateLabel="Revoked"
      signatureLevel={{ term: 'Signature level', description: 'Electronic signature' }}
    />,
  );
  expect(getByText('Revoked').getAttribute('data-status')).toBe('danger');
});
