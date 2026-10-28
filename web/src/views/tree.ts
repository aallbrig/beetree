import { api, ApiError } from '../api';
import { currentUsername, isSignedIn } from '../auth';
import { outline, type Outline } from '../components/outline';
import { clear, formatDate, h, toast } from '../dom';
import { navigate, type Cleanup, type Ctx } from '../router';
import { World } from '../scene/world';
import { parseSpec } from '../spec';
import type { Comment } from '../types';

export async function treeView(ctx: Ctx): Promise<Cleanup | undefined> {
  const id = ctx.params[0] ?? '';
  const tree = await api.getTree(id);
  const spec = tree.spec ? parseSpec(tree.spec) : undefined;
  const mine = isSignedIn() && (await api.me().catch(() => undefined))?.userId === tree.ownerId;

  const likeBtn = h('button', {
    class: 'btn',
    type: 'button',
    'aria-pressed': String(tree.likedByMe === true),
  });
  let liked = tree.likedByMe === true;
  let likes = tree.likeCount;
  const paintLike = (): void => {
    likeBtn.textContent = `${liked ? 'Liked' : 'Like'} (${likes})`;
    likeBtn.setAttribute('aria-pressed', String(liked));
  };
  paintLike();
  likeBtn.addEventListener('click', () => {
    if (!isSignedIn()) {
      navigate(`/signin?next=${encodeURIComponent(`/tree/${id}`)}`);
      return;
    }
    (liked ? api.unlike(id) : api.like(id)).then(
      (r) => {
        liked = r.liked;
        likes = r.likeCount;
        paintLike();
      },
      () => {
        toast('Could not update your like.', 'err');
      },
    );
  });

  const canvas = h('canvas', { 'aria-label': '3D preview of this tree running' });
  const stage = h(
    'div',
    { class: 'stage' },
    canvas,
    h('div', { class: 'hintbar' }, 'Click the ground to move the intruder.'),
  );
  let ol: Outline | undefined;
  const outlineHost = h('div');
  if (spec) {
    ol = outline(spec.tree);
    outlineHost.append(ol.el);
  }

  const commentHeading = h('h2', {}, `Comments (${tree.commentCount})`);
  const commentList = h('div');
  const paintComments = (items: Comment[]): void => {
    commentHeading.textContent = `Comments (${items.length})`;
    clear(commentList);
    if (!items.length) commentList.append(h('p', { class: 'muted' }, 'No comments yet.'));
    for (const c of items) {
      const del = h('button', { class: 'btn small danger', type: 'button' }, 'Delete');
      del.addEventListener('click', () => {
        api.deleteComment(id, c.id).then(loadComments, () => {
          toast('Could not delete.', 'err');
        });
      });
      commentList.append(
        h(
          'div',
          { class: 'comment' },
          h(
            'div',
            { class: 'small muted' },
            h('a', { href: `#/user/${c.authorId}` }, c.author),
            ` · ${formatDate(c.createdAt)} `,
            mine || c.author === currentUsername() ? del : null,
          ),
          h('div', {}, c.body),
        ),
      );
    }
  };
  const loadComments = (): Promise<void> =>
    api.comments(id).then(
      (r) => {
        paintComments(r.items);
      },
      () => {
        commentList.textContent = 'Could not load comments.';
      },
    );

  const box = h('textarea', {
    'aria-label': 'Add a comment',
    maxlength: '2000',
    placeholder: 'Say something kind and useful…',
    style: 'min-height:80px',
  });
  const post = h('button', { class: 'btn primary', type: 'button' }, 'Post comment');
  post.addEventListener('click', () => {
    if (!isSignedIn()) {
      navigate(`/signin?next=${encodeURIComponent(`/tree/${id}`)}`);
      return;
    }
    if (!box.value.trim()) return;
    api.addComment(id, box.value).then(
      () => {
        box.value = '';
        return loadComments();
      },
      (e: unknown) => {
        toast(e instanceof ApiError ? e.message : 'Could not post.', 'err');
      },
    );
  });

  ctx.root.append(
    h(
      'div',
      { class: 'studio' },
      h(
        'div',
        { class: 'stack' },
        h(
          'div',
          { class: 'card stack' },
          h('h1', {}, tree.name),
          h(
            'div',
            { class: 'small muted' },
            'by ',
            h('a', { href: `#/user/${tree.ownerId}` }, tree.ownerName),
            ` · ${formatDate(tree.createdAt)}${tree.visibility === 'private' ? ' · private' : ''}`,
          ),
          h('p', {}, tree.description),
          h('div', {}, ...tree.tags.map((x) => h('span', { class: 'tag' }, x))),
          h(
            'div',
            { class: 'row' },
            likeBtn,
            isSignedIn()
              ? h(
                  'a',
                  { class: 'btn', href: mine ? `#/builder/${id}` : `#/builder?fork=${id}` },
                  mine ? 'Edit' : 'Fork in builder',
                )
              : h('a', { class: 'btn', href: '#/signup' }, 'Sign up to fork'),
          ),
        ),
        h('div', { class: 'card' }, h('h2', {}, 'Structure'), outlineHost),
        spec
          ? h(
              'details',
              { class: 'card' },
              h('summary', {}, 'View YAML'),
              h('pre', { class: 'code' }, tree.spec ?? ''),
            )
          : null,
        h('div', { class: 'card stack' }, commentHeading, commentList, box, post),
      ),
      stage,
    ),
  );
  void loadComments();

  let world: World | undefined;
  let timer: ReturnType<typeof setInterval> | undefined;
  if (spec) {
    try {
      world = await World.create(canvas);
      world.setBeacon({ x: 5, z: -1 });
      world.start();
      const agent = await world.addAgent({
        id: 'a',
        name: tree.name,
        model: 'character-e',
        x: 0,
        z: 0,
        spec: spec.tree,
      });
      timer = setInterval(() => {
        ol?.setStatuses(agent.runtime.statuses);
      }, 120);
    } catch {
      clear(stage);
      stage.append(h('p', { style: 'padding:24px' }, '3D preview unavailable in this browser.'));
    }
  }
  return () => {
    clearInterval(timer);
    world?.dispose();
  };
}
