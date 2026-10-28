import { h, formatDate } from '../dom';
import type { Tree } from '../types';

export function treeCard(t: Tree): HTMLElement {
  return h(
    'article',
    { class: 'card stack' },
    h('h3', {}, h('a', { href: `#/tree/${t.id}` }, t.name)),
    h(
      'div',
      { class: 'small muted' },
      'by ',
      h('a', { href: `#/user/${t.ownerId}` }, t.ownerName),
      ` · ${formatDate(t.createdAt)}`,
    ),
    h('p', { class: 'small' }, t.description || 'No description.'),
    h('div', {}, ...t.tags.map((x) => h('span', { class: 'tag' }, x))),
    h(
      'div',
      { class: 'small muted' },
      `${t.likeCount} likes · ${t.commentCount} comments${t.visibility === 'private' ? ' · private' : ''}`,
    ),
  );
}
