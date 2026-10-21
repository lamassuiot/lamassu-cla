import type { ReactNode } from 'react';
import styles from './section-label.module.css';

export interface SectionLabelProps {
  children: ReactNode;
}

export function SectionLabel({ children }: SectionLabelProps) {
  return <p className={styles.label}>{children}</p>;
}
