import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { addSignup, isValidEmail, readSignups, STORAGE_KEY } from './communitree';

describe('communitree', () => {
  let store: Record<string, string>;
  beforeEach(() => {
    store = {};
    vi.stubGlobal('localStorage', {
      getItem: (k: string): string | null => store[k] ?? null,
      setItem: (k: string, v: string): void => {
        store[k] = v;
      },
    });
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('validates emails', () => {
    expect(isValidEmail(' a@b.co ')).toBe(true);
    expect(isValidEmail('nope')).toBe(false);
    expect(isValidEmail('a@b')).toBe(false);
  });

  it('stores emails as an array, lowercased, without duplicates', () => {
    expect(addSignup('Bee@Tree.com', 'signup', new Date('2026-01-02T03:04:05Z'))).toBe(true);
    expect(addSignup('bee@tree.com', 'homepage')).toBe(false);
    expect(JSON.parse(store[STORAGE_KEY] ?? '')).toEqual([
      { email: 'bee@tree.com', joinedAt: '2026-01-02T03:04:05.000Z', source: 'signup' },
    ]);
  });

  it('survives corrupt storage', () => {
    store[STORAGE_KEY] = '{not json';
    expect(readSignups()).toEqual([]);
    store[STORAGE_KEY] = '{"a":1}';
    expect(readSignups()).toEqual([]);
  });
});
