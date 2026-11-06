export const REPORT_REASONS = [
  'Spam or advertising',
  'Abusive or harassing',
  'Off topic',
  'Other',
] as const;

export interface ReportBody {
  postId?: string;
  reason: string;
}

/** Build the report request body; null when no reason was chosen. Omit postId to report the thread. */
export function reportBody(reason: string, postId?: string): ReportBody | null {
  const r = reason.trim().slice(0, 500);
  if (!r) return null;
  return postId ? { postId, reason: r } : { reason: r };
}

/** One report as returned by GET /v1/forum/reports (moderators only). */
export interface ReportItem {
  id: string;
  threadId: string;
  postId?: string;
  reporterId: string;
  reason: string;
  createdAt: string;
}

export interface QueueRow {
  key: string;
  threadId: string;
  target: 'thread' | 'post';
  count: number;
  reasons: string[];
  latest: string;
}

/** Group reports by reported thread or post, most recently reported first. */
export function queueRows(items: readonly ReportItem[]): QueueRow[] {
  const rows = new Map<string, QueueRow>();
  for (const r of items) {
    const key = `${r.threadId}/${r.postId ?? ''}`;
    const row = rows.get(key) ?? {
      key,
      threadId: r.threadId,
      target: r.postId ? 'post' : 'thread',
      count: 0,
      reasons: [],
      latest: r.createdAt,
    };
    row.count += 1;
    if (!row.reasons.includes(r.reason)) row.reasons.push(r.reason);
    if (r.createdAt > row.latest) row.latest = r.createdAt;
    rows.set(key, row);
  }
  return [...rows.values()].sort((a, b) =>
    a.latest < b.latest ? 1 : a.latest > b.latest ? -1 : 0,
  );
}
