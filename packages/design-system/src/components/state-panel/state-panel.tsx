import type { ReactNode } from 'react';
import styles from './state-panel.module.css';

export interface StatePanelProps {
  title: string;
  description?: string;
  // A recovery action, for example a retry button or a link to request access.
  action?: ReactNode;
  // Matches the surrounding heading structure; 1 when the state is the whole page.
  headingLevel?: 1 | 2 | 3 | 4;
}

type Kind = 'loading' | 'empty' | 'error' | 'denied';

function Panel({
  kind,
  title,
  description,
  action,
  headingLevel = 2,
}: StatePanelProps & { kind: Kind }) {
  const Heading = `h${headingLevel}` as const;
  const content = (
    <>
      <Heading className={styles.title}>{title}</Heading>
      {description && <p className={styles.description}>{description}</p>}
      {action}
    </>
  );
  // The loading panel is a polite status region, so its arrival is announced without moving focus.
  return kind === 'loading' ? (
    <div className={`${styles.panel} ${styles.loading}`} role="status" aria-busy="true">
      {content}
    </div>
  ) : (
    <section className={`${styles.panel} ${styles[kind]}`} aria-label={title}>
      {content}
    </section>
  );
}

// FR-UI-02: every data-driven view handles loading, empty, error, and permission-denied states.
export function LoadingState(props: Omit<StatePanelProps, 'action'>) {
  return <Panel kind="loading" {...props} />;
}

export function EmptyState(props: StatePanelProps) {
  return <Panel kind="empty" {...props} />;
}

export function ErrorState(props: StatePanelProps) {
  return <Panel kind="error" {...props} />;
}

export function PermissionDeniedState(props: StatePanelProps) {
  return <Panel kind="denied" {...props} />;
}
