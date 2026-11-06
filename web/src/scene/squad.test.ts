import { describe, expect, it } from 'vitest';
import { parseSpec } from '../spec';
import type { Anim, Body, Vec2 } from './behaviors';
import { slotOf, Squad } from './squad';

function body(): Body & { anim: Anim; said: string[] } {
  return {
    pos: { x: 0, z: 0 },
    health: 100,
    alert: 0,
    target: { x: 6, z: 0 },
    waypoint: { x: 5, z: 0 },
    leader: undefined,
    commanderAlive: true,
    underFire: false,
    anim: 'idle',
    said: [],
    moveToward(p: Vec2, speed: number, dt: number) {
      const d = Math.hypot(p.x - this.pos.x, p.z - this.pos.z);
      const step = Math.min(d, speed * dt);
      if (d > 0) this.pos.x += ((p.x - this.pos.x) / d) * step;
      return d - step;
    },
    moveAwayFrom() {},
    faceToward() {},
    nextWaypoint() {},
    play(a: Anim) {
      this.anim = a;
    },
    say(t: string) {
      this.said.push(t);
    },
    strike() {},
  };
}

const tree = parseSpec(`tree:
  type: parallel
  name: squad
  children:
    - {type: action, name: hold_position, parameters: {slot: leader}}
    - {type: action, name: chase_flank, parameters: {slot: flanker}}
    - {type: action, name: follow_leader_close, parameters: {slot: support}}
`).tree;

describe('squad slots', () => {
  it('one tree gives three agents distinct actions', () => {
    const [l, f, s] = [body(), body(), body()];
    s.leader = { x: 3, z: 0 };
    const squad = new Squad(tree, { leader: l, flanker: f, support: s });
    for (let i = 0; i < 5; i++) squad.tick(0.1);
    expect(l.anim).toBe('idle');
    expect(f.anim).toBe('sprint');
    expect(s.anim).toBe('walk');
    expect(l.pos.x).toBe(0);
    expect(f.pos.x).toBeGreaterThan(1);
  });

  it('an unnamed slot means the leader, and an empty slot fails the leaf', () => {
    expect(slotOf({ type: 'action', name: 'a' })).toBe('leader');
    expect(slotOf({ type: 'action', name: 'a', parameters: { slot: 'sniper' } })).toBeUndefined();
    const l = body();
    const squad = new Squad(tree, { leader: l });
    expect(squad.tick(0.1)).toBe('failure');
  });

  it('an old single-agent tree runs unchanged on the leader slot', () => {
    const old = parseSpec('tree: {type: action, name: chase_target}').tree;
    const l = body();
    new Squad(old, { leader: l }).tick(0.1);
    expect(l.anim).toBe('sprint');
  });
});

describe('Explore Demo scene 3', () => {
  it('is listed, its agents share one tree, and each fills a distinct slot', async () => {
    const { sceneDefById, SCENES } = await import('./registry');
    const scene = sceneDefById('squad');
    expect(scene?.title).toBe('Demo scene 3');
    expect(SCENES.map((s) => s.title)).toContain('Demo scene 3');
    const agents = scene?.agents ?? [];
    expect(new Set(agents.map((a) => a.treeId)).size).toBe(1);
    expect(agents.map((a) => a.slot)).toEqual(['leader', 'flanker', 'support']);
  });

  it('the squad example tree is valid and gives three agents three different actions', async () => {
    const { COMBAT_TREES } = await import('./combat');
    const { classify } = await import('./behaviors');
    const { findProblems, isLeaf, walk } = await import('../spec');
    const root = parseSpec(COMBAT_TREES['example:squad'] ?? '').tree;
    expect(findProblems(root)).toEqual([]);
    for (const [node] of walk(root)) {
      if (isLeaf(node.type)) expect(classify(node), node.name).not.toBe('generic');
    }
    const [l, f, s] = [body(), body(), body()];
    s.leader = { x: 3, z: 0 };
    const squad = new Squad(root, { leader: l, flanker: f, support: s });
    for (let i = 0; i < 5; i++) squad.tick(0.1);
    expect([l.anim, f.anim, s.anim]).toEqual(['idle', 'sprint', 'walk']);
  });
});
