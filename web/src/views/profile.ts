import { api, ApiError } from '../api';
import { getToken } from '../auth';
import { treeCard } from '../components/cards';
import { formatDate, h, toast } from '../dom';
import type { Ctx } from '../router';

export async function profileView(ctx: Ctx): Promise<void> {
  const [me, stats] = await Promise.all([api.me(), api.myStats()]);
  const name = h('input', {
    type: 'text',
    value: me.displayName,
    maxlength: '32',
    'aria-label': 'Display name',
  });
  const bio = h(
    'textarea',
    { maxlength: '500', 'aria-label': 'Bio', style: 'min-height:90px' },
    me.bio,
  );
  const save = h('button', { class: 'btn primary', type: 'button' }, 'Save profile');
  save.addEventListener('click', () => {
    api.updateMe(name.value.trim(), bio.value).then(
      () => {
        toast('Profile saved.');
      },
      (e: unknown) => {
        toast(e instanceof ApiError ? e.message : 'Could not save.', 'err');
      },
    );
  });
  const copy = h('button', { class: 'btn', type: 'button' }, 'Copy my API token');
  copy.addEventListener('click', () => {
    void (async () => {
      const token = await getToken();
      if (!token) {
        toast('Sign in again first.', 'err');
        return;
      }
      await navigator.clipboard.writeText(token);
      toast('Token copied. It expires in about an hour.');
    })().catch(() => {
      toast('Could not copy.', 'err');
    });
  });
  ctx.root.append(
    h('h1', {}, 'Your profile'),
    h(
      'div',
      { class: 'studio' },
      h(
        'div',
        { class: 'card stack' },
        h(
          'div',
          {},
          h('label', {}, 'Display name'),
          name,
          h(
            'div',
            { class: 'hint' },
            '3–32 letters, digits, _ or -. Shown on your trees, comments and posts.',
          ),
        ),
        h('div', {}, h('label', {}, 'Bio'), bio),
        h('div', {}, h('label', {}, 'Email'), h('div', {}, me.email ?? '')),
        h('div', { class: 'small muted' }, `Member since ${formatDate(me.createdAt)}`),
        save,
      ),
      h(
        'div',
        { class: 'stack' },
        h(
          'div',
          { class: 'card stack' },
          h('h2', {}, 'Engagement'),
          h(
            'p',
            {},
            `${stats.likes} likes and ${stats.comments} comments on your ${stats.trees} trees (${stats.publicTrees} shared).`,
          ),
          h('p', {}, `${stats.forumThreads} forum threads and ${stats.forumPosts} replies.`),
        ),
        h(
          'div',
          { class: 'card stack' },
          h('h2', {}, 'Developer access'),
          h(
            'p',
            { class: 'small muted' },
            'Use a token with the REST API or the MCP server. Tokens last about an hour.',
          ),
          copy,
          h('a', { href: '/docs/api/' }, 'API docs →'),
          h('a', { href: '/docs/mcp/' }, 'MCP docs →'),
        ),
      ),
    ),
  );
}

export async function userView(ctx: Ctx): Promise<void> {
  const { profile, trees } = await api.user(ctx.params[0] ?? '');
  ctx.root.append(
    h('h1', {}, profile.displayName),
    h('p', { class: 'muted' }, profile.bio || 'No bio yet.'),
    h('h2', {}, 'Shared trees'),
    trees.length
      ? h('div', { class: 'grid' }, ...trees.map(treeCard))
      : h('p', { class: 'muted' }, 'Nothing shared yet.'),
  );
}
