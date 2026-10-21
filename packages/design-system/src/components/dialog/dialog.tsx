import { type ReactNode, useId } from 'react';
import styles from './dialog.module.css';
import { useModalDialog } from './use-modal-dialog';

export interface DialogProps {
  open: boolean;
  // Called on Escape; set `open` to false to close.
  onClose: () => void;
  title: string;
  description?: string;
  children?: ReactNode;
  // Must include a way to close the dialog, so pointer and touch users are not trapped.
  actions: ReactNode;
}

export function Dialog({ open, onClose, title, description, children, actions }: DialogProps) {
  const modal = useModalDialog({ open, onClose, title });
  const titleId = useId();
  const descriptionId = useId();
  return (
    <dialog
      {...modal}
      className={styles.dialog}
      aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined}
    >
      <h2 id={titleId} className={styles.title}>
        {title}
      </h2>
      {description && (
        <p id={descriptionId} className={styles.description}>
          {description}
        </p>
      )}
      {children}
      <div className={styles.actions}>{actions}</div>
    </dialog>
  );
}
