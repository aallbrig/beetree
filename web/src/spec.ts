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
  success_policy?: string;
  failure_policy?: string;
  ref?: string;
  file?: string;
  children?: NodeSpec[];
}

/** Top-level sections the builder has no editor for yet; kept so import then export loses nothing. */
export const PASSTHROUGH_KEYS = ['blackboard', 'custom_nodes', 'notation', 'subtrees'] as const;

export interface TreeSpec {
  version: string;
  metadata: {
    name: string;
    description?: string;
    author?: string;
    license?: string;
    tags?: string[];
  };
  blackboard?: unknown;
  custom_nodes?: unknown;
  notation?: unknown;
  subtrees?: unknown;
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
  if (typeof v.success_policy === 'string') node.success_policy = v.success_policy;
  if (typeof v.failure_policy === 'string') node.failure_policy = v.failure_policy;
  if (typeof v.ref === 'string') node.ref = v.ref;
  if (typeof v.file === 'string') node.file = v.file;
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
  if (typeof meta.license === 'string') metadata.license = meta.license;
  if (Array.isArray(meta.tags))
    metadata.tags = meta.tags.filter((t): t is string => typeof t === 'string');
  const spec: TreeSpec = {
    version: typeof doc.version === 'string' ? doc.version : '1.0',
    metadata,
    tree: toNode(doc.tree, 'tree'),
  };
  for (const k of PASSTHROUGH_KEYS) if (doc[k] !== undefined) spec[k] = doc[k];
  return spec;
}

