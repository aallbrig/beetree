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
