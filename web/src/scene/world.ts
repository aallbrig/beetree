import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { GLTFLoader, type GLTF } from 'three/examples/jsm/loaders/GLTFLoader.js';
import { clone as cloneSkinned } from 'three/examples/jsm/utils/SkeletonUtils.js';
import { Runtime, type Status } from '../engine/bt';
import type { NodeSpec } from '../spec';
import { AgentBehaviors, type Anim, type Body, type Vec2 } from './behaviors';

export const CHARACTERS = [
  'character-a',
  'character-b',
  'character-c',
  'character-d',
  'character-e',
  'character-f',
  'character-g',
  'character-h',
] as const;
const PROPS = [
  'tree_default',
  'tree_oak',
  'tree_fat',
  'tree_pineDefaultA',
  'tree_pineTallA',
  'tree_simple',
  'tree_thin',
  'rock_largeA',
  'rock_smallA',
  'rock_tallA',
  'stump_round',
  'log',
  'plant_bush',
  'mushroom_red',
  'flower_redA',
  'flower_yellowA',
  'flower_purpleA',
  'grass_large',
] as const;
const HALF = 45;

const loader = new GLTFLoader();
const cache = new Map<string, Promise<GLTF>>();
function loadGltf(url: string): Promise<GLTF> {
  let p = cache.get(url);
  if (!p) {
    p = loader.loadAsync(url);
    cache.set(url, p);
  }
  return p;
}
const asset = (path: string): string => `${import.meta.env.BASE_URL}assets/kenney/${path}`;

