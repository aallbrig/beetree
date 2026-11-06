import { describe, expect, it } from 'vitest';
import { editAction, type KeyLike } from './shortcuts';

const key = (k: string, mods: Partial<KeyLike> = {}): KeyLike => ({
  key: k,
  ctrlKey: false,
  metaKey: false,
  altKey: false,
  shiftKey: false,
  ...mods,
});

describe('editAction', () => {
  it('maps node-editing keys', () => {
    expect(editAction(key('Delete'))).toBe('remove');
    expect(editAction(key('d', { ctrlKey: true }))).toBe('duplicate');
    expect(editAction(key('D', { metaKey: true }))).toBe('duplicate');
    expect(editAction(key('ArrowUp', { altKey: true }))).toBe('up');
    expect(editAction(key('ArrowDown', { altKey: true }))).toBe('down');
    expect(editAction(key('ArrowLeft', { altKey: true }))).toBe('outdent');
    expect(editAction(key('ArrowRight', { altKey: true }))).toBe('indent');
  });

  it('ignores unmodified arrows, other letters and shifted combos', () => {
    expect(editAction(key('ArrowUp'))).toBeUndefined();
    expect(editAction(key('z', { ctrlKey: true }))).toBeUndefined();
    expect(editAction(key('Delete', { shiftKey: true }))).toBeUndefined();
    expect(editAction(key('d', { ctrlKey: true, altKey: true }))).toBeUndefined();
  });
});
