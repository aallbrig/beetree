import { parse, stringify } from 'yaml';

export const NODE_TYPES = [
  'action',
  'condition',
  'sequence',
  'selector',
  'parallel',
  'decorator',
] as const;
export type NodeType = (typeof NODE_TYPES)[number];

export const DECORATORS = [
  'repeat',
  'negate',
  'always_succeed',
  'always_fail',
  'until_fail',
  'until_succeed',
  'timeout',
  'cooldown',
  'retry',
] as const;

export interface NodeSpec {
  type: NodeType;
  name: string;
  node?: string;
  description?: string;
  decorator?: string;
  parameters?: Record<string, unknown>;
  children?: NodeSpec[];
}

export interface TreeSpec {
  version: string;
  metadata: { name: string; description?: string; author?: string; tags?: string[] };
  tree: NodeSpec;
}

/** A path from the root: [] is the root, [0, 2] is the root's first child's third child. */
export type Path = readonly number[];

export const isComposite = (t: NodeType): boolean =>
  t === 'sequence' || t === 'selector' || t === 'parallel';
export const isLeaf = (t: NodeType): boolean => t === 'action' || t === 'condition';

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v);
}

function isNodeType(v: unknown): v is NodeType {
  return NODE_TYPES.some((t) => t === v);
}

function toNode(v: unknown, where: string): NodeSpec {
  if (!isRecord(v)) throw new Error(`${where}: expected a node object`);
  const { type, name } = v;
  if (!isNodeType(type)) throw new Error(`${where}: unknown node type ${JSON.stringify(type)}`);
  if (typeof name !== 'string' || name === '') throw new Error(`${where}: name is required`);
  const node: NodeSpec = { type, name };
  if (typeof v.node === 'string') node.node = v.node;
  if (typeof v.description === 'string') node.description = v.description;
  if (typeof v.decorator === 'string') node.decorator = v.decorator;
  if (isRecord(v.parameters)) node.parameters = v.parameters;
  if (Array.isArray(v.children)) {
    node.children = v.children.map((c: unknown, i) => toNode(c, `${where}.children[${i}]`));
  }
  return node;
}

/** Parse .beetree.yaml text into the editor's model; throws a readable Error. */
export function parseSpec(text: string): TreeSpec {
  const doc: unknown = parse(text);
  if (!isRecord(doc)) throw new Error('spec must be a YAML mapping');
  const meta = isRecord(doc.metadata) ? doc.metadata : {};
  const metadata: TreeSpec['metadata'] = { name: typeof meta.name === 'string' ? meta.name : '' };
  if (typeof meta.description === 'string') metadata.description = meta.description;
  if (typeof meta.author === 'string') metadata.author = meta.author;
  if (Array.isArray(meta.tags))
    metadata.tags = meta.tags.filter((t): t is string => typeof t === 'string');
  return {
    version: typeof doc.version === 'string' ? doc.version : '1.0',
    metadata,
    tree: toNode(doc.tree, 'tree'),
  };
}

export function stringifySpec(spec: TreeSpec): string {
  return stringify(spec, { lineWidth: 0 });
}

export function newSpec(name = 'my-tree'): TreeSpec {
  return {
    version: '1.0',
    metadata: { name, description: 'A new behavior tree' },
    tree: {
      type: 'selector',
      name: 'root',
      children: [
        {
          type: 'sequence',
          name: 'chase',
          children: [
            { type: 'condition', name: 'has_target' },
            { type: 'action', name: 'chase_target' },
          ],
        },
        { type: 'action', name: 'patrol' },
      ],
    },
  };
}

export function getNode(root: NodeSpec, path: Path): NodeSpec | undefined {
  let cur: NodeSpec | undefined = root;
  for (const i of path) cur = cur?.children?.[i];
  return cur;
}

export function countNodes(n: NodeSpec): number {
  return 1 + (n.children ?? []).reduce((sum, c) => sum + countNodes(c), 0);
}

export function allNames(n: NodeSpec, into = new Set<string>()): Set<string> {
  into.add(n.name);
  for (const c of n.children ?? []) allNames(c, into);
  return into;
}

export function uniqueName(root: NodeSpec, base: string): string {
  const names = allNames(root);
  if (!names.has(base)) return base;
  let i = 2;
  while (names.has(`${base}_${i}`)) i++;
  return `${base}_${i}`;
}

export function makeNode(root: NodeSpec, type: NodeType): NodeSpec {
  const base = type === 'decorator' ? 'repeat_node' : `new_${type}`;
  const node: NodeSpec = { type, name: uniqueName(root, base) };
  if (isComposite(type) || type === 'decorator') node.children = [];
  if (type === 'decorator') node.decorator = 'repeat';
  return node;
}

/** Add child under the node at parentPath. Returns the child's path or an error message. */
export function addChild(root: NodeSpec, parentPath: Path, child: NodeSpec): Path | string {
  const parent = getNode(root, parentPath);
  if (!parent) return 'parent not found';
  if (isLeaf(parent.type)) return `${parent.type} nodes cannot have children`;
  if (parent.type === 'decorator' && (parent.children?.length ?? 0) >= 1)
    return 'a decorator has exactly one child';
  parent.children ??= [];
  parent.children.push(child);
  return [...parentPath, parent.children.length - 1];
}

