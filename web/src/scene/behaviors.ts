import type { LeafHandler, Status } from '../engine/bt';
import type { NodeSpec } from '../spec';

export type Anim =
  'idle' | 'walk' | 'sprint' | 'attack-melee-right' | 'emote-yes' | 'emote-no' | 'die';

export interface Vec2 {
  x: number;
  z: number;
}

/** What a behavior can see and do. The 3D world implements this; tests fake it. */
export interface Body {
  pos: Vec2;
  health: number;
  alert: number;
  target: Vec2 | undefined;
  waypoint: Vec2;
  /** Move toward a point at speed (units/s); returns remaining distance. */
  moveToward(p: Vec2, speed: number, dt: number): number;
  moveAwayFrom(p: Vec2, speed: number, dt: number): void;
  faceToward(p: Vec2, dt: number): void;
  nextWaypoint(): void;
  play(anim: Anim): void;
  say(text: string): void;
  strike(kind: 'melee' | 'ranged'): void;
}

export const RANGES = {
  detect: 12,
  attack: 2.4,
  ranged: 9,
  arrive: 0.6,
  lowHealth: 30,
  alert: 20,
} as const;

const dist = (a: Vec2, b: Vec2): number => Math.hypot(a.x - b.x, a.z - b.z);

/** Names are matched on both the node's `name` and its `node` implementation id. */
function labelOf(n: NodeSpec): string {
  return `${n.name} ${n.node ?? ''}`.toLowerCase();
}

type Kind =
  | 'sees'
  | 'inRange'
  | 'lowHealth'
  | 'alerted'
  | 'hasWaypoint'
  | 'patrol'
  | 'moveWaypoint'
  | 'nextWaypoint'
  | 'chase'
  | 'attack'
  | 'shoot'
  | 'flee'
  | 'wait'
  | 'face'
  | 'emote'
  | 'heal'
  | 'calm'
  | 'generic';

// order matters: more specific patterns first
const PATTERNS: [RegExp, Kind][] = [
  [/in_?range|within_?range|target_in/, 'inRange'],
  [/health_?(is_?)?(low|below)|low_?health|is_?dying|is_?hurt|wounded|critical/, 'lowHealth'],
  [/has_?target|see_?target|can_?see|visible|detect|spotted|sees_?enemy/, 'sees'],
  [/alert/, 'alerted'],
  [/has_?waypoint/, 'hasWaypoint'],
  [/next_?waypoint|advance_?waypoint|cycle_?waypoint/, 'nextWaypoint'],
  [/move_?to_?waypoint|walk_?to_?waypoint|go_?to_?waypoint/, 'moveWaypoint'],
  [/patrol|wander|roam/, 'patrol'],
  [/flee|cover|escape|retreat|run_?away|hide/, 'flee'],
  [/shoot|ranged|fire|throw|cast/, 'shoot'],
  [/attack|strike|melee|hit|slash|punch|bite/, 'attack'],
  [/chase|pursue|approach|move_?to_?target|follow|charge|close_?in/, 'chase'],
  [/face|rotate|turn|look_?at|aim/, 'face'],
  [/heal|regen|recover|drink|rest_?up/, 'heal'],
  [/calm|clear_?alert|set_?alert|reset_?alert/, 'calm'],
  [/wait|idle|pause|search|look_?around|stand|linger|stay/, 'wait'],
  [/emote|animation|animate|taunt|celebrate|play/, 'emote'],
];

export function classify(node: NodeSpec): Kind {
  const l = labelOf(node);
  for (const [re, kind] of PATTERNS) if (re.test(l)) return kind;
  return 'generic';
}

function seconds(node: NodeSpec, fallback: number): number {
  const p = node.parameters;
  for (const k of ['seconds', 'duration', 'time']) {
    const v = p?.[k];
    if (typeof v === 'number' && v > 0) return Math.min(v, 10);
  }
  return fallback;
}

/** Game-AI semantics for common leaf names, so any tree does something visible. */
export class AgentBehaviors implements LeafHandler {
  private elapsed = new Map<string, number>();

  constructor(private readonly body: Body) {}

  tick(node: NodeSpec, id: string, dt: number): Status {
    const b = this.body;
    const kind = classify(node);
    const t = this.elapsed.get(id) ?? 0;
    const timed = (total: number, anim: Anim, label: string): Status => {
      b.play(anim);
      b.say(label);
      if (t + dt >= total) {
        this.elapsed.delete(id);
        return 'success';
      }
      this.elapsed.set(id, t + dt);
      return 'running';
    };
    const ok = (v: boolean): Status => (v ? 'success' : 'failure');

    switch (kind) {
      case 'sees':
        return ok(b.target !== undefined && dist(b.pos, b.target) < RANGES.detect);
      case 'inRange':
        return ok(b.target !== undefined && dist(b.pos, b.target) <= RANGES.attack);
      case 'lowHealth':
        return ok(b.health < RANGES.lowHealth);
      case 'alerted':
        return ok(
          b.alert > 0.3 || (b.target !== undefined && dist(b.pos, b.target) < RANGES.alert),
        );
      case 'hasWaypoint':
        return 'success';
      case 'nextWaypoint':
        b.nextWaypoint();
        return 'success';
      case 'moveWaypoint':
      case 'patrol': {
        const left = b.moveToward(b.waypoint, 2.2, dt);
        b.play('walk');
        b.say(node.name);
        if (left < RANGES.arrive) {
          if (kind === 'patrol') b.nextWaypoint();
          return 'success';
        }
        return 'running';
      }
      case 'chase': {
        if (!b.target) return 'failure';
        const left = b.moveToward(b.target, 3.6, dt);
        b.play('sprint');
        b.say(node.name);
        return left <= RANGES.attack ? 'success' : 'running';
      }
      case 'attack':
      case 'shoot': {
        if (!b.target) return 'failure';
        const reach = kind === 'attack' ? RANGES.attack * 1.3 : RANGES.ranged;
        if (dist(b.pos, b.target) > reach) return 'failure';
        b.faceToward(b.target, dt);
        if (t === 0) b.strike(kind === 'attack' ? 'melee' : 'ranged');
        return timed(0.8, 'attack-melee-right', node.name);
      }
      case 'flee': {
        const from = b.target ?? { x: b.pos.x - 1, z: b.pos.z };
        b.moveAwayFrom(from, 4, dt);
        b.play('sprint');
        b.say(node.name);
        const total = seconds(node, 2.5);
        if (t + dt >= total) {
          this.elapsed.delete(id);
          return 'success';
        }
        this.elapsed.set(id, t + dt);
        return 'running';
      }
      case 'face':
        if (b.target) b.faceToward(b.target, dt * 6);
        b.say(node.name);
        return 'success';
      case 'wait':
        return timed(seconds(node, 1.2), 'idle', node.name);
      case 'heal':
        b.health = Math.min(100, b.health + 40 * dt);
        return timed(seconds(node, 1.5), 'emote-yes', node.name);
      case 'calm':
        b.alert = 0;
        return 'success';
      case 'emote':
        return timed(seconds(node, 1), 'emote-yes', node.name);
      case 'generic':
        // unknown leaf: conditions pass, actions "work" for a moment so the tree stays alive
        if (node.type === 'condition') return 'success';
        return timed(seconds(node, 0.8), 'emote-yes', node.name);
    }
  }

  abort(_node: NodeSpec, id: string): void {
    this.elapsed.delete(id);
  }
}
