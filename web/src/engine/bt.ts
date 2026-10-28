import type { NodeSpec, Path } from '../spec';

export type Status = 'success' | 'failure' | 'running';

/** Implements the leaves (actions and conditions) of a tree for one agent. */
export interface LeafHandler {
  /** Tick a leaf. Actions may return 'running' across ticks. */
  tick(node: NodeSpec, id: string, dt: number): Status;
  /** Called when a leaf that was running is no longer being ticked (preempted). */
  abort(node: NodeSpec, id: string): void;
}

export const pathId = (path: Path): string => (path.length === 0 ? 'r' : `r.${path.join('.')}`);

interface DecoratorState {
  count: number;
  elapsed: number;
}

/**
 * A reactive behavior tree runtime: every tick re-evaluates from the root, so
 * higher-priority branches can preempt running ones (the "reactive pattern" in
 * the BeeTree spec). Leaves keep their own continuity via the LeafHandler.
 */
export class Runtime {
  /** Last status of each node id from the most recent tick, for visualisation. */
  readonly statuses = new Map<string, Status>();
  private runningLeaves = new Map<string, NodeSpec>();
  private visitedRunning = new Map<string, NodeSpec>();
  private deco = new Map<string, DecoratorState>();

  constructor(
    private readonly root: NodeSpec,
    private readonly leaf: LeafHandler,
  ) {}

  tick(dt: number): Status {
    this.statuses.clear();
    this.visitedRunning = new Map();
    const result = this.exec(this.root, [], dt);
    for (const [id, node] of this.runningLeaves) {
      if (!this.visitedRunning.has(id)) this.leaf.abort(node, id);
    }
    this.runningLeaves = this.visitedRunning;
    return result;
  }

  reset(): void {
    for (const [id, node] of this.runningLeaves) this.leaf.abort(node, id);
    this.runningLeaves.clear();
    this.deco.clear();
    this.statuses.clear();
  }

  private exec(node: NodeSpec, path: Path, dt: number): Status {
    const id = pathId(path);
    const status = this.run(node, path, id, dt);
    this.statuses.set(id, status);
    return status;
  }

  private run(node: NodeSpec, path: Path, id: string, dt: number): Status {
    const kids = node.children ?? [];
    switch (node.type) {
      case 'action':
      case 'condition': {
        const s = this.leaf.tick(node, id, dt);
        if (s === 'running') this.visitedRunning.set(id, node);
        return s;
      }
      case 'sequence': {
        for (let i = 0; i < kids.length; i++) {
          const child = kids[i];
          if (!child) continue;
          const s = this.exec(child, [...path, i], dt);
          if (s !== 'success') return s;
        }
        return 'success';
      }
      case 'selector': {
        for (let i = 0; i < kids.length; i++) {
          const child = kids[i];
          if (!child) continue;
          const s = this.exec(child, [...path, i], dt);
          if (s !== 'failure') return s;
        }
        return 'failure';
      }
      case 'parallel': {
        let running = false;
        let failed = false;
        for (let i = 0; i < kids.length; i++) {
          const child = kids[i];
          if (!child) continue;
          const s = this.exec(child, [...path, i], dt);
          if (s === 'failure') failed = true;
          if (s === 'running') running = true;
        }
        if (failed) return 'failure';
        return running ? 'running' : 'success';
      }
      case 'decorator':
        return this.decorate(node, path, id, dt);
    }
  }

  private decorate(node: NodeSpec, path: Path, id: string, dt: number): Status {
    const child = node.children?.[0];
    if (!child) return 'failure';
    const st = this.deco.get(id) ?? { count: 0, elapsed: 0 };
    this.deco.set(id, st);
    const num = (key: string, fallback: number): number => {
      const v = node.parameters?.[key];
      return typeof v === 'number' ? v : fallback;
    };
    const cs = (): Status => this.exec(child, [...path, 0], dt);
    switch (node.decorator) {
      case 'negate': {
        const s = cs();
        return s === 'running' ? s : s === 'success' ? 'failure' : 'success';
      }
      case 'always_succeed':
        return cs() === 'running' ? 'running' : 'success';
      case 'always_fail':
        return cs() === 'running' ? 'running' : 'failure';
      case 'until_fail': {
        return cs() === 'failure' ? 'success' : 'running';
      }
      case 'until_succeed': {
        return cs() === 'success' ? 'success' : 'running';
      }
      case 'retry': {
        const s = cs();
        if (s !== 'failure') {
          if (s === 'success') st.count = 0;
          return s;
        }
        st.count++;
        if (st.count >= num('attempts', 3)) {
          st.count = 0;
          return 'failure';
        }
        return 'running';
      }
      case 'timeout': {
        const s = cs();
        if (s !== 'running') {
          st.elapsed = 0;
          return s;
        }
        st.elapsed += dt;
        if (st.elapsed >= num('seconds', 5)) {
          st.elapsed = 0;
          return 'failure';
        }
        return 'running';
      }
      case 'cooldown': {
        if (st.elapsed > 0) {
          st.elapsed -= dt;
          return 'failure';
        }
        const s = cs();
        if (s !== 'running') st.elapsed = num('seconds', 3);
        return s;
      }
      case 'repeat':
      default: {
        // runs the child forever; one child tick per runtime tick
        cs();
        return 'running';
      }
    }
  }
}
