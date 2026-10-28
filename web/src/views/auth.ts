import {
  AuthError,
  confirmForgotPassword,
  confirmSignUp,
  forgotPassword,
  resendCode,
  signIn,
  signUp,
  notifySignedIn,
} from '../auth';
import { h, toast } from '../dom';
import { navigate, type Ctx } from '../router';

function field(label: string, input: HTMLInputElement, hint?: string): HTMLElement {
  const id = `f-${label.toLowerCase().replace(/\W+/g, '-')}`;
  input.id = id;
  return h(
    'div',
    { class: 'field' },
    h('label', { for: id }, label),
    input,
    hint ? h('div', { class: 'hint' }, hint) : null,
  );
}

const input = (type: string, name: string, extra: Record<string, string> = {}): HTMLInputElement =>
  h('input', { type, name, required: true, ...extra });

/** Runs an async submit handler with a disabled button and an inline error. */
function submitter(form: HTMLFormElement, error: HTMLElement, fn: () => Promise<void>): void {
  form.addEventListener('submit', (e) => {
    e.preventDefault();
    error.textContent = '';
    const btn = form.querySelector('button[type=submit]');
    if (btn instanceof HTMLButtonElement) btn.disabled = true;
    fn()
      .catch((err: unknown) => {
        error.textContent =
          err instanceof AuthError || err instanceof Error ? err.message : 'Something went wrong.';
      })
      .finally(() => {
        if (btn instanceof HTMLButtonElement) btn.disabled = false;
      });
  });
}

function shell(ctx: Ctx, title: string, form: HTMLFormElement, footer: (Node | string)[]): void {
  ctx.root.append(
    h(
      'div',
      { class: 'form card' },
      h('h1', {}, title),
      form,
      h('p', { class: 'small muted', style: 'margin-top:16px' }, ...footer),
    ),
  );
}

const link = (href: string, text: string): HTMLAnchorElement => h('a', { href }, text);

export function signInView(ctx: Ctx): void {
  const user = input('text', 'username', { autocomplete: 'username' });
  const pass = input('password', 'password', { autocomplete: 'current-password' });
  const err = h('div', { class: 'error', role: 'alert' });
  const form = h(
    'form',
    {},
    field('Username or email', user),
    field('Password', pass),
    err,
    h('button', { class: 'btn primary', type: 'submit' }, 'Sign in'),
  );
  submitter(form, err, async () => {
    try {
      await signIn(user.value.trim(), pass.value);
    } catch (e) {
      if (e instanceof AuthError && e.code === 'UserNotConfirmedException') {
        navigate(`/confirm?u=${encodeURIComponent(user.value.trim())}`);
        return;
      }
      throw e;
    }
    notifySignedIn();
    navigate(ctx.query.get('next') ?? '/dashboard');
  });
  shell(ctx, 'Sign in', form, [
    link('#/forgot', 'Forgot your password?'),
    ' · ',
    link('#/signup', 'Create an account'),
  ]);
}

export function signUpView(ctx: Ctx): void {
  const user = input('text', 'username', {
    autocomplete: 'username',
    minlength: '3',
    maxlength: '32',
    pattern: '[A-Za-z0-9_\\-]{3,32}',
  });
  const email = input('email', 'email', { autocomplete: 'email' });
  const dob = input('date', 'birthdate', { autocomplete: 'bday' });
  const pass = input('password', 'password', { autocomplete: 'new-password', minlength: '10' });
  const adult = h('input', { type: 'checkbox', required: true, id: 'adult' });
  const err = h('div', { class: 'error', role: 'alert' });
  const form = h(
    'form',
    {},
    field('Username', user, '3–32 letters, digits, _ or -. This is your public name.'),
    field('Email', email, 'We send a confirmation code here.'),
    field('Date of birth', dob, 'You must be 18 or older.'),
    field(
      'Password',
      pass,
      'At least 10 characters, with upper and lower case letters and a number.',
    ),
    h(
      'div',
      { class: 'field row' },
      adult,
      h(
        'label',
        { for: 'adult', style: 'margin:0;font-weight:400' },
        'I confirm that I am 18 years of age or older.',
      ),
    ),
    err,
    h('button', { class: 'btn primary', type: 'submit' }, 'Create account'),
  );
  submitter(form, err, async () => {
    await signUp({
      username: user.value.trim(),
      email: email.value.trim(),
      password: pass.value,
      birthdate: dob.value,
    });
    toast('Check your email for a confirmation code.');
    navigate(`/confirm?u=${encodeURIComponent(user.value.trim())}`);
  });
  shell(ctx, 'Create your account', form, ['Already registered? ', link('#/signin', 'Sign in')]);
}

export function confirmView(ctx: Ctx): void {
  const user = input('text', 'username', { value: ctx.query.get('u') ?? '' });
  const code = input('text', 'code', {
    inputmode: 'numeric',
    autocomplete: 'one-time-code',
    maxlength: '8',
  });
  const err = h('div', { class: 'error', role: 'alert' });
  const resend = h('button', { class: 'btn small', type: 'button' }, 'Resend code');
  resend.addEventListener('click', () => {
    resendCode(user.value.trim()).then(
      () => {
        toast('A new code is on its way.');
      },
      (e: unknown) => {
        err.textContent = e instanceof Error ? e.message : 'Could not resend.';
      },
    );
  });
  const form = h(
    'form',
    {},
    field('Username', user),
    field('Confirmation code', code, 'Sent to your email.'),
    err,
    h(
      'div',
      { class: 'row' },
      h('button', { class: 'btn primary', type: 'submit' }, 'Confirm email'),
      resend,
    ),
  );
  submitter(form, err, async () => {
    await confirmSignUp(user.value.trim(), code.value);
    toast('Email confirmed. You can sign in now.');
    navigate('/signin');
  });
  shell(ctx, 'Confirm your email', form, [link('#/signin', 'Back to sign in')]);
}

export function forgotView(ctx: Ctx): void {
  const user = input('text', 'username', { autocomplete: 'username' });
  const code = input('text', 'code', { inputmode: 'numeric', autocomplete: 'one-time-code' });
  const pass = input('password', 'password', { autocomplete: 'new-password', minlength: '10' });
  const err = h('div', { class: 'error', role: 'alert' });
  const step2 = h(
    'div',
    { class: 'hide' },
    field('Code from email', code),
    field(
      'New password',
      pass,
      'At least 10 characters, with upper and lower case letters and a number.',
    ),
  );
  const submit = h('button', { class: 'btn primary', type: 'submit' }, 'Send reset code');
  const form = h('form', {}, field('Username or email', user), step2, err, submit);
  let sent = false;
  submitter(form, err, async () => {
    if (!sent) {
      await forgotPassword(user.value.trim());
      sent = true;
      step2.classList.remove('hide');
      code.required = true;
      pass.required = true;
      submit.textContent = 'Set new password';
      toast('If that account exists, a code is on its way.');
      return;
    }
    await confirmForgotPassword(user.value.trim(), code.value, pass.value);
    toast('Password updated. Please sign in.');
    navigate('/signin');
  });
  code.required = false;
  pass.required = false;
  shell(ctx, 'Reset your password', form, [link('#/signin', 'Back to sign in')]);
}
