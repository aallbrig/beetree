/** Who a damage timer drains: one agent id, or every agent of a faction. */
export type TimerTarget = { agent: string } | { faction: string };

export interface DamageTimerConfig {
  on: boolean;
  /** Seconds between ticks, 1 to 20. */
  period: number;
  /** Health removed per tick. */
  damage: number;
  target: TimerTarget;
}

export const TIMER_DEFAULTS = { period: 5, damage: 10, minPeriod: 1, maxPeriod: 20 } as const;

export interface Damageable {
  id: string;
  faction?: string;
  health: number;
}

/** Drains health on a fixed period while on. Pure state so it is testable without a renderer. */
export class DamageTimer {
  on = false;
  period: number = TIMER_DEFAULTS.period;
  damage: number = TIMER_DEFAULTS.damage;
  target: TimerTarget = { faction: '' };
  private elapsed = 0;

  setPeriod(seconds: number): void {
    this.period = Math.min(TIMER_DEFAULTS.maxPeriod, Math.max(TIMER_DEFAULTS.minPeriod, seconds));
  }

  setOn(on: boolean): void {
    if (on && !this.on) this.elapsed = 0;
    this.on = on;
  }

  /** Advance by dt seconds; returns how many ticks fired. */
  update(dt: number, agents: Iterable<Damageable>): number {
    if (!this.on) return 0;
    this.elapsed += dt;
    let ticks = 0;
    while (this.elapsed >= this.period) {
      this.elapsed -= this.period;
      ticks++;
      for (const a of agents) {
        if (!this.hits(a)) continue;
        a.health = Math.max(0, a.health - this.damage);
      }
    }
    return ticks;
  }

  private hits(a: Damageable): boolean {
    return 'agent' in this.target ? a.id === this.target.agent : a.faction === this.target.faction;
  }
}
