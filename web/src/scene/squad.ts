import { Runtime, type LeafHandler, type Status } from '../engine/bt';
import type { NodeSpec } from '../spec';
import { AgentBehaviors, type Body } from './behaviors';

export const SLOTS = ['leader', 'flanker', 'support'] as const;
export type Slot = (typeof SLOTS)[number];

/** The slot a leaf acts on: `parameters.slot`, or the leader when it names none. */
export function slotOf(node: NodeSpec): Slot | undefined {
  const s = node.parameters?.slot;
  if (s === undefined) return 'leader';
  return SLOTS.find((x) => x === s);
}

/**
 * One tree, several agents. The tree ticks once per frame; each leaf acts on
 * the body in the slot it names, so one tree gives each agent its own action.
 */
export class Squad {
  private readonly runtime: Runtime;
  private readonly behaviors = new Map<Slot, AgentBehaviors>();

  constructor(root: NodeSpec, bodies: Partial<Record<Slot, Body>>) {
    for (const slot of SLOTS) {
      const body = bodies[slot];
      if (body) this.behaviors.set(slot, new AgentBehaviors(body));
    }
    const route: LeafHandler = {
      tick: (node, id, dt) => this.tickLeaf(node, id, dt),
      abort: (node, id) => {
        this.abortLeaf(node, id);
      },
    };
    this.runtime = new Runtime(root, route);
  }

  tick(dt: number): Status {
    return this.runtime.tick(dt);
  }

  get statuses(): ReadonlyMap<string, Status> {
    return this.runtime.statuses;
  }

  /** LeafHandler: route the leaf to its slot's agent. An empty slot fails the leaf. */
  private tickLeaf(node: NodeSpec, id: string, dt: number): Status {
    const slot = slotOf(node);
    const b = slot && this.behaviors.get(slot);
    return b ? b.tick(node, id, dt) : 'failure';
  }

  private abortLeaf(node: NodeSpec, id: string): void {
    const slot = slotOf(node);
    if (slot) this.behaviors.get(slot)?.abort(node, id);
  }
}