function mulberry32(seed: number): () => number {
  let a = seed;
  return () => {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

const STATUS_COLOR: Record<Status, string> = {
  success: '#4ade80',
  failure: '#f87171',
  running: '#fbbf24',
};

function makeLabel(): { sprite: THREE.Sprite; set: (text: string, color: string) => void } {
  const canvas = document.createElement('canvas');
  canvas.width = 512;
  canvas.height = 96;
  const tex = new THREE.CanvasTexture(canvas);
  tex.colorSpace = THREE.SRGBColorSpace;
  const sprite = new THREE.Sprite(
    new THREE.SpriteMaterial({ map: tex, depthTest: false, transparent: true }),
  );
  sprite.scale.set(5, 0.94, 1);
  sprite.renderOrder = 10;
  let last = '';
  return {
    sprite,
    set(text, color) {
      const key = text + color;
      if (key === last) return;
      last = key;
      const ctx = canvas.getContext('2d');
      if (!ctx) return;
      ctx.clearRect(0, 0, 512, 96);
      ctx.fillStyle = 'rgba(20,18,12,0.78)';
      ctx.beginPath();
      ctx.roundRect(8, 12, 496, 72, 30);
      ctx.fill();
      ctx.fillStyle = color;
      ctx.beginPath();
      ctx.arc(46, 48, 14, 0, Math.PI * 2);
      ctx.fill();
      ctx.fillStyle = '#fff7e0';
      ctx.font = '600 38px system-ui, sans-serif';
      ctx.textBaseline = 'middle';
      ctx.fillText(text.length > 22 ? `${text.slice(0, 21)}…` : text, 76, 50);
      tex.needsUpdate = true;
    },
  };
}

export interface AgentOptions {
  id: string;
  name: string;
  model: string;
  x: number;
  z: number;
  spec: NodeSpec;
}

export class Agent implements Body {
  readonly group = new THREE.Group();
  get pos(): Vec2 {
    return this.group.position;
  }
  health = 100;
  alert = 0;
  waypoint: Vec2;
  runtime: Runtime;
  rootStatus: Status = 'running';
  said = '';
  private readonly behaviors: AgentBehaviors;
  private mixer: THREE.AnimationMixer | undefined;
  private actions = new Map<string, THREE.AnimationAction>();
  private currentAnim = '';
  private wantAnim: Anim = 'idle';
  private heading = 0;
  private waypoints: Vec2[];
  private wpIndex = 0;
  private readonly label = makeLabel();
  private readonly ring: THREE.Mesh;
  private dead = false;

  constructor(
    private readonly world: World,
    readonly opts: AgentOptions,
    gltf: GLTF,
  ) {
    this.group.position.set(opts.x, 0, opts.z);
    this.waypoints = [0, 1, 2, 3].map((i) => ({
      x: clamp(opts.x + Math.cos((i * Math.PI) / 2) * 6),
      z: clamp(opts.z + Math.sin((i * Math.PI) / 2) * 6),
    }));
    this.waypoint = this.waypoints[0] ?? { x: opts.x, z: opts.z };

    const model = cloneSkinned(gltf.scene);
    model.scale.setScalar(2);
    model.traverse((o) => {
      if (o instanceof THREE.Mesh) o.castShadow = true;
    });
    this.group.add(model);
    this.mixer = new THREE.AnimationMixer(model);
    for (const clip of gltf.animations) this.actions.set(clip.name, this.mixer.clipAction(clip));
    this.play('idle');

    this.ring = new THREE.Mesh(
      new THREE.RingGeometry(0.9, 1.1, 40),
      new THREE.MeshBasicMaterial({
        color: '#fbbf24',
        side: THREE.DoubleSide,
        transparent: true,
        opacity: 0.9,
      }),
    );
    this.ring.rotation.x = -Math.PI / 2;
    this.ring.position.y = 0.03;
    this.ring.visible = false;
    this.group.add(this.ring);
    this.label.sprite.position.y = 4.2;
    this.group.add(this.label.sprite);

    this.behaviors = new AgentBehaviors(this);
    this.runtime = new Runtime(opts.spec, this.behaviors);
  }

  get target(): Vec2 | undefined {
    return this.world.beacon;
  }

  setSelected(on: boolean): void {
    this.ring.visible = on;
  }

  setSpec(spec: NodeSpec): void {
    this.runtime.reset();
    this.runtime = new Runtime(spec, this.behaviors);
  }

  resetState(): void {
    this.health = 100;
    this.alert = 0;
    this.dead = false;
    this.runtime.reset();
  }

  update(dt: number): void {
    if (this.health <= 0 && !this.dead) {
      this.dead = true;
    }
    this.wantAnim = 'idle';
    this.rootStatus = this.runtime.tick(dt);
    const color = STATUS_COLOR[this.rootStatus];
    this.label.set(`${this.opts.name}: ${this.said || '…'}`, color);
    this.said = '';
    this.applyAnim();
    this.mixer?.update(dt);
    this.group.rotation.y = lerpAngle(this.group.rotation.y, this.heading, Math.min(1, dt * 10));
  }

  // ---- Body --------------------------------------------------------------
  moveToward(p: Vec2, speed: number, dt: number): number {
    const dx = p.x - this.pos.x;
    const dz = p.z - this.pos.z;
    const d = Math.hypot(dx, dz);
    if (d > 0.001) {
      this.heading = Math.atan2(dx, dz);
      const step = Math.min(d, speed * dt);
      this.group.position.x = clamp(this.pos.x + (dx / d) * step);
      this.group.position.z = clamp(this.pos.z + (dz / d) * step);
    }
    return d;
  }
  moveAwayFrom(p: Vec2, speed: number, dt: number): void {
    const dx = this.pos.x - p.x;
    const dz = this.pos.z - p.z;
    const d = Math.hypot(dx, dz) || 1;
    this.heading = Math.atan2(dx, dz);
    this.group.position.x = clamp(this.pos.x + (dx / d) * speed * dt);
    this.group.position.z = clamp(this.pos.z + (dz / d) * speed * dt);
  }
  faceToward(p: Vec2): void {
    this.heading = Math.atan2(p.x - this.pos.x, p.z - this.pos.z);
  }
  nextWaypoint(): void {
    this.wpIndex = (this.wpIndex + 1) % this.waypoints.length;
    this.waypoint = this.waypoints[this.wpIndex] ?? this.waypoint;
  }
  play(anim: Anim): void {
    this.wantAnim = anim;
  }
  say(text: string): void {
    this.said = text;
  }
  strike(kind: 'melee' | 'ranged'): void {
    const t = this.target;
    if (t) this.world.spawnFlash(kind === 'ranged' ? this.pos : t, t, kind === 'ranged');
  }

  private applyAnim(): void {
    if (this.wantAnim === this.currentAnim) return;
    this.play2(this.wantAnim);
  }
  private play2(name: string): void {
    const next = this.actions.get(name) ?? this.actions.get('idle');
    if (!next) return;
    const prev = this.actions.get(this.currentAnim);
    next.reset().fadeIn(0.18).play();
    prev?.fadeOut(0.18);
    this.currentAnim = name;
  }

  dispose(): void {
    this.mixer?.stopAllAction();
    this.group.removeFromParent();
  }
}

const clamp = (v: number): number => Math.max(-HALF, Math.min(HALF, v));
function lerpAngle(a: number, b: number, t: number): number {
  let d = (b - a) % (Math.PI * 2);
  if (d > Math.PI) d -= Math.PI * 2;
  if (d < -Math.PI) d += Math.PI * 2;
  return a + d * t;
}

export type PointerMode = 'beacon' | 'place' | 'none';

interface Flash {
  mesh: THREE.Mesh;
  mat: THREE.MeshBasicMaterial;
  from: THREE.Vector3;
  to: THREE.Vector3;
  life: number;
  max: number;
}

export class World {
  readonly scene = new THREE.Scene();
  readonly agents = new Map<string, Agent>();
  beacon: Vec2 | undefined;
  mode: PointerMode = 'beacon';
  onGroundClick: ((x: number, z: number) => void) | undefined;
  onAgentClick: ((id: string) => void) | undefined;

  private readonly renderer: THREE.WebGLRenderer;
  private readonly camera = new THREE.PerspectiveCamera(50, 1, 0.1, 400);
  private readonly controls: OrbitControls;
  private last = 0;
  private readonly ground: THREE.Mesh;
  private readonly beaconMesh = new THREE.Group();
  private readonly flashes: Flash[] = [];
  private readonly ray = new THREE.Raycaster();
  private raf = 0;
  private running = false;
  private readonly ro: ResizeObserver;
  private downAt: { x: number; y: number } | undefined;

  constructor(private readonly canvas: HTMLCanvasElement) {
    this.renderer = new THREE.WebGLRenderer({ canvas, antialias: true });
    this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    this.renderer.shadowMap.enabled = true;
    this.scene.background = new THREE.Color('#bfe3f5');
    this.scene.fog = new THREE.Fog('#bfe3f5', 40, 110);

    this.scene.add(new THREE.HemisphereLight('#ffffff', '#6b8f4a', 1.1));
    const sun = new THREE.DirectionalLight('#fff3d6', 2.2);
    sun.position.set(18, 28, 12);
    sun.castShadow = true;
    sun.shadow.mapSize.set(1024, 1024);
    Object.assign(sun.shadow.camera, {
      left: -35,
      right: 35,
      top: 35,
      bottom: -35,
      near: 1,
      far: 90,
    });
    this.scene.add(sun);

    this.ground = new THREE.Mesh(
      new THREE.CircleGeometry(60, 64),
      new THREE.MeshStandardMaterial({ color: '#7fb05b', roughness: 1 }),
    );
    this.ground.rotation.x = -Math.PI / 2;
    this.ground.receiveShadow = true;
    this.scene.add(this.ground);

    const cone = new THREE.Mesh(
      new THREE.ConeGeometry(0.5, 1.4, 16),
      new THREE.MeshStandardMaterial({ color: '#ef4444', emissive: '#7f1d1d' }),
    );
    cone.rotation.x = Math.PI;
    cone.position.y = 1.6;
    const halo = new THREE.Mesh(
      new THREE.RingGeometry(0.7, 0.95, 32),
      new THREE.MeshBasicMaterial({ color: '#ef4444', side: THREE.DoubleSide }),
    );
    halo.rotation.x = -Math.PI / 2;
    halo.position.y = 0.04;
    this.beaconMesh.add(cone, halo);
    this.beaconMesh.visible = false;
    this.scene.add(this.beaconMesh);

    this.camera.position.set(0, 13, 19);
    this.controls = new OrbitControls(this.camera, canvas);
    this.controls.enableDamping = true;
    this.controls.maxPolarAngle = Math.PI / 2.15;
    this.controls.minDistance = 6;
    this.controls.maxDistance = 70;
    this.controls.target.set(0, 1, 0);

    canvas.addEventListener('pointerdown', this.onDown);
    canvas.addEventListener('pointerup', this.onUp);
    this.ro = new ResizeObserver(() => {
      this.resize();
    });
    this.ro.observe(canvas);
    this.resize();
  }

  /** Create a world and populate it with scenery. Throws if WebGL is unavailable. */
  static async create(canvas: HTMLCanvasElement): Promise<World> {
    const w = new World(canvas);
    await w.scatterProps();
    return w;
  }

  private async scatterProps(): Promise<void> {
    const rand = mulberry32(7);
    const loaded = await Promise.all(
      PROPS.map((n) => loadGltf(asset(`nature/${n}.glb`)).catch(() => undefined)),
    );
    for (let i = 0; i < 70; i++) {
      const idx = Math.floor(rand() * PROPS.length);
      const g = loaded[idx];
      if (!g) continue;
      const a = rand() * Math.PI * 2;
      const r = 12 + rand() * 32;
      const prop = g.scene.clone(true);
      prop.position.set(Math.cos(a) * r, 0, Math.sin(a) * r);
      prop.rotation.y = rand() * Math.PI * 2;
      prop.scale.setScalar(2.6 + rand() * 1.6);
      prop.traverse((o) => {
        if (o instanceof THREE.Mesh) {
          o.castShadow = true;
          o.receiveShadow = true;
        }
      });
      this.scene.add(prop);
    }
  }

  async addAgent(opts: AgentOptions): Promise<Agent> {
    this.removeAgent(opts.id);
    const model = (CHARACTERS as readonly string[]).includes(opts.model)
      ? opts.model
      : CHARACTERS[0];
    const gltf = await loadGltf(asset(`characters/${model}.glb`));
    const agent = new Agent(this, opts, gltf);
    this.agents.set(opts.id, agent);
    this.scene.add(agent.group);
    return agent;
  }

  removeAgent(id: string): void {
    this.agents.get(id)?.dispose();
    this.agents.delete(id);
  }

  setBeacon(p: Vec2 | undefined): void {
    this.beacon = p;
    this.beaconMesh.visible = p !== undefined;
    if (p) this.beaconMesh.position.set(p.x, 0, p.z);
  }

  spawnFlash(from: Vec2, to: Vec2, travel: boolean): void {
    const mat = new THREE.MeshBasicMaterial({
      color: travel ? '#fde047' : '#fb923c',
      transparent: true,
      opacity: 0.9,
    });
    const mesh = new THREE.Mesh(new THREE.SphereGeometry(travel ? 0.18 : 0.45, 12, 12), mat);
    const a = new THREE.Vector3(from.x, 1.4, from.z);
    const b = new THREE.Vector3(to.x, 1.2, to.z);
    mesh.position.copy(travel ? a : b);
    this.scene.add(mesh);
    this.flashes.push({ mesh, mat, from: a, to: b, life: 0, max: travel ? 0.35 : 0.25 });
  }

  resetAgents(): void {
    for (const a of this.agents.values()) a.resetState();
  }

  start(): void {
    if (this.running) return;
    this.running = true;
    this.last = performance.now();
    const loop = (): void => {
      if (!this.running) return;
      this.raf = requestAnimationFrame(loop);
      const now = performance.now();
      this.frame(Math.min((now - this.last) / 1000, 0.05));
      this.last = now;
    };
    loop();
  }

  private frame(dt: number): void {
    for (const a of this.agents.values()) a.update(dt);
    this.separate();
    for (let i = this.flashes.length - 1; i >= 0; i--) {
      const f = this.flashes[i];
      if (!f) continue;
      f.life += dt;
      const k = f.life / f.max;
      f.mesh.position.lerpVectors(f.from, f.to, Math.min(k, 1));
      f.mat.opacity = Math.max(0, 1 - k);
      if (k >= 1) {
        f.mesh.removeFromParent();
        this.flashes.splice(i, 1);
      }
    }
    this.beaconMesh.rotation.y += dt * 1.5;
    this.controls.update();
    this.renderer.render(this.scene, this.camera);
  }

  /** Keep agents from standing inside each other: push overlapping pairs apart. */
  private separate(): void {
    const list = [...this.agents.values()];
    const min = 1.6;
    for (let i = 0; i < list.length; i++) {
      for (let j = i + 1; j < list.length; j++) {
        const a = list[i];
        const b = list[j];
        if (!a || !b) continue;
        const dx = b.pos.x - a.pos.x;
        const dz = b.pos.z - a.pos.z;
        const d = Math.hypot(dx, dz);
        if (d >= min) continue;
        const push = (min - d) / 2;
        const nx = d > 0.001 ? dx / d : 1;
        const nz = d > 0.001 ? dz / d : 0;
        a.group.position.x = clamp(a.group.position.x - nx * push);
        a.group.position.z = clamp(a.group.position.z - nz * push);
        b.group.position.x = clamp(b.group.position.x + nx * push);
        b.group.position.z = clamp(b.group.position.z + nz * push);
      }
    }
  }

  private resize(): void {
    const w = this.canvas.clientWidth || 800;
    const h = this.canvas.clientHeight || 500;
    this.renderer.setSize(w, h, false);
    this.camera.aspect = w / h;
    this.camera.updateProjectionMatrix();
  }

  private ndc(e: PointerEvent): THREE.Vector2 {
    const r = this.canvas.getBoundingClientRect();
    return new THREE.Vector2(
      ((e.clientX - r.left) / r.width) * 2 - 1,
      -((e.clientY - r.top) / r.height) * 2 + 1,
    );
  }

  private readonly onDown = (e: PointerEvent): void => {
    this.downAt = { x: e.clientX, y: e.clientY };
  };

  private readonly onUp = (e: PointerEvent): void => {
    const d = this.downAt;
    this.downAt = undefined;
    if (!d || Math.hypot(e.clientX - d.x, e.clientY - d.y) > 5) return; // it was an orbit drag
    this.ray.setFromCamera(this.ndc(e), this.camera);
    const agentHit = this.ray.intersectObjects(
      [...this.agents.values()].map((a) => a.group),
      true,
    )[0];
    if (agentHit && this.onAgentClick) {
      for (const [id, a] of this.agents)
        if (a.group.getObjectById(agentHit.object.id)) this.onAgentClick(id);
      return;
    }
    const hit = this.ray.intersectObject(this.ground)[0];
    if (!hit) return;
    const p = { x: clamp(hit.point.x), z: clamp(hit.point.z) };
    if (this.mode === 'beacon') this.setBeacon(p);
    else if (this.mode === 'place') this.onGroundClick?.(p.x, p.z);
  };

  dispose(): void {
    this.running = false;
    cancelAnimationFrame(this.raf);
    this.ro.disconnect();
    this.canvas.removeEventListener('pointerdown', this.onDown);
    this.canvas.removeEventListener('pointerup', this.onUp);
    for (const a of this.agents.values()) a.dispose();
    this.agents.clear();
    this.controls.dispose();
    this.renderer.dispose();
  }
}
