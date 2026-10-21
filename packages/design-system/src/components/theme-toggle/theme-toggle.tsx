import { useState } from 'react';
import { getThemePreference, setThemePreference, type ThemePreference } from '../../theme';
import { RadioGroup } from '../radio/radio';

export interface ThemeToggleProps {
  legend: string;
  labels: Readonly<Record<ThemePreference, string>>;
}

const order: readonly ThemePreference[] = ['light', 'dark', 'system'];

// Uses the F-04 theme module; it never reads or writes storage itself.
export function ThemeToggle({ legend, labels }: ThemeToggleProps) {
  const [preference, setPreference] = useState<ThemePreference>(() => getThemePreference());
  return (
    <RadioGroup
      legend={legend}
      name="cla-theme"
      options={order.map((value) => ({ value, label: labels[value] }))}
      value={preference}
      onChange={(next) => {
        setPreference(next);
        setThemePreference(next);
      }}
    />
  );
}
