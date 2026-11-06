import { api, ApiError } from '../api';
import { isSignedIn } from '../auth';
import { clear, formatDate, h, toast } from '../dom';
import { queueRows, reportBody, REPORT_REASONS } from '../report';
import { navigate, type Ctx } from '../router';

const plural = (n: number, one: string, many: string): string => `${n} ${n === 1 ? one : many}`;

/** A character counter under a text field; goes amber near the limit. */
function counted(field: HTMLInputElement | HTMLTextAreaElement, max: number): HTMLElement {
  const n = h('div', { class: 'hint counter', 'aria-live': 'off' });
  const paint = (): void => {
    n.textContent = `${field.value.length} / ${max}`;
    n.classList.toggle('near', field.value.length > max * 0.9);
  };
  field.addEventListener('input', paint);
  paint();
  return n;
}

/** Runs an async post once at a time, so a double tap cannot post twice. */
function once(btn: HTMLButtonElement, busy: string, run: () => Promise<void>): void {
  const label = btn.textContent;
  btn.addEventListener('click', () => {
    if (btn.disabled) return;
    btn.disabled = true;
    btn.textContent = busy;
    run()
      .catch((e: unknown) => {
        toast(e instanceof ApiError ? e.message : 'Could not post.', 'err');
      })
      .finally(() => {
        btn.disabled = false;
        btn.textContent = label;
      });
  });
}

/** A "Report" button that opens an inline reason picker; postId omitted reports the thread. */
function reportControl(threadId: string, postId?: string): HTMLElement {
  const wrap = h('span', { class: 'report' });
  const open = h('button', { class: 'btn small', type: 'button' }, 'Report');
  open.addEventListener('click', () => {
    if (!isSignedIn()) {
      navigate(`/signin?next=${encodeURIComponent(`/forum/${threadId}`)}`);
      return;
    }
    const pick = h(
      'select',
      { 'aria-label': 'Reason for report' },
      h('option', { value: '' }, 'Choose a reason'),
      ...REPORT_REASONS.map((r) => h('option', { value: r }, r)),
    );
    const send = h('button', { class: 'btn small primary', type: 'button' }, 'Send report');
    send.addEventListener('click', () => {
      const body = reportBody(pick.value, postId);
      if (!body) {
        toast('Choose a reason first.', 'err');
        return;
      }
      api.reportThread(threadId, body).then(
        () => {
          clear(wrap);
          wrap.append(h('span', { class: 'small muted' }, 'Reported. A moderator will review it.'));
        },
        (e: unknown) => {
          toast(e instanceof ApiError ? e.message : 'Could not report.', 'err');
        },
      );
    });
    clear(wrap);
    wrap.append(pick, ' ', send);
  });
  wrap.append(open);
  return wrap;
}

/** The moderator review queue; resolves to null for everyone else. */
async function reviewQueue(): Promise<HTMLElement | null> {
  if (!isSignedIn()) return null;
  let rows;
  try {
    if (!(await api.me()).moderator) return null;
    rows = queueRows((await api.recentReports()).items);
  } catch {
    return null;
  }
  const card = h('section', {
    class: 'card stack',
    'aria-label': 'Review queue',
    style: 'margin-bottom:20px',
  });
  card.append(h('h2', {}, 'Review queue'));
  if (!rows.length) card.append(h('p', { class: 'muted' }, 'No reports to review.'));
  for (const r of rows) {
    card.append(
      h(
        'div',
        {},
        h('a', { href: `#/forum/${r.threadId}` }, `Reported ${r.target}`),
        h(
          'span',
          { class: 'small muted' },
          ` · ${r.count} ${r.count === 1 ? 'report' : 'reports'} · ${r.reasons.join(', ')} · ${formatDate(r.latest)}`,
        ),
      ),
    );
  }
  return card;
}

export async function forumView(ctx: Ctx): Promise<void> {
  const queue = await reviewQueue();
  const threads = (await api.threads()).items;
  const list = h('div', { class: 'stack' });
  for (const t of threads) {
    list.append(
      h(
        'article',
        { class: 'card' },
        h('h3', {}, h('a', { href: `#/forum/${t.id}` }, t.title)),
        h(
          'p',
          { class: 'excerpt' },
          t.body.length > 140 ? `${t.body.slice(0, 140).trimEnd()}…` : t.body,
        ),
        h(
          'div',
          { class: 'small muted' },
          `${t.author} · ${plural(t.postCount, 'reply', 'replies')} · active ${formatDate(t.lastActivity)}`,
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
  once(go, 'Posting…', async () => {
    if (!title.value.trim() || !body.value.trim()) {
      toast('Add a title and some text first.', 'err');
      return;
    }
    const t = await api.createThread(title.value, body.value);
    navigate(`/forum/${t.id}`);
  });
  ctx.root.append(
    h('h1', {}, 'Forum'),
    h('p', { class: 'muted' }, 'Ask for help, show off your trees, and trade AI design patterns.'),
    isSignedIn()
      ? h(
          'div',
          { class: 'card stack', style: 'margin-bottom:20px' },
          h('h2', {}, 'New thread'),
          h(
            'p',
            { class: 'small muted' },
            'Plain text only. Links, images and formatting are not supported.',
          ),
          title,
          counted(title, 120),
          body,
          counted(body, 5000),
          go,
        )
      : h(
          'div',
          { class: 'hero-note' },
          h('a', { href: '#/signin?next=%2Fforum' }, 'Sign in'),
          ' to start a thread or reply.',
        ),
    ...(queue ? [queue] : []),
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
            ` · ${formatDate(p.createdAt)} · `,
            reportControl(id, p.id),
          ),
          h('div', { class: 'post-body' }, p.body),
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
  once(send, 'Posting…', async () => {
    if (!isSignedIn()) {
      navigate(`/signin?next=${encodeURIComponent(`/forum/${id}`)}`);
      return;
    }
    if (!reply.value.trim()) {
      toast('Write something first.', 'err');
      return;
    }
    await api.addPost(id, reply.value);
    const d = await api.thread(id);
    reply.value = '';
    paint(d.posts);
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
        `${data.thread.author} · ${formatDate(data.thread.createdAt)} · `,
        reportControl(id),
      ),
      h('div', { class: 'post-body' }, data.thread.body),
    ),
    posts,
    h(
      'div',
      { class: 'card stack' },
      h('h2', {}, 'Reply'),
      h('p', { class: 'small muted' }, 'Plain text only.'),
      reply,
      counted(reply, 5000),
      send,
    ),
  );
}
