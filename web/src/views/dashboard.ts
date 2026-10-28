import { api } from '../api';
import { formatDate, h, toast } from '../dom';
import { type Ctx } from '../router';

export async function dashboardView(ctx: Ctx): Promise<void> {
  const [me, stats, trees, scenes] = await Promise.all([
    api.me(),
    api.myStats(),
    api.myTrees(),
    api.scenes(),
  ]);
  const stat = (label: string, n: number): HTMLElement =>
    h('div', { class: 'card stat' }, h('b', {}, n), h('span', { class: 'muted small' }, label));

  const rows = trees.items.map((t) => {
    const del = h('button', { class: 'btn small danger', type: 'button' }, 'Delete');
    del.addEventListener('click', () => {
      if (!confirm(`Delete "${t.name}"? This cannot be undone.`)) return;
      api.deleteTree(t.id).then(
        () => {
          toast('Deleted.');
          document.getElementById(`row-${t.id}`)?.remove();
        },
        () => {
          toast('Could not delete.', 'err');
        },
      );
    });
    return h(
      'tr',
      { id: `row-${t.id}` },
      h('td', {}, h('a', { href: `#/tree/${t.id}` }, t.name)),
      h('td', {}, t.visibility === 'public' ? 'Public' : '🔒 Private'),
      h('td', {}, `♥ ${t.likeCount}`),
      h('td', {}, `💬 ${t.commentCount}`),
      h('td', {}, formatDate(t.updatedAt)),
      h(
        'td',
        { class: 'row' },
        h('a', { class: 'btn small', href: `#/builder/${t.id}` }, 'Edit'),
        del,
      ),
    );
  });

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
            h('tbody', {}, ...rows),
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
