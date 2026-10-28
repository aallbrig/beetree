import { loadConfig } from './config';

const STORE_KEY = 'beetree.session';
const REFRESH_SKEW_MS = 60_000;

export interface Session {
  idToken: string;
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  username: string;
}

export class AuthError extends Error {
  constructor(
    readonly code: string,
    message: string,
  ) {
    super(message);
  }
}

interface CognitoResult {
  AuthenticationResult?: {
    IdToken: string;
    AccessToken: string;
    RefreshToken?: string;
    ExpiresIn: number;
  };
}

const FRIENDLY: Record<string, string> = {
  UsernameExistsException: 'That username is already taken.',
  InvalidPasswordException:
    'Password must be at least 10 characters with upper and lower case letters and a number.',
  UserNotFoundException: 'Incorrect username or password.',
  NotAuthorizedException: 'Incorrect username or password.',
  UserNotConfirmedException: 'Please confirm your email first.',
  CodeMismatchException: 'That code is not correct.',
  ExpiredCodeException: 'That code has expired. Request a new one.',
  LimitExceededException: 'Too many attempts. Please wait a bit and try again.',
  TooManyRequestsException: 'Too many attempts. Please wait a bit and try again.',
  InvalidParameterException:
    'Some of the details were not accepted. Please check them and try again.',
  AliasExistsException: 'That email address is already in use.',
};

async function call(target: string, body: Record<string, unknown>): Promise<unknown> {
  const cfg = await loadConfig();
  const res = await fetch(`https://cognito-idp.${cfg.region}.amazonaws.com/`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-amz-json-1.1',
      'X-Amz-Target': `AWSCognitoIdentityProviderService.${target}`,
    },
    body: JSON.stringify({ ClientId: cfg.clientId, ...body }),
  });
  const data: unknown = await res.json().catch(() => ({}));
  if (!res.ok) {
    const o = typeof data === 'object' && data !== null ? (data as Record<string, unknown>) : {};
    const type = typeof o.__type === 'string' ? (o.__type.split('#').pop() ?? '') : '';
    // the pre-sign-up trigger rejects under-18s with a message we want to show verbatim
    const raw = typeof o.message === 'string' ? o.message : '';
    const message = raw.includes('18 or older')
      ? 'You must be 18 or older to create an account.'
      : (FRIENDLY[type] ?? (raw || 'Something went wrong.'));
    throw new AuthError(type, message);
  }
  return data;
}

export function isAdultOn(birthdate: string, today: Date): boolean {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(birthdate);
  if (!m) return false;
  const [, y, mo, d] = m;
  const cutoff = new Date(
    Date.UTC(today.getUTCFullYear() - 18, today.getUTCMonth(), today.getUTCDate()),
  );
  const born = new Date(Date.UTC(Number(y), Number(mo) - 1, Number(d)));
  return !Number.isNaN(born.getTime()) && born <= cutoff;
}

export interface SignUpInput {
  username: string;
  email: string;
  password: string;
  birthdate: string;
}

export async function signUp(i: SignUpInput): Promise<void> {
  if (!isAdultOn(i.birthdate, new Date()))
    throw new AuthError('Underage', 'You must be 18 or older to create an account.');
  await call('SignUp', {
    Username: i.username,
    Password: i.password,
    UserAttributes: [
      { Name: 'email', Value: i.email },
      { Name: 'birthdate', Value: i.birthdate },
    ],
  });
}

export async function confirmSignUp(username: string, code: string): Promise<void> {
  await call('ConfirmSignUp', { Username: username, ConfirmationCode: code.trim() });
}

export async function resendCode(username: string): Promise<void> {
  await call('ResendConfirmationCode', { Username: username });
}

export async function forgotPassword(username: string): Promise<void> {
  await call('ForgotPassword', { Username: username });
}

export async function confirmForgotPassword(
  username: string,
  code: string,
  password: string,
): Promise<void> {
  await call('ConfirmForgotPassword', {
    Username: username,
    ConfirmationCode: code.trim(),
    Password: password,
  });
}

function decodeUsername(idToken: string): string {
  try {
    const payload = idToken.split('.')[1] ?? '';
    const json = atob(payload.replace(/-/g, '+').replace(/_/g, '/'));
    const claims: unknown = JSON.parse(json);
    if (typeof claims === 'object' && claims !== null) {
      const u = (claims as Record<string, unknown>)['cognito:username'];
      if (typeof u === 'string') return u;
    }
  } catch {
    // fall through
  }
  return '';
}

function toSession(r: CognitoResult, fallbackRefresh = ''): Session {
  const a = r.AuthenticationResult;
  if (!a)
    throw new AuthError(
      'NoResult',
      'Sign-in needs an extra step that this app does not support yet.',
    );
  return {
    idToken: a.IdToken,
    accessToken: a.AccessToken,
    refreshToken: a.RefreshToken ?? fallbackRefresh,
    expiresAt: Date.now() + a.ExpiresIn * 1000,
    username: decodeUsername(a.IdToken),
  };
}

function isCognitoResult(v: unknown): v is CognitoResult {
  return typeof v === 'object' && v !== null;
}

export async function signIn(username: string, password: string): Promise<Session> {
  const data = await call('InitiateAuth', {
    AuthFlow: 'USER_PASSWORD_AUTH',
    AuthParameters: { USERNAME: username, PASSWORD: password },
  });
  if (!isCognitoResult(data))
    throw new AuthError('Unexpected', 'Unexpected response from sign-in.');
  const s = toSession(data);
  save(s);
  return s;
}

export function load(): Session | undefined {
  try {
    const raw = localStorage.getItem(STORE_KEY);
    if (!raw) return undefined;
    const v: unknown = JSON.parse(raw);
    if (
      typeof v === 'object' &&
      v !== null &&
      'idToken' in v &&
      'refreshToken' in v &&
      'expiresAt' in v
    )
      return v as Session;
  } catch {
    // storage unavailable or corrupt: treat as signed out
  }
  return undefined;
}

function save(s: Session): void {
  try {
    localStorage.setItem(STORE_KEY, JSON.stringify(s));
  } catch {
    // private mode: session lives for this page only
  }
  current = s;
}

let current: Session | undefined = load();
const listeners = new Set<() => void>();

export function onAuthChange(fn: () => void): void {
  listeners.add(fn);
}

function notify(): void {
  listeners.forEach((fn) => {
    fn();
  });
}

export function signOut(): void {
  current = undefined;
  try {
    localStorage.removeItem(STORE_KEY);
  } catch {
    // ignore
  }
  notify();
}

export function isSignedIn(): boolean {
  return current !== undefined;
}

export function currentUsername(): string {
  return current?.username ?? '';
}

/** Returns a token that is valid now, refreshing it if it is about to expire. */
export async function getToken(): Promise<string | undefined> {
  if (!current) return undefined;
  if (current.expiresAt - Date.now() > REFRESH_SKEW_MS) return current.idToken;
  try {
    const data = await call('InitiateAuth', {
      AuthFlow: 'REFRESH_TOKEN_AUTH',
      AuthParameters: { REFRESH_TOKEN: current.refreshToken },
    });
    if (!isCognitoResult(data)) throw new AuthError('Unexpected', 'bad refresh');
    save(toSession(data, current.refreshToken));
    return current.idToken;
  } catch {
    signOut();
    return undefined;
  }
}

export function notifySignedIn(): void {
  notify();
}
