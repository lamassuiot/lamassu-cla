import { type SyntheticEvent, useLayoutEffect, useRef } from 'react';

export interface ModalDialogOptions {
  open: boolean;
  onClose: () => void;
  // The accessible name; an empty title fails in development (F-05 error cases).
  title: string;
}

// Shared by the dialog, confirmation dialog, and drawer. A native modal `dialog` makes the rest of
// the page inert, keeps focus out of it, and focuses its first focusable element (AC-05-5); this
// hook only syncs it with the `open` prop.
export function useModalDialog({ open, onClose, title }: ModalDialogOptions) {
  if (import.meta.env.DEV && !title.trim()) {
    throw new Error('A modal dialog needs a non-empty title as its accessible name.');
  }
  const ref = useRef<HTMLDialogElement>(null);
  const opener = useRef<HTMLElement | null>(null);
  const isOpen = useRef(open);

  useLayoutEffect(() => {
    isOpen.current = open;
    const dialog = ref.current;
    if (!dialog) {
      return;
    }
    if (open && !dialog.open) {
      opener.current =
        document.activeElement instanceof HTMLElement ? document.activeElement : null;
      dialog.showModal();
    } else if (!open && dialog.open) {
      dialog.close();
      // Browsers that do not restore focus on close natively still return it to the opener.
      opener.current?.focus();
    }
  }, [open]);

  // Unmounting an open dialog would otherwise drop focus to the document. The dialog is still
  // modal here, so it is closed first; otherwise the opener is inert and cannot take focus.
  useLayoutEffect(() => {
    const dialog = ref.current;
    return () => {
      if (isOpen.current) {
        isOpen.current = false;
        dialog?.close();
        opener.current?.focus();
      }
    };
  }, []);

  return {
    ref,
    // Escape: the caller decides, through `open`, whether the dialog closes.
    onCancel: (event: SyntheticEvent<HTMLDialogElement>) => {
      event.preventDefault();
      onClose();
    },
    // The browser can still close the dialog itself, for example after a repeated Escape.
    onClose: () => {
      if (isOpen.current) {
        onClose();
      }
    },
  };
}
