import './style.css';
import { currentTheme, showTheme, toggleTheme } from './theme';
import { clearDraft } from './draft';
import { applyPendingConsent } from './consent';
import { currentUsername, isSignedIn, onAuthChange, signOut } from './auth';
import { clear, h } from './dom';
import { navigate, render, route, setGuard, start } from './router';
import { confirmView, forgotView, signInView, signUpView } from './views/auth';
import { communityView } from './views/community';
import { dashboardView } from './views/dashboard';
import { builderView } from './views/builder';
import { exploreView } from './views/explore';
import { forumView, threadView } from './views/forum';
import { profileView, userView } from './views/profile';
import { buildView } from './views/build';
import { scenesView } from './views/scenes';
import { treeView } from './views/tree';
import { REPO_URL, repoHref } from './repo';

window.addEventListener('unhandledrejection', (e: PromiseRejectionEvent) => {
  console.error('unhandled rejection', e.reason);
});
window.addEventListener('error', (e: ErrorEvent) => {
  console.error('uncaught', e.message);
});

showTheme(currentTheme());

const app = document.getElementById('app');
if (!app) throw new Error('missing #app');

let menuOpen = false;
let openGroup = '';
const narrow = window.matchMedia('(max-width: 800px)');
const nav = h('nav', { class: 'nav', 'aria-label': 'Main' });
const main = h('main', { id: 'main' });
const version = h('small', { class: 'ver' });
const themeBtn = h('button', { class: 'linklike', type: 'button' });
function paintTheme(): void {
  const dark = currentTheme() === 'dark';
  themeBtn.textContent = dark ? 'Light mode' : 'Dark mode';
  themeBtn.setAttribute('aria-label', dark ? 'Switch to light mode' : 'Switch to dark mode');
}
themeBtn.addEventListener('click', () => {
  toggleTheme();
  paintTheme();
});
paintTheme();
const foot = h(
  'footer',
  { class: 'foot' },
  h(
    'p',
    {},
    'BeeTreeCraft · ',
    ...(repoHref(REPO_URL)
      ? [h('a', { href: repoHref(REPO_URL) ?? '' }, 'Source on GitHub'), ' · ']
      : []),
    h('a', { href: '/references/' }, 'References and credits'),
    ' · ',
    h('a', { href: '/sitemap.xml' }, 'Sitemap'),
    ' ',
    version,
    ' · ',
    themeBtn,
  ),
);
app.append(nav, main, foot);
void fetch('/version.json')
  .then((r) => r.json() as Promise<{ version?: string }>)
  .then((v) => {
    if (!v.version) return;
    version.textContent = '· ';
    const tag = repoHref(REPO_URL, `releases/tag/${v.version}`);
    version.append(tag ? h('a', { href: tag }, v.version) : v.version);
  })
  .catch(() => undefined);

function renderNav(): void {
  const scrolled = nav.scrollTop;
  document.body.classList.toggle('menu-open', menuOpen);
  clear(nav);
  const here = location.hash.replace(/^#/, '').split('?')[0] ?? '';
  const l = (href: string, text: string): HTMLAnchorElement =>
    h('a', { class: `link${here.startsWith(href) ? ' active' : ''}`, href: `#${href}` }, text);
  const doc = (href: string, text: string): HTMLAnchorElement =>
    h('a', { class: 'link', href }, text);
  const group = (label: string, items: HTMLAnchorElement[]): HTMLDivElement =>
    h(
      'div',
      { class: `grp${openGroup === label ? ' open' : ''}` },
      h(
        'button',
        {
          class: 'grp-label',
          type: 'button',
          'aria-haspopup': 'true',
          'aria-expanded': String(openGroup === label),
          onClick: () => {
            if (!narrow.matches) return;
            openGroup = openGroup === label ? '' : label;
            renderNav();
          },
        },
        label,
      ),
      h('div', { class: 'grp-menu' }, ...items),
    );
  const menuBtn = h('button', {
    class: 'btn small menu-toggle',
    type: 'button',
    'aria-controls': 'nav-links',
    'aria-expanded': String(menuOpen),
    onClick: () => {
      menuOpen = !menuOpen;
      renderNav();
    },
  });
  menuBtn.setAttribute('aria-label', menuOpen ? 'Close menu' : 'Menu');
  menuBtn.append(h('span', { class: 'bars', 'aria-hidden': 'true' }));

  const links = h(
    'div',
    { class: `links${menuOpen ? ' open' : ''}`, id: 'nav-links' },
    l('/build', 'Build'),
    l('/explore', 'Explore'),
    group('Docs', [
      doc('/docs/', 'Overview'),
      doc('/docs/manual/', 'Website manual'),
      doc('/docs/trees/', 'The tree format'),
      doc('/docs/cli/', 'CLI'),
      doc('/docs/api/', 'REST API'),
      doc('/docs/api/reference/', 'API explorer (Swagger)'),
      doc('/docs/mcp/', 'MCP server'),
      doc('/learn/behavior-trees/', 'Behavior trees & history'),
      doc('/learn/building-blocks/', 'Building blocks'),
      doc('/resources/', 'Resources'),
      doc('/references/', 'References and credits'),
    ]),
    group('Community', [l('/community', 'Community trees'), l('/forum', 'Forum')]),
    ...(isSignedIn() ? [l('/dashboard', 'My trees')] : []),
    h('span', { class: 'spacer' }),
    ...(isSignedIn()
      ? [
          l('/profile', currentUsername() || 'Profile'),
          h(
            'button',
            {
              class: 'btn small',
              onClick: () => {
                clearDraft();
                signOut();
                navigate('/explore');
              },
            },
            'Sign out',
          ),
        ]
      : [
          l('/signin', 'Sign in'),
          h('a', { class: 'btn primary small', href: '#/signup' }, 'Sign up'),
        ]),
  );
  nav.append(h('a', { class: 'brand', href: '/' }, 'BeeTreeCraft'), menuBtn, links);
  nav.scrollTop = scrolled;
}

setGuard((path) => (isSignedIn() ? undefined : `/signin?next=${encodeURIComponent(path)}`));

route('/build', buildView);
route('/explore', exploreView, { wide: true });
route('/community', communityView);
route('/tree/([\\w-]+)', treeView, { private: true });
route('/forum', forumView, { private: true });
route('/forum/([\\w-]+)', threadView, { private: true });
route('/user/([\\w-]+)', userView);
route('/signin', signInView);
route('/signup', signUpView);
route('/confirm', confirmView);
route('/forgot', forgotView);
route('/dashboard', dashboardView, { private: true });
// scratch mode: anyone can build; saving, validating and codegen ask for sign-in
route('/builder', builderView, { wide: true });
route('/builder/([\\w-]+)', builderView, { private: true, wide: true });
route('/scenes', scenesView, { private: true, wide: true });
route('/scenes/([\\w-]+)', scenesView, { private: true, wide: true });
route('/profile', profileView, { private: true });

onAuthChange(() => {
  if (isSignedIn()) void applyPendingConsent();
  renderNav();
  void render();
});
window.addEventListener('keydown', (e) => {
  if (e.key === 'Escape' && menuOpen) {
    menuOpen = false;
    renderNav();
  }
});
window.addEventListener('hashchange', () => {
  menuOpen = false;
  renderNav();
});
renderNav();
start(main);
