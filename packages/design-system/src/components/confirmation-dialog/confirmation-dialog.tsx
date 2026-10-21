import { Button } from '../button/button';
import { Dialog } from '../dialog/dialog';

export interface ConfirmationDialogProps {
  open: boolean;
  title: string;
  // States what confirming will do.
  description: string;
  confirmLabel: string;
  cancelLabel: string;
  onConfirm: () => void;
  // Called by the cancel button and by Escape; neither confirms.
  onCancel: () => void;
  destructive?: boolean;
}

export function ConfirmationDialog({
  open,
  title,
  description,
  confirmLabel,
  cancelLabel,
  onConfirm,
  onCancel,
  destructive = false,
}: ConfirmationDialogProps) {
  return (
    <Dialog
      open={open}
      onClose={onCancel}
      title={title}
      description={description}
      actions={
        <>
          {/* Cancel comes first and takes the initial focus, so Enter right away never confirms. */}
          <Button variant="secondary" onClick={onCancel}>
            {cancelLabel}
          </Button>
          <Button variant={destructive ? 'danger' : 'primary'} onClick={onConfirm}>
            {confirmLabel}
          </Button>
        </>
      }
    />
  );
}
