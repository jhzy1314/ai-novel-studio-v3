# API 参考

- Base URL：`http://127.0.0.1:8082`
- 编码：UTF-8
- 普通响应：`{ "item": ... }`、`{ "items": [...] }` 或 `{ "status": "ok" }`
- 错误响应：`{ "error": "中文错误信息" }`
- 除登录与静态资源外，所有 `/api/*` 在配置了 `auth_password` 时需要 `ai_novel_sid` 会话 Cookie。

完整路由注册见 `internal/interfaces/api/server.go`。下面按功能列出主要接口。

## 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/login` | `{password}` 登录，成功后写入 HttpOnly Cookie |
| GET | `/api/auth/check` | 查询 `{authenticated, required}` |
| GET | `/api/version` | 实例身份探针：`{product, version, edition}`（无需认证） |

## AI 生成

### POST `/api/generate`

SSE 流式生成。请求体：

```json
{
  "project_id": "",
  "chapter_id": "",
  "user_demand": "本章剧情要求",
  "selected_text": "",
  "outline": "",
  "world_setting": "",
  "character_setting": "",
  "history_content": "",
  "material_text": "",
  "target_word": 1200,
  "run_mode": "auto",
  "prose_style": "cold",
  "context_scope": "smart",
  "web_search": false,
  "role_models": {"thinker": "deepseek-v4-pro"},
  "role_thinking": {"worker": false}
}
```

`run_mode`：`auto` / `free` / `draft` / `orchestrated` / `manual` / `strict` / `art` / `collab` / `light`。

响应为 `text/event-stream`，每条事件：

```text
data: {"type":"stage","stage":"规划大纲","role":"thinker","model":"deepseek-v4-pro"}

data: {"type":"token","text":"正文增量..."}

data: {"type":"done","final_text":"完整终稿","word_count":1234}
```

事件类型：`plan`、`stage`、`token`、`warning`、`estimate`、`error`、`done`。

### POST `/api/prose/analyze`

纯规则文笔分析，不调用模型。

```json
{ "content": "本章正文……" }
```

响应：

```json
{
  "score": 86,
  "level": "出色",
  "summary": "文笔分 86（出色）……",
  "words": 242,
  "sentences": 15,
  "paragraphs": 7,
  "dialogue_ratio": 0.04,
  "avg_sentence_len": 15.1,
  "sentence_len_std": 11.4,
  "sensory_density": 53.7,
  "adverb_density": 20.7,
  "metrics": [{"key":"rhythm","label":"句长波动","value":11.4,"display":"±11.4 字","target":"6–24","score":100}],
  "issues": [{"severity":"warning","category":"副词密度","description":"……","suggestion":"……"}]
}
```

### 其它 AI 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/generate/candidates` | 生成 3 个续写候选 |
| POST | `/api/precheck` | 需求与目标字数预检 |
| POST | `/api/verify` | 逻辑自检 / 校验官审查 |
| POST | `/api/rag/preview` | 预览本次会注入的 RAG 记忆 |
| POST | `/api/ai-tells` | 检测 AI 味并给出问题清单 |
| POST | `/api/ai-polish` | 去 AI 味 / 润色；可选 `style`（cold/classical/cinematic/earthy/suspense） |
| POST | `/api/prose/analyze` | 文笔实验室：句长/节奏/对话/感官/副词/AI 味分析与 0–100 分 |
| POST | `/api/tools/execute` | Helper 工具（扩写、摘要、改写等） |
| POST | `/api/project-tools` | 项目管理工具（取料包、出场统计、钩子检查等） |
| POST | `/api/concept/ask` | 构思 Agent 对话 |
| POST | `/api/concept/complete` | 生成完整构思方案 |
| POST | `/api/characters/relations` | 分析人物关系 |
| POST | `/api/foreshadows/scan` | 扫描全书伏笔 |

