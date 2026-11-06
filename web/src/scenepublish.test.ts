import { describe, expect, it, vi } from 'vitest';
import { ApiError, sharedTrees } from './api';
import { publishPrompt, saveSceneConfirmed } from './scenepublish';
import type { Scene } from './types';

const scene: Scene = {
  id: 's1',
  name: 'arena',
  visibility: 'public',
  agents: [],
  createdAt: '',
  updatedAt: '',
};
const body = { name: 'arena', visibility: 'public', agents: [] };
const needs = new ApiError(409, 'confirm', [{ id: 't1', name: 'Guard' }]);

describe('sharedTrees', () => {
  it('reads id and name pairs and ignores junk', () => {
    expect(sharedTrees({ trees: [{ id: 'a', name: 'A' }, 3, { id: 1 }] })).toEqual([
      { id: 'a', name: 'A' },
    ]);
    expect(sharedTrees({})).toEqual([]);
    expect(sharedTrees(null)).toEqual([]);
  });
});

describe('publishPrompt', () => {
  it('names every tree and counts them', () => {
    expect(publishPrompt([{ id: '1', name: 'Guard' }])).toContain('1 private tree public: "Guard"');
    const two = publishPrompt([
      { id: '1', name: 'A' },
      { id: '2', name: 'B' },
    ]);
    expect(two).toContain('2 private trees public: "A", "B"');
  });
});

describe('saveSceneConfirmed', () => {
  it('saves straight through when nothing needs confirming', async () => {
    const save = vi.fn().mockResolvedValue(scene);
    const ask = vi.fn();
    await saveSceneConfirmed(save, undefined, body, ask);
    expect(ask).not.toHaveBeenCalled();
    expect(save).toHaveBeenCalledTimes(1);
  });
  it('asks, then resends with confirm on accept', async () => {
    const save = vi.fn().mockRejectedValueOnce(needs).mockResolvedValueOnce(scene);
    const ask = vi.fn().mockReturnValue(true);
    await expect(saveSceneConfirmed(save, 's1', body, ask)).resolves.toBe(scene);
    expect(ask).toHaveBeenCalledWith(expect.stringContaining('"Guard"'));
    expect(save).toHaveBeenLastCalledWith('s1', { ...body, confirm: true });
  });
  it('does not resend when declined', async () => {
    const save = vi.fn().mockRejectedValue(needs);
    await expect(saveSceneConfirmed(save, 's1', body, () => false)).rejects.toThrow(/stay private/);
    expect(save).toHaveBeenCalledTimes(1);
  });
  it('passes other errors through', async () => {
    const boom = new ApiError(400, 'bad');
    await expect(
      saveSceneConfirmed(vi.fn().mockRejectedValue(boom), undefined, body, () => true),
    ).rejects.toBe(boom);
  });
});
