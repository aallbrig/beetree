/** localStorage key for the builder's unsaved draft tree. */
export const DRAFT_KEY = 'beetree.draft';

/** Remove the unsaved draft so the next person on this browser starts clean. */
export function clearDraft(): void {
  try {
    localStorage.removeItem(DRAFT_KEY);
  } catch {
    // storage unavailable
  }
}
