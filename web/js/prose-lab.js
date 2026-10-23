/* ============ prose-lab.js：v3.0 文笔实验室 ============ */
var ProseLab = {
  report: null,
  reportChapterId: '',
  busy: false,
  styles: [
    { id: '', label: '默认文笔（只去 AI 味）' },
    { id: 'cold', label: '冷峻克制' },
    { id: 'classical', label: '古典雅致' },
    { id: 'cinematic', label: '电影感' },
    { id: 'earthy', label: '烟火气' },
    { id: 'suspense', label: '悬疑张力' }
  ],

  render: function () {
    var el = document.getElementById('page-prose');
    if (!el) return;
    var ch = Store.state.currentChapter;
    // 切换章节后不展示上一章的报告，避免误读
    if ((ch ? ch.id : '') !== this.reportChapterId) {
      this.report = null;
      this.reportChapterId = ch ? ch.id : '';
    }
    var style = (Store.state.composer && Store.state.composer.proseStyle) || '';
    var options = this.styles.map(function (s) {
      return '<option value="' + s.id + '"' + (s.id === style ? ' selected' : '') + '>' + s.label + '</option>';
    }).join('');
    var html = '<div class="prose-lab">' +
      '<div class="prose-lab-head">' +
        '<div><h3>✒️ 文笔实验室</h3><p>规则分析零 token；精修会调用当前主笔模型。</p></div>' +
      '</div>' +
      '<div class="prose-lab-controls">' +
        '<select id="proseLabStyle" onchange="ProseLab.onStyleChange(this.value)">' + options + '</select>' +
        '<button class="btn btn-primary btn-sm" id="proseAnalyzeBtn" onclick="ProseLab.analyze()">分析本章</button>' +
      '</div>' +
      '<div id="proseReport">' + (this.report ? '' : this.emptyHtml()) + '</div>' +
    '</div>';
    el.innerHTML = html;
    if (this.report) this.renderReport();
  },

  emptyHtml: function () {
    var ch = Store.state.currentChapter;
    if (!ch) return '<div class="prose-empty">先在左侧选择一个章节。<br>文笔实验室会分析当前编辑器的正文。</div>';
    return '<div class="prose-empty">点击「分析本章」，看看这一章的<br>句长、节奏、对话、感官和 AI 味。</div>';
  },

  onStyleChange: function (value) {
    if (!Store.state.composer) Store.state.composer = {};
    Store.state.composer.proseStyle = value;
    try { Store.savePrefs(); } catch (e) {}
    var sel = document.getElementById('proseStyleSelect');
    if (sel) sel.value = value;
  },

  analyze: async function () {
    var content = Editor.getText();
    if (!content || !content.trim()) { UI.toast('编辑器内容为空', 'warn'); return; }
    var box = document.getElementById('proseReport');
    var btn = document.getElementById('proseAnalyzeBtn');
    if (btn) { btn.disabled = true; btn.textContent = '分析中…'; }
    if (box) box.innerHTML = '<div class="prose-loading"><span class="spin"></span>正在读这一章的节奏…</div>';
    try {
      var r = await API.post('/api/prose/analyze', { content: content });
      this.report = r;
      var ch = Store.state.currentChapter;
      this.reportChapterId = ch ? ch.id : '';
      this.renderReport();
    } catch (e) {
      if (box) box.innerHTML = '<div class="prose-empty">分析失败：' + esc(e && e.message || '未知错误') + '</div>';
    } finally {
      if (btn) { btn.disabled = false; btn.textContent = '分析本章'; }
    }
  },

  renderReport: function () {
    var box = document.getElementById('proseReport');
    var r = this.report;
    if (!box || !r) return;
    var score = Math.max(0, Math.min(100, r.score || 0));
    var metrics = (r.metrics || []).map(function (m) {
      var s = Math.max(0, Math.min(100, m.score || 0));
      return '<div class="prose-metric' + (s < 65 ? ' is-low' : '') + '">' +
        '<div class="prose-metric-top"><span>' + esc(m.label) + '</span><b>' + esc(m.display) + '</b></div>' +
        '<div class="prose-bar"><i style="width:' + s + '%"></i></div>' +
        '</div>';
    }).join('');
    var issues = (r.issues || []).map(function (it) {
      var sev = it.severity === 'warning' ? 'warning' : 'info';
      return '<div class="prose-issue ' + sev + '">' +
        '<b>' + esc(it.category || '建议') + '</b>' +
        (it.description ? '<p>' + esc(it.description) + '</p>' : '') +
        (it.suggestion ? '<p class="prose-suggest">→ ' + esc(it.suggestion) + '</p>' : '') +
        '</div>';
    }).join('');
    var html = '<div class="prose-score-card">' +
      '<div class="prose-score-ring" style="--score:' + score + '"><b>' + score + '</b><span>文笔分</span></div>' +
      '<div class="prose-score-info"><b>' + esc(r.level || '') + '</b><p>' + esc(r.summary || '') + '</p>' +
      '<p style="margin-top:4px;opacity:.75">' + (r.words || 0) + ' 字 · ' + (r.sentences || 0) + ' 句 · ' + (r.paragraphs || 0) + ' 段</p></div>' +
      '</div>' +
      '<div class="prose-metrics">' + metrics + '</div>' +
      (issues ? '<div class="prose-section-title"><span>可改进项</span><span>' + (r.issues || []).length + '</span></div><div class="prose-issues">' + issues + '</div>' : '') +
      '<div class="prose-actions">' +
        '<button class="btn btn-primary btn-sm" id="prosePolishBtn" onclick="ProseLab.polish()">✨ 按当前风格精修</button>' +
        '<button class="btn btn-ghost btn-sm" onclick="ProseLab.analyze()">重新分析</button>' +
      '</div>';
    box.innerHTML = html;
  },

  polish: async function () {
    var content = Editor.getText();
    if (!content || !content.trim()) { UI.toast('编辑器内容为空', 'warn'); return; }
    if (this.busy) { UI.toast('正在精修，请稍候', 'warn'); return; }
    var styleEl = document.getElementById('proseLabStyle');
    var style = styleEl ? styleEl.value : '';
    var p = Store.state.currentProject;
    var btn = document.getElementById('prosePolishBtn');
    this.busy = true;
    if (btn) { btn.disabled = true; btn.textContent = '精修中…'; }
    UI.toast('正在精修本章，可能需要 1–3 分钟…', 'info');
    try {
      var r = await API.post('/api/ai-polish', {
        content: content,
        project_id: p ? p.id : '',
        style: style
      });
      if (!r || !r.text) throw new Error('模型没有返回精修结果');
      this.showResult(content, r);
    } catch (e) {
      UI.toast('精修失败：' + (e && e.message || '未知错误'), 'error');
    } finally {
      this.busy = false;
      if (btn) { btn.disabled = false; btn.textContent = '✨ 按当前风格精修'; }
    }
  },

  showResult: function (original, result) {
    var self = this;
    var oldLen = Array.from(original).length;
    var newLen = Array.from(result.text || '').length;
    var body = '<div class="prose-result-meta"><span>原 ' + oldLen + ' 字 → 精修 ' + newLen + ' 字</span><span>模型：' + esc(result.model || '—') + '</span></div>' +
      '<textarea class="prose-result-text" id="prosePolishText">' + esc(result.text || '') + '</textarea>' +
      '<div style="font-size:10.5px;color:var(--muted);margin-top:6px;line-height:1.6">替换前会自动把当前正文保存为一个版本；不满意可以在版本历史里回滚。</div>';
    UI.modal({
      title: '✨ 文笔精修预览',
      wide: 'min(880px, 92vw)',
      body: body,
      actions: [
        { id: 'cancel', label: '取消' },
        { id: 'copy', label: '复制结果', onClick: function () {
          var t = document.getElementById('prosePolishText').value;
          if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(t).then(function () { UI.toast('已复制', 'success'); });
          } else {
            var ta = document.getElementById('prosePolishText');
            ta.select(); document.execCommand('copy');
            UI.toast('已复制', 'success');
          }
        }},
        { id: 'replace', label: '替换本章', cls: 'btn-primary', onClick: async function (m, ov) {
          var t = document.getElementById('prosePolishText').value;
          if (!t || !t.trim()) { UI.toast('精修结果为空', 'warn'); return; }
          var p = Store.state.currentProject;
          if (p) {
            try {
              await API.saveVersion(p.id, '文笔精修前备份 ' + new Date().toLocaleString('zh-CN', { hour12: false }), Editor.getText());
            } catch (e) { console.warn('[prose] backup failed', e); }
          }
          Editor.setContent(t);
          Editor.syncChapterContent();
          ov.remove();
          UI.toast('已替换本章，精修前内容已存为版本', 'success');
          setTimeout(function () { self.analyze(); }, 300);
        }}
      ]
    });
  }
};

window.ProseLab = ProseLab;
