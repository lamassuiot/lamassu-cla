import { createContext, type ReactNode, useCallback, useContext, useMemo, useState } from 'react';
import styles from './live-region.module.css';

export type Politeness = 'polite' | 'assertive';
export type Announce = (message: string, politeness?: Politeness) => void;

const AnnounceContext = createContext<Announce | null>(null);

interface Messages {
  polite: string;
  assertive: string;
  // Alternates so that repeating the same message is announced again.
  toggle: boolean;
}

// Mount once near the root. The regions exist before any message, so screen readers track them.
export function LiveRegionProvider({ children }: { children: ReactNode }) {
  const [messages, setMessages] = useState<Messages>({ polite: '', assertive: '', toggle: false });

  const announce = useCallback<Announce>((message, politeness = 'polite') => {
    setMessages((previous) => ({
      ...previous,
      [politeness]: previous.toggle ? `${message}\u00a0` : message,
      toggle: !previous.toggle,
    }));
  }, []);

  const value = useMemo(() => announce, [announce]);

  return (
    <AnnounceContext.Provider value={value}>
      {children}
      <div className={styles.region} role="status" aria-live="polite" aria-atomic="true">
        {messages.polite}
      </div>
      <div className={styles.region} role="alert" aria-live="assertive" aria-atomic="true">
        {messages.assertive}
      </div>
    </AnnounceContext.Provider>
  );
}

export function useAnnounce(): Announce {
  const announce = useContext(AnnounceContext);
  if (!announce) {
    throw new Error('useAnnounce must be used inside a LiveRegionProvider.');
  }
  return announce;
}
