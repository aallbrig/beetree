import { api, ApiError } from '../api';
import { isSignedIn } from '../auth';
import { clear, formatDate, h, toast } from '../dom';
import { navigate, type Ctx } from '../router';

export async function forumView(ctx: Ctx): Promise<void> {
  const threads = (await api.threads()).items;
  const list = h('div', { class: 'stack' });
  for (const t of threads) {
    list.append(
      h(
        'article',
        { class: 'card' },
        h('h3', {}, h('a', { href: `#/forum/${t.id}` }, t.title)),
        h(
          'div',
          { class: 'small muted' },
          `${t.author} · ${t.postCount} replies · active ${formatDate(t.lastActivity)}`,
        ),
      ),
    );
  }
  if (!threads.length)
    list.append(h('p', { class: 'muted' }, 'No threads yet. Start the conversation!'));

  const title = h('input', {
    type: 'text',
    maxlength: '120',
    placeholder: 'Title',
    'aria-label': 'Title',
  });
  const body = h('textarea', {
    maxlength: '5000',
    placeholder: 'What do you want to discuss?',
    'aria-label': 'Body',
  });
  const go = h('button', { class: 'btn primary', type: 'button' }, 'Start thread');
  go.addEventListener('click', () => {
    api.createThread(title.value, body.value).then(
      (t) => {
        navigate(`/forum/${t.id}`);
      },
      (e: unknown) => {
        toast(e instanceof ApiError ? e.message : 'Could not post.', 'err');
      },
    );
  });
  ctx.root.append(
    h('h1', {}, 'Forum'),
    h('p', { class: 'muted' }, 'Ask for help, show off your trees, and trade AI design patterns.'),
    isSignedIn()
      ? h(
          'div',
          { class: 'card stack', style: 'margin-bottom:20px' },
          h('h2', {}, 'New thread'),
          title,
          body,
          go,
        )
      : h(
          'div',
          { class: 'hero-note' },
          h('a', { href: '#/signin?next=%2Fforum' }, 'Sign in'),
          ' to start a thread or reply.',
        ),
    list,
  );
}

export async function threadView(ctx: Ctx): Promise<void> {
  const id = ctx.params[0] ?? '';
  const data = await api.thread(id);
  const posts = h('div');
  const paint = (items: typeof data.posts): void => {
    clear(posts);
    for (const p of items) {
      posts.append(
        h(
          'div',
          { class: 'card', style: 'margin-bottom:12px' },
          h(
            'div',
            { class: 'small muted' },
            h('a', { href: `#/user/${p.authorId}` }, p.author),
            ` · ${formatDate(p.createdAt)}`,
          ),
          h('div', {}, p.body),
        ),
      );
    }
  };
  paint(data.posts);
  const reply = h('textarea', {
    maxlength: '5000',
    placeholder: 'Write a reply…',
    'aria-label': 'Reply',
  });
  const send = h('button', { class: 'btn primary', type: 'button' }, 'Reply');
  send.addEventListener('click', () => {
    if (!isSignedIn()) {
      navigate(`/signin?next=${encodeURIComponent(`/forum/${id}`)}`);
      return;
    }
    api.addPost(id, reply.value).then(
      () =>
        api.thread(id).then((d) => {
          reply.value = '';
          paint(d.posts);
        }),
      (e: unknown) => {
        toast(e instanceof ApiError ? e.message : 'Could not reply.', 'err');
      },
    );
  });
  ctx.root.append(
    h('p', {}, h('a', { href: '#/forum' }, '← All threads')),
    h(
      'div',
      { class: 'card stack', style: 'margin-bottom:20px' },
      h('h1', {}, data.thread.title),
      h(
        'div',
        { class: 'small muted' },
        `${data.thread.author} · ${formatDate(data.thread.createdAt)}`,
      ),
      h('p', {}, data.thread.body),
    ),
    posts,
    h('div', { class: 'card stack' }, h('h2', {}, 'Reply'), reply, send),
  );
}
