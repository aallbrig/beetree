import { api } from '../api';
import { sampleById } from '../samples';
import { parseSpec, type NodeSpec } from '../spec';
import type { SceneAgent } from '../types';
import type { World } from './world';

/** Resolve a tree id (a sample or a saved tree) to its root node. */
export async function specFor(treeId: string | undefined): Promise<NodeSpec | undefined> {
  if (!treeId) return undefined;
  const sample = sampleById(treeId);
  if (sample) return parseSpec(sample.yaml).tree;
  const tree = await api.getTree(treeId);
  return tree.spec ? parseSpec(tree.spec).tree : undefined;
}

const FALLBACK: NodeSpec = { type: 'action', name: 'idle_about' };

export interface LoadedAgent extends SceneAgent {
  id: string;
}

/** Replace the world's agents with the given placements. Trees that fail to load fall back to idling. */
export async function populate(world: World, agents: SceneAgent[]): Promise<LoadedAgent[]> {
  for (const id of [...world.agents.keys()]) world.removeAgent(id);
  const out: LoadedAgent[] = [];
  let n = 0;
  for (const a of agents) {
    const spec = await specFor(a.treeId).catch(() => undefined);
    const id = `a${n++}`;
    await world.addAgent({
      id,
      name: a.name,
      model: a.model,
      x: a.x,
      z: a.z,
      spec: spec ?? FALLBACK,
    });
    out.push({ ...a, id });
  }
  return out;
}

export const DEMO_SCENE: SceneAgent[] = [
  { name: 'Patroller', treeId: 'sample:patrol', model: 'character-b', x: -7, z: -3 },
  { name: 'Guard', treeId: 'sample:guard', model: 'character-d', x: 6, z: -4 },
  { name: 'Coward', treeId: 'sample:coward', model: 'character-f', x: -1, z: -6 },
];
