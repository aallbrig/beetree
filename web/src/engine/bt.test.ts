import { describe, expect, it } from 'vitest';
import { parseSpec } from '../spec';
import type { NodeSpec } from '../spec';
import { Runtime, type LeafHandler, type Status } from './bt';

/** Leaves return whatever the script says by name, and record aborts. */
function scripted(script: Record<string, Status | (() => Status)>) {
  const aborted: string[] = [];
  const ticks: string[] = [];
  const handler: LeafHandler = {
    tick(node) {
      ticks.push(node.name);
      const v = script[node.name] ?? 'success';
      return typeof v === 'function' ? v() : v;
    },
    abort(node) {
      aborted.push(node.name);
    },
  };
  return { handler, aborted, ticks };
}

const tree = (yaml: string): NodeSpec => parseSpec(yaml).tree;

describe('Runtime', () => {
  it('sequence stops at the first non-success', () => {
    const t = tree(`tree:
  type: sequence
  name: s
  children:
    - {type: condition, name: a}
    - {type: action, name: b}
    - {type: action, name: c}`);
    const s = scripted({ b: 'failure' });
    expect(new Runtime(t, s.handler).tick(0.1)).toBe('failure');
    expect(s.ticks).toEqual(['a', 'b']);
  });

  it('selector returns the first non-failure', () => {
    const t = tree(`tree:
  type: selector
  name: s
  children:
    - {type: condition, name: a}
    - {type: action, name: b}
    - {type: action, name: c}`);
    const s = scripted({ a: 'failure', b: 'running' });
    const rt = new Runtime(t, s.handler);
    expect(rt.tick(0.1)).toBe('running');
    expect(s.ticks).toEqual(['a', 'b']);
    expect(rt.statuses.get('r')).toBe('running');
    expect(rt.statuses.get('r.0')).toBe('failure');
  });

  it('reactive selector preempts a running branch and aborts it', () => {
    const t = tree(`tree:
  type: selector
  name: root
  children:
    - type: sequence
      name: flee
      children:
        - {type: condition, name: scared}
        - {type: action, name: run}
    - {type: action, name: patrol}`);
    let scared: Status = 'failure';
    const s = scripted({ scared: () => scared, patrol: 'running', run: 'running' });
    const rt = new Runtime(t, s.handler);
    expect(rt.tick(0.1)).toBe('running');
    expect(s.aborted).toEqual([]);
    scared = 'success';
    rt.tick(0.1); // flee now wins; patrol was running and gets aborted
    expect(s.aborted).toEqual(['patrol']);
    scared = 'failure';
    rt.tick(0.1);
    expect(s.aborted).toEqual(['patrol', 'run']);
  });

  it('parallel fails on any failure, runs while any child runs', () => {
    const t = tree(`tree:
  type: parallel
  name: p
  children:
    - {type: action, name: a}
    - {type: action, name: b}`);
    expect(new Runtime(t, scripted({ b: 'running' }).handler).tick(0)).toBe('running');
    expect(new Runtime(t, scripted({ a: 'failure', b: 'running' }).handler).tick(0)).toBe(
      'failure',
    );
    expect(new Runtime(t, scripted({}).handler).tick(0)).toBe('success');
  });

  const deco = (d: string, params = '') =>
    tree(`tree:
  type: decorator
  name: d
  decorator: ${d}
${params}  children:
    - {type: action, name: a}`);

  it('negate and always_* decorators', () => {
    expect(new Runtime(deco('negate'), scripted({}).handler).tick(0)).toBe('failure');
    expect(new Runtime(deco('negate'), scripted({ a: 'failure' }).handler).tick(0)).toBe('success');
    expect(new Runtime(deco('always_succeed'), scripted({ a: 'failure' }).handler).tick(0)).toBe(
      'success',
    );
    expect(new Runtime(deco('always_fail'), scripted({}).handler).tick(0)).toBe('failure');
  });

  it('repeat keeps running and re-ticks the child', () => {
    const s = scripted({});
    const rt = new Runtime(deco('repeat'), s.handler);
    expect(rt.tick(0)).toBe('running');
    expect(rt.tick(0)).toBe('running');
    expect(s.ticks).toEqual(['a', 'a']);
  });

  it('until_fail / until_succeed', () => {
    let n = 0;
    const rt = new Runtime(
      deco('until_fail'),
      scripted({ a: () => (++n < 3 ? 'success' : 'failure') }).handler,
    );
    expect([rt.tick(0), rt.tick(0), rt.tick(0)]).toEqual(['running', 'running', 'success']);
    const rt2 = new Runtime(deco('until_succeed'), scripted({ a: 'failure' }).handler);
    expect(rt2.tick(0)).toBe('running');
  });

  it('retry gives up after N attempts', () => {
    const rt = new Runtime(
      deco('retry', '  parameters: {attempts: 2}\n'),
      scripted({ a: 'failure' }).handler,
    );
    expect([rt.tick(0), rt.tick(0)]).toEqual(['running', 'failure']);
  });

  it('timeout fails a leaf that runs too long', () => {
    const rt = new Runtime(
      deco('timeout', '  parameters: {seconds: 1}\n'),
      scripted({ a: 'running' }).handler,
    );
    expect(rt.tick(0.6)).toBe('running');
    expect(rt.tick(0.6)).toBe('failure');
  });

  it('cooldown blocks re-entry', () => {
    const rt = new Runtime(deco('cooldown', '  parameters: {seconds: 1}\n'), scripted({}).handler);
    expect(rt.tick(0.1)).toBe('success');
    expect(rt.tick(0.5)).toBe('failure');
    expect(rt.tick(0.6)).toBe('failure');
    expect(rt.tick(0.1)).toBe('success');
  });

  it('reset aborts running leaves', () => {
    const s = scripted({ a: 'running' });
    const rt = new Runtime(deco('repeat'), s.handler);
    rt.tick(0);
    rt.reset();
    expect(s.aborted).toEqual(['a']);
  });
});
