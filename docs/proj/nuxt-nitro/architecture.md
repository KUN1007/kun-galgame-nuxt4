# 项目架构文档

## 项目概览

鲲 Galgame 论坛是一个基于 Nuxt 4 的全栈 Web 应用，定位为视觉小说（Galgame）社区论坛。项目使用 Vue 3 Composition API 构建前端、Nitro 作为后端服务器、PostgreSQL + Prisma 作为数据层、Redis 作为缓存和 Token 存储。

- **版本**: 5.0.70
- **包管理器**: pnpm 10.17.1
- **协议**: AGPL-3.0
- **多语言支持**: en-us, ja-jp, zh-cn, zh-tw

## 顶层目录结构

```
kun-galgame-forum/
├── app/                 # 前端应用代码（页面、组件、状态管理、样式等）
├── server/              # 后端服务代码（API、WebSocket、定时任务等）
├── shared/              # 前后端共享代码（类型定义、工具函数）
├── prisma/              # 数据库 Schema 和生成的 Prisma Client
├── lib/                 # 通用库（S3 客户端、图标配置、标签映射）
├── scripts/             # 构建脚本、数据迁移、站点地图生成等
├── public/              # 静态资源（图片、字体、背景等）
├── docs/                # 项目文档
├── algorithms/          # 算法工具
├── backup/              # 备份脚本
├── nuxt.config.ts       # Nuxt 配置文件
├── prisma.config.ts     # Prisma 配置
├── package.json         # 依赖和脚本
├── eslint.config.mjs    # ESLint 配置
├── postcss.config.mjs   # PostCSS 配置
├── .prettierrc          # Prettier 配置
├── .env.example         # 环境变量模板
└── ecosystem.config.cjs # PM2 部署配置
```

## app/ 前端目录

```
app/
├── app.vue              # 根组件
├── pages/               # 页面路由（基于文件的自动路由）
├── components/          # 可复用 Vue 组件（自动导入）
├── composables/         # Vue 3 Composable 函数
├── store/               # Pinia 状态管理
│   ├── modules/         # 持久化 Store
│   ├── temp/            # 临时 Store（会话级别）
│   └── types/           # Store 类型定义
├── middleware/           # 路由中间件（auth、prevent）
├── plugins/             # 客户端插件（字体、指令、域名检查等）
├── validations/         # Zod 验证 Schema（前后端共用）
├── constants/           # 常量和枚举映射
├── config/              # 应用配置（限制、主题、上传等）
├── error/               # 错误处理（i18n 错误码、消息提示）
├── styles/              # 全局样式（Tailwind CSS、编辑器样式）
├── utils/               # 前端工具函数
└── widget/              # Widget 工具
```

## server/ 后端目录

```
server/
├── api/                 # API 端点（基于文件的路由，按资源分组）
│   ├── auth/            # 认证相关
│   ├── galgame/         # Galgame CRUD
│   ├── topic/           # 讨论话题
│   ├── user/            # 用户管理
│   ├── message/         # 消息系统
│   ├── admin/           # 管理后台
│   └── ...              # 其他资源模块
├── utils/               # 服务端工具（JWT、错误处理、Zod 解析、上传等）
├── plugins/             # 服务端插件（Redis、Socket.IO）
├── middleware/           # 服务端中间件（URL 重定向）
├── socket/              # WebSocket 事件处理
├── tasks/               # Nitro 定时任务
├── routes/              # 动态路由（RSS）
└── env/                 # 环境变量加载
```

## shared/ 共享目录

```
shared/
├── types/               # TypeScript 类型定义（按业务模块拆分）
│   ├── utils/           # 工具类型（JWT Payload 等）
│   └── *.ts             # 各业务模块类型
├── utils/               # 工具函数（格式化、校验、防抖、节流等）
└── *.d.ts               # 全局类型声明（语言、分页、用户等）
```

## 技术栈

| 层级 | 技术 |
|------|------|
| 框架 | Nuxt 4 (Vue 3) |
| 样式 | Tailwind CSS 4 |
| 状态管理 | Pinia + pinia-plugin-persistedstate |
| 数据库 | PostgreSQL |
| ORM | Prisma 7 |
| 缓存 | Redis |
| 认证 | JWT (双 Token 机制) |
| 实时通信 | Socket.IO |
| 文件存储 | AWS S3 兼容存储 |
| 验证 | Zod |
| 编辑器 | Milkdown (Markdown) + CodeMirror |
| 图标 | @nuxt/icon (SVG 模式) |
| 主题 | @nuxtjs/color-mode (明暗切换) |
| 数学公式 | KaTeX |
| 图表 | ApexCharts |
| 部署 | PM2 |
| 分析 | Umami |

## Nuxt 模块

项目使用的 Nuxt 模块（`nuxt.config.ts`）：

- `@nuxt/image` - 图片优化
- `@nuxt/icon` - SVG 图标（服务端/客户端分别打包）
- `@nuxt/eslint` - 代码检查
- `@nuxtjs/color-mode` - 深色/浅色主题切换
- `@pinia/nuxt` - 状态管理
- `@dxup/nuxt` - UI 组件库
- `pinia-plugin-persistedstate/nuxt` - Store 持久化
- `nuxt-schema-org` - 结构化数据 SEO
- `nuxt-umami` - 流量分析

## 环境变量

关键环境变量分类（参见 `.env.example`）：

| 类别 | 变量 |
|------|------|
| 应用 | `KUN_GALGAME_URL`, `KUN_GALGAME_API` |
| Redis | `REDIS_HOST`, `REDIS_PORT` |
| JWT | `JWT_ISS`, `JWT_AUD`, `JWT_SECRET` |
| 数据库 | `KUN_DATABASE_URL` |
| 邮件 | `KUN_VISUAL_NOVEL_EMAIL_*` |
| 图床 (S3) | `KUN_VISUAL_NOVEL_IMAGE_BED_*` |
| 文件存储 (S3) | `KUN_VISUAL_NOVEL_S3_STORAGE_*` |
| CDN | `KUN_CF_CACHE_*` |
| 分析 | `KUN_VISUAL_NOVEL_FORUM_UMAMI_ID` |

## 定时任务

通过 Nitro 的 `scheduledTasks` 配置：

- `0 0 * * *` (`reset-daily`) - 每日重置（签到、上传计数等）
- `0 * * * *` (`cleanup-toolset-resource`) - 每小时清理工具集资源
