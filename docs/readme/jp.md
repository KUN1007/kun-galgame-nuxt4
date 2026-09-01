![kun-galgame-forum](https://kungal.com/kungalgame.webp)

### **[English](../../README.md)** | **[日本語](jp.md)** | **[简体中文](chs.md)** | **[繁體中文](cht.md)**

**お問い合わせ：[Telegram](https://telegram.me/kungalgame) | [Discord](https://discord.com/invite/5F4FS2cXhX)**

画像はゲーム [Ark Order](https://apps.qoo-app.com/en/app/9593) のキャラクター「KUN（鯤）」です。

> **AI 支援開発について：** 本プロジェクトでは **5.1.0** 以降、Codex や Claude Code などの LLM 支援開発ツールを使用しています。**5.0.70** までのコードはすべて手作業で書かれました。最後の完全手書き版は [v5.0.70（commit b4ad59e）](https://github.com/KunMoe/kun-galgame-forum/tree/b4ad59eb77d3eaf36d082aa528651039816e1dfa) です。

# 鯤 Galgame フォーラム

## プロジェクト紹介

鯤 Galgame は、ビジュアルノベルと Galgame を愛する人々のコミュニティです。現在、次の公開サイトがあります。

- [鯤 Galgame フォーラム](https://www.kungal.com)（本プロジェクト）
- [鯤 Galgame スタンプ集](https://sticker.kungal.com)
- [鯤 Galgame 開発ドキュメント](https://soft.moe/kun-visualnovel-docs/kun-forum.html)
- [鯤 Galgame ナビゲーション](https://nav.kungal.org/)
- [鯤 Galgame パッチサイト](https://www.moyu.moe)
- [鯤 Galgame フォーラム ステータスページ](https://down.kungal.com/)

詳しくは[鯤 Galgame について](https://www.kungal.com/kungalgame)をご覧ください。

## 機能

- **ビジュアルノベルカタログ** — 共有 Catalog レジストリによるコミュニティ投稿と審査。VNDB 参照、多言語メタデータ、評価、タグ、エンジン、会社、発売情報に対応
- **リソース共有** — ゲームパッチ、翻訳、音声パックなどを共有し、提供者、プラットフォーム、言語で絞り込み
- **ディスカッションフォーラム** — トピック、返信、ネストコメント、投票、リアクション、推薦、お気に入り、下書き、モデレーションフロー
- **共同編集** — Catalog 編集エンジンによる Git 形式の提案、レビュー、改訂履歴、リバート、貢献者記録
- **プライベートメッセージとチャット** — ダイレクトメッセージ、連絡先、既読状態、リアクション、送信取消
- **萌萌点（Moemoepoint）** — エコシステム共通の貢献ポイント。正規の残高と台帳は OAuth が管理
- **リッチコンテンツ編集** — Milkdown と CodeMirror を基盤にした共通 KUN エディター。KaTeX、シンタックスハイライト、メンション、画像アップロードに対応
- **外観のカスタマイズ** — ライト・ダークモード、ページ透明度、フォント、背景画像、コンテンツレーティング設定
- **検索と発見** — Meilisearch による検索、ランキング、アクティビティフィード、発売カレンダー、コレクション、推薦
- **SEO とフィード** — Nuxt SSR、Schema.org メタデータ、分割サイトマップ、正規化リダイレクト、トピックとビジュアルノベルの RSS

## アーキテクチャ

本リポジトリは Go API と Nuxt フロントエンドを収めた **pnpm workspace モノレポ**です。KUN エコシステムの ID・共通サービス基盤である [`nextmoe-infra`](https://github.com/next-moe/nextmoe-infra) の下流アプリケーションでもあります。

| パッケージ | 役割                                                                                                                                                                                     |
| ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `apps/api` | **Go Fiber v3 BFF / REST API** — フォーラムデータと操作を所有し、レスポンスを集約し、共通サービスを呼び出し、定期同期ジョブを実行                                                        |
| `apps/web` | **Nuxt 4 SSR フロントエンド** — Vue ページ、コンポーネント、Pinia 状態、検証、SEO、ブラウザー操作。Nitro はフィード、サイトマップデータ、コンテンツ画像処理、旧 URL のリダイレクトを提供 |

フォーラムはトピック、返信、メッセージ、リソース、評価、クイズ、コレクション、ツールセット、ローカルカウンター、自身の PostgreSQL schema を所有します。次の共通機能は infra が唯一の正規情報源です。

| 機能                                          | 正規の情報源                                                                             |
| --------------------------------------------- | ---------------------------------------------------------------------------------------- |
| ID とユーザープロフィール                     | OAuth。数値ユーザー ID はすべてのサービスで共通                                          |
| ブラウザーログイン                            | Redis を使う不透明な `kungal_session` Cookie。OAuth token はブラウザー保存領域に置かない |
| 萌萌点の残高と台帳                            | OAuth。フォーラムには必要な場合だけキャッシュビューを保持                                |
| ビジュアルノベルのメタデータ、claim、編集履歴 | Catalog。新規投稿はレジストリが発行した work ID をフォーラムの gid として採用            |
| アバターとコンテンツ画像                      | コンテンツアドレス方式の `image_service`                                                 |
| Galgame とリソースのコメントスレッド          | 共通 Community サービス                                                                  |
| 通報と措置                                    | 共通 Trust サービス                                                                      |
| ツールセットのアーカイブ                      | 共通 Artifact サービス                                                                   |

廃止された Galgame Wiki 系は、新機能向けの契約ではありません。Catalog 切り替えの完了までは一部の互換読み取りと通知画面が残りますが、新しい投稿、claim、審査、編集フローは Catalog を対象にします。

## 技術スタック

| レイヤー               | 技術                                                                                                     |
| ---------------------- | -------------------------------------------------------------------------------------------------------- |
| フロントエンド         | [Nuxt 4](https://nuxt.com/) + Vue 3 SSR（Nitro node-server）                                             |
| UI                     | `@kungal/ui-nuxt` と `@kungal/ui-vue`                                                                    |
| エディター             | `@kungal/editor-nuxt`、Milkdown、CodeMirror                                                              |
| スタイル               | [Tailwind CSS 4](https://tailwindcss.com/)                                                               |
| 状態管理               | [Pinia](https://pinia.vuejs.org/) と永続化状態                                                           |
| バックエンド API       | [Go 1.26](https://go.dev/) + [Fiber v3](https://gofiber.io/)                                             |
| データベース           | PostgreSQL + [GORM](https://gorm.io/)。明示的な生 SQL migration を使用し、AutoMigrate は実行しない       |
| キャッシュとセッション | Redis                                                                                                    |
| 検索                   | [Meilisearch](https://www.meilisearch.com/)                                                              |
| 認証                   | OAuth を背後に持つ BFF セッション。90 日間のスライディング有効期限                                       |
| 共通サービス           | Catalog、Community、Trust、Image、Artifact、OAuth クライアント                                           |
| スケジューラー         | [robfig/cron](https://github.com/robfig/cron) による日次メンテナンスと Catalog イベント同期              |
| 検証とテスト           | Zod、Vitest、Vue Test Utils、Go tests                                                                    |
| デプロイ               | Docker イメージを GHCR に公開し、[Dokploy](https://dokploy.com/) でデプロイ。旧 PM2 スクリプトも利用可能 |
| アクセス解析           | [Umami](https://umami.is/)                                                                               |

## プロジェクト構成

```text
├── apps/
│   ├── api/                 # Go Fiber v3 BFF / REST API
│   │   ├── cmd/             # server、migrate、ワンオフ保守ツール
│   │   ├── internal/        # ドメイン handler、service、repository、model
│   │   ├── migrations/      # 明示的な .up.sql / .down.sql migration
│   │   └── pkg/             # 共通クライアント、設定、エラー、権限、ユーティリティ
│   └── web/                 # Nuxt 4 SSR フロントエンド
│       ├── app/             # ページ、コンポーネント、composable、Pinia store、スタイル
│       ├── server/          # Nitro フィード、サイトマップソース、ミドルウェア、リダイレクト
│       └── shared/          # 共通 TypeScript 型とユーティリティ
├── docker/                  # Dockerfile、環境変数サンプル、デプロイ説明
├── docker-compose*.yml      # ローカル統合環境と本番環境の構成
├── scripts/                 # 旧 PM2 ライフサイクルスクリプト
└── docs/                    # プロジェクト文書と生成された読み取り専用の契約ミラー
```

## はじめに

**前提条件：** Corepack 付き Node.js 24+、Go 1.26+、PostgreSQL、Redis、Meilisearch。ローカルですべての機能を使うには、`nextmoe-infra` の OAuth、Catalog、Image、Artifact、Community、Trust サービスも必要です。

最初にローカル共通基盤を起動します。

```bash
cd /path/to/nextmoe-infra
docker compose -f docker-compose.dev.yml --profile full up -d
# 任意：本番に近い形で匿名化されたデータを使い、ローカル DB を更新します。
./scripts/refresh-dev-db.sh
```

次にフォーラムを設定して起動します。

```bash
corepack enable
pnpm install

cp apps/api/.env.example apps/api/.env
cp apps/web/.env.example apps/web/.env

# 本リポジトリの通常のローカル DB migration を適用します。
pnpm migrate

# API：http://127.0.0.1:2334、Web：http://127.0.0.1:2333
pnpm dev
```

コミット済みの環境変数サンプルはローカル infra stack を参照します。サービス profile、データベース更新、ローカル認証情報については [infra 開発環境ガイド](https://github.com/next-moe/nextmoe-infra/blob/main/docs/dev-environment.md) を参照してください。リポジトリ横断の ID migration には追加の順序要件があります。詳しくは [docs/migration/user/README.md](../migration/user/README.md) を参照してください。

Infra のネットワークが利用可能になった後、フォーラムをコンテナで実行することもできます。

```bash
docker compose run --rm migrate
docker compose up -d kungal-api web
```

コンテナとデプロイの完全な手順は [docker/README.md](../../docker/README.md) を参照してください。

## スクリプト

| コマンド                                                         | 説明                                                 |
| ---------------------------------------------------------------- | ---------------------------------------------------- |
| `pnpm dev`                                                       | API と Web を同時に起動                              |
| `pnpm dev:web` / `pnpm dev:api`                                  | いずれか一方のアプリを起動                           |
| `pnpm build`                                                     | Go API のあとに Nuxt フロントエンドをビルド          |
| `pnpm lint` / `pnpm lint:fix`                                    | フロントエンドの ESLint 問題を検査または修正         |
| `pnpm typecheck`                                                 | フロントエンドの `vue-tsc` 型チェックを実行          |
| `pnpm -F web test`                                               | フロントエンドの Vitest テストを実行                 |
| `pnpm test:api`                                                  | Go テストを実行                                      |
| `pnpm vet`                                                       | `go vet` を実行                                      |
| `pnpm format`                                                    | Prettier と gofmt で両アプリを整形                   |
| `pnpm migrate` / `pnpm migrate:down`                             | 本リポジトリの DB migration を適用またはロールバック |
| `pnpm prod:deploy` / `prod:start` / `prod:stop` / `prod:restart` | 旧 PM2 ライフサイクルスクリプトを実行                |
| `pnpm prod:logs`                                                 | 旧 PM2 ログを表示                                    |

## 開発上の境界

- `docs/oauth/`、`docs/image_service/`、`docs/artifact/` は生成された読み取り専用の契約ミラーです。原本は `nextmoe-infra` で変更し、`kungal-docs` を通じて同期します。
- データベース schema の変更には `apps/api/migrations/` の番号付き migration が必要です。API 起動時に GORM AutoMigrate は実行されません。
- フロントエンドではローカル代替コンポーネントを作る前に KunUI を使用してください。KunUI は上流パッケージであり、本リポジトリでは変更しません。
- フロントエンドとバックエンドのレスポンス構造を一致させてください。Go API は安定した `{ code, message, data }` envelope を返します。

## 参加 / お問い合わせ

- [Telegram グループ](https://telegram.me/kungalgame)
- [Twitter / X](https://twitter.com/kungalgame)
- [GitHub リポジトリ](https://github.com/KunMoe/kun-galgame-forum)
- [Discord グループ](https://discord.com/invite/5F4FS2cXhX)
- [YouTube チャンネル](https://youtube.com/@kungalgame)
- [Bilibili](https://space.bilibili.com/1748455574)

## ライセンス

本プロジェクトは `AGPL-3.0` ライセンスで公開されています。
