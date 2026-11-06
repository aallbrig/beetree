import { api } from '../api';
import { isSignedIn } from '../auth';
import { outline, type Outline } from '../components/outline';
import { clear, h, toast } from '../dom';
import type { Ctx, Cleanup } from '../router';
import { populate, type LoadedAgent } from '../scene/loader';
import { DEFAULT_SCENE_ID, SCENES, sceneDefById } from '../scene/registry';
import { World } from '../scene/world';
import type { SceneAgent } from '../types';

export async function exploreView(ctx: Ctx): Promise<Cleanup | undefined> {
  const canvas = h('canvas', { 'aria-label': '3D scene of agents running behavior trees' });
  const intruderBtn = h(
    'button',
    { class: 'btn small', type: 'button', 'aria-pressed': 'true' },
    'Intruder',
  );
  const resetBtn = h('button', { class: 'btn small', type: 'button' }, '↺ Reset');
  const stage = h(
    'div',
    { class: 'stage' },
    canvas,
    h('div', { class: 'overlay' }, intruderBtn, resetBtn),
    h(
      'div',
      { class: 'hintbar' },
      'Click the ground to move the red intruder. Drag to orbit, scroll to zoom.',
    ),
  );

  const sceneSelect = h(
    'select',
    { 'aria-label': 'Scene' },
    ...SCENES.map((s) => h('option', { value: s.id }, s.title)),
  );
  const health = h('input', {
    type: 'range',
    min: '0',
    max: '100',
    value: '100',
    'aria-label': 'Agent health',
  });
  const treeList = h('div', { class: 'stack' });
  const panel = h(
    'div',
    { class: 'card stack' },
    h('h2', {}, 'Explore'),
    h(
      'p',
      { class: 'muted' },
      'Agents below run real behavior trees. Move the intruder and change health to watch the trees react.',
    ),
    h('div', {}, h('label', {}, 'Scene'), sceneSelect),
    h('div', {}, h('label', {}, 'Agent health (try dropping it)'), health),
    treeList,
    h(
      'p',
      { class: 'small' },
      isSignedIn()
        ? h('a', { href: '#/scenes' }, 'Arrange your own agents →')
        : h('a', { href: '#/signup' }, 'Sign up to build your own trees →'),
    ),
  );
  ctx.root.append(h('div', { class: 'studio' }, panel, stage));

  let world: World;
  try {
    world = await World.create(canvas);
  } catch {
    clear(stage);
    stage.append(
      h(
        'p',
        { style: 'padding:24px' },
        'Your browser could not start WebGL, so the 3D scene is unavailable.',
      ),
    );
    return undefined;
  }
  world.mode = 'beacon';
  world.setBeacon({ x: 0, z: 2 });
  world.start();

  let loaded: LoadedAgent[] = [];
  interface TreeSection {
    details: HTMLDetailsElement;
    outline: Outline;
  }
  const sections = new Map<string, TreeSection>();

  /** Highlight one agent in the scene and bring its tree into view. */
  const showAgent = (id: string): void => {
    const target = world.agents.get(id);
    for (const other of world.agents.values()) other.setSelected(other === target);
    for (const [key, sec] of sections) sec.details.classList.toggle('selected', key === id);
    const sec = sections.get(id);
    if (sec) {
      sec.details.open = true;
      sec.details.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
    }
  };

  const load = async (agents: SceneAgent[]): Promise<void> => {
    loaded = await populate(world, agents);
    clear(treeList);
    sections.clear();
    for (const a of loaded) {
      const agent = world.agents.get(a.id);
      if (!agent) continue;
      const tree = outline(agent.opts.spec);
      const details = h(
        'details',
        { class: 'tree-section', open: true },
        h(
          'summary',
          {
            onClick: () => {
              showAgent(a.id);
            },
          },
          h('strong', {}, `${a.name}'s tree`),
        ),
        tree.el,
      );
      sections.set(a.id, { details, outline: tree });
      treeList.append(details);
    }
    const first = loaded[0];
    if (first) showAgent(first.id);
  };
  world.onAgentClick = showAgent;

  const myScenes = new Map<string, SceneAgent[]>();
  const sceneParam = ctx.query.get('scene');
  try {
    if (isSignedIn()) {
      for (const s of (await api.scenes()).items) {
        myScenes.set(s.id, s.agents);
        sceneSelect.append(h('option', { value: s.id }, s.name));
      }
    }
    if (sceneParam && !myScenes.has(sceneParam)) {
      const s = await api.getScene(sceneParam);
      myScenes.set(s.id, s.agents);
      sceneSelect.append(h('option', { value: s.id }, `${s.name} (shared)`));
    }
  } catch {
    toast('Could not load saved scenes.', 'err');
  }
  const agentsFor = (id: string): SceneAgent[] =>
    myScenes.get(id) ?? sceneDefById(id)?.agents ?? sceneDefById(DEFAULT_SCENE_ID)?.agents ?? [];
  if (sceneParam && sceneDefById(sceneParam)) sceneSelect.value = sceneParam;
  if (sceneParam && myScenes.has(sceneParam)) sceneSelect.value = sceneParam;
  await load(agentsFor(sceneSelect.value));

  sceneSelect.addEventListener('change', () => {
    load(agentsFor(sceneSelect.value)).catch(() => {
      toast('Could not load that scene.', 'err');
    });
  });
  health.addEventListener('input', () => {
    for (const a of world.agents.values()) a.health = Number(health.value);
  });
  intruderBtn.addEventListener('click', () => {
    const on = world.beacon === undefined;
    world.setBeacon(on ? { x: 0, z: 2 } : undefined);
    intruderBtn.setAttribute('aria-pressed', String(on));
  });
  resetBtn.addEventListener('click', () => {
    world.resetAgents();
    health.value = '100';
  });

  const timer = setInterval(() => {
    for (const [id, sec] of sections) {
      const agent = world.agents.get(id);
      if (agent && sec.details.open) sec.outline.setStatuses(agent.runtime.statuses);
    }
  }, 120);

  return () => {
    clearInterval(timer);
    world.dispose();
  };
}
