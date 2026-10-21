import { type InputHTMLAttributes, useId } from 'react';
import styles from './choice.module.css';

export interface CheckboxProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, 'className' | 'style' | 'type' | 'id'> {
  label: string;
  hint?: string;
}

// The hint sits outside the label so that it describes the control without becoming its name.
export function Checkbox({ label, hint, ...props }: CheckboxProps) {
  const id = useId();
  const hintId = hint ? `${id}-hint` : undefined;
  return (
    <div className={styles.row}>
      <input
        {...props}
        id={id}
        type="checkbox"
        className={styles.input}
        aria-describedby={hintId}
      />
      <div>
        <label className={styles.text} htmlFor={id}>
          {label}
        </label>
        {hint && (
          <p className={styles.hint} id={hintId}>
            {hint}
          </p>
        )}
      </div>
    </div>
  );
}
