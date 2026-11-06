/** Frozen credit for a remix. state is derived by the API from the live original. */
export interface Lineage {
  kind: string;
  id: string;
  title: string;
  ownerId: string;
  ownerName: string;
  at: string;
  state?: 'public' | 'private' | 'deleted';
}
export interface Tree {
  id: string;
  ownerId: string;
  ownerName: string;
  name: string;
  description: string;
  visibility: 'public' | 'private';
  tags: string[];
  spec?: string;
  likeCount: number;
  commentCount: number;
  likedByMe?: boolean;
  createdAt: string;
  updatedAt: string;
  remixedFrom?: Lineage;
}
export interface Comment {
  id: string;
  treeId: string;
  authorId: string;
  author: string;
  body: string;
  createdAt: string;
}
export interface Profile {
  userId: string;
  displayName: string;
  bio: string;
  email?: string;
  moderator?: boolean;
  avatar?: string;
  marketingOptIn?: boolean;
  createdAt: string;
}
export interface Stats {
  trees: number;
  publicTrees: number;
  likes: number;
  comments: number;
  forumThreads: number;
  forumPosts: number;
  scenes: number;
}
export interface Thread {
  id: string;
  title: string;
  author: string;
  authorId: string;
  body: string;
  postCount: number;
  createdAt: string;
  lastActivity: string;
}
export interface Post {
  id: string;
  threadId: string;
  author: string;
  authorId: string;
  body: string;
  createdAt: string;
}
export interface SceneAgent {
  name: string;
  treeId?: string;
  /** Side the agent fights for; scene timers can target a whole faction. */
  faction?: string;
  /** Built-in scenes only: 'commander' lets allies of the faction know when it falls. */
  role?: string;
  /** Built-in squad scenes only: the slot (leader, flanker, support) this agent fills in a shared tree. */
  slot?: string;
  model: string;
  x: number;
  z: number;
}
export interface Scene {
  id: string;
  name: string;
  visibility: 'public' | 'private';
  agents: SceneAgent[];
  createdAt: string;
  updatedAt: string;
}
export interface ValidateResult {
  valid: boolean;
  errors: string[];
  nodeCount: number;
}
export interface Items<T> {
  items: T[];
}
