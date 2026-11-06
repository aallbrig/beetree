import { ApiError, type SharedTree } from './api';
import type { Scene } from './types';

type SceneBody = { name: string; visibility: string; agents: Scene['agents']; confirm?: boolean };
type Save = (id: string | undefined, body: SceneBody) => Promise<Scene>;

/** The confirmation text shown before a public scene publishes its private trees. */
export function publishPrompt(trees: SharedTree[]): string {
  const names = trees.map((t) => `"${t.name}"`).join(', ');
  const noun = trees.length === 1 ? 'tree' : 'trees';
  return `Sharing this scene also makes ${String(trees.length)} private ${noun} public: ${names}. Anyone with the link will be able to read ${trees.length === 1 ? 'it' : 'them'}. Continue?`;
}

/**
 * Saves a scene. When the API answers that publishing needs confirmation, asks
 * the person and, on accept, saves again with confirm set. Declining rejects
 * with an ApiError so callers keep their normal error path.
 */
export async function saveSceneConfirmed(
  save: Save,
  id: string | undefined,
  body: SceneBody,
  ask: (message: string) => boolean,
): Promise<Scene> {
  try {
    return await save(id, body);
  } catch (e) {
    if (!(e instanceof ApiError) || e.status !== 409 || e.trees.length === 0) throw e;
    if (!ask(publishPrompt(e.trees))) {
      throw new ApiError(0, 'Scene not saved: its private trees stay private.');
    }
    return save(id, { ...body, confirm: true });
  }
}
