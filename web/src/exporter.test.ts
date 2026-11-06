import { describe, expect, it } from 'vitest';
import { parseSpec, stringifySpec, newSpec } from './spec';
import { fileName } from './exporter';

describe('fileName', () => {
  it('slugs the tree name', () => {
    expect(fileName('Guard Patrol!', 'beetree.yaml')).toBe('guard-patrol.beetree.yaml');
  });
  it('falls back when the name is empty', () => {
    expect(fileName('  ', 'cs')).toBe('tree.cs');
  });
});

describe('exported YAML round trip', () => {
  it('parses back to the same spec', () => {
    const spec = newSpec();
    expect(parseSpec(stringifySpec(spec))).toEqual(spec);
  });
});
