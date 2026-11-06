export type Theme = 'light' | 'dark';

const KEY = 'beetreecraft-theme';

/** The OS preference, used until the visitor chooses. */
function systemTheme(): Theme {
  try {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  } catch {
    return 'light';
  }
}

/** The saved choice, or the OS preference, or light. */
export function currentTheme(): Theme {
  try {
    const saved = localStorage.getItem(KEY);
    if (saved === 'dark' || saved === 'light') return saved;
  } catch {
    // storage blocked: fall through to the OS preference
  }
  return systemTheme();
}

/** Show a theme without saving it, so the OS preference keeps applying. */
export function showTheme(theme: Theme): void {
  document.documentElement.dataset.theme = theme;
}

export function applyTheme(theme: Theme): void {
  showTheme(theme);
  try {
    localStorage.setItem(KEY, theme);
  } catch {
    // storage blocked: the choice lasts for this page only
  }
}

export function toggleTheme(): Theme {
  const next: Theme = currentTheme() === 'dark' ? 'light' : 'dark';
  applyTheme(next);
  return next;
}
