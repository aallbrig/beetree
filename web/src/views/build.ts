import { h } from '../dom';
import type { Ctx } from '../router';

/** The Build hub: the one obvious place to start authoring. */
export function buildView(ctx: Ctx): void {
  ctx.root.append(
    h('h1', {}, 'Build'),
    h('p', { class: 'muted' }, 'Choose what you want to make. You can always come back here.'),
    h(
      'div',
      { class: 'grid' },
      h(
        'article',
        { class: 'card stack' },
        h('h2', {}, 'A behavior tree'),
        h(
          'p',
          { class: 'muted' },
          'Compose nodes visually, watch a character run the tree live, validate it, and export engine code.',
        ),
        h('a', { class: 'btn primary', href: '#/builder' }, 'New tree'),
      ),
      h(
        'article',
        { class: 'card stack' },
        h('h2', {}, 'A scene of agents'),
        h(
          'p',
          { class: 'muted' },
          'Place several agents in a 3D scene and give each its own tree. Write a new tree for any agent without leaving the scene.',
        ),
        h('a', { class: 'btn primary', href: '#/scenes' }, 'New scene'),
      ),
    ),
    h(
      'p',
      { class: 'muted' },
      'Looking for what you already made? ',
      h('a', { href: '#/dashboard' }, 'My trees and scenes'),
      '.',
    ),
  );
}
