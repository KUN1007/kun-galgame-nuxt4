![kun-galgame-forum](https://kungal.com/kungalgame.webp)

### **[English](../../README.md)** | **[日本語](jp.md)** | **[简体中文](chs.md)** | **[繁體中文](cht.md)**

**聯絡我們：[Telegram](https://telegram.me/kungalgame) | [Discord](https://discord.com/invite/5F4FS2cXhX)**

圖片來自遊戲 [方舟指令](https://apps.qoo-app.com/en/app/9593) 中的角色鯤（KUN）。

> **AI 輔助開發說明：** 本專案自 **5.1.0** 版本起使用 Codex、Claude Code 等 LLM 輔助開發工具。**5.0.70** 及之前版本的所有程式碼均完全由人工編寫。最後一個完全手寫的版本是 [v5.0.70（commit b4ad59e）](https://github.com/KunMoe/kun-galgame-forum/tree/b4ad59eb77d3eaf36d082aa528651039816e1dfa)。

# 鯤 Galgame 論壇

## 專案簡介

鯤 Galgame 是由視覺小說與 Galgame 愛好者共同組成的社群，目前包含以下公開網站：

- [鯤 Galgame 論壇](https://www.kungal.com)（本專案）
- [鯤 Galgame 表情包](https://sticker.kungal.com)
- [鯤 Galgame 開發文件](https://soft.moe/kun-visualnovel-docs/kun-forum.html)
- [鯤 Galgame 導航](https://nav.kungal.org/)
- [鯤 Galgame 補丁站](https://www.moyu.moe)
- [鯤 Galgame 論壇狀態頁](https://down.kungal.com/)

更多資訊請造訪[關於鯤 Galgame](https://www.kungal.com/zh-tw/kungalgame)。

## 功能

- **視覺小說目錄** — 由共用 Catalog 註冊表承載社群投稿與審核，支援 VNDB 參照、多語言中繼資料、評分、標籤、引擎、會社與發售資訊
- **資源分享** — 分享遊戲補丁、漢化、語音包等資源，支援提供者追蹤以及平台、語言篩選
- **討論論壇** — 話題、回覆、子留言、投票、表態、推票、收藏、草稿與內容審核流程
- **協作編輯** — 透過 Catalog 編輯引擎提供 Git 風格的提案、審查、修訂歷史、回復與貢獻者記錄
- **私訊與聊天** — 私訊、聯絡人、已讀狀態、表態與收回功能
- **萌萌點** — 全生態統一的貢獻積分，權威餘額與流水由 OAuth 持有
- **豐富內容編輯** — 基於 Milkdown 與 CodeMirror 的共用鯤編輯器，支援 KaTeX、程式碼高亮、@ 提及與圖片上傳
- **個人化外觀** — 深色與淺色模式，以及頁面透明度、字型、背景圖片和內容分級偏好
- **搜尋與探索** — 基於 Meilisearch 的搜尋、排行榜、動態流、發售日曆、收藏集與推薦
- **SEO 與訂閱** — Nuxt SSR、Schema.org 中繼資料、分片網站地圖、規範化重新導向，以及話題和視覺小說 RSS

## 架構

本倉庫是包含 Go API 與 Nuxt 前端的 **pnpm workspace monorepo**，也是 [`nextmoe-infra`](https://github.com/next-moe/nextmoe-infra) 的下游應用程式。Infra 是鯤生態的身分與共用服務中樞。

| 套件       | 職責                                                                                                                                 |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `apps/api` | **Go Fiber v3 BFF 與 REST API** — 持有論壇資料和互動、聚合回應、呼叫共用服務，並執行定時同步工作                                     |
| `apps/web` | **Nuxt 4 SSR 前端** — Vue 頁面、元件、Pinia 狀態、驗證、SEO 與瀏覽器互動；Nitro 提供訂閱、網站地圖資料、內容圖片處理及舊路由重新導向 |

論壇持有話題、回覆、訊息、資源、評分、測驗、收藏集、工具集、本地計數器及自身 PostgreSQL schema。以下共用能力仍以 infra 為唯一權威來源：

| 能力                             | 權威資料來源                                                                    |
| -------------------------------- | ------------------------------------------------------------------------------- |
| 身分與使用者資料                 | OAuth；各服務的數字使用者 ID 始終一致                                           |
| 瀏覽器登入                       | 由 Redis 支撐的不透明 `kungal_session` Cookie；OAuth token 不進入瀏覽器儲存空間 |
| 萌萌點餘額與流水                 | OAuth；論壇只在必要處保留本地快取檢視                                           |
| 視覺小說中繼資料、認領與編輯歷史 | Catalog；新投稿直接領養註冊表簽發的 work ID 作為論壇 gid                        |
| 頭像與內容圖片                   | 內容定址的 `image_service`                                                      |
| Galgame 與資源留言區             | 共用 Community 服務                                                             |
| 檢舉與處置                       | 共用 Trust 服務                                                                 |
| 工具集封存檔                     | 共用 Artifact 服務                                                              |

已退役的 Galgame Wiki 家族不再是新功能的契約。Catalog 切換收尾期間仍有少量相容讀取和通知頁面，但新的投稿、認領、審核與編輯流程均以 Catalog 為目標。

## 技術堆疊

| 層             | 技術                                                                                    |
| -------------- | --------------------------------------------------------------------------------------- |
| 前端           | [Nuxt 4](https://nuxt.com/) + Vue 3 SSR（Nitro node-server）                            |
| UI             | `@kungal/ui-nuxt` 與 `@kungal/ui-vue`                                                   |
| 編輯器         | `@kungal/editor-nuxt`、Milkdown 與 CodeMirror                                           |
| 樣式           | [Tailwind CSS 4](https://tailwindcss.com/)                                              |
| 狀態管理       | [Pinia](https://pinia.vuejs.org/) 與持久化狀態                                          |
| 後端 API       | [Go 1.26](https://go.dev/) + [Fiber v3](https://gofiber.io/)                            |
| 資料庫         | PostgreSQL + [GORM](https://gorm.io/)，使用明確的原生 SQL migration，不執行 AutoMigrate |
| 快取與工作階段 | Redis                                                                                   |
| 搜尋           | [Meilisearch](https://www.meilisearch.com/)                                             |
| 身分驗證       | OAuth 支撐的 BFF 工作階段，90 天滑動有效期                                              |
| 共用服務       | Catalog、Community、Trust、Image、Artifact 與 OAuth 用戶端                              |
| 定時工作       | [robfig/cron](https://github.com/robfig/cron)，負責每日維護與 Catalog 事件同步          |
| 驗證與測試     | Zod、Vitest、Vue Test Utils 與 Go tests                                                 |
| 部署           | Docker 映像發布到 GHCR，並透過 [Dokploy](https://dokploy.com/) 部署；仍保留舊 PM2 腳本  |
| 流量分析       | [Umami](https://umami.is/)                                                              |

## 專案結構

```text
├── apps/
│   ├── api/                 # Go Fiber v3 BFF / REST API
│   │   ├── cmd/             # server、migrate 與一次性維護工具
│   │   ├── internal/        # 領域 handler、service、repository 與 model
│   │   ├── migrations/      # 明確的 .up.sql / .down.sql migration
│   │   └── pkg/             # 共用用戶端、設定、錯誤、權限與工具
│   └── web/                 # Nuxt 4 SSR 前端
│       ├── app/             # 頁面、元件、composable、Pinia store 與樣式
│       ├── server/          # Nitro 訂閱、網站地圖來源、中介軟體與重新導向
│       └── shared/          # 共用 TypeScript 型別與工具
├── docker/                  # Dockerfile、環境變數範例與部署說明
├── docker-compose*.yml      # 本地整合與正式環境編排
├── scripts/                 # 舊 PM2 生命週期腳本
└── docs/                    # 專案文件與生成的唯讀契約鏡像
```

## 快速開始

**前置需求：** 帶有 Corepack 的 Node.js 24+、Go 1.26+、PostgreSQL、Redis 與 Meilisearch。完整本地功能還需要 `nextmoe-infra` 提供的 OAuth、Catalog、Image、Artifact、Community 與 Trust 服務。

先啟動本地共用平台：

```bash
cd /path/to/nextmoe-infra
docker compose -f docker-compose.dev.yml --profile full up -d
# 選用：以形狀接近真實資料、經過去識別化的內容重新整理本地資料庫。
./scripts/refresh-dev-db.sh
```

然後設定並啟動論壇：

```bash
corepack enable
pnpm install

cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env

# 執行本倉庫例行的本地資料庫 migration。
pnpm migrate

# API：http://127.0.0.1:2334，Web：http://127.0.0.1:2333
pnpm dev
```

倉庫內的環境變數範例已指向本地 infra stack。服務 profile、資料庫重新整理與本地憑據請參閱 [infra 開發環境指南](https://github.com/next-moe/nextmoe-infra/blob/main/docs/dev-environment.md)。跨倉庫身分遷移另有順序要求，詳見 [docs/migration/user/README.md](../migration/user/README.md)。

Infra 網路可用後，也可以透過容器執行論壇：

```bash
docker compose run --rm migrate
docker compose up -d kungal-api web
```

完整的容器與部署流程請參閱 [docker/README.md](../../docker/README.md)。

## 腳本指令

| 指令                                                             | 說明                                      |
| ---------------------------------------------------------------- | ----------------------------------------- |
| `pnpm dev`                                                       | 同時啟動 API 與 Web                       |
| `pnpm dev:web` / `pnpm dev:api`                                  | 單獨啟動一個應用程式                      |
| `pnpm build`                                                     | 先建置 Go API，再建置 Nuxt 前端           |
| `pnpm lint` / `pnpm lint:fix`                                    | 檢查或修正前端 ESLint 問題                |
| `pnpm typecheck`                                                 | 執行前端 `vue-tsc` 型別檢查               |
| `pnpm -F web test`                                               | 執行前端 Vitest 測試                      |
| `pnpm test:api`                                                  | 執行 Go 測試                              |
| `pnpm vet`                                                       | 執行 `go vet`                             |
| `pnpm format`                                                    | 使用 Prettier 與 gofmt 格式化兩個應用程式 |
| `pnpm migrate` / `pnpm migrate:down`                             | 執行或回復本倉庫的資料庫 migration        |
| `pnpm prod:deploy` / `prod:start` / `prod:stop` / `prod:restart` | 執行舊 PM2 生命週期腳本                   |
| `pnpm prod:logs`                                                 | 查看舊 PM2 日誌                           |

## 開發邊界

- `docs/oauth/`、`docs/image_service/` 與 `docs/artifact/` 是生成的唯讀契約鏡像。應在 `nextmoe-infra` 修改來源文件，再透過 `kungal-docs` 同步。
- 資料庫 schema 變更必須在 `apps/api/migrations/` 中新增編號 migration；API 啟動時不會執行 GORM AutoMigrate。
- 前端開發應優先使用 KunUI 元件，而不是建立本地替代品。KunUI 是上游套件，不在本倉庫內修改。
- 前後端回應結構必須保持一致。Go API 使用穩定的 `{ code, message, data }` 信封。

## 參與 / 聯絡我們

- [Telegram 群組](https://telegram.me/kungalgame)
- [Twitter / X](https://twitter.com/kungalgame)
- [GitHub 倉庫](https://github.com/KunMoe/kun-galgame-forum)
- [Discord 群組](https://discord.com/invite/5F4FS2cXhX)
- [YouTube 頻道](https://youtube.com/@kungalgame)
- [Bilibili](https://space.bilibili.com/1748455574)

## License

本專案遵循 `AGPL-3.0` 開源授權條款。
