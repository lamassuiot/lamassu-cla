import { type ReactNode, useId } from 'react';
import { Button } from '../button/button';
import { useModalDialog } from '../dialog/use-modal-dialog';
import styles from './drawer.module.css';

export interface DrawerProps {
  open: boolean;
  // Called on Escape and by the close button; set `open` to false to close.
  onClose: () => void;
  title: string;
  closeLabel: string;
  side?: 'start' | 'end';
  children?: ReactNode;
}

export function Drawer({ open, onClose, title, closeLabel, side = 'end', children }: DrawerProps) {
  const modal = useModalDialog({ open, onClose, title });
  const titleId = useId();
  return (
    <dialog {...modal} className={`${styles.drawer} ${styles[side]}`} aria-labelledby={titleId}>
      <div className={styles.header}>
        <h2 id={titleId} className={styles.title}>
          {title}
        </h2>
        <Button variant="ghost" onClick={onClose}>
          {closeLabel}
        </Button>
      </div>
      {children}
    </dialog>
  );
}
