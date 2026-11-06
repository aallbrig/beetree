import { describe, expect, it } from 'vitest';
import { readdirSync, readFileSync } from 'node:fs';
import { parse } from 'yaml';
import {
  addChild,
  addVar,
  addCustomNode,
  formatParams,
  getCustomNodes,
  parseList,
  parseParams,
  removeCustomNode,
  updateCustomNode,
  getBlackboard,
  parseDefault,
  removeVar,
  updateVar,
  countNodes,
  type NodeSpec,
  findNodes,
  branchIds,
  revealPath,
  visiblePaths,
  duplicateAt,
  indent,
  outdent,
  getNode,
  makeNode,
  moveSibling,
  setPolicy,
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

describe('example trees survive import then export', () => {
  const dir = new URL('../../examples/', import.meta.url);
  const files = readdirSync(dir).filter((f) => f.endsWith('.beetree.yaml'));
  it('finds the scene 2 trees', () => {
    for (const n of ['trooper', 'skirmisher', 'shieldbearer', 'commander'])
      expect(files).toContain(`${n}.beetree.yaml`);
  });
  for (const f of files) {
    it(`${f} loses nothing`, () => {
      const text = readFileSync(new URL(f, dir), 'utf8');
      expect(parse(stringifySpec(parseSpec(text)))).toEqual(parse(text));
    });
  }
});

describe('parallel policies', () => {
  const make = () => {
    const root = makeNode(newSpec().tree, 'parallel');
    return root;
  };
  it('sets, exports and clears success and failure policies', () => {
    const spec = newSpec();
    spec.tree = make();
    expect(setPolicy(spec.tree, [], 'success_policy', 'require_all')).toBeUndefined();
    expect(setPolicy(spec.tree, [], 'failure_policy', 'require_one')).toBeUndefined();
    const out = parse(stringifySpec(spec)) as {
      tree: { success_policy: string; failure_policy: string };
    };
    expect(out.tree.success_policy).toBe('require_all');
    expect(out.tree.failure_policy).toBe('require_one');
    setPolicy(spec.tree, [], 'success_policy', '');
    expect(spec.tree.success_policy).toBeUndefined();
  });
  it('rejects unknown values and non-parallel nodes', () => {
    const n = make();
    expect(setPolicy(n, [], 'success_policy', 'bogus')).toMatch(/must be one of/);
    const seq = makeNode(newSpec().tree, 'sequence');
    expect(setPolicy(seq, [], 'success_policy', 'require_all')).toMatch(/only parallel/);
  });
});

describe('blackboard editing', () => {
  it('adds, renames and removes a variable and exports it', () => {
    const spec = newSpec('bb');
    expect(addVar(spec, 'health', 'int')).toBe('health');
    expect(addVar(spec, 'health')).toBe('health_2');
    expect(
      updateVar(spec, 'health', { default: parseDefault('int', '100'), description: 'hp' }),
    ).toBe(true);
    expect(updateVar(spec, 'health', { name: 'hp' })).toBe(true);
    expect(updateVar(spec, 'hp', { name: 'health_2' })).toBe(false);
    let back = parseSpec(stringifySpec(spec));
    expect(getBlackboard(back)).toEqual([
      { name: 'hp', type: 'int', default: 100, description: 'hp' },
      { name: 'health_2', type: 'string' },
    ]);
    expect(removeVar(back, 'hp')).toBe(true);
    expect(removeVar(back, 'hp')).toBe(false);
    expect(removeVar(back, 'health_2')).toBe(true);
    back = parseSpec(stringifySpec(back));
    expect(back.blackboard).toBeUndefined();
  });

  it('clears an empty default and parses typed defaults', () => {
    const spec = newSpec('bb');
    addVar(spec, 'speed', 'float');
    updateVar(spec, 'speed', { default: 2.5 });
    updateVar(spec, 'speed', { default: parseDefault('float', '') });
    expect(getBlackboard(spec)[0]).toEqual({ name: 'speed', type: 'float' });
    expect(parseDefault('bool', 'true')).toBe(true);
    expect(parseDefault('string', '12')).toBe('12');
  });
});

describe('custom node editing', () => {
  it('parses and formats parameter text', () => {
    const ps = parseParams('animation_name:string, speed:float=1.5, loop:bool=true, bare');
    expect(ps).toEqual([
      { name: 'animation_name', type: 'string' },
      { name: 'speed', type: 'float', default: 1.5 },
      { name: 'loop', type: 'bool', default: true },
      { name: 'bare', type: 'string' },
    ]);
    expect(parseParams(formatParams(ps))).toEqual(ps);
    expect(parseList(' a, ,b ')).toEqual(['a', 'b']);
  });

  it('adds, edits, round-trips and removes a custom node', () => {
    const spec = newSpec('cn');
    expect(addCustomNode(spec, 'PlayAnimation')).toBe('PlayAnimation');
    expect(addCustomNode(spec, 'PlayAnimation', 'condition')).toBe('PlayAnimation_2');
    expect(
      updateCustomNode(spec, 'PlayAnimation', {
        description: 'Play a clip',
        parameters: parseParams('speed:float=1'),
        blackboard_reads: ['target'],
      }),
    ).toBe(true);
    expect(updateCustomNode(spec, 'PlayAnimation_2', { name: 'PlayAnimation' })).toBe(false);
    expect(updateCustomNode(spec, 'nope', { description: 'x' })).toBe(false);
    const back = parseSpec(stringifySpec(spec));
    expect(getCustomNodes(back)[0]).toEqual({
      name: 'PlayAnimation',
      type: 'action',
      description: 'Play a clip',
      parameters: [{ name: 'speed', type: 'float', default: 1 }],
      blackboard_reads: ['target'],
    });
    updateCustomNode(back, 'PlayAnimation', { parameters: [], description: '' });
    expect(getCustomNodes(back)[0]).toEqual({
      name: 'PlayAnimation',
      type: 'action',
      blackboard_reads: ['target'],
    });
    expect(removeCustomNode(back, 'PlayAnimation')).toBe(true);
    expect(removeCustomNode(back, 'PlayAnimation_2')).toBe(true);
    expect(back.custom_nodes).toBeUndefined();
  });

  it('keeps the full-enemy-ai custom nodes through import and export', () => {
    const text = readFileSync(
      new URL('../../examples/full-enemy-ai.beetree.yaml', import.meta.url),
      'utf8',
    );
    const spec = parseSpec(text);
    const defs = getCustomNodes(spec);
    expect(defs.length).toBeGreaterThan(0);
    expect(getCustomNodes(parseSpec(stringifySpec(spec)))).toEqual(defs);
  });
});

describe('collapse and find in large trees', () => {
  const idOf = (p: readonly number[]): string => (p.length === 0 ? 'r' : `r.${p.join('.')}`);
  const big = (): NodeSpec => ({
    type: 'selector',
    name: 'root',
    children: Array.from({ length: 20 }, (_, i) => ({
      type: 'sequence' as const,
      name: `branch-${i}`,
      children: Array.from({ length: 20 }, (_, j) => ({
        type: j % 2 ? ('condition' as const) : ('action' as const),
        name: `leaf-${i}-${j}`,
      })),
    })),
  });

  it('finds a node by name anywhere in 400+ nodes, case-insensitively', () => {
    const root = big();
    expect(countNodes(root)).toBe(421);
    expect(findNodes(root, 'LEAF-13-7')).toEqual([[13, 7]]);
    expect(findNodes(root, 'branch-1')).toHaveLength(11);
    expect(findNodes(root, '  ')).toEqual([]);
    expect(findNodes(root, 'nope')).toEqual([]);
  });

  it('collapsing a subtree hides its descendants and keeps the node', () => {
    const root = big();
    const all = visiblePaths(root, new Set(), idOf);
    expect(all).toHaveLength(421);
    const collapsed = new Set(['r.3']);
    const vis = visiblePaths(root, collapsed, idOf);
    expect(vis).toHaveLength(421 - 20);
    expect(vis.some((p) => p.join() === '3')).toBe(true);
    expect(vis.some((p) => p.join() === '3,5')).toBe(false);
  });

  it('collapse all folds the root, and revealing a match reopens only its ancestors', () => {
    const root = big();
    const collapsed = branchIds(root, idOf);
    expect(visiblePaths(root, collapsed, idOf)).toHaveLength(1);
    const [hit] = findNodes(root, 'leaf-13-7');
    expect(hit).toBeDefined();
    revealPath(collapsed, hit ?? [], idOf);
    const vis = visiblePaths(root, collapsed, idOf).map((p) => p.join());
    expect(vis).toContain('13,7');
    expect(vis).not.toContain('12,0');
  });
});
