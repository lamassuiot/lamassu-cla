import { describe, expect, it } from 'vitest';
import { applyTheme, getThemePreference, setThemePreference, themeStorageKey } from './theme.ts';

function memoryStorage(initial?: string) {
  const values = new Map<string, string>();
  if (initial !== undefined) values.set(themeStorageKey, initial);
  return {
    values,
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => void values.set(key, value),
    removeItem: (key: string) => void values.delete(key),
  };
}

function throwingStorage() {
  const fail = () => {
    throw new DOMException('blocked', 'SecurityError');
  };
  return { getItem: fail, setItem: fail, removeItem: fail };
}

function fakeRoot(theme?: string) {
  const attributes = new Map<string, string>();
  if (theme !== undefined) attributes.set('data-theme', theme);
  return {
    attributes,
    setAttribute: (name: string, value: string) => void attributes.set(name, value),
    removeAttribute: (name: string) => void attributes.delete(name),
  };
}

describe('theme preference (AC-04-4, AC-04-5)', () => {
  it('uses the system theme when nothing is stored', () => {
    const root = fakeRoot('dark');
    applyTheme(getThemePreference(memoryStorage()), root);
    expect(getThemePreference(memoryStorage())).toBe('system');
    expect(root.attributes.has('data-theme')).toBe(false);
  });

  it.each(['light', 'dark'] as const)('applies a stored %s preference', (theme) => {
    const root = fakeRoot();
    applyTheme(getThemePreference(memoryStorage(theme)), root);
    expect(root.attributes.get('data-theme')).toBe(theme);
  });

  it('stores light and dark, and removes the stored value and attribute for system', () => {
    const storage = memoryStorage();
    const root = fakeRoot();
    setThemePreference('dark', storage, root);
    expect(storage.values.get(themeStorageKey)).toBe('dark');
    expect(root.attributes.get('data-theme')).toBe('dark');

    setThemePreference('system', storage, root);
    expect(storage.values.has(themeStorageKey)).toBe(false);
    expect(root.attributes.has('data-theme')).toBe(false);
  });

  it.each(['system', 'DARK', 'sepia', '"><script>', ''])(
    'treats the stored value %j as system',
    (value) => {
      const root = fakeRoot('light');
      expect(getThemePreference(memoryStorage(value))).toBe('system');
      applyTheme(getThemePreference(memoryStorage(value)), root);
      expect(root.attributes.has('data-theme')).toBe(false);
    },
  );

  it('falls back to system without throwing when storage throws', () => {
    const root = fakeRoot('dark');
    expect(() => getThemePreference(throwingStorage())).not.toThrow();
    expect(getThemePreference(throwingStorage())).toBe('system');
    expect(() => setThemePreference('light', throwingStorage(), root)).not.toThrow();
    expect(root.attributes.get('data-theme')).toBe('light');
  });

  it('falls back to system when storage is unavailable', () => {
    expect(getThemePreference(undefined)).toBe('system');
    const root = fakeRoot();
    setThemePreference('dark', undefined, root);
    expect(root.attributes.get('data-theme')).toBe('dark');
  });

  it('ignores an unexpected preference passed at runtime', () => {
    const storage = memoryStorage('dark');
    const root = fakeRoot('dark');
    setThemePreference('sepia' as never, storage, root);
    expect(storage.values.has(themeStorageKey)).toBe(false);
    expect(root.attributes.has('data-theme')).toBe(false);
  });
});
