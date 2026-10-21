import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { StepIndicator } from './step-indicator';

afterEach(cleanup);

test('the current step is marked and completed steps say so in text', () => {
  const { getByRole, getAllByRole } = render(
    <StepIndicator
      label="Signing progress"
      steps={['Review', 'Sign', 'Verify']}
      current={1}
      completedLabel="completed"
    />,
  );
  const items = getAllByRole('listitem');
  expect(getByRole('list', { name: 'Signing progress' })).toBeTruthy();
  expect(items.map((item) => item.getAttribute('aria-current'))).toEqual([null, 'step', null]);
  expect(items.map((item) => item.textContent)).toEqual(['1Review completed', '2Sign', '3Verify']);
});
