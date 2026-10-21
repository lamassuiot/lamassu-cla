import styles from './navigation.module.css';

export interface NavigationItem {
  // A relative path or an `https:` URL; anything else is rejected (F-05 security considerations).
  href: string;
  label: string;
  current?: boolean;
}

export interface NavigationProps {
  // Names the landmark, for example "Main".
  label: string;
  items: NavigationItem[];
  // Called when a link is followed, for example to close the drawer that contains the navigation.
  onNavigate?: () => void;
}

export function isSafeHref(href: string): boolean {
  if (href.startsWith('/')) {
    return !href.startsWith('//') && !href.startsWith('/\\');
  }
  try {
    return new URL(href).protocol === 'https:';
  } catch {
    return false;
  }
}

export function Navigation({ label, items, onNavigate }: NavigationProps) {
  for (const item of items) {
    if (!isSafeHref(item.href)) {
      throw new Error(`Navigation link "${item.label}" must be a relative path or an https: URL.`);
    }
  }
  return (
    <nav className={styles.navigation} aria-label={label}>
      <ul className={styles.list}>
        {items.map((item) => (
          <li key={item.href}>
            <a
              className={styles.link}
              href={item.href}
              aria-current={item.current ? 'page' : undefined}
              onClick={onNavigate}
            >
              {item.label}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  );
}
