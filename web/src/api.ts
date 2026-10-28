import { getToken } from './auth';
import { loadConfig } from './config';
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

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
  }
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
  me: () => request<Profile>('GET', '/v1/me'),
  updateMe: (displayName: string, bio: string) =>
    request<Profile>('PUT', '/v1/me', { displayName, bio }),
  myTrees: () => request<Items<Tree>>('GET', '/v1/me/trees'),
  myStats: () => request<Stats>('GET', '/v1/me/stats'),
  user: (id: string) => request<{ profile: Profile; trees: Tree[] }>('GET', `/v1/users/${id}`),
  threads: () => request<Items<Thread>>('GET', '/v1/forum/threads'),
  thread: (id: string) =>
    request<{ thread: Thread; posts: Post[] }>('GET', `/v1/forum/threads/${id}`),
  createThread: (title: string, body: string) =>
    request<Thread>('POST', '/v1/forum/threads', { title, body }),
  addPost: (id: string, body: string) =>
    request<Post>('POST', `/v1/forum/threads/${id}/posts`, { body }),
  scenes: () => request<Items<Scene>>('GET', '/v1/scenes'),
  getScene: (id: string) => request<Scene>('GET', `/v1/scenes/${id}`),
  saveScene: (
    id: string | undefined,
    s: { name: string; visibility: string; agents: SceneAgent[] },
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
