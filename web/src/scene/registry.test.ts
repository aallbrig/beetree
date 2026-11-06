import { describe, expect, it } from 'vitest';
import { DEFAULT_SCENE_ID, SCENES, sceneDefById } from './registry';

describe('scene registry', () => {
  it('lists Demo scene 1 first with unique ids and agents', () => {
    expect(SCENES[0]?.title).toBe('Demo scene 1');
    expect(new Set(SCENES.map((s) => s.id)).size).toBe(SCENES.length);
    for (const s of SCENES) expect(s.agents.length).toBeGreaterThan(0);
  });
  it('finds the default scene', () => {
    expect(sceneDefById(DEFAULT_SCENE_ID)?.agents).toHaveLength(3);
    expect(sceneDefById('nope')).toBeUndefined();
  });
});
