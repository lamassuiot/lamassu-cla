import type { SelectHTMLAttributes } from 'react';
import styles from '../input/control.module.css';

export interface SelectOption {
  value: string;
  label: string;
}

export interface SelectProps
  extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'className' | 'style' | 'children'> {
  options: readonly SelectOption[];
}

export function Select({ options, ...props }: SelectProps) {
  return (
    <select {...props} className={styles.control}>
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}
