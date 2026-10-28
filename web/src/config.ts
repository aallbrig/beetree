export interface AppConfig {
  apiUrl: string;
  region: string;
  clientId: string;
  userPoolId: string;
}

function isConfig(v: unknown): v is AppConfig {
  if (typeof v !== 'object' || v === null) return false;
  const o = v as Record<string, unknown>;
  return ['apiUrl', 'region', 'clientId', 'userPoolId'].every(
    (k) => typeof o[k] === 'string' && o[k] !== '',
  );
}

let cached: AppConfig | undefined;

/** Load runtime config written by the deploy (so the bundle is environment-neutral). */
export async function loadConfig(): Promise<AppConfig> {
  if (cached) return cached;
  const res = await fetch(`${import.meta.env.BASE_URL}config.json`, { cache: 'no-cache' });
  const data: unknown = await res.json();
  if (!isConfig(data)) throw new Error('config.json is missing required fields');
  cached = data;
  return data;
}
