import { describe, expect, it } from 'vitest';
import { filterSortTrees, renameTree } from './treelist';
import type { Tree } from './types';

const mk = (name: string, likeCount: number, updatedAt: string, tags: string[] = []): Tree => ({
  id: name,
  ownerId: 'u',
  ownerName: 'u',
  name,
  description: 'd',
  visibility: 'private',
  tags,
  spec: 'spec: yes',
  likeCount,
  commentCount: 0,
  createdAt: updatedAt,
  updatedAt,
});
const guard = mk('Guard', 1, '2026-01-02', ['patrol']);
const trees = [guard, mk('Archer', 5, '2026-01-03'), mk('Bee', 5, '2026-01-01')];

describe('filterSortTrees', () => {
  it('searches names and tags case-insensitively', () => {
    expect(filterSortTrees(trees, 'ARCH', 'name').map((t) => t.name)).toEqual(['Archer']);
    expect(filterSortTrees(trees, 'patrol', 'name').map((t) => t.name)).toEqual(['Guard']);
  });
  it('sorts by recent, name and likes', () => {
    expect(filterSortTrees(trees, '', 'recent').map((t) => t.name)).toEqual([
      'Archer',
      'Guard',
      'Bee',
    ]);
    expect(filterSortTrees(trees, '', 'name').map((t) => t.name)).toEqual([
      'Archer',
      'Bee',
      'Guard',
    ]);
    expect(filterSortTrees(trees, '', 'likes').map((t) => t.name)).toEqual([
      'Archer',
      'Bee',
      'Guard',
    ]);
  });
});

describe('renameTree', () => {
  it('persists the new name and keeps spec and tags', async () => {
    const saved: Record<string, unknown>[] = [];
    const api = {
      getTree: () => Promise.resolve(guard),
      updateTree: (_id: string, t: Record<string, unknown>) => {
        saved.push(t);
        return Promise.resolve({ ...guard, name: String(t.name) });
      },
    };
    const out = await renameTree(api as never, 'Guard', '  Sentry ');
    expect(out.name).toBe('Sentry');
    expect(saved[0]).toMatchObject({ name: 'Sentry', spec: 'spec: yes', tags: ['patrol'] });
  });
  it('rejects an empty name without saving', async () => {
    const api = {
      getTree: () => Promise.reject(new Error('no')),
      updateTree: () => Promise.reject(new Error('no')),
    };
    await expect(renameTree(api, 'x', '  ')).rejects.toThrow('empty');
  });
});
