import { describe, expect, it } from 'vitest';
import {
  addChild,
  countNodes,
  duplicateAt,
  indent,
  outdent,
  getNode,
  makeNode,
  moveSibling,
  newSpec,
  parseSpec,
  removeAt,
  rename,
  findProblems,
  stringifySpec,
  parseValidatorError,
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

describe('findProblems', () => {
  it('points at the childless sequence by path', () => {
    const spec = newSpec();
    const res = addChild(spec.tree, [], makeNode(spec.tree, 'sequence'));
    expect(Array.isArray(res)).toBe(true);
    const found = findProblems(spec.tree).filter((p) => p.message.includes('needs at least one'));
    expect(found.map((p) => p.path)).toContainEqual(res);
  });
});

describe('findProblems with several bad nodes', () => {
  it('reports one problem per bad node, each naming its node', () => {
    const spec = newSpec();
    for (let i = 0; i < 3; i++) addChild(spec.tree, [], makeNode(spec.tree, 'sequence'));
    const found = findProblems(spec.tree).filter((p) => p.message.includes('needs at least one'));
    expect(found).toHaveLength(3);
    const names = new Set(found.map((p) => p.message.split(':')[0]));
    expect(names.size).toBe(3);
  });
});

describe('findProblems with duplicate names', () => {
  it('flags the second node that reuses a name, with a fix', () => {
    const spec = parseSpec(
      'tree:\n  type: sequence\n  name: root\n  children:\n    - type: action\n      name: go\n    - type: action\n      name: go\n',
    );
    const found = findProblems(spec.tree).filter((p) => p.message.includes('more than once'));
    expect(found).toHaveLength(1);
    expect(found[0]?.path).toEqual([1]);
    expect(found[0]?.fix).toMatch(/rename/i);
  });
});

describe('restructuring', () => {
  it('duplicates a subtree with unique names', () => {
    const spec = newSpec();
    const res = duplicateAt(spec.tree, [0]);
    expect(res).toEqual([1]);
    expect(getNode(spec.tree, [1])?.name).toBe('chase_2');
    expect(getNode(spec.tree, [1, 0])?.name).toBe('has_target_2');
    const names = [...walk(spec.tree)].map(([n]) => n.name);
    expect(new Set(names).size).toBe(names.length);
    expect(typeof duplicateAt(spec.tree, [])).toBe('string');
  });

  it('outdents a node to follow its parent', () => {
    const spec = newSpec();
    expect(outdent(spec.tree, [0, 1])).toEqual([1]);
    expect(getNode(spec.tree, [1])?.name).toBe('chase_target');
    expect(typeof outdent(spec.tree, [0])).toBe('string');
  });

  it('indents a node into the sibling above, rejecting leaves', () => {
    const spec = newSpec();
    expect(indent(spec.tree, [1])).toEqual([0, 2]);
    expect(getNode(spec.tree, [0, 2])?.name).toBe('patrol');
    const leaf = newSpec();
    expect(typeof indent(leaf.tree, [0, 1])).toBe('string');
    expect(getNode(leaf.tree, [0, 1])?.name).toBe('chase_target');
  });
});

describe('parseValidatorError', () => {
  it('maps a server error to the node path it names', () => {
    expect(parseValidatorError('tree: type is required')).toEqual({
      path: [],
      message: 'type is required',
    });
    expect(
      parseValidatorError(
        'tree.children[1].children[0]: composite "sequence" must have at least one child',
      ),
    ).toEqual({
      path: [1, 0],
      message: 'composite "sequence" must have at least one child',
    });
  });

  it('leaves errors that name no node unpathed', () => {
    expect(parseValidatorError('metadata.name is required')).toEqual({
      path: null,
      message: 'metadata.name is required',
    });
  });
});

describe('findProblems fix hints', () => {
  it('gives every problem kind a non-empty fix', () => {
    const spec = parseSpec(
      'tree:\n  type: sequence\n  name: root\n  children:\n    - type: action\n      name: go\n    - type: action\n      name: go\n    - type: sequence\n      name: empty\n    - type: decorator\n      name: deco\n',
    );
    const found = findProblems(spec.tree);
    const kinds = ['more than once', 'needs at least one', 'needs exactly one'];
    for (const k of kinds) {
      const p = found.find((x) => x.message.includes(k));
      expect(p, k).toBeDefined();
      expect(p?.fix.length).toBeGreaterThan(10);
    }
  });
});
