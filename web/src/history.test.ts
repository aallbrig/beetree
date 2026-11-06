import { describe, expect, it } from 'vitest';
import { History } from './history';
import {
  addChild,
  countNodes,
  makeNode,
  newSpec,
  parseSpec,
  removeAt,
  stringifySpec,
} from './spec';

describe('History', () => {
  it('undoes and redoes in order', () => {
    const h = new History('a');
    h.record('b');
    h.record('c');
    expect(h.undo()).toBe('b');
    expect(h.undo()).toBe('a');
    expect(h.undo()).toBeUndefined();
    expect(h.redo()).toBe('b');
    expect(h.canRedo).toBe(true);
  });

  it('ignores unchanged snapshots and clears redo on a new edit', () => {
    const h = new History('a');
    expect(h.record('a')).toBe(false);
    expect(h.canUndo).toBe(false);
    h.record('b');
    h.undo();
    h.record('c');
    expect(h.canRedo).toBe(false);
  });

  it('caps the stack', () => {
    const h = new History('0', 3);
    for (let i = 1; i <= 10; i++) h.record(String(i));
    let steps = 0;
    while (h.undo() !== undefined) steps++;
    expect(steps).toBe(3);
  });

  it('restores a tree after an accidental delete', () => {
    const spec = newSpec();
    const h = new History(stringifySpec(spec));
    const before = countNodes(spec.tree);
    const added = addChild(spec.tree, [], makeNode(spec.tree, 'action')) as number[];
    h.record(stringifySpec(spec));
    removeAt(spec.tree, added);
    h.record(stringifySpec(spec));
    const back = parseSpec(h.undo() as string);
    expect(countNodes(back.tree)).toBe(before + 1);
  });
});
