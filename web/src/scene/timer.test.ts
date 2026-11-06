import { describe, expect, it } from 'vitest';
import { DamageTimer, TIMER_DEFAULTS, type Damageable } from './timer';
import { SCENES, sceneHasControl } from './registry';

const squad = (): Damageable[] => [
  { id: 'a0', faction: 'allies', health: 100 },
  { id: 'a1', faction: 'allies', health: 100 },
  { id: 'a2', faction: 'enemies', health: 100 },
];

describe('DamageTimer', () => {
  it('defaults to a 5 second period and 10 damage', () => {
    const t = new DamageTimer();
    expect(t.period).toBe(TIMER_DEFAULTS.period);
    expect(t.damage).toBe(10);
  });
  it('drains the target faction at the set period', () => {
    const t = new DamageTimer();
    t.target = { faction: 'allies' };
    t.setPeriod(2);
    t.setOn(true);
    const agents = squad();
    t.update(1.9, agents);
    expect(agents[0]?.health).toBe(100);
    t.update(0.2, agents);
    expect(agents.map((a) => a.health)).toEqual([90, 90, 100]);
    t.update(4, agents);
    expect(agents[0]?.health).toBe(70);
  });
  it('drains a single agent and floors at zero', () => {
    const t = new DamageTimer();
    t.target = { agent: 'a2' };
    t.damage = 60;
    t.setPeriod(1);
    t.setOn(true);
    const agents = squad();
    t.update(2, agents);
    expect(agents.map((a) => a.health)).toEqual([100, 100, 0]);
  });
  it('stops when off and clamps the period to 1 to 20', () => {
    const t = new DamageTimer();
    t.target = { faction: 'allies' };
    t.setPeriod(0);
    expect(t.period).toBe(1);
    t.setPeriod(99);
    expect(t.period).toBe(20);
    t.setPeriod(1);
    t.setOn(true);
    const agents = squad();
    t.update(1, agents);
    t.setOn(false);
    expect(t.update(10, agents)).toBe(0);
    expect(agents[0]?.health).toBe(90);
  });
});

describe('scene controls', () => {
  it('shows the timer only on scenes that declare it', () => {
    const declared = SCENES.filter((s) => s.controls.includes('timer'));
    for (const s of SCENES) expect(sceneHasControl(s, 'timer')).toBe(declared.includes(s));
    const base = SCENES[0];
    if (!base) throw new Error('no scenes');
    expect(sceneHasControl({ ...base, controls: ['timer'] }, 'timer')).toBe(true);
    expect(sceneHasControl(base, 'health')).toBe(true);
  });
});
