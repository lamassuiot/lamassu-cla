import type { ReactNode } from 'react';
import styles from './card.module.css';

export interface CardProps {
  children: ReactNode;
  // `section` and `article` should be named, for example with `aria-labelledby` on their heading.
  as?: 'div' | 'section' | 'article';
  'aria-labelledby'?: string;
}

export function Card({ children, as: Element = 'div', ...aria }: CardProps) {
  return (
    <Element className={styles.card} {...aria}>
      {children}
    </Element>
  );
}
