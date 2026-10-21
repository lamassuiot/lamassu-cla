import { useLayoutEffect, useRef } from 'react';
import { Button } from '../button/button';
import styles from './pagination.module.css';

export interface PaginationProps {
  // Names the navigation landmark, for example "Agreements pages".
  label: string;
  previousLabel: string;
  nextLabel: string;
  // The API pages with an opaque `next_cursor`, so there are no page numbers; the caller keeps the
  // cursors of earlier pages.
  hasPrevious: boolean;
  hasNext: boolean;
  onPrevious: () => void;
  onNext: () => void;
}

export function Pagination({
  label,
  previousLabel,
  nextLabel,
  hasPrevious,
  hasNext,
  onPrevious,
  onNext,
}: PaginationProps) {
  const navRef = useRef<HTMLElement>(null);
  const pressed = useRef<'previous' | 'next' | null>(null);

  // A button that becomes disabled loses focus; move it to the other button instead of the document.
  useLayoutEffect(() => {
    const [previous, next] = navRef.current?.querySelectorAll('button') ?? [];
    if (pressed.current === 'next' && !hasNext) {
      previous?.focus();
    } else if (pressed.current === 'previous' && !hasPrevious) {
      next?.focus();
    }
    pressed.current = null;
  }, [hasPrevious, hasNext]);

  return (
    <nav ref={navRef} className={styles.pagination} aria-label={label}>
      <Button
        variant="secondary"
        disabled={!hasPrevious}
        onClick={() => {
          pressed.current = 'previous';
          onPrevious();
        }}
      >
        {previousLabel}
      </Button>
      <Button
        variant="secondary"
        disabled={!hasNext}
        onClick={() => {
          pressed.current = 'next';
          onNext();
        }}
      >
        {nextLabel}
      </Button>
    </nav>
  );
}
