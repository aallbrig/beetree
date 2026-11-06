import { api, ApiError } from '../api';
import { outline, type Outline } from '../components/outline';
import { clear, h, toast } from '../dom';
import { copyText, downloadText, fileName } from '../exporter';
import { History } from '../history';
import { navigate, type Cleanup, type Ctx } from '../router';
import { pathId } from '../engine/bt';
import { World, type Agent } from '../scene/world';
import {
  addChild,
  countNodes,
  DECORATORS,
  getNode,
  isLeaf,
  makeNode,
  moveSibling,
  newSpec,
  NODE_TYPES,
  parseSpec,
  removeAt,
  duplicateAt,
  indent,
  outdent,
  rename,
  findProblems,
  stringifySpec,
  type NodeType,
  type Path,
  type TreeSpec,
} from '../spec';
import { runAll, SimWalker, type Policy } from '../walker';

const DRAFT_KEY = 'beetree.draft';

export async function builderView(ctx: Ctx): Promise<Cleanup | undefined> {
  const treeId = ctx.params[0];
  let spec: TreeSpec = newSpec();
  let visibility: 'public' | 'private' = 'private';
  let selected: Path = [];

  const forkId = ctx.query.get('fork');
  // Set when a scene sent us here to author a tree for one of its agents.
  const forScene = ctx.query.get('scene');
  const forAgent = Number(ctx.query.get('agent') ?? '');
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
    'Intruder',
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
  const dupBtn = h('button', { class: 'btn small', type: 'button' }, 'Duplicate');
  const outBtn = h(
    'button',
    { class: 'btn small', type: 'button', title: 'Move out of its parent' },
    'Outdent',
  );
  const inBtn = h(
    'button',
    { class: 'btn small', type: 'button', title: 'Move into the sibling above' },
    'Indent',
  );
  const undoBtn = h(
    'button',
    { class: 'btn small', type: 'button', 'aria-label': 'Undo', title: 'Undo (Ctrl+Z)' },
    'Undo',
  );
  const redoBtn = h(
    'button',
    { class: 'btn small', type: 'button', 'aria-label': 'Redo', title: 'Redo (Ctrl+Shift+Z)' },
    'Redo',
  );
  const design = h(
    'div',
    { class: 'stack' },
    h(
      'div',
      { class: 'row' },
      typeSel,
      addBtn,
      rmBtn,
      dupBtn,
      upBtn,
      downBtn,
      outBtn,
      inBtn,
      undoBtn,
      redoBtn,
    ),
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
      h(
        'div',
        { class: 'row' },
        h('button', { class: 'btn small', type: 'button', onClick: applyYaml }, 'Apply YAML'),
        h(
          'button',
          {
            class: 'btn small',
            type: 'button',
            'data-export': 'copy-yaml',
            onClick: () => void copyOut(yamlArea.value),
          },
          'Copy YAML',
        ),
        h(
          'button',
          {
            class: 'btn small',
            type: 'button',
            'data-export': 'download-yaml',
            onClick: () => {
              downloadText(fileName(spec.metadata.name, 'beetree.yaml'), yamlArea.value);
            },
          },
          'Download YAML',
        ),
      ),
    ),
  );
  const copyOut = async (text: string): Promise<void> => {
    toast((await copyText(text)) ? 'Copied to the clipboard.' : 'Could not copy.', 'ok');
  };
  const genBtn = h('button', { class: 'btn small', type: 'button' }, 'Generate');
  addTab(
    'code',
    'Engine code',
    h('div', { class: 'stack' }, h('div', { class: 'row' }, engineSel, genBtn), codeOut),
  );

  // ---- debug tab: step through, or run all ------------------------------
  let walker: SimWalker | undefined;
  let walkerSnapshot = '';
  const dbgPrompt = h('div', { class: 'stack', role: 'status', 'aria-live': 'polite' });
  const dbgTrace = h('pre', { class: 'code', 'aria-label': 'Trace' });
  const stepBtn = h('button', { class: 'btn small', type: 'button' }, 'Step through');
  const runBtn = h('button', { class: 'btn small', type: 'button' }, 'Run all');
  const policySel = h(
    'select',
    { 'aria-label': 'Answer every leaf with' },
    h('option', { value: 'success' }, 'every leaf succeeds'),
    h('option', { value: 'failure' }, 'every leaf fails'),
    h('option', { value: 'running' }, 'every leaf is running'),
    h('option', { value: 'cycle' }, 'leaves cycle success, failure, running'),
  );
  const stopBtn = h('button', { class: 'btn small', type: 'button' }, 'Stop debugging');
  addTab(
    'debug',
    'Debug',
    h(
      'div',
      { class: 'stack' },
      h(
        'p',
        { class: 'hint' },
        'Step through pauses at each leaf so you pick its status. Composites resolve from their children.',
      ),
      h('div', { class: 'row' }, stepBtn, runBtn, policySel, stopBtn),
      dbgPrompt,
      dbgTrace,
    ),
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
  if (forScene) {
    ctx.root.append(
      h(
        'p',
        { class: 'note' },
        'Writing a tree for a scene agent. Saving assigns it to the agent and returns you to the scene. ',
        h('a', { href: `#/scenes/${forScene}` }, 'Back to the scene without saving'),
      ),
    );
  }
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

  const history = new History(stringifySpec(spec));

  function refresh(): void {
    syncMeta();
    history.record(stringifySpec(spec));
    undoBtn.disabled = !history.canUndo;
    redoBtn.disabled = !history.canRedo;
    ol = outline(spec.tree, (p) => {
      selected = p;
      refresh();
    });
    ol.setSelected(selected);
    clear(outlineHost);
    outlineHost.append(ol.el);
    renderProps();
    const found = findProblems(spec.tree);
    ol.setProblems(found);
    const issues = found.map((p) => p.message);
    clear(status);
    if (issues.length) status.append(...issues.map((m) => h('div', {}, m)));
    else status.textContent = `${countNodes(spec.tree)} nodes`;
    status.style.color = issues.length ? 'var(--warn)' : 'var(--muted)';
    agent?.setSpec(spec.tree);
    if (walker) {
      // selecting a node rebuilds the outline; only a real edit ends the session
      if (stringifySpec(spec) !== walkerSnapshot) stopDebug();
      else showDebug(walker);
    }
    persistDraft();
  }

  function showDebug(w: SimWalker): void {
    walker = w;
    walkerSnapshot = stringifySpec(spec);
    ol?.setStatuses(w.statuses());
    ol?.setDebug(w.visited(), w.current ? pathId(w.current.path) : undefined);
    clear(dbgPrompt);
    const cur = w.current;
    if (w.state === 'waiting' && cur) {
      dbgPrompt.append(
        h(
          'div',
          {},
          `Paused at ${cur.node.type} `,
          h('strong', {}, cur.node.name),
          '. Its status?',
        ),
        h(
          'div',
          { class: 'row' },
          ...(['success', 'failure', 'running'] as const).map((st) =>
            h(
              'button',
              {
                class: 'btn small',
                type: 'button',
                onClick: () => {
                  w.resolve(st);
                  showDebug(w);
                },
              },
              st,
            ),
          ),
        ),
      );
    } else if (w.state === 'complete') {
      dbgPrompt.append(h('div', {}, 'Finished: ', h('strong', {}, w.result ?? '')));
    }
    dbgTrace.textContent = w.trace
      .map(
        (s) =>
          `${'  '.repeat(s.depth)}${s.event === 'enter' ? '>' : '<'} ${s.name}${s.status ? ` -> ${s.status}` : ''}`,
      )
      .join('\n');
  }
  function stopDebug(): void {
    walker = undefined;
    clear(dbgPrompt);
    dbgTrace.textContent = '';
    ol?.setDebug(new Set(), undefined);
  }
  stepBtn.addEventListener('click', () => {
    const w = new SimWalker(spec.tree);
    w.step();
    showDebug(w);
  });
  runBtn.addEventListener('click', () => {
    showDebug(runAll(spec.tree, policySel.value as Policy));
  });
  stopBtn.addEventListener('click', stopDebug);

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

  function restore(snapshot: string | undefined): void {
    if (snapshot === undefined) return;
    spec = parseSpec(snapshot);
    nameIn.value = spec.metadata.name;
    descIn.value = spec.metadata.description ?? '';
    tagsIn.value = (spec.metadata.tags ?? []).join(', ');
    while (selected.length > 0 && !getNode(spec.tree, selected)) selected = selected.slice(0, -1);
    refresh();
  }
  undoBtn.addEventListener('click', () => {
    restore(history.undo());
  });
  redoBtn.addEventListener('click', () => {
    restore(history.redo());
  });
  const onKey = (e: KeyboardEvent): void => {
    if (!(e.ctrlKey || e.metaKey) || e.altKey) return;
    const t = e.target;
    if (t instanceof HTMLInputElement || t instanceof HTMLTextAreaElement) return;
    if (t instanceof HTMLSelectElement) return;
    const k = e.key.toLowerCase();
    if (k === 'z') restore(e.shiftKey ? history.redo() : history.undo());
    else if (k === 'y') restore(history.redo());
    else return;
    e.preventDefault();
  };
  document.addEventListener('keydown', onKey);

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
  const applyMove = (res: Path | string): void => {
    if (typeof res === 'string') {
      toast(res, 'err');
      return;
    }
    selected = res;
    refresh();
  };
  dupBtn.addEventListener('click', () => {
    applyMove(duplicateAt(spec.tree, selected));
  });
  outBtn.addEventListener('click', () => {
    applyMove(outdent(spec.tree, selected));
  });
  inBtn.addEventListener('click', () => {
    applyMove(indent(spec.tree, selected));
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
        else status.textContent = r.errors.join(' · ');
      },
      () => {
        toast('Could not reach the validator.', 'err');
      },
    );
  });

  /** Points the scene's agent at the tree just saved, then returns to the scene. */
  const attachToScene = async (sid: string, agent: number, tid: string): Promise<void> => {
    try {
      const sc = await api.getScene(sid);
      const target = sc.agents[agent];
      if (target) target.treeId = tid;
      await api.saveScene(sid, { name: sc.name, visibility: sc.visibility, agents: sc.agents });
      toast('Saved and assigned to the agent.');
      navigate(`/scenes/${sid}`);
    } catch {
      toast('Tree saved, but it could not be assigned to the scene.', 'err');
    }
  };

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
        if (forScene && Number.isInteger(forAgent)) {
          void attachToScene(forScene, forAgent, t.id);
          return;
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
              h(
                'div',
                { class: 'row' },
                h(
                  'button',
                  {
                    class: 'btn small',
                    type: 'button',
                    'data-export': 'copy-code',
                    onClick: () => void copyOut(f.content),
                  },
                  'Copy code',
                ),
                h(
                  'button',
                  {
                    class: 'btn small',
                    type: 'button',
                    'data-export': 'download-code',
                    onClick: () => {
                      downloadText(f.path.split('/').pop() ?? 'code.txt', f.content);
                    },
                  },
                  'Download code',
                ),
              ),
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
    if (agent && !walker) ol?.setStatuses(agent.runtime.statuses);
  }, 120);
  return () => {
    clearInterval(timer);
    document.removeEventListener('keydown', onKey);
    world?.dispose();
  };
}
