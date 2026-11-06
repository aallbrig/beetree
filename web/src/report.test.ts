import { describe, expect, it } from 'vitest';
import { REPORT_REASONS, reportBody } from './report';

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
