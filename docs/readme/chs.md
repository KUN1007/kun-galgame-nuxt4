![kun-galgame-forum](https://kungal.com/kungalgame.webp)

### **[English](../../README.md)** | **[日本語](jp.md)** | **[简体中文](chs.md)** | **[繁體中文](cht.md)**

**联系我们：[Telegram](https://telegram.me/kungalgame) | [Discord](https://discord.com/invite/5F4FS2cXhX)**

图片来自游戏 [方舟指令](https://apps.qoo-app.com/en/app/9593) 中的角色鲲（KUN）。

> **AI 辅助开发说明：** 本项目自 **5.1.0** 版本起使用 Codex、Claude Code 等 LLM 辅助开发工具。**5.0.70** 及之前版本的所有代码均完全由手工编写。最后一个完全手写的版本是 [v5.0.70（commit b4ad59e）](https://github.com/KunMoe/kun-galgame-forum/tree/b4ad59eb77d3eaf36d082aa528651039816e1dfa)。

# 鲲 Galgame 论坛

## 项目简介

鲲 Galgame 是由视觉小说与 Galgame 爱好者共同组成的社区，目前包含以下公开站点：

- [鲲 Galgame 论坛](https://www.kungal.com)（本项目）
- [鲲 Galgame 表情包](https://sticker.kungal.com)
- [鲲 Galgame 开发文档](https://soft.moe/kun-visualnovel-docs/kun-forum.html)
- [鲲 Galgame 导航](https://nav.kungal.org/)
- [鲲 Galgame 补丁站](https://www.moyu.moe)
- [鲲 Galgame 论坛状态页](https://down.kungal.com/)

更多信息请访问[关于鲲 Galgame](https://www.kungal.com/kungalgame)。

## 特性

- **视觉小说目录** — 由共享 Catalog 注册表承载社区投稿与审核，支持 VNDB 引用、多语言元数据、评分、标签、引擎、会社与发售信息
- **资源分享** — 分享游戏补丁、汉化、语音包等资源，支持提供者追踪以及平台、语言筛选
- **讨论论坛** — 话题、回复、子评论、投票、表态、推票、收藏、草稿与内容审核工作流
- **协作编辑** — 通过 Catalog 编辑引擎提供 Git 风格的提案、评审、修订历史、回滚与贡献者记录
- **私信与聊天** — 私信、联系人、已读状态、表态与撤回功能
- **萌萌点** — 全生态统一的贡献积分，权威余额与流水由 OAuth 持有
- **富内容编辑** — 基于 Milkdown 与 CodeMirror 的共享鲲编辑器，支持 KaTeX、代码高亮、@ 提及与图片上传
- **个性化外观** — 深色与浅色模式，以及页面透明度、字体、背景图和内容分级偏好
- **搜索与发现** — 基于 Meilisearch 的搜索、排行榜、动态流、发售日历、收藏夹与推荐
- **SEO 与订阅** — Nuxt SSR、Schema.org 元数据、分片站点地图、规范化重定向，以及话题和视觉小说 RSS

## 架构

本仓库是包含 Go API 与 Nuxt 前端的 **pnpm workspace monorepo**，也是 [`nextmoe-infra`](https://github.com/next-moe/nextmoe-infra) 的下游应用。Infra 是鲲生态的身份与共享服务中枢。

| 包         | 职责                                                                                                                               |
| ---------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `apps/api` | **Go Fiber v3 BFF 与 REST API** — 持有论坛数据和交互，聚合响应，调用共享服务，并运行定时同步任务                                   |
| `apps/web` | **Nuxt 4 SSR 前端** — Vue 页面、组件、Pinia 状态、校验、SEO 与浏览器交互；Nitro 提供订阅、站点地图数据、内容图片处理及旧路由重定向 |

论坛持有话题、回复、消息、资源、评分、测验、收藏夹、工具集、本地计数器及自身 PostgreSQL schema。以下共享能力仍以 infra 为唯一权威来源：

| 能力                           | 权威数据源                                                               |
| ------------------------------ | ------------------------------------------------------------------------ |
| 身份与用户资料                 | OAuth；各服务的数字用户 ID 始终一致                                      |
| 浏览器登录                     | Redis 支撑的不透明 `kungal_session` Cookie；OAuth token 不进入浏览器存储 |
| 萌萌点余额与流水               | OAuth；论坛只在必要处保留本地缓存视图                                    |
| 视觉小说元数据、认领与编辑历史 | Catalog；新投稿直接领养注册表签发的 work ID 作为论坛 gid                 |
| 头像与内容图片                 | 内容寻址的 `image_service`                                               |
| Galgame 与资源评论区           | 共享 Community 服务                                                      |
| 举报与处置                     | 共享 Trust 服务                                                          |
| 工具集归档文件                 | 共享 Artifact 服务                                                       |

已退役的 Galgame Wiki 家族不再是新功能的契约。Catalog 切换收尾期间仍有少量兼容读取和通知页面，但新的投稿、认领、审核与编辑流程均以 Catalog 为目标。

## 技术栈

| 层         | 技术                                                                                   |
| ---------- | -------------------------------------------------------------------------------------- |
| 前端       | [Nuxt 4](https://nuxt.com/) + Vue 3 SSR（Nitro node-server）                           |
| UI         | `@kungal/ui-nuxt` 与 `@kungal/ui-vue`                                                  |
| 编辑器     | `@kungal/editor-nuxt`、Milkdown 与 CodeMirror                                          |
| 样式       | [Tailwind CSS 4](https://tailwindcss.com/)                                             |
| 状态管理   | [Pinia](https://pinia.vuejs.org/) 与持久化状态                                         |
| 后端 API   | [Go 1.26](https://go.dev/) + [Fiber v3](https://gofiber.io/)                           |
| 数据库     | PostgreSQL + [GORM](https://gorm.io/)，使用显式原生 SQL migration，不运行 AutoMigrate  |
| 缓存与会话 | Redis                                                                                  |
| 搜索       | [Meilisearch](https://www.meilisearch.com/)                                            |
| 身份验证   | OAuth 支撑的 BFF 会话，90 天滑动有效期                                                 |
| 共享服务   | Catalog、Community、Trust、Image、Artifact 与 OAuth 客户端                             |
| 定时任务   | [robfig/cron](https://github.com/robfig/cron)，负责每日维护与 Catalog 事件同步         |
| 校验与测试 | Zod、Vitest、Vue Test Utils 与 Go tests                                                |
| 部署       | Docker 镜像发布到 GHCR，并通过 [Dokploy](https://dokploy.com/) 部署；仍保留旧 PM2 脚本 |
| 流量分析   | [Umami](https://umami.is/)                                                             |

## 项目结构

```text
├── apps/
│   ├── api/                 # Go Fiber v3 BFF / REST API
│   │   ├── cmd/             # server、migrate 与一次性维护工具
│   │   ├── internal/        # 领域 handler、service、repository 与 model
│   │   ├── migrations/      # 显式 .up.sql / .down.sql migration
│   │   └── pkg/             # 共享客户端、配置、错误、权限与工具
│   └── web/                 # Nuxt 4 SSR 前端
│       ├── app/             # 页面、组件、组合式函数、Pinia store 与样式
│       ├── server/          # Nitro 订阅、站点地图源、中间件与重定向
│       └── shared/          # 共享 TypeScript 类型与工具
├── docker/                  # Dockerfile、环境变量示例与部署说明
├── docker-compose*.yml      # 本地集成与生产环境编排
├── scripts/                 # 旧 PM2 生命周期脚本
└── docs/                    # 项目文档与生成的只读契约镜像
```

## 快速开始

**前置依赖：** 带 Corepack 的 Node.js 24+、Go 1.26+、PostgreSQL、Redis 与 Meilisearch。完整本地功能还需要 `nextmoe-infra` 提供的 OAuth、Catalog、Image、Artifact、Community 与 Trust 服务。

先启动本地共享平台：

```bash
cd /path/to/nextmoe-infra
docker compose -f docker-compose.dev.yml --profile full up -d
# 可选：使用形状接近真实数据、经过脱敏的内容刷新本地数据库。
./scripts/refresh-dev-db.sh
```

然后配置并启动论坛：

```bash
corepack enable
pnpm install

cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env

# 执行本仓库常规的本地数据库 migration。
pnpm migrate

# API：http://127.0.0.1:2334，Web：http://127.0.0.1:2333
pnpm dev
```

仓库内的环境变量示例已经指向本地 infra 栈。服务 profile、数据库刷新与本地凭据参见 [infra 开发环境指南](https://github.com/next-moe/nextmoe-infra/blob/main/docs/dev-environment.md)。跨仓库身份迁移还有额外的顺序要求，详见 [docs/migration/user/README.md](../migration/user/README.md)。

Infra 网络可用后，也可以用容器运行论坛：

```bash
docker compose run --rm migrate
docker compose up -d kungal-api web
```

完整的容器与部署流程参见 [docker/README.md](../../docker/README.md)。

## 脚本命令

| 命令                                                             | 描述                                  |
| ---------------------------------------------------------------- | ------------------------------------- |
| `pnpm dev`                                                       | 同时启动 API 与 Web                   |
| `pnpm dev:web` / `pnpm dev:api`                                  | 单独启动一个应用                      |
| `pnpm build`                                                     | 先构建 Go API，再构建 Nuxt 前端       |
| `pnpm lint` / `pnpm lint:fix`                                    | 检查或修复前端 ESLint 问题            |
| `pnpm typecheck`                                                 | 运行前端 `vue-tsc` 类型检查           |
| `pnpm -F web test`                                               | 运行前端 Vitest 测试                  |
| `pnpm test:api`                                                  | 运行 Go 测试                          |
| `pnpm vet`                                                       | 运行 `go vet`                         |
| `pnpm format`                                                    | 使用 Prettier 与 gofmt 格式化两个应用 |
| `pnpm migrate` / `pnpm migrate:down`                             | 执行或回滚本仓库的数据库 migration    |
| `pnpm prod:deploy` / `prod:start` / `prod:stop` / `prod:restart` | 运行旧 PM2 生命周期脚本               |
| `pnpm prod:logs`                                                 | 查看旧 PM2 日志                       |

## 开发边界

- `docs/oauth/`、`docs/image_service/` 与 `docs/artifact/` 是生成的只读契约镜像。应在 `nextmoe-infra` 修改源文档，再通过 `kungal-docs` 同步。
- 数据库 schema 变更必须在 `apps/api/migrations/` 中增加编号 migration；API 启动时不会运行 GORM AutoMigrate。
- 前端开发应优先使用 KunUI 组件，而不是创建本地替代品。KunUI 是上游包，不在本仓库内修改。
- 前后端响应结构必须保持一致。Go API 使用稳定的 `{ code, message, data }` 信封。

## 加入 / 联系我们

- [Telegram 群组](https://telegram.me/kungalgame)
- [Twitter / X](https://twitter.com/kungalgame)
- [GitHub 仓库](https://github.com/KunMoe/kun-galgame-forum)
- [Discord 群组](https://discord.com/invite/5F4FS2cXhX)
- [YouTube 频道](https://youtube.com/@kungalgame)
- [Bilibili](https://space.bilibili.com/1748455574)

## License

本项目遵循 `AGPL-3.0` 开源协议。
