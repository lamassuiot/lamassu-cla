import { useId } from 'react';
import { Button } from '../button/button';
import { useAnnounce } from '../live-region/live-region';
import styles from './hash-display.module.css';

export interface HashDisplayProps {
  // Names the value, for example "Document SHA-256".
  label: string;
  value: string;
  copyLabel: string;
  copiedMessage: string;
  // Should say that the value can be selected and copied manually.
  copyFailedMessage: string;
}

// Requires a surrounding LiveRegionProvider; copy results are announced through it (AC-05-7).
export function HashDisplay({
  label,
  value,
  copyLabel,
  copiedMessage,
  copyFailedMessage,
}: HashDisplayProps) {
  const labelId = useId();
  const announce = useAnnounce();

  const copy = async () => {
    try {
      // `navigator.clipboard` is missing outside secure contexts; that is a failure too.
      await navigator.clipboard.writeText(value);
      announce(copiedMessage);
    } catch {
      announce(copyFailedMessage, 'assertive');
    }
  };

  return (
    <figure className={styles.hash} aria-labelledby={labelId}>
      <figcaption id={labelId} className={styles.label}>
        {label}
      </figcaption>
      <code className={styles.value}>{value}</code>
      <Button variant="ghost" aria-describedby={labelId} onClick={copy}>
        {copyLabel}
      </Button>
    </figure>
  );
}
