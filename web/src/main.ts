import './style.css';
import { applyTheme, currentTheme, toggleTheme } from './theme';
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

window.addEventListener('unhandledrejection', (e: PromiseRejectionEvent) => {
  console.error('unhandled rejection', e.reason);
});
window.addEventListener('error', (e: ErrorEvent) => {
  console.error('uncaught', e.message);
});

applyTheme(currentTheme());

const app = document.getElementById('app');
if (!app) throw new Error('missing #app');

let menuOpen = false;
const nav = h('nav', { class: 'nav', 'aria-label': 'Main' });
const main = h('main', { id: 'main' });
app.append(nav, main);

function renderNav(): void {
  clear(nav);
  const here = location.hash.replace(/^#/, '').split('?')[0] ?? '';
  const l = (href: string, text: string): HTMLAnchorElement =>
    h('a', { class: `link${here.startsWith(href) ? ' active' : ''}`, href: `#${href}` }, text);
  const themeBtn = h('button', {
    class: 'btn small theme-toggle',
    type: 'button',
    onClick: () => {
      toggleTheme();
      renderNav();
    },
  });
  const dark = currentTheme() === 'dark';
  themeBtn.textContent = dark ? 'Light mode' : 'Dark mode';
  themeBtn.setAttribute('aria-label', dark ? 'Switch to light mode' : 'Switch to dark mode');

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
  menuBtn.textContent = menuOpen ? 'Close' : 'Menu';

  const links = h(
    'div',
    { class: `links${menuOpen ? ' open' : ''}`, id: 'nav-links' },
    l('/build', 'Build'),
    l('/explore', 'Explore'),
    l('/community', 'Community'),
    l('/forum', 'Forum'),
    ...(isSignedIn() ? [l('/dashboard', 'My trees')] : []),
    h('a', { class: 'link', href: '/docs/' }, 'Docs'),
    h('span', { class: 'spacer' }),
    themeBtn,
    ...(isSignedIn()
      ? [
          l('/profile', currentUsername() || 'Profile'),
          h(
            'button',
            {
              class: 'btn small',
              onClick: () => {
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
}

setGuard((path) => (isSignedIn() ? undefined : `/signin?next=${encodeURIComponent(path)}`));

route('/build', buildView);
route('/explore', exploreView, { wide: true });
route('/community', communityView);
route('/tree/([\\w-]+)', treeView);
route('/forum', forumView);
route('/forum/([\\w-]+)', threadView);
route('/user/([\\w-]+)', userView);
route('/signin', signInView);
route('/signup', signUpView);
route('/confirm', confirmView);
route('/forgot', forgotView);
route('/dashboard', dashboardView, { private: true });
route('/builder', builderView, { private: true, wide: true });
route('/builder/([\\w-]+)', builderView, { private: true, wide: true });
route('/scenes', scenesView, { private: true, wide: true });
route('/scenes/([\\w-]+)', scenesView, { private: true, wide: true });
route('/profile', profileView, { private: true });

onAuthChange(() => {
  renderNav();
  void render();
});
window.addEventListener('hashchange', () => {
  menuOpen = false;
  renderNav();
});
renderNav();
start(main);
