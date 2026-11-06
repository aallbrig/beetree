(function () {
  var KEY = 'beetreecraft-theme';
  var root = document.documentElement;
  function saved() {
    try {
      var v = localStorage.getItem(KEY);
      if (v === 'dark' || v === 'light') return v;
    } catch (e) {
      /* storage blocked: use the OS preference */
    }
    try {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    } catch (e) {
      return 'light';
    }
  }
  root.setAttribute('data-theme', saved());

  document.addEventListener('DOMContentLoaded', function () {
    var btn = document.getElementById('theme-toggle');
    var menu = document.getElementById('menu-toggle');
    var links = document.getElementById('nav-links');
    function paint() {
      var dark = root.getAttribute('data-theme') === 'dark';
      if (btn) {
        btn.textContent = dark ? 'Light mode' : 'Dark mode';
        btn.setAttribute('aria-label', dark ? 'Switch to light mode' : 'Switch to dark mode');
      }
    }
    if (btn) {
      btn.addEventListener('click', function () {
        var next = root.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
        root.setAttribute('data-theme', next);
        try {
          localStorage.setItem(KEY, next);
        } catch (e) {
          /* storage blocked */
        }
        paint();
      });
    }
    if (menu && links) {
      menu.addEventListener('click', function () {
        var open = links.classList.toggle('open');
        menu.setAttribute('aria-expanded', String(open));
        menu.setAttribute('aria-label', open ? 'Close menu' : 'Menu');
      });
    }
    paint();

    var blocks = document.querySelectorAll('.codetabs');
    if (blocks.length) {
      var LKEY = 'beetreecraft-lang';
      var current = 'agnostic';
      try {
        current = localStorage.getItem(LKEY) || 'agnostic';
      } catch (e) {
        /* storage blocked */
      }
      var select = function (id) {
        current = id;
        blocks.forEach(function (block) {
          var has = block.querySelector('.codetab[data-tab="' + id + '"]');
          var active = has ? id : 'agnostic';
          block.querySelectorAll('.codetab').forEach(function (t) {
            t.hidden = t.getAttribute('data-tab') !== active;
          });
          block.querySelectorAll('.codetabs-bar button').forEach(function (b) {
            var on = b.getAttribute('data-tab') === active;
            b.setAttribute('aria-selected', String(on));
            b.tabIndex = on ? 0 : -1;
          });
        });
        try {
          localStorage.setItem(LKEY, id);
        } catch (e) {
          /* storage blocked */
        }
      };
      blocks.forEach(function (block) {
        var bar = block.querySelector('.codetabs-bar');
        block.querySelectorAll('.codetab').forEach(function (t) {
          var b = document.createElement('button');
          b.type = 'button';
          b.setAttribute('role', 'tab');
          b.setAttribute('data-tab', t.getAttribute('data-tab'));
          b.textContent = t.getAttribute('data-label');
          b.addEventListener('click', function () {
            select(t.getAttribute('data-tab'));
          });
          bar.appendChild(b);
        });
        bar.hidden = false;
        block.classList.add('js');
      });
      select(current);
    }
  });
})();

// Copy buttons: <button data-copy="#selector"> copies that element's text.
document.addEventListener('click', function (e) {
  var btn = e.target.closest && e.target.closest('[data-copy]');
  if (!btn) return;
  var src = document.querySelector(btn.getAttribute('data-copy'));
  if (!src || !navigator.clipboard) return;
  navigator.clipboard.writeText(src.textContent.trim()).then(function () {
    var old = btn.textContent;
    btn.textContent = 'Copied';
    setTimeout(function () { btn.textContent = old; }, 1500);
  });
});
