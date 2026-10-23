/* ============ command-palette.js：v3.0 命令面板（Ctrl+K） ============ */
var CommandPalette = {
  _ov: null,
  _active: 0,
  _items: [],

  all: function () {
    return [
      { icon: '📁', title: '新建项目', hint: 'Project', kw: '项目 小说 project new', run: function () { if (typeof ProjectUI !== 'undefined') ProjectUI.showCreate(); } },
      { icon: '📄', title: '新建章节', hint: 'Chapter', kw: '章节 新建 chapter', run: function () { if (typeof ChapterUI !== 'undefined') ChapterUI.addChapter(); } },
      { icon: '💾', title: '保存版本', hint: 'Ctrl+S', kw: '保存 版本 save version', run: function () { if (typeof ManualSave !== 'undefined') ManualSave.saveDraft(); } },
      { icon: '✒️', title: '打开文笔实验室', hint: 'Prose Lab', kw: '文笔 实验室 润色 prose polish', run: function () { if (typeof RightPanel !== 'undefined') RightPanel.switch('prose'); } },
      { icon: '✨', title: 'AI 味检测', hint: '规则检测', kw: 'ai 味 检测 去味 aigc', run: function () { if (typeof Tools !== 'undefined') Tools.detectAIGC(); } },
      { icon: '🎯', title: '专注模式', hint: 'Focus', kw: '专注 全屏 focus zen', run: function () { if (typeof Editor !== 'undefined') Editor.toggleFocus(); } },
      { icon: '🔎', title: '全文搜索', hint: 'Ctrl+F', kw: '搜索 查找 search', run: function () { if (typeof ChapterUI !== 'undefined') ChapterUI.searchAllChapters(); } },
      { icon: '🧼', title: '切换纯净模式', hint: 'Clean', kw: '纯净 简洁 clean', run: function () { if (typeof UI !== 'undefined') UI.toggleClean(); } },
      { icon: '🌌', title: '切换到极光金墨主题', hint: 'Aurora', kw: '主题 极光 aurora theme', run: function () { if (typeof Themes !== 'undefined') Themes.apply('aurora'); } },
      { icon: '🌙', title: '切换到经典深色主题', hint: 'Dark', kw: '主题 深色 dark theme', run: function () { if (typeof Themes !== 'undefined') Themes.apply('dark'); } },
      { icon: '☀️', title: '切换到经典浅色主题', hint: 'Light', kw: '主题 浅色 light theme', run: function () { if (typeof Themes !== 'undefined') Themes.apply('light'); } },
      { icon: '🔑', title: '打开 API / 模型设置', hint: 'Settings', kw: 'api 模型 key 设置 model', run: function () { if (typeof ModelSettings !== 'undefined') ModelSettings.showQuickKey(); } },
      { icon: '📖', title: '打开新手指南', hint: 'Guide', kw: '指南 帮助 help guide', run: function () { if (typeof UI !== 'undefined') UI.showGuide(); } },
      { icon: '📤', title: '导出作品', hint: 'Export', kw: '导出 export txt docx', run: function () { if (typeof ProjectUI !== 'undefined') ProjectUI.showExport(); } }
    ];
  },

  show: function () {
    if (this._ov) { var inp = document.getElementById('cmdInput'); if (inp) inp.focus(); return; }
    var ov = document.createElement('div');
    ov.className = 'cmd-overlay';
    ov.innerHTML = '<div class="cmd-modal" role="dialog" aria-label="命令面板">' +
      '<input class="cmd-input" id="cmdInput" autocomplete="off" placeholder="输入命令或搜索：新建项目 / 文笔实验室 / 专注模式…">' +
      '<div class="cmd-list" id="cmdList"></div>' +
      '<div class="cmd-foot"><span><kbd>↑</kbd><kbd>↓</kbd> 选择 &nbsp; <kbd>Enter</kbd> 执行</span><span><kbd>Esc</kbd> 关闭</span></div>' +
      '</div>';
    this._ov = ov;
    this._items = [];
    this._active = 0;
    document.body.appendChild(ov);
    ov.onclick = function (e) { if (e.target === ov) CommandPalette.close(); };
    var input = document.getElementById('cmdInput');
    var self = this;
    input.oninput = function () { self.render(input.value); };
    input.onkeydown = function (e) {
      if (e.key === 'ArrowDown') { e.preventDefault(); self.move(1); }
      else if (e.key === 'ArrowUp') { e.preventDefault(); self.move(-1); }
      else if (e.key === 'Enter') { e.preventDefault(); self.runActive(); }
      else if (e.key === 'Escape') { e.preventDefault(); self.close(); }
    };
    this.render('');
    setTimeout(function () { input.focus(); }, 30);
  },

  render: function (q) {
    q = (q || '').trim().toLowerCase();
    var all = this.all();
    this._items = all.filter(function (it) {
      if (!q) return true;
      return (it.title + ' ' + it.kw + ' ' + it.hint).toLowerCase().indexOf(q) >= 0;
    }).slice(0, 12);
    if (this._active >= this._items.length) this._active = 0;
    var list = document.getElementById('cmdList');
    if (!list) return;
    if (!this._items.length) {
      list.innerHTML = '<div class="cmd-empty">没有匹配的命令</div>';
      return;
    }
    var self = this;
    list.innerHTML = this._items.map(function (it, i) {
      return '<div class="cmd-item' + (i === self._active ? ' active' : '') + '" data-i="' + i + '">' +
        '<span class="cmd-ico">' + it.icon + '</span><span class="cmd-title">' + esc(it.title) + '</span><span class="cmd-hint">' + esc(it.hint) + '</span></div>';
    }).join('');
    list.querySelectorAll('.cmd-item').forEach(function (el) {
      el.onmouseenter = function () { self._active = parseInt(el.getAttribute('data-i')) || 0; self.highlight(); };
      el.onclick = function () { self._active = parseInt(el.getAttribute('data-i')) || 0; self.runActive(); };
    });
  },

  highlight: function () {
    var list = document.getElementById('cmdList');
    if (!list) return;
    list.querySelectorAll('.cmd-item').forEach(function (el, i) {
      el.classList.toggle('active', i === this._active);
    }, this);
    var cur = list.querySelector('.cmd-item.active');
    if (cur && cur.scrollIntoView) cur.scrollIntoView({ block: 'nearest' });
  },

  move: function (d) {
    if (!this._items.length) return;
    this._active = (this._active + d + this._items.length) % this._items.length;
    this.highlight();
  },

  runActive: function () {
    var it = this._items[this._active];
    if (!it) return;
    this.close();
    setTimeout(function () { it.run(); }, 30);
  },

  close: function () {
    if (this._ov) { this._ov.remove(); this._ov = null; }
  }
};

document.addEventListener('keydown', function (e) {
  if ((e.ctrlKey || e.metaKey) && !e.altKey && (e.key === 'k' || e.key === 'K')) {
    e.preventDefault();
    CommandPalette.show();
  } else if (e.key === 'Escape' && CommandPalette._ov) {
    CommandPalette.close();
  }
});

window.CommandPalette = CommandPalette;
