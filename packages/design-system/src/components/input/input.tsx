import type { InputHTMLAttributes } from 'react';
import styles from './control.module.css';

export type InputProps = Omit<
  InputHTMLAttributes<HTMLInputElement>,
  'className' | 'style' | 'type'
> & {
  type?: 'text' | 'email' | 'search' | 'url' | 'tel' | 'number' | 'password';
};

// Use inside FormField, which supplies the id, label, and descriptions.
export function Input({ type = 'text', ...props }: InputProps) {
  return <input {...props} type={type} className={styles.control} />;
}
