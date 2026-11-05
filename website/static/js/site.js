(function () {
  var KEY = 'beetreecraft-theme';
  var root = document.documentElement;
  function saved() {
    try {
      return localStorage.getItem(KEY) === 'dark' ? 'dark' : 'light';
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
        menu.textContent = open ? 'Close' : 'Menu';
      });
    }
    paint();

    var picker = document.querySelector('.langpicker');
    if (picker) {
      var LKEY = 'beetreecraft-lang';
      var pick = function (lang) {
        if (lang === 'all') document.body.removeAttribute('data-lang');
        else document.body.setAttribute('data-lang', lang);
        picker.querySelectorAll('button').forEach(function (b) {
          b.setAttribute('aria-pressed', String(b.getAttribute('data-pick') === lang));
        });
        try {
          localStorage.setItem(LKEY, lang);
        } catch (e) {
          /* storage blocked */
        }
      };
      picker.hidden = false;
      picker.addEventListener('click', function (e) {
        var t = e.target;
        if (t instanceof HTMLElement && t.dataset.pick) pick(t.dataset.pick);
      });
      var stored = 'all';
      try {
        stored = localStorage.getItem(LKEY) || 'all';
      } catch (e) {
        /* storage blocked */
      }
      pick(stored);
    }
  });
})();
