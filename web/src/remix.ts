import { h } from './dom';
import type { Lineage } from './types';

export interface Segment {
  text: string;
  href?: string;
}

/**
 * The history line as text segments, by the state of the original. Only a
 * public original links to its page; the owner's profile is linked in every
 * state; a deleted original shows no title.
 */
export function historySegments(l: Lineage): Segment[] {
  const owner: Segment = { text: l.ownerName, href: `#/user/${l.ownerId}` };
  switch (l.state) {
    case 'deleted':
      return [{ text: `Remixed from a deleted ${l.kind} by ` }, owner];
    case 'private':
      return [{ text: 'Remixed from ' }, { text: l.title }, { text: ', now private, by ' }, owner];
    default:
      return [
        { text: 'Remixed from ' },
        { text: l.title, ...(l.state === 'public' ? { href: `#/${l.kind}/${l.id}` } : {}) },
        { text: ' by ' },
        owner,
      ];
  }
}

export function historyLine(l: Lineage): HTMLElement {
  return h(
    'p',
    { class: 'small muted history-line' },
    ...historySegments(l).map((s) => (s.href ? h('a', { href: s.href }, s.text) : s.text)),
  );
}

/** Why a remix cannot be published yet, or undefined when it can. */
export function publishBlockReason(identicalToOriginal: boolean): string | undefined {
  return identicalToOriginal
    ? 'This remix is identical to the original. Change the tree before sharing it.'
    : undefined;
}
