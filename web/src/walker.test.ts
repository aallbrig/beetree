import { readFileSync, writeFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { SAMPLES } from './samples';
import { parseSpec } from './spec';
import { runAll, SimWalker, type Policy } from './walker';

const POLICIES: Policy[] = ['success', 'failure', 'running', 'cycle'];
const TEMPLATES = new URL('./testdata/walker-templates.json', import.meta.url);
const GOLDEN = new URL('./testdata/walker-golden.json', import.meta.url);

interface GoldenCase {
  template: string;
  policy: string;
  result: string;
  trace: { name: string; type: string; depth: number; event: string; status: string }[];
}

const templates = (): Record<string, unknown> =>
  Object.fromEntries(SAMPLES.map((s) => [s.id, parseSpec(s.yaml).tree]));

describe('SimWalker', () => {
  it('shares its template trees with the Go parity test', () => {
    // Regenerate with: UPDATE_WALKER_FIXTURES=1 npx vitest run walker
    if (process.env.UPDATE_WALKER_FIXTURES)
      writeFileSync(TEMPLATES, JSON.stringify(templates(), null, 2) + '\n');
    expect(JSON.parse(readFileSync(TEMPLATES, 'utf8'))).toEqual(templates());
  });

  it('matches the Go walker trace on every template and policy', () => {
    const golden = JSON.parse(readFileSync(GOLDEN, 'utf8')) as GoldenCase[];
    expect(golden).toHaveLength(SAMPLES.length * POLICIES.length);
    for (const c of golden) {
      const sample = SAMPLES.find((s) => s.id === c.template);
      expect(sample, c.template).toBeDefined();
      const w = runAll(parseSpec(sample?.yaml ?? '').tree, c.policy as Policy);
      expect(w.result, `${c.template}/${c.policy}`).toBe(c.result);
      expect(
        w.trace.map((s) => ({
          name: s.name,
          type: s.type,
          depth: s.depth,
          event: s.event,
          status: s.status ?? '',
        })),
        `${c.template}/${c.policy}`,
      ).toEqual(c.trace);
    }
  });

  it('pauses at each leaf and resolves composites', () => {
    const tree = parseSpec(SAMPLES[1]?.yaml ?? '').tree; // guard
    const w = new SimWalker(tree);
    w.step();
    expect(w.state).toBe('waiting');
    expect(w.current?.node.name).toBe('can_see_target');
    w.resolve('failure'); // combat fails, selector moves to patrol
    expect(w.current?.node.name).toBe('move_to_waypoint');
    expect(w.statuses().get('r.0')).toBe('failure');
    expect(w.visited().has('r.1')).toBe(true);
    w.resolve('success');
    w.resolve('success');
    w.resolve('success');
    expect(w.state).toBe('complete');
    expect(w.result).toBe('success');
  });

  it('ignores resolve when not waiting, and resets', () => {
    const w = new SimWalker(parseSpec(SAMPLES[0]?.yaml ?? '').tree);
    w.resolve('success');
    expect(w.trace).toHaveLength(0);
    w.step();
    w.reset();
    expect(w.state).toBe('ready');
  });
});
