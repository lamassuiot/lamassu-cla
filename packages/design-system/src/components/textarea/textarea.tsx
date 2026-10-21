import type { TextareaHTMLAttributes } from 'react';
import styles from '../input/control.module.css';

export type TextareaProps = Omit<
  TextareaHTMLAttributes<HTMLTextAreaElement>,
  'className' | 'style'
>;

export function Textarea(props: TextareaProps) {
  return <textarea rows={4} {...props} className={`${styles.control} ${styles.textarea}`} />;
}
