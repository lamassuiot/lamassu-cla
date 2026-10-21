import { cleanup, render } from '@testing-library/react';
import { useState } from 'react';
import { afterEach, expect, test } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Checkbox } from '../checkbox/checkbox';
import { RadioGroup } from './radio';

afterEach(cleanup);

function Group() {
  const [value, setValue] = useState<'a' | 'b' | 'c' | undefined>('a');
  return (
    <>
      <RadioGroup
        legend="Choose"
        name="choice"
        value={value}
        onChange={setValue}
        options={[
          { value: 'a', label: 'Alpha' },
          { value: 'b', label: 'Bravo' },
          { value: 'c', label: 'Charlie', hint: 'Third option' },
        ]}
      />
      <Checkbox label="After the group" />
    </>
  );
}

test('radios: Tab enters at the selection, arrows move it, Tab leaves the group', async () => {
  const { getByRole } = render(<Group />);
  expect(getByRole('group', { name: 'Choose' })).toBeTruthy();

  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('radio', { name: 'Alpha' }));

  await userEvent.keyboard('{ArrowDown}');
  expect(getByRole('radio', { name: 'Bravo' })).toHaveProperty('checked', true);
  await userEvent.keyboard('{ArrowDown}');
  const charlie = getByRole('radio', { name: 'Charlie' });
  expect(charlie).toHaveProperty('checked', true);
  expect(document.getElementById(charlie.getAttribute('aria-describedby') ?? '')?.textContent).toBe(
    'Third option',
  );

  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('checkbox', { name: 'After the group' }));
});

test('checkbox: Space toggles it and the label is the click target', async () => {
  const { getByRole, getByText } = render(<Checkbox label="I agree" hint="Required to sign" />);
  const checkbox = getByRole('checkbox', { name: 'I agree' });
  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(checkbox);
  await userEvent.keyboard(' ');
  expect(checkbox).toHaveProperty('checked', true);
  await userEvent.click(getByText('I agree'));
  expect(checkbox).toHaveProperty('checked', false);
});

test('a disabled radio group is skipped by Tab', async () => {
  const { getByRole } = render(
    <>
      <RadioGroup
        legend="Locked"
        name="locked"
        disabled
        value="a"
        onChange={() => undefined}
        options={[{ value: 'a', label: 'Only' }]}
      />
      <Checkbox label="Next" />
    </>,
  );
  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('checkbox', { name: 'Next' }));
});
