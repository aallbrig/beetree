import type { Status } from '../engine/bt';
import { pathId } from '../engine/bt';
import { h } from '../dom';
import { countNodes, type NodeSpec, type Path, type Problem } from '../spec';

const ICONS: Record<string, string> = {
  action: '▶',
  condition: '?',
  sequence: '→',
  selector: '⑂',
  parallel: '⇉',
  decorator: '◇',
};

export interface Outline {
  el: HTMLElement;
  setStatuses(statuses: ReadonlyMap<string, Status>): void;
  setSelected(path: Path | undefined): void;
  /** Marks nodes on the debugger's visited path, and the leaf it is paused on. */
  setDebug(visited: ReadonlySet<string>, waiting: string | undefined): void;
  setProblems(problems: readonly Problem[]): void;
}

/** A keyboard-accessible, collapsible tree outline with live status dots. */
export interface OutlineView {
  /** Ids of branches shown folded; the set is edited in place by the toggle. */
  collapsed: Set<string>;
  /** Ids to highlight as search matches. */
  matches?: ReadonlySet<string>;
  onToggle?: () => void;
}

export function outline(
  root: NodeSpec,
  onSelect?: (path: Path) => void,
  view?: OutlineView,
): Outline {
  const buttons = new Map<string, HTMLButtonElement>();

  const build = (node: NodeSpec, path: Path): HTMLLIElement => {
    const id = pathId(path);
    const label =
      node.type === 'decorator' ? `${node.name} (${node.decorator ?? 'repeat'})` : node.name;
    const btn = h(
      'button',
      { class: 'node', type: 'button', 'data-id': id, title: node.description ?? node.type },
      h('span', { class: 'icon', 'aria-hidden': 'true' }, ICONS[node.type] ?? '•'),
      h('span', {}, label),
      h('span', { class: 'hint' }, node.node ? `· ${node.node}` : ''),
      h('span', { class: 'dot', 'aria-hidden': 'true' }),
    );
    if (onSelect)
      btn.addEventListener('click', () => {
        onSelect(path);
      });
    buttons.set(id, btn);
    if (view?.matches?.has(id)) btn.classList.add('match');
    const kids = node.children ?? [];
    let row: HTMLElement = btn;
    const folded = view?.collapsed.has(id) ?? false;
    if (view && kids.length > 0) {
      btn.setAttribute('aria-expanded', String(!folded));
      const toggle = h(
        'button',
        {
          class: 'fold',
          type: 'button',
          'aria-label': `${folded ? 'Expand' : 'Collapse'} ${node.name}`,
        },
        folded ? `+${countNodes(node) - 1}` : '−',
      );
      toggle.addEventListener('click', () => {
        if (folded) view.collapsed.delete(id);
        else view.collapsed.add(id);
        view.onToggle?.();
      });
      row = h('div', { class: 'row' }, toggle, btn);
    }
    const li = h('li', {}, row);
    if (kids.length > 0 && !folded) {
      li.append(h('ul', { class: 'outline' }, ...kids.map((c, i) => build(c, [...path, i]))));
    }
    return li;
  };

  const el = h('ul', { class: 'outline root', role: 'tree' }, build(root, []));
  return {
    el,
    setStatuses(statuses) {
      for (const [id, btn] of buttons) {
        const s = statuses.get(id);
        if (s) btn.dataset.status = s;
        else delete btn.dataset.status;
      }
    },
    setProblems(problems) {
      const byId = new Map(problems.map((p) => [pathId(p.path), `${p.message}. Fix: ${p.fix}`]));
      for (const [bid, btn] of buttons) {
        const msg = byId.get(bid);
        if (msg) {
          btn.dataset.problem = msg;
          btn.title = msg;
        } else {
          delete btn.dataset.problem;
        }
      }
    },
    setDebug(visited, waiting) {
      for (const [id, btn] of buttons) {
        btn.classList.toggle('visited', visited.has(id));
        btn.classList.toggle('waiting', id === waiting);
      }
    },
    setSelected(path) {
      const id = path ? pathId(path) : '';
      for (const [bid, btn] of buttons) {
        if (bid === id) btn.setAttribute('aria-current', 'true');
        else btn.removeAttribute('aria-current');
      }
    },
  };
}
