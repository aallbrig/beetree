import { api, ApiError } from '../api';
import { clear, h, toast } from '../dom';
import { navigate, type Cleanup, type Ctx } from '../router';
import { populate } from '../scene/loader';
import { CHARACTERS, World } from '../scene/world';
import { SAMPLES } from '../samples';
import type { SceneAgent } from '../types';

/** The 3D scene studio: position agents, give each a behavior tree, save and share. */
export async function scenesView(ctx: Ctx): Promise<Cleanup | undefined> {
  const sceneId = ctx.params[0];
  const existing = sceneId ? await api.getScene(sceneId) : undefined;
  const mine = (await api.myTrees()).items;
  let agents: SceneAgent[] = existing?.agents ?? [
    { name: 'Agent 1', treeId: 'sample:guard', model: 'character-a', x: 0, z: 0 },
  ];
  let selected = 0;

  const canvas = h('canvas', { 'aria-label': '3D scene editor' });
  const stage = h(
    'div',
    { class: 'stage' },
    canvas,
    h(
      'div',
      { class: 'hintbar' },
      'Click the ground to move the selected agent. Click an agent to select it.',
    ),
  );

  const nameIn = h('input', {
    type: 'text',
    value: existing?.name ?? 'My scene',
    maxlength: '80',
    'aria-label': 'Scene name',
  });
  const visIn = h(
    'select',
    { 'aria-label': 'Visibility' },
    h('option', { value: 'private' }, 'Private'),
    h('option', { value: 'public' }, 'Public (shareable link)'),
  );
  visIn.value = existing?.visibility ?? 'private';
  const list = h('div', { class: 'stack' });
  const form = h('div', { class: 'stack' });
  const addBtn = h('button', { class: 'btn small', type: 'button' }, '+ Add agent');
  const saveBtn = h('button', { class: 'btn primary', type: 'button' }, 'Save scene');

  ctx.root.append(
    h(
      'div',
      { class: 'studio' },
      h(
        'div',
        { class: 'card stack' },
        h('h2', {}, existing ? 'Edit scene' : 'New scene'),
        h('div', {}, h('label', {}, 'Name'), nameIn),
        h('div', {}, h('label', {}, 'Visibility'), visIn),
        h('div', { class: 'row' }, h('h3', { style: 'margin:0' }, 'Agents'), addBtn),
        list,
        form,
        h(
          'div',
          { class: 'row' },
          saveBtn,
          existing
            ? h('a', { class: 'btn', href: `#/explore?scene=${existing.id}` }, 'Watch it')
            : null,
        ),
      ),
      stage,
    ),
  );

  let world: World;
  try {
    world = await World.create(canvas);
  } catch {
    clear(stage);
    stage.append(
      h('p', { style: 'padding:24px' }, 'WebGL is unavailable, so the 3D editor cannot run.'),
    );
    return undefined;
  }
  world.mode = 'place';
  world.setBeacon({ x: 0, z: 14 });
  world.start();

  let savedId = sceneId;
  /** Saves the scene (creating it if new) and resolves its id. */
  const ensureSaved = async (): Promise<string> => {
    const s = await api.saveScene(savedId, {
      name: nameIn.value.trim() || 'My scene',
      visibility: visIn.value,
      agents,
    });
    savedId = s.id;
    return s.id;
  };

  const treeOptions = (): HTMLOptionElement[] => [
    ...SAMPLES.map((s) => h('option', { value: s.id }, `Sample: ${s.title}`)),
    ...mine.map((t) => h('option', { value: t.id }, `Mine: ${t.name}`)),
  ];

  const reload = async (): Promise<void> => {
    await populate(world, agents);
    highlight();
  };
  const highlight = (): void => {
    [...world.agents.values()].forEach((a, i) => {
      a.setSelected(i === selected);
    });
  };

  const paint = (): void => {
    clear(list);
    agents.forEach((a, i) => {
      list.append(
        h(
          'button',
          {
            class: 'btn small',
            type: 'button',
            'aria-pressed': String(i === selected),
            onClick: () => {
              selected = i;
              paint();
              highlight();
            },
          },
          a.name,
        ),
      );
    });
    clear(form);
    const a = agents[selected];
    if (!a) return;
    const name = h('input', {
      type: 'text',
      value: a.name,
      maxlength: '40',
      'aria-label': 'Agent name',
    });
    const model = h(
      'select',
      { 'aria-label': 'Character' },
      ...CHARACTERS.map((c) => h('option', { value: c }, c.replace('character-', 'Character '))),
    );
    model.value = a.model;
    const tree = h('select', { 'aria-label': 'Behavior tree' }, ...treeOptions());
    tree.value = a.treeId ?? 'sample:guard';
    const remove = h('button', { class: 'btn small danger', type: 'button' }, 'Remove agent');
    name.addEventListener('change', () => {
      a.name = name.value.trim() || a.name;
      paint();
      void reload();
    });
    model.addEventListener('change', () => {
      a.model = model.value;
      void reload();
    });
    tree.addEventListener('change', () => {
      a.treeId = tree.value;
      void reload();
    });
    const author = (path: string): void => {
      ensureSaved().then(
        (id) => {
          navigate(`${path}${path.includes('?') ? '&' : '?'}scene=${id}&agent=${selected}`);
        },
        (e: unknown) => {
          toast(e instanceof ApiError ? e.message : 'Could not save the scene first.', 'err');
        },
      );
    };
    const isMine = mine.some((t) => t.id === a.treeId);
    const newTree = h(
      'button',
      {
        class: 'btn small',
        type: 'button',
        onClick: () => {
          author('/builder');
        },
      },
      'Write a new tree for this agent',
    );
    const editTree = isMine
      ? h(
          'button',
          {
            class: 'btn small',
            type: 'button',
            onClick: () => {
              author(`/builder/${a.treeId ?? ''}`);
            },
          },
          'Edit this tree',
        )
      : null;
    remove.addEventListener('click', () => {
      agents = agents.filter((_, i) => i !== selected);
      selected = Math.max(0, selected - 1);
      paint();
      void reload();
    });
    form.append(
      h('div', {}, h('label', {}, 'Name'), name),
      h('div', {}, h('label', {}, 'Character'), model),
      h('div', {}, h('label', {}, 'Behavior tree'), tree),
      h('div', { class: 'row' }, newTree, editTree),
      h('p', { class: 'hint' }, `Position: ${a.x.toFixed(1)}, ${a.z.toFixed(1)}`),
      remove,
    );
  };

  world.onGroundClick = (x, z) => {
    const a = agents[selected];
    if (!a) return;
    a.x = x;
    a.z = z;
    paint();
    void reload();
  };
  world.onAgentClick = (id) => {
    const idx = Number(id.slice(1));
    if (!Number.isNaN(idx)) {
      selected = idx;
      paint();
      highlight();
    }
  };
  addBtn.addEventListener('click', () => {
    if (agents.length >= 24) {
      toast('A scene holds up to 24 agents.', 'err');
      return;
    }
    agents.push({
      name: `Agent ${agents.length + 1}`,
      treeId: 'sample:patrol',
      model: CHARACTERS[agents.length % CHARACTERS.length] ?? 'character-a',
      x: (agents.length % 5) * 4 - 8,
      z: Math.floor(agents.length / 5) * 4 - 4,
    });
    selected = agents.length - 1;
    paint();
    void reload();
  });
  saveBtn.addEventListener('click', () => {
    saveBtn.disabled = true;
    api
      .saveScene(sceneId, { name: nameIn.value.trim(), visibility: visIn.value, agents })
      .then((s) => {
        toast('Scene saved.');
        if (!sceneId) navigate(`/scenes/${s.id}`);
      })
      .catch((e: unknown) => {
        toast(e instanceof ApiError ? e.message : 'Could not save.', 'err');
      })
      .finally(() => {
        saveBtn.disabled = false;
      });
  });

  paint();
  await reload();
  return () => {
    world.dispose();
  };
}
