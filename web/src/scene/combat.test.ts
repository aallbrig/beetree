import { describe, expect, it } from 'vitest';
import { Runtime } from '../engine/bt';
import { findProblems, isLeaf, parseSpec, walk } from '../spec';
import { AgentBehaviors, classify, type Anim, type Body, type Vec2 } from './behaviors';
import { COMBAT_TREES } from './combat';
import { specFor } from './loader';
import { sceneDefById, SCENES } from './registry';
import { DamageTimer } from './timer';

class Fighter implements Body {
  pos: Vec2;
  health = 100;
  alert = 0;
  target: Vec2 | undefined;
  waypoint: Vec2 = { x: 0, z: 0 };
  leader: Vec2 | undefined;
  commanderAlive = true;
  underFire = false;
  anim: Anim = 'idle';
  start: Vec2;
  constructor(
    readonly id: string,
    readonly faction: string,
    x: number,
    z: number,
  ) {
    this.pos = { x, z };
    this.start = { x, z };
  }
  moveToward(p: Vec2, speed: number, dt: number): number {
    const d = Math.hypot(p.x - this.pos.x, p.z - this.pos.z);
    const step = Math.min(d, speed * dt);
    if (d > 0) {
      this.pos.x += ((p.x - this.pos.x) / d) * step;
      this.pos.z += ((p.z - this.pos.z) / d) * step;
    }
    return d - step;
  }
  moveAwayFrom(p: Vec2, speed: number, dt: number): void {
    const d = Math.hypot(this.pos.x - p.x, this.pos.z - p.z) || 1;
    this.pos.x += ((this.pos.x - p.x) / d) * speed * dt;
    this.pos.z += ((this.pos.z - p.z) / d) * speed * dt;
  }
  faceToward(): void {}
  nextWaypoint(): void {}
  play(a: Anim): void {
    this.anim = a;
  }
  say(): void {}
  strike(): void {}
  runtime(tree: string): Runtime {
    return new Runtime(parseSpec(COMBAT_TREES[tree] ?? '').tree, new AgentBehaviors(this));
  }
}

const distBetween = (a: Vec2, b: Vec2): number => Math.hypot(a.x - b.x, a.z - b.z);

describe('scene 2 data', () => {
  it('is listed as Demo scene 2 and every tree loads, validates and is not generic', async () => {
    const scene = sceneDefById('combat');
    expect(scene?.title).toBe('Demo scene 2');
    expect(SCENES.map((s) => s.title)).toContain('Demo scene 2');
    for (const a of scene?.agents ?? []) {
      const spec = await specFor(a.treeId);
      expect(spec, a.name).toBeDefined();
    }
    for (const [id, yaml] of Object.entries(COMBAT_TREES)) {
      const tree = parseSpec(yaml).tree;
      expect(findProblems(tree), id).toEqual([]);
      for (const [node] of walk(tree)) {
        if (isLeaf(node.type)) expect(classify(node), `${id}/${node.name}`).not.toBe('generic');
      }
    }
  });

  it('has two factions and exactly one Commander', () => {
    const agents = sceneDefById('combat')?.agents ?? [];
    expect(new Set(agents.map((a) => a.faction))).toEqual(new Set(['allies', 'enemies']));
    expect(agents.filter((a) => a.role === 'commander')).toHaveLength(1);
  });
});

describe('scene 2 behavior', () => {
  it('a Trooper holds the line, then retreats once the damage timer drains it', () => {
    const t = new Fighter('t', 'allies', 0, 9);
    t.target = { x: 0, z: 2 }; // enemy within rifle range
    const rt = t.runtime('example:trooper');
    for (let i = 0; i < 10; i++) rt.tick(0.1);
    expect(t.anim).toBe('attack-melee-right'); // shooting, not retreating
    const before = distBetween(t.pos, t.target);

    const timer = new DamageTimer();
    timer.target = { faction: 'allies' };
    timer.damage = 40;
    timer.setPeriod(1);
    timer.setOn(true);
    for (let i = 0; i < 25; i++) timer.update(0.1, [t]);
    for (let i = 0; i < 4; i++) timer.update(1, [t]);
    expect(t.health).toBeLessThan(30);
    for (let i = 0; i < 20; i++) rt.tick(0.1);
    expect(t.anim).toBe('sprint');
    expect(distBetween(t.pos, t.target)).toBeGreaterThan(before);
  });

  it('Skirmishers harass while the Commander lives and flee when it falls', () => {
    const s = new Fighter('s', 'enemies', 0, -9);
    s.target = { x: 0, z: 0 };
    s.leader = { x: 0, z: -12 };
    const rt = s.runtime('example:skirmisher');
    for (let i = 0; i < 5; i++) rt.tick(0.1);
    expect(s.anim).toBe('attack-melee-right');
    const near = distBetween(s.pos, s.target);

    s.commanderAlive = false;
    for (let i = 0; i < 10; i++) rt.tick(0.1);
    expect(s.anim).toBe('sprint');
    expect(distBetween(s.pos, s.target)).toBeGreaterThan(near);
  });

  it('a Commander charges, then retreats to recharge when hurt', () => {
    const c = new Fighter('c', 'enemies', 0, -12);
    c.target = { x: 0, z: 0 };
    const rt = c.runtime('example:commander');
    for (let i = 0; i < 5; i++) rt.tick(0.1);
    expect(c.anim).toBe('sprint');
    c.health = 10;
    for (let i = 0; i < 5; i++) rt.tick(0.1);
    expect(c.anim).toBe('sprint');
    const away = distBetween(c.pos, c.target);
    for (let i = 0; i < 10; i++) rt.tick(0.1);
    expect(distBetween(c.pos, c.target)).toBeGreaterThan(away);
  });
});
