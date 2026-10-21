import { type Status, statusForState } from './status';
import styles from './status-badge.module.css';

// The label is always rendered, so status is never conveyed by color alone.
export type StatusBadgeProps = { label: string } & (
  | { state: string; status?: never }
  | { status: Status; state?: never }
);

export function StatusBadge({ label, state, status }: StatusBadgeProps) {
  const resolved = status ?? statusForState(state ?? '');
  return (
    <span className={`${styles.badge} ${styles[resolved]}`} data-status={resolved}>
      {label}
    </span>
  );
}
