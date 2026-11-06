import type { SceneAgent } from '../types';

/** A built-in Explore scene: data, not code, so new scenes are added by appending an entry. */
export interface SceneDef {
  id: string;
  title: string;
  agents: SceneAgent[];
  /** Ids of the controls the scene shows. Scene 1 keeps the health slider and the intruder. */
  controls: string[];
}

export const SCENES: SceneDef[] = [
  {
    id: 'demo',
    title: 'Demo scene 1',
    controls: ['health', 'intruder'],
    agents: [
      { name: 'Patroller', treeId: 'sample:guard', model: 'character-b', x: -7, z: -3 },
      { name: 'Guard', treeId: 'sample:guard', model: 'character-d', x: 6, z: -4 },
      { name: 'Coward', treeId: 'sample:coward', model: 'character-f', x: -1, z: -6 },
    ],
  },
];

export function sceneHasControl(scene: SceneDef | undefined, control: string): boolean {
  return scene?.controls.includes(control) ?? false;
}

export const DEFAULT_SCENE_ID = 'demo';

export function sceneDefById(id: string): SceneDef | undefined {
  return SCENES.find((s) => s.id === id);
}
