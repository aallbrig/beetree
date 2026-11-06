import { api } from './api';
import { currentUsername } from './auth';

/**
 * The sign-up form records the newsletter choice here because the account does
 * not exist (and cannot take an API call) until the email is confirmed. The
 * first sign-in saves it to the profile and clears it.
 */
const KEY = 'btc.pending-consent';

export function rememberConsent(username: string, optIn: boolean): void {
  try {
    localStorage.setItem(KEY, JSON.stringify({ username: username.toLowerCase(), optIn }));
  } catch {
    /* storage blocked: the profile page still offers the choice */
  }
}

export async function applyPendingConsent(): Promise<void> {
  let raw: string | null;
  try {
    raw = localStorage.getItem(KEY);
  } catch {
    return;
  }
  if (!raw) return;
  const me = currentUsername().toLowerCase();
  try {
    const p = JSON.parse(raw) as { username?: string; optIn?: boolean };
    if (!me || p.username !== me) return;
    if (p.optIn) {
      const profile = await api.me();
      await api.updateMe(profile.displayName, profile.bio, {
        marketingOptIn: true,
        marketingSource: 'signup',
      });
    }
    localStorage.removeItem(KEY);
  } catch {
    /* try again at the next sign-in */
  }
}
