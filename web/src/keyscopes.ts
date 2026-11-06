/** API key scopes: presets, the custom grid, expiry choices and key-list labels. */

export type Preset = 'read-only' | 'read-write' | 'full';
export type Expiry = '30d' | '90d' | '365d' | 'never';

export const RESOURCES: Record<string, readonly string[]> = {
  trees: ['read', 'write', 'delete'],
  scenes: ['read', 'write', 'delete'],
  forum: ['read', 'write', 'delete'],
  profile: ['read', 'write'],
  keys: ['read', 'write'],
};

export const PRESETS: Record<Preset, { label: string; help: string; scopes: string[] }> = {
  'read-only': {
    label: 'Read-only',
    help: 'Read trees, scenes, forum and profile.',
    scopes: ['trees:read', 'scenes:read', 'forum:read', 'profile:read'],
  },
  'read-write': {
    label: 'Read/write',
    help: 'Create and edit, but never delete.',
    scopes: ['trees:write', 'scenes:write', 'forum:write', 'profile:write'],
  },
  full: {
    label: 'Full',
    help: 'Everything except managing keys.',
    scopes: ['trees:delete', 'scenes:delete', 'forum:delete', 'profile:write'],
  },
};

export const EXPIRIES: { value: Expiry; label: string }[] = [
  { value: '30d', label: '30 days' },
  { value: '90d', label: '90 days (default)' },
  { value: '365d', label: '1 year' },
  { value: 'never', label: 'Never' },
];

/** The scope for one grid cell: ticking an action grants it and every lower one. */
export function scopeFor(resource: string, action: string): string {
  return `${resource}:${action}`;
}

/** Collapse a set of ticked grid cells to the highest action per resource. */
export function gridToScopes(ticked: Iterable<string>): string[] {
  const best = new Map<string, number>();
  for (const sc of ticked) {
    const [res, act] = sc.split(':');
    const actions = res === undefined ? undefined : RESOURCES[res];
    const rank = act === undefined || !actions ? -1 : actions.indexOf(act);
    if (res === undefined || rank < 0) continue;
    best.set(res, Math.max(best.get(res) ?? -1, rank));
  }
  const out: string[] = [];
  for (const [res, rank] of best) {
    const action = RESOURCES[res]?.[rank];
    if (action) out.push(`${res}:${action}`);
  }
  return out.sort();
}

export interface KeyBody {
  name: string;
  expiry: Expiry;
  preset?: Preset;
  scopes?: string[];
}

/** Build the create-key request; null while the form is incomplete. */
export function keyBody(
  name: string,
  choice: Preset | 'custom' | '',
  ticked: Iterable<string>,
  expiry: Expiry,
): KeyBody | null {
  const n = name.trim();
  if (n === '' || choice === '') return null;
  if (choice !== 'custom') return { name: n, preset: choice, expiry };
  const scopes = gridToScopes(ticked);
  return scopes.length === 0 ? null : { name: n, scopes, expiry };
}

/** Chips for a key: one per resource at its highest action, e.g. "trees: write". */
export function scopeChips(scopes: readonly string[]): string[] {
  const merged = gridToScopes(scopes);
  return merged.map((s) => s.replace(':', ': '));
}

export function expiryText(expiresAt: string | undefined, now: Date = new Date()): string {
  if (!expiresAt) return 'never expires';
  const t = new Date(expiresAt);
  if (Number.isNaN(t.getTime())) return 'expiry unknown';
  if (t.getTime() <= now.getTime()) return 'expired';
  const days = Math.ceil((t.getTime() - now.getTime()) / 86_400_000);
  return days === 1 ? 'expires in 1 day' : `expires in ${String(days)} days`;
}

export function lastUsedText(lastUsedAt: string | undefined, now: Date = new Date()): string {
  if (!lastUsedAt) return 'never used';
  const t = new Date(lastUsedAt);
  if (Number.isNaN(t.getTime())) return 'never used';
  const hours = Math.floor((now.getTime() - t.getTime()) / 3_600_000);
  if (hours < 1) return 'used within the last hour';
  if (hours < 24) return `used ${String(hours)} hours ago`;
  const days = Math.floor(hours / 24);
  return days === 1 ? 'used 1 day ago' : `used ${String(days)} days ago`;
}
