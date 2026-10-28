import { describe, expect, it } from 'vitest';
import { Runtime } from '../engine/bt';
import { parseSpec } from '../spec';
import { AgentBehaviors, classify, RANGES, type Anim, type Body, type Vec2 } from './behaviors';

function fakeBody(
  over: Partial<Body> = {},
): Body & { anim: Anim; strikes: string[]; waypointIdx: number } {
  const b = {
    pos: { x: 0, z: 0 },
    health: 100,
    alert: 0,
    target: undefined as Vec2 | undefined,
    waypoint: { x: 5, z: 0 },
    anim: 'idle' as Anim,
    strikes: [] as string[],
    waypointIdx: 0,
    moveToward(p: Vec2, speed: number, dt: number) {
      const d = Math.hypot(p.x - this.pos.x, p.z - this.pos.z);
      const step = Math.min(d, speed * dt);
      if (d > 0) {
        this.pos.x += ((p.x - this.pos.x) / d) * step;
        this.pos.z += ((p.z - this.pos.z) / d) * step;
      }
      return d - step;
    },
    moveAwayFrom() {},
    faceToward() {},
    nextWaypoint() {
      this.waypointIdx++;
    },
    play(a: Anim) {
      this.anim = a;
    },
    say() {},
    strike(k: 'melee' | 'ranged') {
      this.strikes.push(k);
    },
    ...over,
  };
  return b;
}

const enemyAi = parseSpec(`tree:
  type: selector
  name: enemy_ai
  children:
    - type: sequence
      name: flee
      children:
        - {type: condition, name: is_dying, node: IsHealthBelow}
        - {type: action, name: run_to_cover, node: MoveTo}
    - type: sequence
      name: combat
      children:
        - {type: condition, name: has_target}
        - type: selector
          name: attack_or_close
          children:
            - type: sequence
              name: melee
              children:
                - {type: condition, name: in_range}
                - {type: action, name: strike}
            - {type: action, name: chase}
    - type: decorator
      name: patrol_forever
      decorator: repeat
      children:
        - type: sequence
          name: cycle
          children:
            - {type: action, name: walk_to_waypoint}
            - {type: action, name: wait_briefly}
            - {type: action, name: next_waypoint}
`).tree;

describe('classify', () => {
  it('recognises common AI vocabulary by name or implementation id', () => {
    const c = (name: string, node?: string) =>
      classify({ type: 'action', name, ...(node ? { node } : {}) });
    expect(c('chase_target')).toBe('chase');
    expect(c('move_to_target')).toBe('chase');
    expect(c('ranged_attack')).toBe('shoot');
    expect(c('melee_attack')).toBe('attack');
    expect(c('whatever', 'MoveToWaypoint')).toBe('moveWaypoint');
    expect(c('is_dying')).toBe('lowHealth');
    expect(c('in_range')).toBe('inRange');
    expect(c('can_see_target')).toBe('sees');
    expect(c('zzz')).toBe('generic');
  });
});

describe('agent behaviour through the runtime', () => {
  const run = (body: ReturnType<typeof fakeBody>, ticks: number) => {
    const rt = new Runtime(enemyAi, new AgentBehaviors(body));
    for (let i = 0; i < ticks; i++) rt.tick(0.1);
    return rt;
  };

  it('patrols toward its waypoint when nothing is around', () => {
    const body = fakeBody();
    run(body, 10);
    expect(body.pos.x).toBeGreaterThan(1);
    expect(body.anim).toBe('walk');
  });

  it('chases an intruder that comes into detection range', () => {
    const body = fakeBody({ target: { x: RANGES.detect - 2, z: 0 } });
    run(body, 10);
    expect(body.anim).toBe('sprint');
    expect(body.pos.x).toBeGreaterThan(2);
  });

  it('attacks once the intruder is in range', () => {
    const body = fakeBody({ target: { x: 1.5, z: 0 } });
    run(body, 3);
    expect(body.anim).toBe('attack-melee-right');
    expect(body.strikes).toEqual(['melee']);
  });

  it('flees when health is low, even with an intruder nearby', () => {
    const body = fakeBody({ health: 10, target: { x: 1.5, z: 0 } });
    const rt = run(body, 2);
    expect(body.anim).toBe('sprint');
    expect(body.strikes).toEqual([]);
    expect(rt.statuses.get('r.0')).toBe('running');
  });

  it('unknown actions still run and finish rather than hanging the tree', () => {
    const t = parseSpec('tree: {type: action, name: do_the_thing}').tree;
    const rt = new Runtime(t, new AgentBehaviors(fakeBody()));
    const seen = new Set<string>();
    for (let i = 0; i < 12; i++) seen.add(rt.tick(0.1));
    expect(seen).toEqual(new Set(['running', 'success']));
  });
});
