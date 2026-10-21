import { type ReactNode, useId } from 'react';
import styles from './form-field.module.css';

export interface FieldControlProps {
  id: string;
  required: boolean | undefined;
  'aria-describedby': string | undefined;
  'aria-invalid': true | undefined;
}

export interface FormFieldProps {
  label: string;
  hint?: string;
  // Present when the value is invalid; shown below the control and announced with it.
  error?: string;
  required?: boolean;
  children: (control: FieldControlProps) => ReactNode;
}

// Wires the label, hint, and error to the control; required state is part of the caller's label.
export function FormField({ label, hint, error, required, children }: FormFieldProps) {
  const id = useId();
  const hintId = hint ? `${id}-hint` : undefined;
  const errorId = error ? `${id}-error` : undefined;
  const describedBy = [hintId, errorId].filter(Boolean).join(' ') || undefined;
  return (
    <div className={styles.field}>
      <label className={styles.label} htmlFor={id}>
        {label}
      </label>
      {hint && (
        <p className={styles.hint} id={hintId}>
          {hint}
        </p>
      )}
      {children({
        id,
        required,
        'aria-describedby': describedBy,
        'aria-invalid': error ? true : undefined,
      })}
      {error && (
        <p className={styles.error} id={errorId}>
          {error}
        </p>
      )}
    </div>
  );
}
