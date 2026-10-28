// External (not inline) so the site's Content-Security-Policy can stay script-src 'self'.
window.addEventListener('load', function () {
  var el = document.getElementById('swagger-ui');
  window.SwaggerUIBundle({
    url: el.dataset.spec,
    dom_id: '#swagger-ui',
    deepLinking: true,
    persistAuthorization: false,
  });
});
