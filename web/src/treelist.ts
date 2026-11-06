import type { TreeInput } from './api';
import type { Tree } from './types';

export type TreeSort = 'recent' | 'name' | 'likes';

export function filterSortTrees(trees: Tree[], query: string, sort: TreeSort): Tree[] {
  const q = query.trim().toLowerCase();
  const hit = (t: Tree): boolean =>
    q === '' ||
    t.name.toLowerCase().includes(q) ||
    t.tags.some((tag) => tag.toLowerCase().includes(q));
  const out = trees.filter(hit);
  const by: Record<TreeSort, (a: Tree, b: Tree) => number> = {
    recent: (a, b) => b.updatedAt.localeCompare(a.updatedAt),
    name: (a, b) => a.name.localeCompare(b.name),
    likes: (a, b) => b.likeCount - a.likeCount || a.name.localeCompare(b.name),
  };
  return out.sort(by[sort]);
}

export interface RenameApi {
  getTree: (id: string) => Promise<Tree>;
  updateTree: (id: string, t: TreeInput) => Promise<Tree>;
}

/** Rename keeps every other field; the list endpoint omits the spec, so fetch the full tree first. */
export async function renameTree(api: RenameApi, id: string, name: string): Promise<Tree> {
  const next = name.trim();
  if (next === '') throw new Error('Name cannot be empty.');
  const t = await api.getTree(id);
  return api.updateTree(id, {
    name: next,
    description: t.description,
    visibility: t.visibility,
    tags: t.tags,
    spec: t.spec ?? '',
  });
}
