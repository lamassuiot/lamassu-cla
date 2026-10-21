import type { Theme } from './tokens/semantic.ts';

export type ThemePreference = Theme | 'system';

export const themeStorageKey = 'cla-theme';

type PreferenceStorage = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>;
type ThemeRoot = Pick<Element, 'setAttribute' | 'removeAttribute'>;

// Accessing localStorage itself throws when storage is blocked.
function defaultStorage(): PreferenceStorage | undefined {
  try {
    return globalThis.localStorage;
  } catch {
    return undefined;
  }
}

function isTheme(value: unknown): value is Theme {
  return value === 'light' || value === 'dark';
}

export function getThemePreference(storage = defaultStorage()): ThemePreference {
  try {
    const stored = storage?.getItem(themeStorageKey);
    return isTheme(stored) ? stored : 'system';
  } catch {
    return 'system';
  }
}

// `system` removes the attribute, so `prefers-color-scheme` in tokens.css selects the theme.
export function applyTheme(
  preference: ThemePreference = getThemePreference(),
  root: ThemeRoot = document.documentElement,
): void {
  if (isTheme(preference)) {
    root.setAttribute('data-theme', preference);
  } else {
    root.removeAttribute('data-theme');
  }
}

export function setThemePreference(
  preference: ThemePreference,
  storage = defaultStorage(),
  root: ThemeRoot = document.documentElement,
): void {
  const theme = isTheme(preference) ? preference : 'system';
  try {
    if (theme === 'system') {
      storage?.removeItem(themeStorageKey);
    } else {
      storage?.setItem(themeStorageKey, theme);
    }
  } catch {
    // The choice is not remembered, but it still applies to this page.
  }
  applyTheme(theme, root);
}
