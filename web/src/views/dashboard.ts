import { api } from '../api';
import { formatDate, h, toast } from '../dom';
import { type Ctx } from '../router';
import type { Tree } from '../types';
import { filterSortTrees, renameTree, type TreeSort } from '../treelist';

export async function dashboardView(ctx: Ctx): Promise<void> {
  const [me, stats, trees, scenes] = await Promise.all([
    api.me(),
    api.myStats(),
    api.myTrees(),
    api.scenes(),
  ]);
  const stat = (label: string, n: number): HTMLElement =>
    h('div', { class: 'card stat' }, h('b', {}, n), h('span', { class: 'muted small' }, label));

  const tbody = h('tbody', {});
  const search = h('input', {
    type: 'search',
    id: 'tree-search',
    placeholder: 'Search by name or tag',
    'aria-label': 'Search your trees',
  });
  const sortSel = h(
    'select',
    { id: 'tree-sort', 'aria-label': 'Sort your trees' },
    h('option', { value: 'recent' }, 'Recently updated'),
    h('option', { value: 'name' }, 'Name'),
    h('option', { value: 'likes' }, 'Most liked'),
  );

  const makeRow = (t: Tree): HTMLElement => {
    const del = h('button', { class: 'btn small danger', type: 'button' }, 'Delete');
    del.addEventListener('click', () => {
      if (!confirm(`Delete "${t.name}"? This cannot be undone.`)) return;
      api.deleteTree(t.id).then(
        () => {
          toast('Deleted.');
          trees.items = trees.items.filter((x) => x.id !== t.id);
          refresh();
        },
        () => {
          toast('Could not delete.', 'err');
        },
      );
    });
    const rename = h('button', { class: 'btn small', type: 'button' }, 'Rename');
    rename.addEventListener('click', () => {
      const next = prompt('New name', t.name);
      if (next === null || next.trim() === '' || next.trim() === t.name) return;
      renameTree(api, t.id, next).then(
        (saved) => {
          t.name = saved.name;
          t.updatedAt = saved.updatedAt;
          toast('Renamed.');
          refresh();
        },
        () => {
          toast('Could not rename.', 'err');
        },
      );
    });
    return h(
      'tr',
      { id: `row-${t.id}`, class: 'tree-row' },
      h('td', {}, h('a', { href: `#/tree/${t.id}` }, t.name)),
      h('td', {}, t.visibility === 'public' ? 'Public' : 'Private'),
      h('td', {}, `${t.likeCount} likes`),
      h('td', {}, `${t.commentCount} comments`),
      h('td', {}, formatDate(t.updatedAt)),
      h(
        'td',
        { class: 'row' },
        h('a', { class: 'btn small', href: `#/builder/${t.id}` }, 'Edit'),
        rename,
        del,
      ),
    );
  };
  const refresh = (): void => {
    const shown = filterSortTrees(trees.items, search.value, sortSel.value as TreeSort);
    tbody.replaceChildren(
      ...(shown.length
        ? shown.map(makeRow)
        : [h('tr', {}, h('td', { colspan: '6', class: 'muted' }, 'No trees match.'))]),
    );
  };
  search.addEventListener('input', refresh);
  sortSel.addEventListener('change', refresh);
  refresh();

  ctx.root.append(
    h(
      'div',
      { class: 'row', style: 'justify-content:space-between' },
      h('h1', {}, `Welcome, ${me.displayName}`),
      h('a', { class: 'btn primary', href: '#/builder' }, '+ New tree'),
    ),
    h(
      'div',
      {
        class: 'grid',
        style: 'grid-template-columns:repeat(auto-fit,minmax(130px,1fr));margin-bottom:24px',
      },
      stat('Trees', stats.trees),
      stat('Shared', stats.publicTrees),
      stat('Likes received', stats.likes),
      stat('Comments received', stats.comments),
      stat('Forum threads', stats.forumThreads),
      stat('Forum replies', stats.forumPosts),
    ),
    h('h2', {}, 'Your trees'),
    trees.items.length
      ? h(
          'div',
          { class: 'card', style: 'overflow:auto' },
          h('div', { class: 'row', style: 'margin-bottom:12px' }, search, sortSel),
          h(
            'table',
            {},
            h(
              'thead',
              {},
              h(
                'tr',
                {},
                ...['Name', 'Visibility', 'Likes', 'Comments', 'Updated', ''].map((c) =>
                  h('th', {}, c),
                ),
              ),
            ),
            tbody,
          ),
        )
      : h(
          'div',
          { class: 'hero-note' },
          'You have not created a tree yet. ',
          h('a', { href: '#/builder' }, 'Build your first one →'),
        ),
    h('h2', { style: 'margin-top:28px' }, 'Your 3D scenes'),
    scenes.items.length
      ? h(
          'div',
          { class: 'grid' },
          ...scenes.items.map((s) =>
            h(
              'div',
              { class: 'card stack' },
              h('h3', {}, s.name),
              h('div', { class: 'small muted' }, `${s.agents.length} agents · ${s.visibility}`),
              h(
                'div',
                { class: 'row' },
                h('a', { class: 'btn small', href: `#/scenes/${s.id}` }, 'Edit'),
                h('a', { class: 'btn small', href: `#/explore?scene=${s.id}` }, 'Watch'),
              ),
            ),
          ),
        )
      : h(
          'div',
          { class: 'hero-note' },
          'No scenes yet. ',
          h('a', { href: '#/scenes' }, 'Place agents in a 3D world →'),
        ),
  );
}
