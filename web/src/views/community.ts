import { api } from '../api';
import { isSignedIn } from '../auth';
import { treeCard } from '../components/cards';
import { clear, h } from '../dom';
import type { Ctx } from '../router';

export function communityView(ctx: Ctx): void {
  const signedIn = isSignedIn();
  const q = h('input', {
    type: 'search',
    placeholder: 'Search trees, tags, ideas…',
    'aria-label': 'Search',
  });
  const sort = h(
    'select',
    { 'aria-label': 'Sort' },
    h('option', { value: 'recent' }, 'Newest'),
    h('option', { value: 'popular' }, 'Most liked'),
  );
  const results = h('div', { class: 'grid' });
  const note = h('p', { class: 'muted', role: 'status' });
  const load = (): void => {
    note.textContent = 'Loading…';
    const source = signedIn ? api.listTrees : api.publicTrees;
    source(q.value.trim(), sort.value).then(
      (r) => {
        clear(results);
        note.textContent = r.items.length ? '' : 'No trees yet. Be the first to share one!';
        for (const t of r.items) results.append(treeCard(t));
      },
      () => {
        note.textContent = 'Could not load trees right now.';
      },
    );
  };
  let timer: ReturnType<typeof setTimeout> | undefined;
  q.addEventListener('input', () => {
    clearTimeout(timer);
    timer = setTimeout(load, 250);
  });
  sort.addEventListener('change', load);
  ctx.root.append(
    h('h1', {}, 'Community trees'),
    ...(signedIn
      ? []
      : [
          h(
            'div',
            { class: 'note', role: 'note' },
            'You are viewing the daily community snapshot. ',
            h('a', { href: '#/signup' }, 'Create a free account'),
            ' or ',
            h('a', { href: '#/signin?next=%2Fcommunity' }, 'sign in'),
            ' for live results, likes, comments and forking.',
          ),
        ]),
    h(
      'p',
      { class: 'muted' },
      'Behavior trees shared by other creators. Like them, comment, or fork one into your own builder.',
    ),
    h(
      'div',
      { class: 'row', style: 'margin-bottom:16px' },
      h('div', { style: 'flex:1;min-width:220px' }, q),
      sort,
    ),
    note,
    results,
  );
  load();
}
