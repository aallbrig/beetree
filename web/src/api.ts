import { getToken } from './auth';
import { loadConfig } from './config';
import type { ReportBody, ReportItem } from './report';
import type {
  Comment,
  Items,
  Post,
  Profile,
  Scene,
  SceneAgent,
  Stats,
  Thread,
  Tree,
  ValidateResult,
} from './types';

export interface ApiKeyInfo {
  id: string;
  name: string;
  prefix: string;
  createdAt: string;
}

/** A private tree that saving a public scene would also publish. */
export interface SharedTree {
  id: string;
  name: string;
}

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    /** Set on a 409 confirmation_required answer. */
    readonly trees: SharedTree[] = [],
  ) {
    super(message);
  }
}

/** Reads the trees list of a confirmation_required answer; empty for any other body. */
export function sharedTrees(data: unknown): SharedTree[] {
  const list: unknown =
    typeof data === 'object' && data !== null && 'trees' in data ? data.trees : [];
  if (!Array.isArray(list)) return [];
  const out: SharedTree[] = [];
  for (const t of list as unknown[]) {
    if (typeof t !== 'object' || t === null) continue;
    const { id, name } = t as { id?: unknown; name?: unknown };
    if (typeof id === 'string' && typeof name === 'string') out.push({ id, name });
  }
  return out;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const cfg = await loadConfig();
  const headers: Record<string, string> = {};
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  const token = await getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(cfg.apiUrl + path, {
    method,
    headers,
    ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
  });
  if (res.status === 204) return undefined as T;
  const data: unknown = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err =
      typeof data === 'object' && data !== null && 'error' in data
        ? (data as { error?: { message?: unknown } }).error
        : undefined;
    throw new ApiError(
      res.status,
      typeof err?.message === 'string' ? err.message : `Request failed (${res.status})`,
      sharedTrees(data),
    );
  }
  // JSON enters here: the API contract (openapi.yaml, tested server-side) is the type source.
  return data as T;
}

export interface TreeInput {
  name: string;
  description: string;
  visibility: 'public' | 'private';
  tags: string[];
  spec: string;
}

export const api = {
  listTrees: (q = '', sort = 'recent') =>
    request<Items<Tree>>(
      'GET',
      `/v1/trees?${new URLSearchParams({ q, sort, limit: '60' }).toString()}`,
    ),
  getTree: (id: string) => request<Tree>('GET', `/v1/trees/${id}`),
  remixTree: (id: string) => request<Tree>('POST', `/v1/trees/${id}/remix`),
  createTree: (t: TreeInput) => request<Tree>('POST', '/v1/trees', t),
  updateTree: (id: string, t: TreeInput) => request<Tree>('PUT', `/v1/trees/${id}`, t),
  deleteTree: (id: string) => request<undefined>('DELETE', `/v1/trees/${id}`),
  like: (id: string) =>
    request<{ liked: boolean; likeCount: number }>('PUT', `/v1/trees/${id}/like`),
  unlike: (id: string) =>
    request<{ liked: boolean; likeCount: number }>('DELETE', `/v1/trees/${id}/like`),
  comments: (id: string) => request<Items<Comment>>('GET', `/v1/trees/${id}/comments`),
  addComment: (id: string, body: string) =>
    request<Comment>('POST', `/v1/trees/${id}/comments`, { body }),
  deleteComment: (id: string, cid: string) =>
    request<undefined>('DELETE', `/v1/trees/${id}/comments/${cid}`),
  publicTrees: (q: string, sort: string) =>
    request<Items<Tree> & { generatedAt: string }>(
      'GET',
      `/v1/public/trees?q=${encodeURIComponent(q)}&sort=${sort}`,
    ),
  listKeys: () => request<Items<ApiKeyInfo>>('GET', '/v1/keys'),
  createKey: (name: string, preset: 'read-only' | 'read-write' | 'full' = 'full') =>
    request<ApiKeyInfo & { key: string }>('POST', '/v1/keys', { name, preset }),
  deleteKey: (id: string) => request<undefined>('DELETE', `/v1/keys/${id}`),
  me: () => request<Profile>('GET', '/v1/me'),
  updateMe: (
    displayName: string,
    bio: string,
    more: {
      avatar?: string;
      marketingOptIn?: boolean;
      marketingSource?: 'signup' | 'profile';
    } = {},
  ) => request<Profile>('PUT', '/v1/me', { displayName, bio, ...more }),
  myTrees: () => request<Items<Tree>>('GET', '/v1/me/trees'),
  myStats: () => request<Stats>('GET', '/v1/me/stats'),
  user: (id: string) =>
    request<{ profile: Profile; trees: Tree[]; scenes?: Scene[] }>('GET', `/v1/users/${id}`),
  threads: () => request<Items<Thread>>('GET', '/v1/forum/threads'),
  thread: (id: string) =>
    request<{ thread: Thread; posts: Post[] }>('GET', `/v1/forum/threads/${id}`),
  createThread: (title: string, body: string) =>
    request<Thread>('POST', '/v1/forum/threads', { title, body }),
  reportThread: (id: string, body: ReportBody) =>
    request<undefined>('POST', `/v1/forum/threads/${id}/report`, body),
  recentReports: () => request<Items<ReportItem>>('GET', '/v1/forum/reports'),
  addPost: (id: string, body: string) =>
    request<Post>('POST', `/v1/forum/threads/${id}/posts`, { body }),
  scenes: () => request<Items<Scene>>('GET', '/v1/scenes'),
  getScene: (id: string) => request<Scene>('GET', `/v1/scenes/${id}`),
  saveScene: (
    id: string | undefined,
    s: { name: string; visibility: string; agents: SceneAgent[]; confirm?: boolean },
  ) =>
    id ? request<Scene>('PUT', `/v1/scenes/${id}`, s) : request<Scene>('POST', '/v1/scenes', s),
  deleteScene: (id: string) => request<undefined>('DELETE', `/v1/scenes/${id}`),
  validate: (spec: string) => request<ValidateResult>('POST', '/v1/validate', { spec }),
  render: (spec: string, format: string) =>
    request<{ output: string }>('POST', '/v1/render', { spec, format }),
  generate: (spec: string, engine: string) =>
    request<{ files: { path: string; content: string; isStub: boolean }[] }>(
      'POST',
      '/v1/generate',
      { spec, engine },
    ),
};
