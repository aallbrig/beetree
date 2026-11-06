import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { historySegments, publishBlockReason } from './remix';
import type { Lineage } from './types';

const base: Lineage = {
  kind: 'tree',
  id: 't1',
  title: 'Patrol',
  ownerId: 'u1',
  ownerName: 'alice',
  at: '2026-01-01T00:00:00Z',
};

const line = (l: Lineage): string =>
  historySegments(l)
    .map((s) => s.text)
    .join('');
const links = (l: Lineage): (string | undefined)[] =>
  historySegments(l)
    .filter((s) => s.href)
    .map((s) => s.href);

describe('remix history line', () => {
  it('public original links the title and the owner', () => {
    const l = { ...base, state: 'public' as const };
    expect(line(l)).toBe('Remixed from Patrol by alice');
    expect(links(l)).toEqual(['#/tree/t1', '#/user/u1']);
  });
  it('private original keeps the title as text and links only the profile', () => {
    const l = { ...base, state: 'private' as const };
    expect(line(l)).toBe('Remixed from Patrol, now private, by alice');
    expect(links(l)).toEqual(['#/user/u1']);
  });
  it('deleted original shows no title', () => {
    const l = { ...base, state: 'deleted' as const };
    expect(line(l)).toBe('Remixed from a deleted tree by alice');
    expect(line(l)).not.toContain('Patrol');
    expect(links(l)).toEqual(['#/user/u1']);
  });
});

describe('publish gate', () => {
  it('explains the block only for an identical remix', () => {
    expect(publishBlockReason(true)).toContain('identical to the original');
    expect(publishBlockReason(false)).toBeUndefined();
  });
  it('the builder disables the visibility choice and the tree page offers Remix', () => {
    expect(readFileSync('src/views/builder.ts', 'utf8')).toContain('publishBlockReason');
    const tree = readFileSync('src/views/tree.ts', 'utf8');
    expect(tree).toContain('api.remixTree');
    expect(tree).toContain('historyLine');
  });
});
