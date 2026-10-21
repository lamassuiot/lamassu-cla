import { cleanup, render } from '@testing-library/react';
import { afterEach, expect, test } from 'vitest';
import { userEvent } from 'vitest/browser';
import { Input } from '../input/input';
import { Select } from '../select/select';
import { Textarea } from '../textarea/textarea';
import { FormField } from './form-field';

afterEach(cleanup);

test('labels the control and describes it with the hint and error', () => {
  const { getByRole } = render(
    <FormField label="Display name" hint="Shown publicly." error="Enter a display name.">
      {(control) => <Input {...control} />}
    </FormField>,
  );
  const input = getByRole('textbox', { name: 'Display name' });
  expect(input.getAttribute('aria-invalid')).toBe('true');
  const described = (input.getAttribute('aria-describedby') ?? '')
    .split(' ')
    .map((id) => document.getElementById(id)?.textContent);
  expect(described).toEqual(['Shown publicly.', 'Enter a display name.']);
});

test('has no description or invalid state without a hint or error', () => {
  const { getByRole } = render(
    <FormField label="Display name">{(control) => <Input {...control} />}</FormField>,
  );
  const input = getByRole('textbox', { name: 'Display name' });
  expect(input.hasAttribute('aria-describedby')).toBe(false);
  expect(input.hasAttribute('aria-invalid')).toBe(false);
});

test('input, select, and textarea are reachable with Tab and accept keyboard input', async () => {
  const { getByRole } = render(
    <>
      <FormField label="Name">{(control) => <Input {...control} />}</FormField>
      <FormField label="Type">
        {(control) => (
          <Select
            {...control}
            defaultValue="a"
            options={[
              { value: 'a', label: 'First' },
              { value: 'b', label: 'Second' },
            ]}
          />
        )}
      </FormField>
      <FormField label="Reason">{(control) => <Textarea {...control} />}</FormField>
    </>,
  );
  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('textbox', { name: 'Name' }));
  await userEvent.keyboard('Ada');
  expect(getByRole('textbox', { name: 'Name' })).toHaveProperty('value', 'Ada');

  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('combobox', { name: 'Type' }));

  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('textbox', { name: 'Reason' }));
  await userEvent.keyboard('Left the organization');
  expect(getByRole('textbox', { name: 'Reason' })).toHaveProperty('value', 'Left the organization');
});

test('a disabled control is skipped by Tab', async () => {
  const { getByRole } = render(
    <>
      <FormField label="Locked">{(control) => <Input {...control} disabled />}</FormField>
      <FormField label="Open">{(control) => <Input {...control} />}</FormField>
    </>,
  );
  await userEvent.keyboard('{Tab}');
  expect(document.activeElement).toBe(getByRole('textbox', { name: 'Open' }));
});
