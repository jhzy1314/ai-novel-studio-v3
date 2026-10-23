# 架构说明

## 1. 分层

```text
cmd/server
   │  加载 config.yaml → 打开 SQLite → 种子模型/限额 → 装配 Registry、Limiter、Dispatcher
   ▼
internal/interfaces/api          HTTP / SSE / 静态资源（go:embed）
   ▼
internal/domain/pipeline         调度中枢 + 9 条流水线 + 上下文/RAG/审校
   ▼
internal/domain/roles            4 个职能角色 + 系统提示词
   ▼
internal/infrastructure/llm      模型适配（API / 网页 AI）+ 角色模型注册中心
   ▼
internal/infrastructure/database SQLite 仓库、迁移、备份
```

原则：

- 接口层不写业务规则；
- 流水线只负责编排，不直接拼 SQL；
- 模型与角色解耦，角色绑定可在运行时修改；
- 所有数据落本地 SQLite，不依赖外部服务。

## 2. 角色

| 角色 | 主要职责 | 提示词 |
|------|---------|--------|
| `thinker` | 需求拆解、全书/本章大纲、剧情脉络 | `roles.ThinkerPrompt` |
| `worker` | 正文撰写、续写、润色、修订 | `roles.WorkerPrompt` |
| `verifier` | 人设、世界观、时间线、逻辑、语病审查 | `roles.VerifierPrompt` |
| `helper` | 扩写、缩写、摘要、改写等轻任务 | `roles.HelperPrompt` |

每个角色在 `role_models` 中按 priority 绑定一个或多个模型。调用失败时，`Registry` 自动尝试下一级模型。

## 3. 流水线

| 内部值 | 模式 | 链路 |
|--------|------|------|
| `auto` | 智能协同 | 按任务自动选择 |
| `free` | 自由写作 | Worker（单 Agent，可注入文风/人物/前情/世界观/伏笔/素材） |
| `draft` | 快速草稿 | Worker |
| `orchestrated` | 指派 Agent 模型 | Thinker → Worker → Verifier，每角色用指定模型 |
| `manual` | 手动自选模型 | 直调指定模型 |
| `strict` | 严谨模式 | Thinker → Worker → Verifier（高标准） |
| `art` | 文艺创作 | Thinker（极简）→ Worker（自由）→ Verifier（宽松） |
| `collab` | 协同闭环 | Thinker → Worker → Verifier → Thinker 重规划 → Worker 重写 |
| `light` | 轻量快捷 | Helper |

公共能力：

- **上下文构建**：项目设定、人物卡、世界观、前文、章节摘要、素材、文风样本、伏笔。
- **RAG**：按查询检索相关章节片段，注入提示词。
- **字数控制**：目标字数、分段生成、双层字数校验。
- **质量规则**：AI 味标记、元叙事、说教词、填充词、段落节奏。
- **联网检索**：可选，检索结果只作参考，提示词要求不得照抄、不得编造来源。

### 3.1 文笔层（v3.0）

- `pipeline.ProseStyles`：冷峻 / 古典 / 电影感 / 烟火气 / 悬疑 5 个预设；`ProseStyleDirective` 注入生成提示词，`ProseStylePolishDirective` 注入精修提示词。
- `pipeline.AnalyzeProse`：纯规则分析句长、句长波动、对话占比、感官密度、副词密度、段落节奏、AI 味，输出 0–100 分与建议。
- `POST /api/prose/analyze`：文笔实验室后端，不调用模型。
- `POST /api/ai-polish`：`style` 参数控制精修方向；前端替换前自动保存版本。
- 前端 `ProseLab` / `CommandPalette` / `v3.css`：文笔面板、Ctrl+K 命令面板、极光金墨主题与衬线正文。

## 4. 数据模型（主要表）

```text
projects ──< chapters ──< chapter_versions
    │           └──< annotations
    ├──< characters ──< relations
    ├──< world_settings / factions / locations / timeline_events
    ├──< foreshadows
    ├──< materials / style_samples / scene_beats
    ├──< volumes
    └──< versions

models ──< role_models
configs / generation_logs / reading_progress / sessions
```

- 所有业务表以 `project_id` 做项目隔离。
- 章节使用软删除 + 回收站；编辑保存时写版本快照。
- 删除项目会级联清理其下资源。
- 数据库迁移在 `database.Open` 与 `RunMigrations` 中幂等执行。

## 5. 前端

- `web/index.html`：单页壳、模板、内联引导脚本。
- `web/js/*.js`：按职责拆分的普通脚本（无打包步骤），通过全局对象协作。
- `web/css/*.css`：基础样式、主题、精修层、新视觉层。
- 富文本使用 ProseMirror/Tiptap（esm.sh CDN）；加载失败时自动回退到 Markdown 编辑模式。
- 所有静态资源由 `web/static.go` 的 `go:embed` 打进二进制；默认背景图也在其中。

### 写作空间视觉层（2026-10-23）

- `web/css/studio.css` 在现有样式之后加载：共用布局与间距跟随主题变量，极光金墨使用独立色板。
- `web/js/studio-icons.js` 只处理显式标注的 `data-icon` 节点；正文、用户项目名与导入内容不参与替换。
- `.app` 的直接子节点依次为左侧导航、`main`、右侧助手，面板展开时为正文保留实际布局空间。
- `MobileUI.switchRightPage` 将已有面板移动到手机抽屉，关闭或切换桌面时归还，保持 DOM ID 唯一与模块状态一致。
- 空白引导只在未选项目时出现；主题选择器分别限定 rich / Markdown 的作用域。

## 6. 安全与运行

- 默认监听 `127.0.0.1`，仅本机可访问。
- `auth_password` 非空时，除登录/静态资源外的 API 需要会话 Cookie。
- API Key 只存在于本地配置文件或 SQLite；接口返回时脱敏。
- 启动时用 `VACUUM INTO` 备份数据库，避免 WAL 模式下直接复制文件造成不一致。
- 静态资源带启动版本参数，可长期缓存；首页与 API 不缓存。
