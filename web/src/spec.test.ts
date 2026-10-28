import { describe, expect, it } from 'vitest';
import {
  addChild,
  countNodes,
  getNode,
  makeNode,
  moveSibling,
  newSpec,
  parseSpec,
  removeAt,
  rename,
  stringifySpec,
  walk,
} from './spec';

describe('spec', () => {
  it('round-trips through YAML', () => {
    const spec = newSpec('demo');
    const back = parseSpec(stringifySpec(spec));
    expect(back).toEqual(spec);
  });

  it('rejects malformed specs with readable errors', () => {
    expect(() => parseSpec('- not a mapping')).toThrow(/mapping/);
    expect(() => parseSpec('tree:\n  type: wizard\n  name: x')).toThrow(/unknown node type/);
    expect(() => parseSpec('tree:\n  type: action')).toThrow(/name is required/);
    expect(() => parseSpec('tree: [')).toThrow();
  });

  it('adds, finds and removes nodes by path', () => {
    const spec = newSpec();
    const added = addChild(spec.tree, [], makeNode(spec.tree, 'action'));
    expect(Array.isArray(added)).toBe(true);
    expect(countNodes(spec.tree)).toBe(6);
    const path = added as number[];
    expect(getNode(spec.tree, path)?.type).toBe('action');
    expect(removeAt(spec.tree, path)).toBe(true);
    expect(countNodes(spec.tree)).toBe(5);
    expect(removeAt(spec.tree, [])).toBe(false);
  });

  it('refuses invalid structure', () => {
    const spec = newSpec();
    expect(addChild(spec.tree, [1], makeNode(spec.tree, 'action'))).toMatch(/cannot have children/);
    const dec = makeNode(spec.tree, 'decorator');
    addChild(spec.tree, [], dec);
    const decPath = [2];
    expect(typeof addChild(spec.tree, decPath, makeNode(spec.tree, 'action'))).toBe('object');
    expect(addChild(spec.tree, decPath, makeNode(spec.tree, 'action'))).toMatch(
      /exactly one child/,
    );
  });

  it('keeps generated names unique', () => {
    const spec = newSpec();
    const a = makeNode(spec.tree, 'action');
    addChild(spec.tree, [], a);
    const b = makeNode(spec.tree, 'action');
    expect(b.name).not.toBe(a.name);
  });

  it('renames with validation', () => {
    const spec = newSpec();
    expect(rename(spec.tree, [1], 'bad name!')).toMatch(/letters/);
    expect(rename(spec.tree, [1], 'chase')).toMatch(/already used/);
    expect(rename(spec.tree, [1], 'wander')).toBeUndefined();
    expect(getNode(spec.tree, [1])?.name).toBe('wander');
  });

  it('reorders siblings', () => {
    const spec = newSpec();
    const p = moveSibling(spec.tree, [0], 1);
    expect(p).toEqual([1]);
    expect(getNode(spec.tree, [1])?.name).toBe('chase');
    expect(moveSibling(spec.tree, [1], 1)).toEqual([1]);
  });

  it('walks every node depth-first', () => {
    const names = [...walk(newSpec().tree)].map(([n]) => n.name);
    expect(names).toEqual(['root', 'chase', 'has_target', 'chase_target', 'patrol']);
  });
});
