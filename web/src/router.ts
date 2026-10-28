import { clear } from './dom';

export type Cleanup = () => void;
export interface Ctx {
  root: HTMLElement;
  params: string[];
  query: URLSearchParams;
}
/** A view renders into ctx.root and may return a cleanup function (or nothing). */
export type View = (ctx: Ctx) => unknown;

interface Route {
  pattern: RegExp;
  view: View;
  /** requires a signed-in user */
  private?: boolean;
  wide?: boolean;
}

const isCleanup = (v: unknown): v is Cleanup => typeof v === 'function';

const routes: Route[] = [];
let cleanup: Cleanup | undefined;
let guard: ((path: string) => string | undefined) | undefined;
let outlet: HTMLElement;

export function route(
  pattern: string,
  view: View,
  opts: { private?: boolean; wide?: boolean } = {},
): void {
  routes.push({ pattern: new RegExp(`^${pattern}$`), view, ...opts });
}

export function setGuard(fn: (path: string) => string | undefined): void {
  guard = fn;
}

export function navigate(path: string): void {
  if (location.hash === `#${path}`) void render();
  else location.hash = path;
}

function parse(): { path: string; query: URLSearchParams } {
  const raw = location.hash.replace(/^#/, '') || '/explore';
  const [path = '/explore', qs = ''] = raw.split('?');
  return { path, query: new URLSearchParams(qs) };
}

export async function render(): Promise<void> {
  cleanup?.();
  cleanup = undefined;
  const { path, query } = parse();
  const redirect = guard?.(path);
  const match = routes.map((r) => ({ r, m: r.pattern.exec(path) })).find((x) => x.m !== null);
  if (!match?.m) {
    clear(outlet);
    outlet.textContent = 'Page not found.';
    return;
  }
  if (match.r.private && redirect) {
    navigate(redirect);
    return;
  }
  outlet.className = match.r.wide ? 'wide' : '';
  clear(outlet);
  try {
    const result = await match.r.view({ root: outlet, params: match.m.slice(1), query });
    cleanup = isCleanup(result) ? result : undefined;
  } catch (e) {
    clear(outlet);
    outlet.textContent = e instanceof Error ? e.message : 'Something went wrong.';
  }
  window.scrollTo(0, 0);
}

export function start(el: HTMLElement): void {
  outlet = el;
  window.addEventListener('hashchange', () => void render());
  void render();
}
