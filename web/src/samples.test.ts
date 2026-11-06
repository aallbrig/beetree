import { describe, expect, it } from 'vitest';
import { classify } from './scene/behaviors';
import { SAMPLES, sampleById } from './samples';
import { findProblems, isLeaf, parseSpec, walk } from './spec';

describe('built-in templates', () => {
  it('ships four to five templates with unique ids', () => {
    expect(SAMPLES.length).toBeGreaterThanOrEqual(4);
    expect(SAMPLES.length).toBeLessThanOrEqual(5);
    expect(new Set(SAMPLES.map((s) => s.id)).size).toBe(SAMPLES.length);
  });

  it('covers every trope VISION.md names', () => {
    const ids = SAMPLES.map((s) => s.id);
    for (const trope of ['guard', 'coward', 'boss', 'shopkeeper', 'companion'])
      expect(ids, trope).toContain(`sample:${trope}`);
  });

  for (const sample of SAMPLES) {
    it(`${sample.title} parses, has unique node names and runs in 3D`, () => {
      const spec = parseSpec(sample.yaml);
      expect(findProblems(spec.tree), 'builder problems').toEqual([]);
      const names = new Set<string>();
      for (const [node] of walk(spec.tree)) {
        expect(names.has(node.name), `duplicate node name ${node.name}`).toBe(false);
        names.add(node.name);
        // a leaf the 3D preview cannot interpret would just stand still
        if (isLeaf(node.type)) expect(classify(node), node.name).not.toBe('generic');
      }
    });
  }
});

describe('scene tree references', () => {
  it('resolves every built-in sample id used by the demo scene and the add-agent default', async () => {
    const { DEMO_SCENE } = await import('./scene/loader');
    const { readFileSync } = await import('node:fs');
    const scenes = readFileSync(new URL('./views/scenes.ts', import.meta.url), 'utf8');
    const ids = [
      ...DEMO_SCENE.map((a) => a.treeId),
      ...[...scenes.matchAll(/'(sample:[a-z]+)'/g)].map((m) => m[1]),
    ];
    for (const id of ids) expect(sampleById(id ?? ''), String(id)).toBeDefined();
  });
});
