export type EditAction = 'remove' | 'duplicate' | 'up' | 'down' | 'outdent' | 'indent';

export interface KeyLike {
  key: string;
  ctrlKey: boolean;
  metaKey: boolean;
  altKey: boolean;
  shiftKey: boolean;
}

/** Plain-language list of the node-editing shortcuts, shown beside the toolbar. */
export const SHORTCUT_HELP =
  'Shortcuts: Delete removes, Ctrl+D duplicates, Alt+Up/Down reorders, Alt+Left/Right outdents or indents the selected node.';

/** Maps a key press to a node edit, or undefined when it is not an editing shortcut. */
export function editAction(e: KeyLike): EditAction | undefined {
  const mod = e.ctrlKey || e.metaKey;
  if (!mod && !e.altKey && !e.shiftKey && (e.key === 'Delete' || e.key === 'Backspace'))
    return 'remove';
  if (mod && !e.altKey && !e.shiftKey && e.key.toLowerCase() === 'd') return 'duplicate';
  if (e.altKey && !mod && !e.shiftKey) {
    if (e.key === 'ArrowUp') return 'up';
    if (e.key === 'ArrowDown') return 'down';
    if (e.key === 'ArrowLeft') return 'outdent';
    if (e.key === 'ArrowRight') return 'indent';
  }
  return undefined;
}