export function stringifySpec(spec: TreeSpec): string {
  // Canonical section order: version, metadata, blackboard..., tree, subtrees
  const ordered: Record<string, unknown> = { version: spec.version, metadata: spec.metadata };
  for (const k of PASSTHROUGH_KEYS)
    if (k !== 'subtrees' && spec[k] !== undefined) ordered[k] = spec[k];
  ordered.tree = spec.tree;
  if (spec.subtrees !== undefined) ordered.subtrees = spec.subtrees;
  return stringify(ordered, { lineWidth: 0 });
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

export const NODE_CEILING = 25;

/** Soft size warning so authors keep trees small enough to read whole. */
export function sizeWarning(count: number, ceiling = NODE_CEILING): string | undefined {
  if (count <= ceiling) return undefined;
  return `${count} nodes, over the ${ceiling}-node target. Fix: extract a branch into a subtree or merge similar nodes`;
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

/** Wrap the node in a new composite or decorator that takes its place. Returns the node's new path or an error. */
export function wrapAt(root: NodeSpec, path: Path, type: NodeType): Path | string {
  if (isLeaf(type)) return 'choose a sequence, selector, parallel or decorator to wrap with';
  const node = getNode(root, path);
  if (!node) return 'node not found';
  const wrapper = makeNode(root, type);
  wrapper.children = [node];
  if (path.length === 0) {
    const inner: NodeSpec = { ...node };
    const fresh: NodeSpec = { ...wrapper, children: [inner] };
    root.type = fresh.type;
    root.name = fresh.name;
    root.children = fresh.children;
    delete root.node;
    delete root.description;
    delete root.parameters;
    delete root.ref;
    delete root.file;
    delete root.success_policy;
    delete root.failure_policy;
    if (fresh.decorator) root.decorator = fresh.decorator;
    else delete root.decorator;
    return [0];
  }
  const parent = getNode(root, path.slice(0, -1));
  const idx = path[path.length - 1];
  if (!parent?.children || idx === undefined) return 'node not found';
  parent.children[idx] = wrapper;
  return [...path, 0];
}

/** Change a node's type in place: composites swap among themselves, leaves between action and condition. */
export function retype(root: NodeSpec, path: Path, type: NodeType): string | undefined {
  const node = getNode(root, path);
  if (!node) return 'node not found';
  if (node.type === type) return undefined;
  const kind = (t: NodeType): string => (isComposite(t) ? 'composite' : isLeaf(t) ? 'leaf' : t);
  if (kind(node.type) !== kind(type))
    return `a ${node.type} cannot become a ${type}; add the new node and move the children instead`;
  node.type = type;
  if (type !== 'parallel') {
    delete node.success_policy;
    delete node.failure_policy;
  }
  return undefined;
}

/** Add a new node right after the one at path. Returns the new path or an error. */
export function addSibling(root: NodeSpec, path: Path, child: NodeSpec): Path | string {
  if (path.length === 0) return 'the root has no siblings; add a child instead';
  const parent = getNode(root, path.slice(0, -1));
  const idx = path[path.length - 1];
  if (!parent?.children || idx === undefined) return 'node not found';
  if (parent.type === 'decorator') return 'a decorator has exactly one child';
  parent.children.splice(idx + 1, 0, child);
  return [...path.slice(0, -1), idx + 1];
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

export const PARALLEL_POLICIES = ['require_all', 'require_one'] as const;
export type PolicyKey = 'success_policy' | 'failure_policy';

/** Set or clear (empty value) a parallel node's success or failure policy. Returns an error message or undefined. */
export function setPolicy(
  root: NodeSpec,
  path: Path,
  key: PolicyKey,
  value: string,
): string | undefined {
  const node = getNode(root, path);
  if (!node) return 'node not found';
  if (node.type !== 'parallel') return 'only parallel nodes have policies';
  if (value === '') {
    if (key === 'success_policy') delete node.success_policy;
    else delete node.failure_policy;
    return undefined;
  }
  if (!(PARALLEL_POLICIES as readonly string[]).includes(value))
    return `policy must be one of ${PARALLEL_POLICIES.join(', ')}`;
  node[key] = value;
  return undefined;
}

export function* walk(n: NodeSpec, path: Path = []): Generator<[NodeSpec, Path]> {
  yield [n, path];
  for (let i = 0; i < (n.children ?? []).length; i++) {
    const child = n.children?.[i];
    if (child) yield* walk(child, [...path, i]);
  }
}

/** Paths of nodes whose name, implementation or description contains the query (case-insensitive), in tree order. */
export function findNodes(root: NodeSpec, query: string): Path[] {
  const q = query.trim().toLowerCase();
  if (q === '') return [];
  const out: Path[] = [];
  for (const [n, path] of walk(root)) {
    const hay = `${n.name} ${n.node ?? ''} ${n.description ?? ''}`.toLowerCase();
    if (hay.includes(q)) out.push(path);
  }
  return out;
}

/** Ids (see pathId) of every branch node, so "collapse all" is one call. */
export function branchIds(root: NodeSpec, idOf: (p: Path) => string): Set<string> {
  const ids = new Set<string>();
  for (const [n, path] of walk(root)) if ((n.children ?? []).length > 0) ids.add(idOf(path));
  return ids;
}

/** Removes the ids of every ancestor of path from collapsed, so the node at path is visible. */
export function revealPath(collapsed: Set<string>, path: Path, idOf: (p: Path) => string): void {
  for (let i = 0; i < path.length; i++) collapsed.delete(idOf(path.slice(0, i)));
}

/** Paths a collapsed outline still shows: a collapsed node is shown, its descendants are not. */
export function visiblePaths(
  root: NodeSpec,
  collapsed: ReadonlySet<string>,
  idOf: (p: Path) => string,
): Path[] {
  const out: Path[] = [];
  const go = (n: NodeSpec, path: Path): void => {
    out.push(path);
    if (collapsed.has(idOf(path))) return;
    (n.children ?? []).forEach((c, i) => {
      go(c, [...path, i]);
    });
  };
  go(root, []);
  return out;
}

/**
 * Splits a server validator error such as `tree.children[1].children[0]: name is required`
 * into the node path it names and the message. Errors that name no node get a null path.
 */
export function parseValidatorError(err: string): { path: Path | null; message: string } {
  const m = /^tree((?:\.children\[\d+\])*): (.*)$/s.exec(err);
  if (!m) return { path: null, message: err };
  const path = [...(m[1] ?? '').matchAll(/\[(\d+)\]/g)].map((x) => Number(x[1]));
  return { path, message: m[2] ?? '' };
}

export interface Problem {
  path: Path;
  message: string;
  /** A concrete repair the author can apply. */
  fix: string;
}

/** Structural problems the validator would reject, each tied to the offending node. */
export function findProblems(root: NodeSpec): Problem[] {
  const out: Problem[] = [];
  const seen = new Set<string>();
  for (const [n, path] of walk(root)) {
    if (seen.has(n.name))
      out.push({
        path,
        message: `${n.name}: name is used more than once`,
        fix: 'Rename one of the nodes so every name is unique.',
      });
    seen.add(n.name);
    const kids = n.children?.length ?? 0;
    if (isComposite(n.type) && kids === 0)
      out.push({
        path,
        message: `${n.name}: a ${n.type} needs at least one child`,
        fix: `Add a child under ${n.name}, or delete the ${n.type}.`,
      });
    if (n.type === 'decorator' && kids !== 1)
      out.push({
        path,
        message: `${n.name}: a decorator needs exactly one child`,
        fix: `Leave exactly one child under ${n.name}; wrap several in a sequence first.`,
      });
  }
  return out;
}

/** One blackboard variable: a typed value shared by all nodes. */
export interface BlackboardVar {
  name: string;
  type: string;
  default?: unknown;
  description?: string;
}

/** The blackboard variables of a spec; entries that are not well-formed are skipped. */
export function getBlackboard(spec: TreeSpec): BlackboardVar[] {
  if (!Array.isArray(spec.blackboard)) return [];
  const out: BlackboardVar[] = [];
  for (const v of spec.blackboard as unknown[]) {
    if (isRecord(v) && typeof v.name === 'string' && typeof v.type === 'string') {
      const item: BlackboardVar = { ...v, name: v.name, type: v.type };
      out.push(item);
    }
  }
  return out;
}

function setBlackboard(spec: TreeSpec, vars: BlackboardVar[]): void {
  if (vars.length === 0) delete spec.blackboard;
  else spec.blackboard = vars;
}

/** Add a variable with a unique name (name, name_2, ...); returns the name used. */
export function addVar(spec: TreeSpec, name = 'new_var', type = 'string'): string {
  const vars = getBlackboard(spec);
  const taken = new Set(vars.map((v) => v.name));
  let unique = name;
  for (let i = 2; taken.has(unique); i++) unique = `${name}_${i}`;
  setBlackboard(spec, [...vars, { name: unique, type }]);
  return unique;
}

/** Change fields of a variable; an empty default or description removes the key. Returns false if the name is unknown or a rename would collide. */
export function updateVar(
  spec: TreeSpec,
  name: string,
  patch: { name?: string; type?: string; default?: unknown; description?: string },
): boolean {
  const vars = getBlackboard(spec);
  const idx = vars.findIndex((v) => v.name === name);
  const cur = vars[idx];
  if (!cur) return false;
  const next: BlackboardVar = { ...cur };
  if (patch.name !== undefined) {
    const n = patch.name.trim();
    if (n === '' || (n !== name && vars.some((v) => v.name === n))) return false;
    next.name = n;
  }
  if (patch.type !== undefined && patch.type.trim() !== '') next.type = patch.type.trim();
  if ('default' in patch) {
    if (patch.default === undefined || patch.default === '') delete next.default;
    else next.default = patch.default;
  }
  if (patch.description !== undefined) {
    if (patch.description === '') delete next.description;
    else next.description = patch.description;
  }
  vars[idx] = next;
  setBlackboard(spec, vars);
  return true;
}

export function removeVar(spec: TreeSpec, name: string): boolean {
  const vars = getBlackboard(spec);
  const rest = vars.filter((v) => v.name !== name);
  if (rest.length === vars.length) return false;
  setBlackboard(spec, rest);
  return true;
}

/** Read a typed default from the text box: numbers and booleans by declared type, else a string. */
export function parseDefault(type: string, text: string): unknown {
  const t = text.trim();
  if (t === '') return undefined;
  if (type === 'int' || type === 'float') {
    const n = Number(t);
    return Number.isFinite(n) ? n : t;
  }
  if (type === 'bool' || type === 'boolean') {
    if (t === 'true') return true;
    if (t === 'false') return false;
  }
  return t;
}

/** One typed parameter of a custom node. */
export interface CustomParam {
  name: string;
  type: string;
  default?: unknown;
}

/** A user-defined action or condition type, declared once and used by leaves. */
export interface CustomNode {
  name: string;
  type: 'action' | 'condition';
  description?: string;
  parameters?: CustomParam[];
  blackboard_reads?: string[];
  blackboard_writes?: string[];
}

/** The custom node definitions of a spec; entries that are not well-formed are skipped. */
export function getCustomNodes(spec: TreeSpec): CustomNode[] {
  if (!Array.isArray(spec.custom_nodes)) return [];
  const out: CustomNode[] = [];
  for (const v of spec.custom_nodes as unknown[]) {
    if (
      isRecord(v) &&
      typeof v.name === 'string' &&
      (v.type === 'action' || v.type === 'condition')
    ) {
      out.push({ ...v, name: v.name, type: v.type });
    }
  }
  return out;
}

function setCustomNodes(spec: TreeSpec, defs: CustomNode[]): void {
  if (defs.length === 0) delete spec.custom_nodes;
  else spec.custom_nodes = defs;
}

/** Add a custom node with a unique name; returns the name used. */
export function addCustomNode(
  spec: TreeSpec,
  name = 'NewNode',
  type: 'action' | 'condition' = 'action',
): string {
  const defs = getCustomNodes(spec);
  const taken = new Set(defs.map((d) => d.name));
  let unique = name;
  for (let i = 2; taken.has(unique); i++) unique = `${name}_${i}`;
  setCustomNodes(spec, [...defs, { name: unique, type }]);
  return unique;
}

/** Change a custom node; empty description, parameters or lists remove the key. False if unknown or a rename collides. */
export function updateCustomNode(
  spec: TreeSpec,
  name: string,
  patch: {
    name?: string;
    type?: 'action' | 'condition';
    description?: string;
    parameters?: CustomParam[];
    blackboard_reads?: string[];
    blackboard_writes?: string[];
  },
): boolean {
  const defs = getCustomNodes(spec);
  const idx = defs.findIndex((d) => d.name === name);
  const cur = defs[idx];
  if (!cur) return false;
  const next: CustomNode = { ...cur };
  if (patch.name !== undefined) {
    const n = patch.name.trim();
    if (n === '' || (n !== name && defs.some((d) => d.name === n))) return false;
    next.name = n;
  }
  if (patch.type !== undefined) next.type = patch.type;
  if (patch.description !== undefined) {
    if (patch.description === '') delete next.description;
    else next.description = patch.description;
  }
  if (patch.parameters !== undefined) {
    if (patch.parameters.length === 0) delete next.parameters;
    else next.parameters = patch.parameters;
  }
  if (patch.blackboard_reads !== undefined) {
    if (patch.blackboard_reads.length === 0) delete next.blackboard_reads;
    else next.blackboard_reads = patch.blackboard_reads;
  }
  if (patch.blackboard_writes !== undefined) {
    if (patch.blackboard_writes.length === 0) delete next.blackboard_writes;
    else next.blackboard_writes = patch.blackboard_writes;
  }
  defs[idx] = next;
  setCustomNodes(spec, defs);
  return true;
}

export function removeCustomNode(spec: TreeSpec, name: string): boolean {
  const defs = getCustomNodes(spec);
  const rest = defs.filter((d) => d.name !== name);
  if (rest.length === defs.length) return false;
  setCustomNodes(spec, rest);
  return true;
}

/** Read "speed:float=1.5, clip:string" into parameters; blank entries are dropped, type defaults to string. */
export function parseParams(text: string): CustomParam[] {
  const out: CustomParam[] = [];
  for (const part of text.split(',')) {
    const t = part.trim();
    if (t === '') continue;
    const eq = t.indexOf('=');
    const head = (eq === -1 ? t : t.slice(0, eq)).trim();
    const colon = head.indexOf(':');
    const name = (colon === -1 ? head : head.slice(0, colon)).trim();
    if (name === '') continue;
    const type = colon === -1 ? 'string' : head.slice(colon + 1).trim() || 'string';
    const p: CustomParam = { name, type };
    if (eq !== -1) {
      const d = parseDefault(type, t.slice(eq + 1));
      if (d !== undefined) p.default = d;
    }
    out.push(p);
  }
  return out;
}

/** The inverse of parseParams, for the text box. */
export function formatParams(params: CustomParam[] | undefined): string {
  return (params ?? [])
    .map(
      (p) =>
        `${p.name}:${p.type}${p.default === undefined ? '' : `=${typeof p.default === 'string' ? p.default : JSON.stringify(p.default)}`}`,
    )
    .join(', ');
}

/** Split a comma list of blackboard keys. */
export function parseList(text: string): string[] {
  return text
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s !== '');
}
