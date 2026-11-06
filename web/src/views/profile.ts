import { api, ApiError, type ApiKeyInfo } from '../api';
import { AVATAR_IDS, avatarSrc } from '../avatar';
import { getToken } from '../auth';
import { treeCard } from '../components/cards';
import { formatDate, h, toast } from '../dom';
import {
  EXPIRIES,
  PRESETS,
  RESOURCES,
  expiryText,
  keyBody,
  lastUsedText,
  scopeChips,
  scopeFor,
  type Expiry,
  type Preset,
} from '../keyscopes';
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
  let avatar = me.avatar ?? '';
  const preview = h('img', {
    class: 'avatar big',
    alt: 'Your profile picture',
    width: '64',
    height: '64',
    src: avatarSrc(avatar || 'a001'),
  });
  const picker = h('div', {
    class: 'avatar-grid',
    role: 'radiogroup',
    'aria-label': 'Profile picture',
  });
  for (const id of AVATAR_IDS) {
    const b = h(
      'button',
      {
        class: `avatar-opt${id === avatar ? ' on' : ''}`,
        type: 'button',
        role: 'radio',
        'aria-checked': String(id === avatar),
        'aria-label': `Picture ${id.slice(1)}`,
      },
      h('img', { src: avatarSrc(id), alt: '', width: '40', height: '40', loading: 'lazy' }),
    );
    b.addEventListener('click', () => {
      avatar = id;
      preview.src = avatarSrc(id);
      for (const o of picker.children) {
        o.classList.remove('on');
        o.setAttribute('aria-checked', 'false');
      }
      b.classList.add('on');
      b.setAttribute('aria-checked', 'true');
    });
    picker.append(b);
  }
  const news = h('input', { type: 'checkbox', id: 'news', checked: me.marketingOptIn === true });
  const save = h('button', { class: 'btn primary', type: 'button' }, 'Save profile');
  save.addEventListener('click', () => {
    api.updateMe(name.value.trim(), bio.value, { avatar, marketingOptIn: news.checked }).then(
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
  const keyName = h('input', {
    type: 'text',
    maxlength: '60',
    placeholder: 'Key name, e.g. laptop',
    'aria-label': 'API key name',
  });
  const keyList = h('ul', { class: 'keylist' });
  const fresh = h('div', { class: 'note hide', role: 'status' });
  const renderKeys = (keys: ApiKeyInfo[]): void => {
    keyList.replaceChildren(
      ...(keys.length === 0
        ? [h('li', { class: 'muted small' }, 'No keys yet.')]
        : keys.map((k) =>
            h(
              'li',
              { class: 'row' },
              h('span', {}, `${k.name} `),
              h('code', {}, `${k.prefix}…`),
              h('span', { class: 'small muted' }, ` created ${formatDate(k.createdAt)}`),
              h(
                'div',
                { class: 'chips' },
                ...scopeChips(k.scopes ?? []).map((c) => h('span', { class: 'chip' }, c)),
                k.legacy ? h('span', { class: 'chip' }, 'made before scopes') : null,
              ),
              h(
                'div',
                { class: 'small muted' },
                `${expiryText(k.expiresAt)} · ${lastUsedText(k.lastUsedAt)}`,
              ),
              h(
                'button',
                {
                  class: 'btn small danger',
                  type: 'button',
                  onClick: () => {
                    api
                      .deleteKey(k.id)
                      .then(refreshKeys)
                      .catch(() => {
                        toast('Could not revoke the key.', 'err');
                      });
                  },
                },
                'Revoke',
              ),
            ),
          )),
    );
  };
  function refreshKeys(): Promise<void> {
    return api.listKeys().then((r) => {
      renderKeys(r.items);
    });
  }
  const ticked = new Set<string>();
  let choice: Preset | 'custom' | '' = '';
  const grid = h('div', { class: 'scopegrid hide' });
  for (const [res, actions] of Object.entries(RESOURCES)) {
    grid.append(
      h(
        'div',
        { class: 'row' },
        h('strong', {}, res),
        ...actions.map((a) => {
          const box = h('input', { type: 'checkbox', 'aria-label': `${res} ${a}` });
          box.addEventListener('change', () => {
            if (box.checked) ticked.add(scopeFor(res, a));
            else ticked.delete(scopeFor(res, a));
          });
          return h('label', { class: 'small' }, box, ` ${a}`);
        }),
      ),
    );
  }
  const choices: [Preset | 'custom', string, string][] = [
    ...(Object.entries(PRESETS) as [Preset, (typeof PRESETS)[Preset]][]).map(
      ([id, p]): [Preset, string, string] => [id, p.label, p.help],
    ),
    ['custom', 'Custom', 'Pick each resource and action. A higher action includes the lower ones.'],
  ];
  const choiceBox = h(
    'fieldset',
    { class: 'scopechoice' },
    h('legend', { class: 'small' }, 'What may this key do?'),
    ...choices.map(([id, label, help]) => {
      const radio = h('input', { type: 'radio', name: 'keypreset', value: id });
      radio.addEventListener('change', () => {
        choice = id;
        grid.classList.toggle('hide', id !== 'custom');
      });
      return h('label', { class: 'small' }, radio, ` ${label}: ${help}`);
    }),
  );
  const expiry = h(
    'select',
    { 'aria-label': 'Key expiry' },
    ...EXPIRIES.map((e) => h('option', { value: e.value, selected: e.value === '90d' }, e.label)),
  );
  const makeKey = h('button', { class: 'btn primary', type: 'button' }, 'Create API key');
  makeKey.addEventListener('click', () => {
    const body = keyBody(keyName.value, choice, ticked, expiry.value as Expiry);
    if (!body) {
      toast('Name the key and pick what it may do.', 'err');
      return;
    }
    api
      .createKey(body)
      .then(async (k) => {
        keyName.value = '';
        fresh.classList.remove('hide');
        fresh.replaceChildren(
          h('strong', {}, 'Copy this key now. It is shown only once.'),
          h('pre', { class: 'code' }, k.key),
        );
        await refreshKeys();
      })
      .catch((e: unknown) => {
        toast(e instanceof ApiError ? e.message : 'Could not create the key.', 'err');
      });
  });
  void refreshKeys();
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
        h(
          'div',
          {},
          h('label', {}, 'Profile picture'),
          h(
            'div',
            { class: 'row' },
            preview,
            h('span', { class: 'small muted' }, 'Pick one of 100.'),
          ),
          picker,
        ),
        h('div', {}, h('label', {}, 'Email'), h('div', {}, me.email ?? '')),
        h(
          'div',
          { class: 'field row' },
          news,
          h(
            'label',
            { for: 'news', style: 'margin:0;font-weight:400' },
            'Email me BeeTreeCraft news, new features and updates. Untick to unsubscribe.',
          ),
        ),
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
          h('h2', {}, 'API keys'),
          h(
            'p',
            { class: 'small muted' },
            'The REST API and the MCP server need a key. Send it as X-API-Key or as a bearer token. Keep it secret; revoke it here if it leaks.',
          ),
          h('div', { class: 'row' }, keyName),
          choiceBox,
          grid,
          h('div', { class: 'row' }, h('label', { class: 'small' }, 'Expires '), expiry, makeKey),
          h(
            'p',
            { class: 'small muted' },
            'Scopes cannot be edited after a key is made. To change them, create a new key and revoke the old one.',
          ),
          fresh,
          keyList,
          h('h3', {}, 'Short-lived token'),
          h('p', { class: 'small muted' }, 'A sign-in token also works and lasts about an hour.'),
          copy,
          h(
            'div',
            { class: 'doclinks' },
            h('a', { class: 'btn small', href: '/docs/api/' }, 'REST API docs'),
            h('a', { class: 'btn small', href: '/docs/api/reference/' }, 'API explorer'),
            h('a', { class: 'btn small', href: '/docs/mcp/' }, 'MCP docs'),
          ),
        ),
      ),
    ),
  );
}

export async function userView(ctx: Ctx): Promise<void> {
  const { profile, trees, scenes = [] } = await api.user(ctx.params[0] ?? '');
  ctx.root.append(
    h('img', {
      class: 'avatar big',
      alt: `${profile.displayName}'s profile picture`,
      width: '64',
      height: '64',
      src: avatarSrc(profile.avatar ?? 'a001'),
    }),
    h('h1', {}, profile.displayName),
    h('p', { class: 'muted' }, profile.bio || 'No bio yet.'),
    h('h2', {}, 'Shared trees'),
    trees.length
      ? h('div', { class: 'grid' }, ...trees.map(treeCard))
      : h('p', { class: 'muted' }, 'Nothing shared yet.'),
    h('h2', {}, 'Shared scenes'),
    scenes.length
      ? h(
          'ul',
          {},
          ...scenes.map((s) =>
            h('li', {}, `${s.name} (${s.agents.length} agents, ${formatDate(s.createdAt)})`),
          ),
        )
      : h('p', { class: 'muted' }, 'No scenes shared yet.'),
  );
}
