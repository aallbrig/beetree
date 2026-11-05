import './communitree.css';
import { h } from './dom';

/** Shared with website/static/js/communitree.js: same key, same shape. */
export const STORAGE_KEY = 'beetreecraft-communitree';
const SEEN_KEY = 'beetreecraft-communitree-dismissed-signup';

export interface Signup {
  email: string;
  joinedAt: string;
  source: string;
}

const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const isValidEmail = (v: string): boolean => EMAIL.test(v.trim());

function isSignup(v: unknown): v is Signup {
  return (
    typeof v === 'object' && v !== null && typeof (v as Record<string, unknown>).email === 'string'
  );
}

export function readSignups(): Signup[] {
  try {
    const raw: unknown = JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]');
    return Array.isArray(raw) ? raw.filter(isSignup) : [];
  } catch {
    return [];
  }
}

/** Appends the email unless already present (case-insensitive). Returns true if added. */
export function addSignup(email: string, source: string, now = new Date()): boolean {
  const lower = email.trim().toLowerCase();
  const list = readSignups();
  if (list.some((s) => s.email === lower)) return false;
  list.push({ email: lower, joinedAt: now.toISOString(), source });
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(list));
  } catch {
    return false;
  }
  return true;
}

/** Shows the interstitial unless the visitor already joined or closed it this session. */
export function maybeShowCommunitree(source: string, delayMs = 800): () => void {
  let seen = false;
  try {
    seen = sessionStorage.getItem(SEEN_KEY) === '1';
  } catch {
    /* storage blocked */
  }
  if (seen || readSignups().length > 0) return () => undefined;

  let back: HTMLElement | undefined;
  const timer = setTimeout(() => {
    back = open(source);
  }, delayMs);
  return () => {
    clearTimeout(timer);
    back?.remove();
  };
}

function open(source: string): HTMLElement {
  const prev = document.activeElement;
  const input = h('input', {
    type: 'email',
    required: true,
    autocomplete: 'email',
    placeholder: 'you@example.com',
    'aria-label': 'Email address',
  });
  const err = h('p', { class: 'ct-err', role: 'alert' });
  const modal = h('div', {
    class: 'ct-modal',
    role: 'dialog',
    'aria-modal': 'true',
    'aria-labelledby': 'ct-title',
  });
  const back = h('div', { class: 'ct-backdrop' }, modal);

  const onKey = (e: KeyboardEvent): void => {
    if (e.key === 'Escape') done();
  };
  function done(): void {
    try {
      sessionStorage.setItem(SEEN_KEY, '1');
    } catch {
      /* storage blocked */
    }
    document.removeEventListener('keydown', onKey);
    back.remove();
    if (prev instanceof HTMLElement) prev.focus();
  }

  const form = h(
    'form',
    {
      novalidate: true,
      onSubmit: (e: Event) => {
        e.preventDefault();
        if (!isValidEmail(input.value)) {
          err.textContent = 'That email looks a little off.';
          return;
        }
        addSignup(input.value, source);
        const ok = h(
          'button',
          { class: 'btn primary ct-go', type: 'button', onClick: done },
          'Close',
        );
        modal.replaceChildren(
          h('div', { class: 'ct-emoji' }, '\u{1F331}'),
          h('h2', {}, "You're in the communitree!"),
          h('p', {}, 'Thanks for joining. We will be in touch.'),
          ok,
        );
        ok.focus();
      },
    },
    input,
    err,
    h('button', { class: 'btn primary ct-go', type: 'submit' }, 'Count me in'),
  );

  modal.append(
    h('button', { class: 'ct-close', type: 'button', 'aria-label': 'Close', onClick: done }, '×'),
    h('div', { class: 'ct-emoji' }, '\u{1F333}'),
    h('h2', { id: 'ct-title' }, 'Join the communitree'),
    h(
      'p',
      {},
      'Branch out with us. Get new behavior tree patterns, features and community picks, straight to your inbox.',
    ),
    form,
    h('button', { class: 'ct-skip', type: 'button', onClick: done }, 'Maybe later'),
    h('p', { class: 'ct-fine' }, 'Stored only in this browser for now. No spam, ever.'),
  );
  back.addEventListener('click', (e) => {
    if (e.target === back) done();
  });
  document.addEventListener('keydown', onKey);
  document.body.append(back);
  input.focus();
  return back;
}
