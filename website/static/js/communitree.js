/* "Join the communitree": a once-per-session email interstitial.
   Emails are only kept in this browser (localStorage) until a mailing list exists.
   The storage format is shared with web/src/communitree.ts. */
(function () {
  var KEY = 'beetreecraft-communitree';
  var SEEN = 'beetreecraft-communitree-dismissed';

  function read() {
    try {
      var v = JSON.parse(localStorage.getItem(KEY) || '[]');
      return Array.isArray(v) ? v : [];
    } catch (e) {
      return [];
    }
  }
  function add(email, source) {
    var list = read();
    var lower = email.toLowerCase();
    if (!list.some(function (x) { return x && x.email === lower; })) {
      list.push({ email: lower, joinedAt: new Date().toISOString(), source: source });
      try {
        localStorage.setItem(KEY, JSON.stringify(list));
      } catch (e) {
        /* storage blocked */
      }
    }
  }
  function dismissed() {
    try {
      return sessionStorage.getItem(SEEN) === '1';
    } catch (e) {
      return false;
    }
  }
  function el(tag, cls, text) {
    var n = document.createElement(tag);
    if (cls) n.className = cls;
    if (text) n.textContent = text;
    return n;
  }

  function open() {
    var prev = document.activeElement;
    var back = el('div', 'ct-backdrop');
    var modal = el('div', 'ct-modal');
    modal.setAttribute('role', 'dialog');
    modal.setAttribute('aria-modal', 'true');
    modal.setAttribute('aria-labelledby', 'ct-title');
    var title = el('h2', '', 'Join the communitree');
    title.id = 'ct-title';
    var close = el('button', 'ct-close', '×');
    close.type = 'button';
    close.setAttribute('aria-label', 'Close');
    var form = el('form');
    form.noValidate = true;
    var input = el('input');
    input.type = 'email';
    input.required = true;
    input.autocomplete = 'email';
    input.placeholder = 'you@example.com';
    input.setAttribute('aria-label', 'Email address');
    var err = el('p', 'ct-err');
    err.setAttribute('role', 'alert');
    var go = el('button', 'btn primary ct-go', 'Count me in');
    go.type = 'submit';
    var skip = el('button', 'ct-skip', 'Maybe later');
    skip.type = 'button';
    form.append(input, err, go);
    modal.append(
      close,
      title,
      el('p', '', 'Branch out with us. Get new behavior tree patterns, features and community picks, straight to your inbox.'),
      form,
      skip,
      el('p', 'ct-fine', 'Stored only in this browser for now. No spam, ever.')
    );
    back.append(modal);

    function done() {
      try {
        sessionStorage.setItem(SEEN, '1');
      } catch (e) {
        /* storage blocked */
      }
      document.removeEventListener('keydown', onKey);
      back.remove();
      if (prev && prev.focus) prev.focus();
    }
    function onKey(e) {
      if (e.key === 'Escape') done();
    }
    form.addEventListener('submit', function (e) {
      e.preventDefault();
      var v = input.value.trim();
      if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)) {
        err.textContent = 'That email looks a little off.';
        return;
      }
      add(v, 'homepage');
      modal.replaceChildren(
        el('h2', '', "You're in the communitree!"),
        el('p', '', 'Thanks for joining. We will be in touch.')
      );
      var ok = el('button', 'btn primary ct-go', 'Close');
      ok.type = 'button';
      ok.addEventListener('click', done);
      modal.append(ok);
      ok.focus();
    });
    close.addEventListener('click', done);
    skip.addEventListener('click', done);
    back.addEventListener('click', function (e) {
      if (e.target === back) done();
    });
    document.addEventListener('keydown', onKey);
    document.body.append(back);
    input.focus();
  }

  document.addEventListener('DOMContentLoaded', function () {
    if (read().length === 0 && !dismissed()) setTimeout(open, 1500);
  });
})();
