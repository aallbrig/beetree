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
  });
})();
