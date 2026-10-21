import { useId } from 'react';
import type { DetailItem } from '../agreement-status-display/agreement-status-display';
import styles from './audit-event-display.module.css';

export interface AuditEventDisplayProps {
  // Readable name of the event, for example "Agreement activated".
  title: string;
  // Dotted event type, for example `agreement.activated`.
  eventType: string;
  // Machine-readable timestamp (ISO 8601) and its caller-formatted text (D-21).
  dateTime: string;
  timeLabel: string;
  // Only these fields are rendered; the component never fetches data. Fields must not contain
  // personal data (audit metadata rules in the domain model).
  fields: DetailItem[];
  headingLevel?: 2 | 3 | 4;
}

export function AuditEventDisplay({
  title,
  eventType,
  dateTime,
  timeLabel,
  fields,
  headingLevel = 3,
}: AuditEventDisplayProps) {
  const Heading = `h${headingLevel}` as const;
  const titleId = useId();
  return (
    <article className={styles.event} aria-labelledby={titleId}>
      <Heading id={titleId} className={styles.title}>
        {title}
      </Heading>
      <p className={styles.meta}>
        <code>{eventType}</code>
        <time dateTime={dateTime}>{timeLabel}</time>
      </p>
      <dl className={styles.fields}>
        {fields.map((field) => (
          <div key={field.term} className={styles.field}>
            <dt>{field.term}</dt>
            <dd>{field.description}</dd>
          </div>
        ))}
      </dl>
    </article>
  );
}
