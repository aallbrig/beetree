import { readdirSync, readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { parse } from 'yaml';
import { countNodes, findProblems, parseSpec, stringifySpec } from './spec';

const DIR = new URL('../../examples/', import.meta.url);
const FILES = readdirSync(DIR).filter((f) => f.endsWith('.beetree.yaml'));

describe('example trees open in the builder', () => {
  it('finds the examples', () => {
    expect(FILES.length).toBeGreaterThan(0);
  });
  for (const f of FILES) {
    it(`${f} survives import then export unchanged`, () => {
      const text = readFileSync(new URL(f, DIR), 'utf8');
      const spec = parseSpec(text);
      expect(parse(stringifySpec(spec))).toEqual(parse(text));
      expect(findProblems(spec.tree)).toEqual([]);
      expect(countNodes(spec.tree)).toBeGreaterThan(1);
    });
  }
});
