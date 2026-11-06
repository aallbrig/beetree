import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { clearDraft, DRAFT_KEY } from './draft';

describe('draft', () => {
  let store: Record<string, string>;
  beforeEach(() => {
    store = { [DRAFT_KEY]: 'name: x', other: 'keep' };
    vi.stubGlobal('localStorage', {
      removeItem: (k: string): void => {
        Reflect.deleteProperty(store, k);
      },
    });
  });
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('clearDraft removes only the draft key', () => {
    clearDraft();
    expect(store[DRAFT_KEY]).toBeUndefined();
    expect(store['other']).toBe('keep');
  });
});
