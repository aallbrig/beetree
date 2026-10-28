import './style.css';
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
import { scenesView } from './views/scenes';
import { treeView } from './views/tree';

window.addEventListener('unhandledrejection', (e: PromiseRejectionEvent) => {
  console.error('unhandled rejection', e.reason);
});
window.addEventListener('error', (e: ErrorEvent) => {
  console.error('uncaught', e.message);
});

const app = document.getElementById('app');
if (!app) throw new Error('missing #app');

const nav = h('nav', { class: 'nav', 'aria-label': 'Main' });
const main = h('main', { id: 'main' });
app.append(nav, main);

function renderNav(): void {
  clear(nav);
  const here = location.hash.replace(/^#/, '').split('?')[0] ?? '';
  const l = (href: string, text: string): HTMLAnchorElement =>
    h('a', { class: `link${here.startsWith(href) ? ' active' : ''}`, href: `#${href}` }, text);
  nav.append(
    h('a', { class: 'brand', href: '/' }, '🐝 BeeTreeCraft'),
    l('/explore', 'Explore'),
    l('/community', 'Community'),
    l('/forum', 'Forum'),
    ...(isSignedIn() ? [l('/dashboard', 'Dashboard'), l('/scenes', '3D Scenes')] : []),
    h('a', { class: 'link', href: '/docs/' }, 'Docs'),
    h('span', { class: 'spacer' }),
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
}

setGuard((path) => (isSignedIn() ? undefined : `/signin?next=${encodeURIComponent(path)}`));

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
window.addEventListener('hashchange', renderNav);
renderNav();
start(main);
