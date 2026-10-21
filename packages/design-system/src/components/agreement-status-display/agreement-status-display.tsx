import { type ReactNode, useId } from 'react';
import { StatusBadge } from '../status-badge/status-badge';
import styles from './agreement-status-display.module.css';

export interface DetailItem {
  term: string;
  description: ReactNode;
}

export interface AgreementStatusDisplayProps {
  title: string;
  // Domain agreement state, for example `ACTIVE`; it selects the badge color.
  state: string;
  stateLabel: string;
  // Shown exactly as supplied, from the recorded level; this component never adds or upgrades a
  // signature-level claim such as "qualified" (ADR-0004, AC-05-8).
  signatureLevel: { term: string; description: string };
  details?: DetailItem[];
  headingLevel?: 2 | 3 | 4;
}

export function AgreementStatusDisplay({
  title,
  state,
  stateLabel,
  signatureLevel,
  details = [],
  headingLevel = 2,
}: AgreementStatusDisplayProps) {
  const Heading = `h${headingLevel}` as const;
  const titleId = useId();
  return (
    <section className={styles.agreement} aria-labelledby={titleId}>
      <div className={styles.header}>
        <Heading id={titleId} className={styles.title}>
          {title}
        </Heading>
        <StatusBadge state={state} label={stateLabel} />
      </div>
      <dl className={styles.details}>
        <div className={styles.item}>
          <dt>{signatureLevel.term}</dt>
          <dd>{signatureLevel.description}</dd>
        </div>
        {details.map((detail) => (
          <div key={detail.term} className={styles.item}>
            <dt>{detail.term}</dt>
            <dd>{detail.description}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
