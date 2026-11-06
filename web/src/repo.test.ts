import { describe, expect, it } from 'vitest';
import { repoHref } from './repo';

describe('repoHref', () => {
  it('joins a path onto the repository url', () => {
    expect(repoHref('https://example.com/o/r', 'releases')).toBe(
      'https://example.com/o/r/releases',
    );
    expect(repoHref('https://example.com/o/r')).toBe('https://example.com/o/r');
  });
  it('is undefined when the repository url is empty', () => {
    expect(repoHref('', 'releases')).toBeUndefined();
  });
});
