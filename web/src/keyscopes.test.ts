import { describe, expect, it } from 'vitest';
import { PRESETS, expiryText, gridToScopes, keyBody, lastUsedText, scopeChips } from './keyscopes';

describe('presets', () => {
  it('match the server presets and never include keys', () => {
    expect(PRESETS['read-only'].scopes.every((s) => s.endsWith(':read'))).toBe(true);
    expect(PRESETS['read-write'].scopes).not.toContain('trees:delete');
    for (const p of Object.values(PRESETS)) {
      expect(p.scopes.some((s) => s.startsWith('keys:'))).toBe(false);
    }
    expect(PRESETS.full.scopes).toContain('trees:delete');
  });
});

describe('keyBody', () => {
  it('requires a name and a choice', () => {
    expect(keyBody('', 'full', [], '90d')).toBeNull();
    expect(keyBody('laptop', '', [], '90d')).toBeNull();
  });
  it('sends a preset by name with the expiry', () => {
    expect(keyBody(' laptop ', 'read-only', [], 'never')).toEqual({
      name: 'laptop',
      preset: 'read-only',
      expiry: 'never',
    });
  });
  it('rejects an empty custom grid', () => {
    expect(keyBody('ci', 'custom', [], '30d')).toBeNull();
  });
  it('sends custom scopes collapsed to the highest action', () => {
    expect(keyBody('ci', 'custom', ['trees:read', 'trees:write', 'forum:read'], '365d')).toEqual({
      name: 'ci',
      scopes: ['forum:read', 'trees:write'],
      expiry: '365d',
    });
  });
});

describe('gridToScopes', () => {
  it('ignores unknown scopes', () => {
    expect(gridToScopes(['bogus:read', 'trees:fly', 'profile:write'])).toEqual(['profile:write']);
  });
});

describe('scopeChips', () => {
  it('renders one chip per resource', () => {
    expect(scopeChips(['trees:read', 'trees:delete', 'profile:read'])).toEqual([
      'profile: read',
      'trees: delete',
    ]);
  });
});

describe('key list text', () => {
  const now = new Date('2026-11-06T12:00:00Z');
  it('describes expiry', () => {
    expect(expiryText(undefined, now)).toBe('never expires');
    expect(expiryText('2026-11-07T12:00:00Z', now)).toBe('expires in 1 day');
    expect(expiryText('2026-12-06T12:00:00Z', now)).toBe('expires in 30 days');
    expect(expiryText('2026-11-01T12:00:00Z', now)).toBe('expired');
  });
  it('describes last use', () => {
    expect(lastUsedText(undefined, now)).toBe('never used');
    expect(lastUsedText('2026-11-06T11:30:00Z', now)).toBe('used within the last hour');
    expect(lastUsedText('2026-11-06T07:00:00Z', now)).toBe('used 5 hours ago');
    expect(lastUsedText('2026-11-04T12:00:00Z', now)).toBe('used 2 days ago');
  });
});
