/**
 * The one place the app learns the repository URL. Keep it equal to
 * `params.repoUrl` in website/hugo.toml (scripts/check_site.sh enforces that).
 * An empty value removes every repository link from the app.
 */
export const REPO_URL = 'https://github.com/aallbrig/beetreecraft';

/** A repository link target, or undefined when no repository is configured. */
export function repoHref(base: string, path = ''): string | undefined {
  if (base === '') return undefined;
  return path === '' ? base : `${base}/${path}`;
}
