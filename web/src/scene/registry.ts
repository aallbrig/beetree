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
  {
    id: 'combat',
    title: 'Demo scene 2',
    controls: ['timer'],
    agents: [
      {
        name: 'Trooper 1',
        treeId: 'example:trooper',
        faction: 'allies',
        model: 'character-b',
        x: -8,
        z: 8,
      },
      {
        name: 'Trooper 2',
        treeId: 'example:trooper',
        faction: 'allies',
        model: 'character-c',
        x: -3,
        z: 9,
      },
      {
        name: 'Trooper 3',
        treeId: 'example:trooper',
        faction: 'allies',
        model: 'character-d',
        x: 2,
        z: 9,
      },
      {
        name: 'Trooper 4',
        treeId: 'example:trooper',
        faction: 'allies',
        model: 'character-e',
        x: 7,
        z: 8,
      },
      {
        name: 'Commander',
        treeId: 'example:commander',
        faction: 'enemies',
        role: 'commander',
        model: 'character-h',
        x: 0,
        z: -12,
      },
      {
        name: 'Skirmisher 1',
        treeId: 'example:skirmisher',
        faction: 'enemies',
        model: 'character-f',
        x: -6,
        z: -9,
      },
      {
        name: 'Skirmisher 2',
        treeId: 'example:skirmisher',
        faction: 'enemies',
        model: 'character-f',
        x: 6,
        z: -9,
      },
      {
        name: 'Shieldbearer',
        treeId: 'example:shieldbearer',
        faction: 'enemies',
        model: 'character-g',
        x: 0,
        z: -7,
      },
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
