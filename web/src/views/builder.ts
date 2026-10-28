import { api, ApiError } from '../api';
import { outline, type Outline } from '../components/outline';
import { clear, h, toast } from '../dom';
import { navigate, type Cleanup, type Ctx } from '../router';
import { World, type Agent } from '../scene/world';
import {
  addChild,
  countNodes,
  DECORATORS,
  getNode,
  isComposite,
  isLeaf,
  makeNode,
  moveSibling,
  newSpec,
  NODE_TYPES,
  parseSpec,
  removeAt,
  rename,
  stringifySpec,
  walk,
  type NodeType,
  type Path,
  type TreeSpec,
} from '../spec';

const DRAFT_KEY = 'beetree.draft';

function problems(spec: TreeSpec): string[] {
  const out: string[] = [];
  for (const [n] of walk(spec.tree)) {
    const kids = n.children?.length ?? 0;
    if (isComposite(n.type) && kids === 0)
      out.push(`${n.name}: a ${n.type} needs at least one child`);
    if (n.type === 'decorator' && kids !== 1)
      out.push(`${n.name}: a decorator needs exactly one child`);
  }
  return out;
}

export async function builderView(ctx: Ctx): Promise<Cleanup | undefined> {
  const treeId = ctx.params[0];
  let spec: TreeSpec = newSpec();
  let visibility: 'public' | 'private' = 'private';
  let selected: Path = [];

  const forkId = ctx.query.get('fork');
  if (treeId ?? forkId) {
    const t = await api.getTree((treeId ?? forkId) as string);
    if (!t.spec) throw new Error('That tree has no spec.');
    spec = parseSpec(t.spec);
    if (treeId) visibility = t.visibility;
    else spec.metadata.name = `${t.name}-fork`;
  } else {
    try {
      const draft = localStorage.getItem(DRAFT_KEY);
      if (draft) {
        spec = parseSpec(draft);
        toast('Restored your unsaved draft.');
      }
    } catch {
      // unreadable draft: start fresh
    }
  }

  // ---- 3D preview -------------------------------------------------------
  const canvas = h('canvas', { 'aria-label': 'Live 3D preview of this tree' });
  const intruderBtn = h(
    'button',
    { class: 'btn small', type: 'button', 'aria-pressed': 'true' },
    '🎯 Intruder',
  );
  const resetBtn = h('button', { class: 'btn small', type: 'button' }, '↺ Restart');
  const stage = h(
    'div',
    { class: 'stage' },
    canvas,
    h('div', { class: 'overlay' }, intruderBtn, resetBtn),
    h(
      'div',
      { class: 'hintbar' },
      'Your tree is running live. Click the ground to move the intruder.',
    ),
  );

  // ---- form pieces ------------------------------------------------------
  const nameIn = h('input', {
    type: 'text',
    value: spec.metadata.name,
    maxlength: '80',
    'aria-label': 'Tree name',
  });
  const descIn = h(
    'textarea',
    { 'aria-label': 'Description', maxlength: '500', style: 'min-height:60px' },
    spec.metadata.description ?? '',
  );
  const tagsIn = h('input', {
    type: 'text',
    value: (spec.metadata.tags ?? []).join(', '),
    'aria-label': 'Tags',
  });
  const visIn = h(
    'select',
    { 'aria-label': 'Visibility' },
    h('option', { value: 'private' }, 'Private (only me)'),
    h('option', { value: 'public' }, 'Public (shared)'),
  );
  visIn.value = visibility;

  const status = h('div', { class: 'small', role: 'status' });
  const outlineHost = h('div');
  const propsHost = h('div', { class: 'stack' });
  const yamlArea = h('textarea', {
    class: 'code',
    'aria-label': 'YAML',
    spellcheck: 'false',
    style: 'min-height:340px',
  });
  const yamlErr = h('div', { class: 'error', role: 'alert' });
  const codeOut = h('div', { class: 'stack' });
  const engineSel = h(
    'select',
    { 'aria-label': 'Engine' },
    ...['unity', 'unreal', 'godot'].map((e) => h('option', { value: e }, e)),
  );

  const panels: Record<string, HTMLElement> = {};
  const tabs = h('div', { class: 'tabs', role: 'tablist' });
  const showTab = (name: string): void => {
    for (const [k, el] of Object.entries(panels)) el.classList.toggle('hide', k !== name);
    for (const b of tabs.querySelectorAll('button'))
      b.setAttribute('aria-selected', String(b.dataset.tab === name));
    if (name === 'yaml') yamlArea.value = stringifySpec(spec);
  };
  const addTab = (id: string, label: string, el: HTMLElement): void => {
    panels[id] = el;
    tabs.append(
      h(
        'button',
        {
          class: 'tab',
          role: 'tab',
          'data-tab': id,
          onClick: () => {
            showTab(id);
          },
        },
        label,
      ),
    );
  };

  // ---- design tab -------------------------------------------------------
  const typeSel = h(
    'select',
    { 'aria-label': 'New node type' },
    ...NODE_TYPES.map((t) => h('option', { value: t }, t)),
  );
  const addBtn = h('button', { class: 'btn small', type: 'button' }, '+ Add child');
  const rmBtn = h('button', { class: 'btn small danger', type: 'button' }, 'Remove');
  const upBtn = h('button', { class: 'btn small', type: 'button', 'aria-label': 'Move up' }, '↑');
  const downBtn = h(
    'button',
    { class: 'btn small', type: 'button', 'aria-label': 'Move down' },
    '↓',
  );
  const design = h(
    'div',
    { class: 'stack' },
    h('div', { class: 'row' }, typeSel, addBtn, rmBtn, upBtn, downBtn),
    outlineHost,
    propsHost,
  );
  addTab('design', 'Design', design);
  addTab(
    'yaml',
    'YAML',
    h(
      'div',
      { class: 'stack' },
      yamlArea,
      yamlErr,
      h('button', { class: 'btn small', type: 'button', onClick: applyYaml }, 'Apply YAML'),
    ),
  );
  const genBtn = h('button', { class: 'btn small', type: 'button' }, 'Generate');
  addTab(
    'code',
    'Engine code',
    h('div', { class: 'stack' }, h('div', { class: 'row' }, engineSel, genBtn), codeOut),
  );

  const saveBtn = h(
    'button',
    { class: 'btn primary', type: 'button' },
    treeId ? 'Save changes' : 'Save tree',
  );
  const validateBtn = h('button', { class: 'btn', type: 'button' }, 'Validate');
  const left = h(
    'div',
    { class: 'card stack' },
    h('h2', {}, treeId ? 'Edit tree' : 'New tree'),
    h('div', {}, h('label', {}, 'Name'), nameIn),
    h('div', {}, h('label', {}, 'Description'), descIn),
    h(
      'div',
      { class: 'row' },
      h('div', { style: 'flex:1' }, h('label', {}, 'Tags'), tagsIn),
      h('div', { style: 'flex:1' }, h('label', {}, 'Visibility'), visIn),
    ),
    h('div', { class: 'row' }, saveBtn, validateBtn),
    status,
    tabs,
    ...Object.values(panels),
  );
  ctx.root.append(h('div', { class: 'studio' }, left, stage));

  // ---- preview world ----------------------------------------------------
  let world: World | undefined;
  let agent: Agent | undefined;
  let ol: Outline | undefined;
  try {
    world = await World.create(canvas);
    world.mode = 'beacon';
    world.setBeacon({ x: 5, z: -1 });
    world.start();
    agent = await world.addAgent({
      id: 'preview',
      name: spec.metadata.name || 'preview',
      model: 'character-c',
      x: 0,
      z: 0,
      spec: spec.tree,
    });
  } catch {
    clear(stage);
    stage.append(
      h(
        'p',
        { style: 'padding:24px' },
        'WebGL is unavailable, so the live preview is off. You can still edit and save.',
      ),
    );
  }

  // ---- behaviour --------------------------------------------------------
  function syncMeta(): void {
    spec.metadata.name = nameIn.value.trim() || 'untitled';
    spec.metadata.description = descIn.value.trim();
    spec.metadata.tags = tagsIn.value
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean);
  }

  function persistDraft(): void {
    if (treeId) return;
    try {
      localStorage.setItem(DRAFT_KEY, stringifySpec(spec));
    } catch {
      // storage unavailable: drafts just are not remembered
    }
  }

  function refresh(): void {
    syncMeta();
    ol = outline(spec.tree, (p) => {
      selected = p;
      refresh();
    });
    ol.setSelected(selected);
    clear(outlineHost);
    outlineHost.append(ol.el);
    renderProps();
    const issues = problems(spec);
    status.textContent = issues.length
      ? `⚠ ${issues[0] ?? ''}${issues.length > 1 ? ` (+${issues.length - 1} more)` : ''}`
      : `${countNodes(spec.tree)} nodes`;
    status.style.color = issues.length ? 'var(--warn)' : 'var(--muted)';
    agent?.setSpec(spec.tree);
    persistDraft();
  }

  function renderProps(): void {
    clear(propsHost);
    const node = getNode(spec.tree, selected);
    if (!node) return;
    const nameField = h('input', { type: 'text', value: node.name, 'aria-label': 'Node name' });
    const nameErr = h('div', { class: 'error' });
    nameField.addEventListener('change', () => {
      const err = rename(spec.tree, selected, nameField.value);
      nameErr.textContent = err ?? '';
      if (!err) refresh();
    });
    propsHost.append(
      h('h3', {}, `Selected: ${node.type}`),
      h('div', {}, h('label', {}, 'Name'), nameField, nameErr),
    );
    if (isLeaf(node.type)) {
      const impl = h('input', {
        type: 'text',
        value: node.node ?? '',
        placeholder: 'e.g. MoveToTarget',
        'aria-label': 'Implementation id',
      });
      impl.addEventListener('change', () => {
        if (impl.value.trim()) node.node = impl.value.trim();
        else delete node.node;
        refresh();
      });
      propsHost.append(
        h(
          'div',
          {},
          h('label', {}, 'Implementation id'),
          impl,
          h(
            'div',
            { class: 'hint' },
            'The class or function your engine code provides. Names like chase, attack, flee, wait and patrol animate in the preview.',
          ),
        ),
      );
    }
    if (node.type === 'decorator') {
      const dec = h(
        'select',
        { 'aria-label': 'Decorator' },
        ...DECORATORS.map((d) => h('option', { value: d }, d)),
      );
      dec.value = node.decorator ?? 'repeat';
      dec.addEventListener('change', () => {
        node.decorator = dec.value;
        refresh();
      });
      propsHost.append(h('div', {}, h('label', {}, 'Decorator'), dec));
    }
    const desc = h('input', {
      type: 'text',
      value: node.description ?? '',
      'aria-label': 'Node description',
    });
    desc.addEventListener('change', () => {
      if (desc.value.trim()) node.description = desc.value.trim();
      else delete node.description;
      refresh();
    });
    const params = h('input', {
      type: 'text',
      value: node.parameters ? JSON.stringify(node.parameters) : '',
      placeholder: '{"seconds": 2}',
      'aria-label': 'Parameters (JSON)',
    });
    const paramErr = h('div', { class: 'error' });
    params.addEventListener('change', () => {
      paramErr.textContent = '';
      if (!params.value.trim()) {
        delete node.parameters;
        refresh();
        return;
      }
      try {
        const v: unknown = JSON.parse(params.value);
        if (typeof v !== 'object' || v === null || Array.isArray(v))
          throw new Error('not an object');
        node.parameters = v as Record<string, unknown>;
        refresh();
      } catch {
        paramErr.textContent = 'Parameters must be a JSON object, like {"seconds": 2}.';
      }
    });
    propsHost.append(
      h('div', {}, h('label', {}, 'Description'), desc),
      h('div', {}, h('label', {}, 'Parameters (JSON)'), params, paramErr),
    );
  }

  function applyYaml(): void {
    try {
      spec = parseSpec(yamlArea.value);
      yamlErr.textContent = '';
      selected = [];
      nameIn.value = spec.metadata.name;
      descIn.value = spec.metadata.description ?? '';
      tagsIn.value = (spec.metadata.tags ?? []).join(', ');
      refresh();
      toast('YAML applied.');
    } catch (e) {
      yamlErr.textContent = e instanceof Error ? e.message : 'Invalid YAML.';
    }
  }

  addBtn.addEventListener('click', () => {
    const type = typeSel.value as NodeType;
    const res = addChild(spec.tree, selected, makeNode(spec.tree, type));
    if (typeof res === 'string') {
      toast(res, 'err');
      return;
    }
    selected = res;
    refresh();
  });
  rmBtn.addEventListener('click', () => {
    if (!removeAt(spec.tree, selected)) {
      toast('The root cannot be removed.', 'err');
      return;
    }
    selected = selected.slice(0, -1);
    refresh();
  });
  upBtn.addEventListener('click', () => {
    selected = moveSibling(spec.tree, selected, -1);
    refresh();
  });
  downBtn.addEventListener('click', () => {
    selected = moveSibling(spec.tree, selected, 1);
    refresh();
  });
  for (const el of [nameIn, descIn, tagsIn]) el.addEventListener('change', refresh);

  validateBtn.addEventListener('click', () => {
    syncMeta();
    api.validate(stringifySpec(spec)).then(
      (r) => {
        if (r.valid) toast(`Valid — ${r.nodeCount} nodes.`);
        else status.textContent = `✖ ${r.errors.join(' · ')}`;
      },
      () => {
        toast('Could not reach the validator.', 'err');
      },
    );
  });

  saveBtn.addEventListener('click', () => {
    syncMeta();
    saveBtn.disabled = true;
    const body = {
      name: spec.metadata.name,
      description: spec.metadata.description ?? '',
      visibility: visIn.value === 'public' ? ('public' as const) : ('private' as const),
      tags: spec.metadata.tags ?? [],
      spec: stringifySpec(spec),
    };
    (treeId ? api.updateTree(treeId, body) : api.createTree(body))
      .then((t) => {
        toast('Saved.');
        try {
          localStorage.removeItem(DRAFT_KEY);
        } catch {
          // ignore
        }
        if (!treeId) navigate(`/builder/${t.id}`);
      })
      .catch((e: unknown) => {
        toast(e instanceof ApiError ? e.message : 'Could not save.', 'err');
      })
      .finally(() => {
        saveBtn.disabled = false;
      });
  });

  genBtn.addEventListener('click', () => {
    syncMeta();
    clear(codeOut);
    api.generate(stringifySpec(spec), engineSel.value).then(
      (r) => {
        for (const f of r.files) {
          codeOut.append(
            h(
              'details',
              {},
              h('summary', {}, `${f.path}${f.isStub ? ' (your stub)' : ''}`),
              h('pre', { class: 'code' }, f.content),
            ),
          );
        }
      },
      (e: unknown) => {
        codeOut.append(
          h(
            'div',
            { class: 'error' },
            e instanceof ApiError ? e.message : 'Could not generate code.',
          ),
        );
      },
    );
  });

  intruderBtn.addEventListener('click', () => {
    if (!world) return;
    const on = world.beacon === undefined;
    world.setBeacon(on ? { x: 5, z: -1 } : undefined);
    intruderBtn.setAttribute('aria-pressed', String(on));
  });
  resetBtn.addEventListener('click', () => {
    world?.resetAgents();
  });

  showTab('design');
  refresh();
  const timer = setInterval(() => {
    if (agent) ol?.setStatuses(agent.runtime.statuses);
  }, 120);
  return () => {
    clearInterval(timer);
    world?.dispose();
  };
}
