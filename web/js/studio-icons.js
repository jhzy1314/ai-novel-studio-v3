/* Local line icons. Only explicit data-icon nodes are decorated; editor content is untouched. */
var StudioIcons = {
  paths: {
    pen: '<path d="m15 5 4 4M4 20l4-1L20 7a2.8 2.8 0 0 0-4-4L4 15l-1 6Z"/>',
    feather: '<path d="M20 3c-6-1-13 3-13 11v3h3c8 0 12-7 10-14ZM4 21 16 9m-6 6h6"/>',
    plus: '<path d="M12 5v14M5 12h14"/>',
    search: '<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 4 4"/>',
    book: '<path d="M12 5C9 3 5 3 2 4v15c3-1 7-1 10 1 3-2 7-2 10-1V4c-3-1-7-1-10 1Zm0 0v15"/>',
    outline: '<rect x="4" y="3" width="16" height="18" rx="2"/><path d="M8 8h8m-8 4h8m-8 4h5"/>',
    users: '<circle cx="9" cy="8" r="3"/><path d="M3 21v-3a6 6 0 0 1 12 0v3m2-16a3 3 0 0 1 0 6m1 4a5 5 0 0 1 3 4v2"/>',
    globe: '<circle cx="12" cy="12" r="9"/><ellipse cx="12" cy="12" rx="4" ry="9"/><path d="M3 12h18"/>',
    sparkles: '<path d="m12 3 2.6 6.4L21 12l-6.4 2.6L12 21l-2.6-6.4L3 12l6.4-2.6L12 3Zm7 0v4m-2-2h4"/>',
    library: '<path d="M4 4v16M9 4v16m5-15 5-1 3 15-5 1-3-15ZM3 17h7"/>',
    palette: '<path d="M12 3a9 9 0 1 0 0 18h1a2 2 0 0 0 1-3.7c-.7-.4-.5-1.3.3-1.3H17a4 4 0 0 0 4-4 9 9 0 0 0-9-9Z"/><path d="M7 10h.01M10 6.5h.01M15 7h.01M17.5 11h.01" stroke-width="2.8"/>',
    sliders: '<path d="M4 7h6m4 0h6M4 17h10m4 0h2"/><circle cx="12" cy="7" r="2"/><circle cx="16" cy="17" r="2"/>',
    chart: '<path d="M4 4v16h17M8 15v-4m5 4V7m5 8v-6"/>',
    shield: '<path d="m12 3 8 3v6c0 5-8 9-8 9s-8-4-8-9V6l8-3Z"/><path d="m8 12 3 3 5-6"/>',
    bell: '<path d="M6 8a6 6 0 0 1 12 0v7l2 3H4l2-3V8Zm4 13h4"/>',
    focus: '<path d="M8 3H3v5m13-5h5v5M3 16v5h5m13-5v5h-5"/><circle cx="12" cy="12" r="3"/>',
    save: '<path d="M5 3h12l4 4v14H3V3h2Zm2 0v6h9V3M7 21v-8h10v8"/>',
    layers: '<path d="m12 3 10 5-10 5L2 8l10-5ZM2 12l10 5 10-5M2 16l10 5 10-5"/>',
    cpu: '<rect x="6" y="6" width="12" height="12" rx="2"/><path d="M9 2v4m6-4v4M9 18v4m6-4v4M2 9h4m-4 6h4m12-6h4m-4 6h4"/><rect x="10" y="10" width="4" height="4"/>'
  },
  markup: function (name) {
    return '<svg class="studio-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">' + (this.paths[name] || this.paths.pen) + '</svg>';
  },
  init: function () {
    var self = this;
    document.querySelectorAll('[data-icon]').forEach(function (el) {
      el.innerHTML = self.markup(el.getAttribute('data-icon'));
    });
  }
};
document.addEventListener('DOMContentLoaded', function () { StudioIcons.init(); });
