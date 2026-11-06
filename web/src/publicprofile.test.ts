import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

describe('public profiles', () => {
  const main = readFileSync('src/main.ts', 'utf8');
  it('the user page is readable signed out', () => {
    const line = main.split('\n').find((l) => l.includes("route('/user/"));
    expect(line).toBeDefined();
    expect(line).not.toContain('private');
  });
  it('tree cards and tree pages link the owner to the profile', () => {
    expect(readFileSync('src/components/cards.ts', 'utf8')).toContain('#/user/${t.ownerId}');
    expect(readFileSync('src/views/tree.ts', 'utf8')).toContain('#/user/${tree.ownerId}');
  });
});
