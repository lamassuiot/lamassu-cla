import type { ButtonHTMLAttributes } from 'react';
import styles from './button.module.css';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';

// `className` and `style` are not accepted, so pages cannot add one-off button styles.
export interface ButtonProps
  extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'className' | 'style'> {
  variant?: ButtonVariant;
}

export function Button({ variant = 'primary', type = 'button', ...props }: ButtonProps) {
  return <button {...props} type={type} className={`${styles.button} ${styles[variant]}`} />;
}
