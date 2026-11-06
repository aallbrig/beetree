import { describe, expect, it } from 'vitest';
import { REPORT_REASONS, queueRows, reportBody, type ReportItem } from './report';

describe('reportBody', () => {
  it('requires a reason', () => {
    expect(reportBody('   ')).toBeNull();
  });
  it('reports the thread when no post is given', () => {
    expect(reportBody(' Spam ')).toEqual({ reason: 'Spam' });
  });
  it('reports a single post', () => {
    expect(reportBody(REPORT_REASONS[0], 'p1')).toEqual({
      postId: 'p1',
      reason: REPORT_REASONS[0],
    });
  });
  it('clips to the API limit', () => {
    expect(reportBody('x'.repeat(900))?.reason).toHaveLength(500);
  });
});

describe('queueRows', () => {
  const mk = (
    id: string,
    threadId: string,
    postId: string | undefined,
    reason: string,
    at: string,
  ): ReportItem => ({
    id,
    threadId,
    postId,
    reporterId: 'u' + id,
    reason,
    createdAt: at,
  });
  it('groups by target and counts distinct reasons once', () => {
    const rows = queueRows([
      mk('1', 't1', 'p1', 'Spam', '2026-01-01T00:00:00Z'),
      mk('2', 't1', 'p1', 'Spam', '2026-01-03T00:00:00Z'),
      mk('3', 't1', 'p1', 'Off topic', '2026-01-02T00:00:00Z'),
      mk('4', 't1', undefined, 'Other', '2026-01-04T00:00:00Z'),
    ]);
    expect(rows).toHaveLength(2);
    expect(rows[0]).toMatchObject({ target: 'thread', count: 1 });
    expect(rows[1]).toMatchObject({ target: 'post', count: 3, reasons: ['Spam', 'Off topic'] });
  });
  it('is empty for no reports', () => {
    expect(queueRows([])).toEqual([]);
  });
});
