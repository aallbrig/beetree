import { describe, expect, it } from 'vitest';
import { AVATAR_IDS, avatarSvg, avatarSrc, isAvatarId } from './avatar';

describe('avatars', () => {
  it('offers exactly 100 distinct pictures', () => {
    expect(AVATAR_IDS).toHaveLength(100);
    expect(AVATAR_IDS[0]).toBe('a001');
    expect(AVATAR_IDS[99]).toBe('a100');
    expect(new Set(AVATAR_IDS.map(avatarSvg)).size).toBe(100);
  });
  it('validates ids', () => {
    expect(isAvatarId('a050')).toBe(true);
    expect(isAvatarId('a101')).toBe(false);
    expect(isAvatarId('')).toBe(false);
  });
  it('produces an image data URI without script', () => {
    expect(avatarSrc('a017').startsWith('data:image/svg+xml')).toBe(true);
    expect(avatarSvg('a017')).not.toContain('<script');
  });
});
