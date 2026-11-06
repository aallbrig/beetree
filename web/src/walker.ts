import type { Status } from './engine/bt';
import { pathId } from './engine/bt';
import type { NodeSpec, Path } from './spec';

/**
 * Step-through walker. A TypeScript port of cli/internal/walker/walker.go: it
 * descends to each leaf, pauses for the author's answer, and resolves
 * composites from their children. walker.test.ts proves it matches the Go
 * walker's trace on every built-in template.
 */
export type SimState = 'ready' | 'waiting' | 'complete';
export type SimEvent = 'enter' | 'resolve';

export interface SimStep {
  name: string;
  type: string;
  depth: number;
  event: SimEvent;
  /** Set on resolve events. */
  status?: Status;
  /** Set on resolve events: why the node ended with this status. */
  reason?: string;
  path: Path;
}

interface Frame {
  node: NodeSpec;
  path: Path;
  next: number;
  results: Status[];
  depth: number;
}

// Mirrors the Go walker: a node with no children is a leaf, whatever its type.
const leaf = (n: NodeSpec): boolean => (n.children ?? []).length === 0;

function applyDecorator(node: NodeSpec, s: Status): Status {
  switch (node.decorator) {
    case 'negate':
      return s === 'success' ? 'failure' : s === 'failure' ? 'success' : s;
    case 'always_succeed':
      return s === 'running' ? s : 'success';
    case 'always_fail':
      return s === 'running' ? s : 'failure';
    default:
      return s;
  }
}

function evaluate(node: NodeSpec, results: Status[], next: number): Status | undefined {
  const last = results[results.length - 1];
  const count = (node.children ?? []).length;
  if (node.type === 'parallel') {
    if (next < count) return undefined;
    if (results.includes('failure')) return 'failure';
    return results.includes('running') ? 'running' : 'success';
  }
  if (node.type === 'selector') {
    if (last === 'success' || last === 'running') return last;
    return next >= count ? 'failure' : undefined;
  }
  // sequence, and the Go default for anything else
  if (last === 'failure' || last === 'running') return last;
  return next >= count ? 'success' : undefined;
}

/** Why a composite resolved, naming the child that decided it. */
function explain(node: NodeSpec, results: Status[], status: Status): string {
  const child = (node.children ?? [])[results.length - 1]?.name ?? 'a child';
  const exhausted = (node.children ?? []).length === results.length;
  if (node.type === 'parallel') {
    if (status === 'failure') return 'at least one child failed';
    return status === 'running' ? 'at least one child is still running' : 'every child succeeded';
  }
  if (status === 'running') return `${child} is still running`;
  if (node.type === 'selector') {
    return status === 'success' ? `${child} succeeded` : 'every child failed';
  }
  return status === 'failure' ? `${child} failed` : exhausted ? 'every child succeeded' : '';
}

export class SimWalker {
  state: SimState = 'ready';
  current: { node: NodeSpec; path: Path } | undefined;
  trace: SimStep[] = [];
  result: Status | undefined;
  private stack: Frame[] = [];

  constructor(private readonly root: NodeSpec) {}

  /** Start, and descend to the first leaf. A no-op once started. */
  step(): void {
    if (this.state === 'ready') this.descend(this.root, [], 0);
  }

  /** Answer for the leaf the walker is paused on, then run to the next leaf. */
  resolve(status: Status): void {
    const cur = this.current;
    if (this.state !== 'waiting' || !cur) return;
    const resolved = applyDecorator(cur.node, status);
    this.trace.push({
      name: cur.node.name,
      type: cur.node.type,
      depth: this.stack.length,
      event: 'resolve',
      status: resolved,
      reason:
        resolved === status
          ? `answered ${status}`
          : `answered ${status}, ${cur.node.decorator ?? 'decorator'} made it ${resolved}`,
      path: cur.path,
    });
    this.current = undefined;
    this.bubble(resolved);
  }

  reset(): void {
    this.state = 'ready';
    this.current = undefined;
    this.trace = [];
    this.result = undefined;
    this.stack = [];
  }

  /** Statuses by path id of every node resolved so far, for the outline dots. */
  statuses(): Map<string, Status> {
    const m = new Map<string, Status>();
    for (const s of this.trace) if (s.status) m.set(pathId(s.path), s.status);
    return m;
  }

  /** Path ids of every node entered so far: the visited path. */
  visited(): Set<string> {
    return new Set(this.trace.filter((s) => s.event === 'enter').map((s) => pathId(s.path)));
  }

  private descend(node: NodeSpec, path: Path, depth: number): void {
    this.trace.push({ name: node.name, type: node.type, depth, event: 'enter', path });
    if (leaf(node)) {
      this.state = 'waiting';
      this.current = { node, path };
      return;
    }
    this.stack.push({ node, path, next: 0, results: [], depth });
    const first = (node.children ?? [])[0];
    if (first) this.descend(first, [...path, 0], depth + 1);
  }

  private bubble(childResult: Status): void {
    const frame = this.stack[this.stack.length - 1];
    if (!frame) {
      this.state = 'complete';
      this.result = childResult;
      return;
    }
    frame.results.push(childResult);
    frame.next++;
    const resolved = evaluate(frame.node, frame.results, frame.next);
    if (resolved) {
      this.trace.push({
        name: frame.node.name,
        type: frame.node.type,
        depth: frame.depth,
        event: 'resolve',
        status: resolved,
        reason: explain(frame.node, frame.results, resolved),
        path: frame.path,
      });
      this.stack.pop();
      this.bubble(resolved);
      return;
    }
    const child = (frame.node.children ?? [])[frame.next];
    if (child) this.descend(child, [...frame.path, frame.next], frame.depth + 1);
  }
}

/** Answer policies for Run all. Each leaf, in the order asked, gets the next answer. */
export type Policy = 'success' | 'failure' | 'running' | 'cycle';
const CYCLE: Status[] = ['success', 'failure', 'running'];

export function runAll(root: NodeSpec, policy: Policy): SimWalker {
  const w = new SimWalker(root);
  w.step();
  let i = 0;
  while (w.state === 'waiting') {
    w.resolve(policy === 'cycle' ? (CYCLE[i % 3] ?? 'success') : policy);
    i++;
  }
  return w;
}
