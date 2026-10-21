import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import { Button } from '../button/button';
import { useAnnounce } from '../live-region/live-region';
import styles from './toast.module.css';

export type ToastStatus = 'success' | 'info' | 'warning' | 'danger';

export interface ToastInput {
  // The message itself states the outcome, so status is not conveyed by color alone.
  message: string;
  status?: ToastStatus;
}

interface Toast extends Required<ToastInput> {
  id: number;
}

export interface ToastLabels {
  region: string;
  dismiss: string;
}

const ToastContext = createContext<((toast: ToastInput) => void) | null>(null);

// Toasts stay until dismissed (WCAG 2.2.1) and never move focus; each one is announced.
// Requires a surrounding LiveRegionProvider.
export function ToastProvider({ labels, children }: { labels: ToastLabels; children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const nextId = useRef(0);
  const regionRef = useRef<HTMLElement>(null);
  const focusAfterDismiss = useRef<HTMLElement | null>(null);
  const idPrefix = useId();
  const announce = useAnnounce();

  const show = useCallback(
    ({ message, status = 'info' }: ToastInput) => {
      nextId.current += 1;
      const id = nextId.current;
      setToasts((current) => [...current, { id, message, status }]);
      announce(message, status === 'danger' ? 'assertive' : 'polite');
    },
    [announce],
  );

  const dismiss = (id: number) => {
    const region = regionRef.current;
    const index = toasts.findIndex((toast) => toast.id === id);
    const items = region ? [...region.querySelectorAll('li')] : [];
    if (items[index]?.contains(document.activeElement)) {
      // Keep keyboard users in the toast region instead of dropping focus to the document.
      const buttons = items.map((item) => item.querySelector('button'));
      focusAfterDismiss.current = buttons[index + 1] ?? buttons[index - 1] ?? region;
    }
    setToasts((current) => current.filter((toast) => toast.id !== id));
  };

  useLayoutEffect(() => {
    focusAfterDismiss.current?.focus();
    focusAfterDismiss.current = null;
  });

  const value = useMemo(() => show, [show]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <section ref={regionRef} className={styles.region} aria-label={labels.region} tabIndex={-1}>
        <ol className={styles.list}>
          {toasts.map((toast) => {
            const messageId = `${idPrefix}-${toast.id}`;
            return (
              <li key={toast.id} className={`${styles.toast} ${styles[toast.status]}`}>
                <p className={styles.message} id={messageId}>
                  {toast.message}
                </p>
                <Button
                  variant="ghost"
                  aria-describedby={messageId}
                  onClick={() => dismiss(toast.id)}
                >
                  {labels.dismiss}
                </Button>
              </li>
            );
          })}
        </ol>
      </section>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const show = useContext(ToastContext);
  if (!show) {
    throw new Error('useToast must be used inside a ToastProvider.');
  }
  return show;
}
