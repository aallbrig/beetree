/** Snapshot-based undo/redo. Snapshots are opaque strings (the builder uses YAML). */
export class History {
  private past: string[] = [];
  private future: string[] = [];
  private current: string;

  constructor(
    initial: string,
    private readonly limit = 100,
  ) {
    this.current = initial;
  }

  get canUndo(): boolean {
    return this.past.length > 0;
  }

  get canRedo(): boolean {
    return this.future.length > 0;
  }

  /** Records a new state. Returns false (and changes nothing) when it equals the current one. */
  record(snapshot: string): boolean {
    if (snapshot === this.current) return false;
    this.past.push(this.current);
    if (this.past.length > this.limit) this.past.shift();
    this.current = snapshot;
    this.future = [];
    return true;
  }

  undo(): string | undefined {
    const prev = this.past.pop();
    if (prev === undefined) return undefined;
    this.future.push(this.current);
    this.current = prev;
    return prev;
  }

  redo(): string | undefined {
    const next = this.future.pop();
    if (next === undefined) return undefined;
    this.past.push(this.current);
    this.current = next;
    return next;
  }
}
