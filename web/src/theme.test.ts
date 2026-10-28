import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { applyTheme, currentTheme, toggleTheme } from './theme';

describe('theme', () => {
  let store: Record<string, string>;
  let root: { dataset: Record<string, string> };

  beforeEach(() => {
    store = {};
    root = { dataset: {} };
    vi.stubGlobal('localStorage', {
      getItem: (k: string): string | null => store[k] ?? null,
      setItem: (k: string, v: string): void => {
        store[k] = v;
      },
    });
    vi.stubGlobal('document', { documentElement: root });
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('defaults to light', () => {
    expect(currentTheme()).toBe('light');
  });

  it('applies and remembers the choice', () => {
    applyTheme('dark');
    expect(root.dataset.theme).toBe('dark');
    expect(currentTheme()).toBe('dark');
  });

  it('toggles back and forth', () => {
    expect(toggleTheme()).toBe('dark');
    expect(toggleTheme()).toBe('light');
  });

  it('survives blocked storage', () => {
    vi.stubGlobal('localStorage', {
      getItem: (): never => {
        throw new Error('blocked');
      },
      setItem: (): never => {
        throw new Error('blocked');
      },
    });
    expect(currentTheme()).toBe('light');
    expect(() => {
      applyTheme('dark');
    }).not.toThrow();
    expect(root.dataset.theme).toBe('dark');
  });
});
