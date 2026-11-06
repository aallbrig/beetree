/** The 100 preset profile pictures: 10 shapes in 10 colour pairs, drawn as inline SVG so no files ship. */

const PALETTES: [string, string][] = [
  ['#f5b700', '#2b2410'],
  ['#e8a300', '#fff7dc'],
  ['#2f9e6e', '#e8fff4'],
  ['#1f7a8c', '#e6f7fb'],
  ['#3b6fd4', '#eaf0ff'],
  ['#7b4fd0', '#f1eaff'],
  ['#c2410c', '#fff0e6'],
  ['#d63384', '#ffe9f4'],
  ['#4b5563', '#f3f4f6'],
  ['#14532d', '#d9f99d'],
];

// Each shape is drawn on a 64x64 canvas in the foreground colour.
const SHAPES: string[] = [
  '<circle cx="32" cy="32" r="16"/>',
  '<rect x="17" y="17" width="30" height="30" rx="4"/>',
  '<polygon points="32,14 50,48 14,48"/>',
  '<polygon points="32,12 52,32 32,52 12,32"/>',
  '<polygon points="32,12 49,22 49,42 32,52 15,42 15,22"/>',
  '<polygon points="32,11 38,26 54,26 41,36 46,52 32,42 18,52 23,36 10,26 26,26"/>',
  '<path d="M32 50V30M32 30c-10 0-14-8-14-14 10 0 14 6 14 14zm0 8c10 0 14-8 14-14-10 0-14 6-14 14z"/>',
  '<path d="M12 40c6-12 14-12 20 0s14 12 20 0v10H12z"/>',
  '<path d="M36 10 20 36h12l-4 18 16-26H32z"/>',
  '<path d="M44 14a20 20 0 1 0 0 36 16 16 0 0 1 0-36z"/>',
];

export const AVATAR_COUNT = 100;

/** Ids a001 to a100. */
export const AVATAR_IDS: string[] = Array.from(
  { length: AVATAR_COUNT },
  (_, i) => `a${String(i + 1).padStart(3, '0')}`,
);

export function isAvatarId(id: string): boolean {
  return AVATAR_IDS.includes(id);
}

/** SVG markup for an avatar id; unknown ids get the first picture. */
export function avatarSvg(id: string): string {
  const n = Math.max(0, AVATAR_IDS.indexOf(id));
  const [bg, fg] = PALETTES[Math.floor(n / SHAPES.length)] ?? PALETTES[0] ?? ['#ccc', '#333'];
  const shape = SHAPES[n % SHAPES.length] ?? '';
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64" width="64" height="64" role="img"><rect width="64" height="64" rx="32" fill="${bg}"/><g fill="${fg}">${shape}</g></svg>`;
}

/** An <img>-ready data URI. */
export function avatarSrc(id: string): string {
  return `data:image/svg+xml;utf8,${encodeURIComponent(avatarSvg(id))}`;
}
