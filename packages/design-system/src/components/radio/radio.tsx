import { useId } from 'react';
import styles from '../checkbox/choice.module.css';

export interface RadioOption<T extends string> {
  value: T;
  label: string;
  hint?: string;
  disabled?: boolean;
}

export interface RadioGroupProps<T extends string> {
  legend: string;
  name: string;
  options: readonly RadioOption<T>[];
  value: T | undefined;
  onChange: (value: T) => void;
  disabled?: boolean;
}

// Native radios in a fieldset: arrow keys move the selection, Tab enters and leaves the group.
export function RadioGroup<T extends string>({
  legend,
  name,
  options,
  value,
  onChange,
  disabled,
}: RadioGroupProps<T>) {
  const id = useId();
  return (
    <fieldset className={styles.group} disabled={disabled}>
      <legend className={styles.legend}>{legend}</legend>
      {options.map((option) => {
        const optionId = `${id}-${option.value}`;
        const hintId = option.hint ? `${optionId}-hint` : undefined;
        return (
          <div key={option.value} className={styles.row}>
            <input
              id={optionId}
              type="radio"
              name={name}
              value={option.value}
              checked={value === option.value}
              disabled={option.disabled}
              onChange={() => onChange(option.value)}
              className={styles.input}
              aria-describedby={hintId}
            />
            <div>
              <label className={styles.text} htmlFor={optionId}>
                {option.label}
              </label>
              {option.hint && (
                <p className={styles.hint} id={hintId}>
                  {option.hint}
                </p>
              )}
            </div>
          </div>
        );
      })}
    </fieldset>
  );
}
