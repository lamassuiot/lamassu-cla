import type { ReactNode } from 'react';
import type { Status } from '../status-badge/status';
import styles from './timeline-item.module.css';

export function Timeline({ label, children }: { label: string; children: ReactNode }) {
  return (
    <ol className={styles.timeline} aria-label={label}>
      {children}
    </ol>
  );
}

export interface TimelineItemProps {
  title: string;
  // Machine-readable timestamp (ISO 8601) and its formatted text; the caller formats it (D-21).
  dateTime: string;
  timeLabel: string;
  description?: ReactNode;
  // Colors the marker only; the title must state the outcome, so it is never conveyed by color alone.
  status?: Status;
}

export function TimelineItem({
  title,
  dateTime,
  timeLabel,
  description,
  status = 'neutral',
}: TimelineItemProps) {
  return (
    <li className={styles.item}>
      <span className={`${styles.marker} ${styles[status]}`} aria-hidden="true" />
      <div className={styles.content}>
        <p className={styles.title}>{title}</p>
        <time className={styles.time} dateTime={dateTime}>
          {timeLabel}
        </time>
        {description && <div className={styles.description}>{description}</div>}
      </div>
    </li>
  );
}