## 项目与章节

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/projects` | 项目列表 / 新建 |
| GET/PUT/DELETE | `/api/projects/{id}` | 详情 / 更新 / 删除 |
| POST | `/api/projects/{id}/duplicate` | 复制项目 |
| GET | `/api/projects/{id}/search` | 项目内全文搜索 |
| GET | `/api/projects/{id}/stats` | 写作统计 |
| GET/POST | `/api/volumes` | 卷列表 / 新建 |
| POST | `/api/volumes/reorder` | 卷排序 |
| GET/POST | `/api/chapters` | 章节列表 / 新建 |
| GET/PUT/DELETE | `/api/chapters/{id}` | 章节详情 / 更新 / 软删除 |
| POST | `/api/chapters/reorder` | 章节排序 |
| POST | `/api/chapters/{id}/copy` | 复制章节 |
| POST | `/api/chapters/{id}/split` | 拆分章节 |
| POST | `/api/chapters/merge` | 合并章节 |
| GET | `/api/chapters/trash` | 回收站 |
| POST | `/api/chapters/{id}/restore` | 恢复 |
| POST | `/api/chapters/{id}/permanent-delete` | 永久删除 |
| GET/POST | `/api/chapters/{id}/versions` | 章节版本 |
| POST | `/api/chapters/import` | 导入章节 |
| GET | `/api/chapters/export` | 导出章节 |
| GET | `/api/export/docx` `/api/export/txt` `/api/export/md` | 整书导出 |

## 设定与资源

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/characters` | 人物卡 |
| PUT/DELETE | `/api/characters/{id}` | 改 / 删 |
| GET | `/api/characters/duplicates` | 重复人物检测 |
| POST | `/api/characters/merge` | 合并人物 |
| GET/POST | `/api/worldsettings` | 世界观设定 |
| GET/POST | `/api/factions` | 势力 |
| GET/POST | `/api/locations` | 地点 |
| GET/POST | `/api/timeline` | 时间线事件 |
| GET/POST | `/api/relations` | 人物关系 |
| GET/POST | `/api/foreshadows` | 伏笔 |
| GET | `/api/foreshadows/check` | 伏笔回收检查 |
| GET/POST | `/api/scenebeats` | 场景卡 |
| GET/POST | `/api/materials/writing` | 素材库 |
| POST | `/api/materials/upload` | 上传素材（txt/md/epub/docx） |
| GET/DELETE | `/api/materials` | 素材列表 / 删除 |
| GET/POST | `/api/stylesamples` | 文风样本 |
| POST | `/api/stylesamples/import` | 导入拆书样本 |
| GET/POST | `/api/annotations` | 批注 / 高亮 |
| GET/POST | `/api/reading_progress` | 阅读进度 |
| GET/POST | `/api/templates` | 提示词模板 |

## 模型与配置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET/POST | `/api/models` | 模型列表 / 新增 |
| PUT/DELETE | `/api/models/{id}` | 改 / 删 |
| POST | `/api/models/{id}/test` | 测试连通性 |
| PUT | `/api/models/{id}/default` | 设为默认模型 |
| GET/PUT | `/api/roles/{role}/models` | 角色绑定 |
| GET/POST | `/api/webai/models` | 网页 AI 模型 |
| POST | `/api/webai/test` | 测试网页 AI |
| POST | `/api/webai/auto-cookie` | 自动获取 Cookie 会话 |
| GET/PUT | `/api/configs` | 限额配置 |
| GET | `/api/usage` `/api/logs` | 用量 / 调用日志 |

## 外观与静态资源

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/appearance/backgrounds` | 背景库与当前背景 |
| POST | `/api/appearance/backgrounds/set` | 切换背景 |
| POST | `/api/appearance/backgrounds/upload` | 上传背景 |
| POST | `/api/appearance/backgrounds/generate` | Picsum 随机背景 |
| POST | `/api/appearance/backgrounds/random` | 随机背景 |
| GET | `/static/*` | 前端静态资源（嵌入二进制） |
| GET | `/backgrounds/*` | 背景图（优先磁盘上传，回退内置） |
| GET | `/covers/*` | 项目封面 |
