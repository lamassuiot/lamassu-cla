import { type ReactNode, useId, useState } from 'react';
import { Button } from '../button/button';
import { Drawer } from '../drawer/drawer';
import { Navigation, type NavigationProps } from '../navigation/navigation';
import styles from './app-shell.module.css';

export interface AppShellLabels {
  skipToContent: string;
  // Opens the navigation drawer below `bp.lg`.
  menu: string;
  menuTitle: string;
  closeMenu: string;
}

export interface AppShellProps {
  labels: AppShellLabels;
  // Product name or logo link; brand assets are pending D-12.
  brand: ReactNode;
  // Shown as a sidebar from `bp.lg`, and in a drawer below it.
  navigation: Omit<NavigationProps, 'onNavigate'>;
  // Header controls such as the theme toggle or sign-in.
  actions?: ReactNode;
  children: ReactNode;
}

export function AppShell({ labels, brand, navigation, actions, children }: AppShellProps) {
  const mainId = useId();
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <div className={styles.shell}>
      <a className={styles.skip} href={`#${mainId}`}>
        {labels.skipToContent}
      </a>
      <header className={styles.header}>
        <div className={styles.menu}>
          <Button variant="ghost" aria-haspopup="dialog" onClick={() => setMenuOpen(true)}>
            {labels.menu}
          </Button>
        </div>
        <div className={styles.brand}>{brand}</div>
        {actions && <div className={styles.actions}>{actions}</div>}
      </header>
      <div className={styles.body}>
        <div className={styles.sidebar}>
          <Navigation {...navigation} />
        </div>
        <main id={mainId} className={styles.main} tabIndex={-1}>
          {children}
        </main>
      </div>
      <Drawer
        open={menuOpen}
        onClose={() => setMenuOpen(false)}
        title={labels.menuTitle}
        closeLabel={labels.closeMenu}
        side="start"
      >
        {/* Following a link closes the drawer; the destination page takes over. */}
        <Navigation {...navigation} onNavigate={() => setMenuOpen(false)} />
      </Drawer>
    </div>
  );
}