export function removeAt(root: NodeSpec, path: Path): boolean {
  if (path.length === 0) return false;
  const parent = getNode(root, path.slice(0, -1));
  const idx = path[path.length - 1];
  if (!parent?.children || idx === undefined || idx >= parent.children.length) return false;
  parent.children.splice(idx, 1);
  return true;
}

/** Move a node among its siblings by delta (-1 up, +1 down). Returns the new path. */
export function moveSibling(root: NodeSpec, path: Path, delta: -1 | 1): Path {
  const parent = getNode(root, path.slice(0, -1));
  const idx = path[path.length - 1];
  if (!parent?.children || idx === undefined) return path;
  const to = idx + delta;
  const moved = parent.children[idx];
  const swap = parent.children[to];
  if (!moved || !swap) return path;
  parent.children[idx] = swap;
  parent.children[to] = moved;
  return [...path.slice(0, -1), to];
}

/** Deep-copy n, renaming nodes that collide with taken (which accumulates). */
function cloneUnique(n: NodeSpec, taken: Set<string>): NodeSpec {
  let name = n.name;
  if (taken.has(name)) {
    let i = 2;
    while (taken.has(`${n.name}_${i}`)) i++;
    name = `${n.name}_${i}`;
  }
  taken.add(name);
  const copy: NodeSpec = { ...n, name };
  if (n.children) copy.children = n.children.map((c) => cloneUnique(c, taken));
  return copy;
}

/** Duplicate the node (and its subtree) right after itself. Returns the copy's path or an error. */
export function duplicateAt(root: NodeSpec, path: Path): Path | string {
  if (path.length === 0) return 'the root cannot be duplicated';
  const parent = getNode(root, path.slice(0, -1));
  const idx = path[path.length - 1];
  const node = getNode(root, path);
  if (!parent?.children || idx === undefined || !node) return 'node not found';
  if (parent.type === 'decorator') return 'a decorator has exactly one child';
  const copy = cloneUnique(node, allNames(root));
  parent.children.splice(idx + 1, 0, copy);
  return [...path.slice(0, -1), idx + 1];
}

/** Move a node out to become its parent's next sibling. Returns the new path or an error. */
export function outdent(root: NodeSpec, path: Path): Path | string {
  if (path.length < 2) return 'already directly under the root';
  const parentPath = path.slice(0, -1);
  const grand = getNode(root, parentPath.slice(0, -1));
  const parent = getNode(root, parentPath);
  const idx = path[path.length - 1];
  const pidx = parentPath[parentPath.length - 1];
  if (!grand?.children || !parent?.children || idx === undefined || pidx === undefined)
    return 'node not found';
  if (grand.type === 'decorator') return 'a decorator has exactly one child';
  const [node] = parent.children.splice(idx, 1);
  if (!node) return 'node not found';
  grand.children.splice(pidx + 1, 0, node);
  return [...parentPath.slice(0, -1), pidx + 1];
}

/** Move a node into the sibling just above it. Returns the new path or an error. */
export function indent(root: NodeSpec, path: Path): Path | string {
  const idx = path[path.length - 1];
  const parent = getNode(root, path.slice(0, -1));
  if (path.length === 0 || !parent?.children || idx === undefined)
    return 'the root cannot be moved';
  const prev = parent.children[idx - 1];
  const node = parent.children[idx];
  if (!prev || !node) return 'no sibling above to move into';
  const prevPath = [...path.slice(0, -1), idx - 1];
  parent.children.splice(idx, 1);
  const res = addChild(root, prevPath, node);
  if (typeof res === 'string') {
    parent.children.splice(idx, 0, node);
    return res;
  }
  return res;
}

/** Rename a node, keeping names unique. Returns an error message or undefined. */
export function rename(root: NodeSpec, path: Path, name: string): string | undefined {
  const clean = name.trim();
  if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(clean))
    return 'use letters, digits and underscores (not starting with a digit)';
  const node = getNode(root, path);
  if (!node) return 'node not found';
  if (clean !== node.name && allNames(root).has(clean)) return `"${clean}" is already used`;
  node.name = clean;
  return undefined;
}

export function* walk(n: NodeSpec, path: Path = []): Generator<[NodeSpec, Path]> {
  yield [n, path];
  for (let i = 0; i < (n.children ?? []).length; i++) {
    const child = n.children?.[i];
    if (child) yield* walk(child, [...path, i]);
  }
}

export interface Problem {
  path: Path;
  message: string;
}

/** Structural problems the validator would reject, each tied to the offending node. */
export function findProblems(root: NodeSpec): Problem[] {
  const out: Problem[] = [];
  const seen = new Set<string>();
  for (const [n, path] of walk(root)) {
    if (seen.has(n.name))
      out.push({ path, message: `${n.name}: name is used more than once; rename one of them` });
    seen.add(n.name);
    const kids = n.children?.length ?? 0;
    if (isComposite(n.type) && kids === 0)
      out.push({ path, message: `${n.name}: a ${n.type} needs at least one child` });
    if (n.type === 'decorator' && kids !== 1)
      out.push({ path, message: `${n.name}: a decorator needs exactly one child` });
  }
  return out;
}
