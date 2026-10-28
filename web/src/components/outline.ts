import type { Status } from '../engine/bt';
import { pathId } from '../engine/bt';
import { h } from '../dom';
import type { NodeSpec, Path } from '../spec';

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
}

/** A collapsible-free, keyboard-accessible tree outline with live status dots. */
export function outline(root: NodeSpec, onSelect?: (path: Path) => void): Outline {
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
    const li = h('li', {}, btn);
    const kids = node.children ?? [];
    if (kids.length > 0) {
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
    setSelected(path) {
      const id = path ? pathId(path) : '';
      for (const [bid, btn] of buttons) {
        if (bid === id) btn.setAttribute('aria-current', 'true');
        else btn.removeAttribute('aria-current');
      }
    },
  };
}
