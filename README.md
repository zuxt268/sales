# Sales API

ドメイン管理APIサーバー（JWT認証付き）

## 機能

- **ドメイン管理**: CRUD操作、逆引きIP検索、業種自動判定
- **ターゲット管理**: IPベースのターゲット管理機能
- **タスク管理**: タスクの作成、更新、実行機能
- **ログ管理**: 操作ログの記録と取得
- **JWT認証**: Bearer token認証によるセキュアなAPI
- **AI連携**: OpenAI GPTを使用した業種自動判定
- **構造化ログ**: slogによるJSON形式のログ出力
- **APIドキュメント**: Swagger/OpenAPI による対話的なドキュメント
- **開発環境**: Airによるホットリロード対応
- **データベース**: sql-migrateによるマイグレーション管理

## 技術スタック

- **言語**: Go 1.25
- **Webフレームワーク**: Echo v4.13
- **データベース**: MySQL + GORM v1.31
- **キャッシュ**: Redis (go-redis v9)
- **認証**: JWT (golang-jwt/jwt v5)
- **AI**: OpenAI GPT (sashabaranov/go-openai)
- **APIドキュメント**: Swagger/OpenAPI (swaggo)
- **ログ**: slog (標準ライブラリ)
- **テスト**: testcontainers-go v0.39
- **マイグレーション**: sql-migrate v1.8
- **開発ツール**: Air (ホットリロード)

## 必要な環境

- Go 1.25+
- MySQL 5.7+
- Redis 6.0+ (キャッシュ用)
- Docker (testcontainers用、開発環境オプション)

## セットアップ

### 1. リポジトリのクローン

```bash
git clone https://github.com/zuxt268/sales.git
cd sales
```

### 2. 環境変数ファイルの作成

```bash
cp .env.sample .env
```

### 3. 環境変数の設定

`.env` ファイルを編集:

```bash
# データベース設定
DB_USER=root
DB_PASSWORD=your_db_password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=sales

# Redis設定
REDIS_HOST=localhost
REDIS_PORT=6379

# サーバー設定
ADDRESS=:8080

# Swagger設定
# 開発環境: localhost:8080
# 本番環境: sales.hp-standard.com
SWAGGER_HOST=localhost:8080

# ViewDNS API設定
VIEW_DNS_API_URL=https://api.viewdns.info
API_KEY=your_viewdns_api_key

# 認証設定
# パスワードのSHA256ハッシュを生成: echo -n "your_password" | openssl sha256
PASSWORD=your_sha256_password_hash

# JWTシークレットキーを生成: openssl rand -base64 48
JWT_SECRET=your_jwt_secret_key

# OpenAI API設定
OPENAI_API_KEY=your_openai_api_key
```

### 4. データベースマイグレーションの実行

`dbconfig.yml` を作成:

```yaml
development:
  dialect: mysql
  datasource: root:password@tcp(localhost:3306)/sales?parseTime=true
  dir: migrations
```

マイグレーション実行:

```bash
sql-migrate up
```

### 5. Swaggerドキュメントの生成

```bash
swag init -g cmd/sales/main.go
```

## アプリケーションの起動

### 開発環境（ホットリロード）

```bash
air
```

### 本番環境

```bash
go build -o sales cmd/sales/main.go
./sales
```

## 認証について

### JWTトークンの生成

```bash
go run cmd/token/main.go your_password
```

これで1年間有効なJWTトークンが出力されます。

### トークンの使用方法

`Authorization` ヘッダーにトークンを含めてリクエストを送信:

```bash
curl -H "Authorization: Bearer <your_token>" http://localhost:8091/api/domains
```

## APIエンドポイント

> **注意**: 現在JWT認証はコメントアウトされています (cmd/sales/main.go:68)。
> 本番環境では認証を有効化してください。

### ヘルスチェック

- `GET /` - Hello World
- `GET /api/healthcheck` - ヘルスチェック

### ドメイン管理

- `GET /api/domains` - ドメイン一覧取得
- `GET /api/domains/:id` - ドメイン詳細取得
- `PUT /api/domains/:id` - ドメイン情報更新
- `DELETE /api/domains/:id` - ドメイン削除
- `POST /api/fetch` - ViewDNS逆引きIPからドメイン情報取得
- `POST /api/domains/analyze` - ドメイン業種分析

### ターゲット管理

- `GET /api/targets` - ターゲット一覧取得
- `POST /api/targets` - ターゲット作成
- `PUT /api/targets/:id` - ターゲット更新
- `DELETE /api/targets/:id` - ターゲット削除

### タスク管理

- `GET /api/tasks` - タスク一覧取得
- `POST /api/tasks` - タスク作成
- `PUT /api/tasks/:id` - タスク更新
- `DELETE /api/tasks/:id` - タスク削除
- `POST /api/tasks/execute` - 全タスク実行
- `POST /api/tasks/:id/execute` - 個別タスク実行

### ログ管理

- `GET /api/logs` - ログ一覧取得
- `POST /api/logs` - ログ作成

### ドキュメント

- `GET /swagger/*` - Swagger UI

## APIドキュメント（Swagger）

以下のURLで対話的なAPIドキュメントにアクセスできます:

```
http://localhost:8091/swagger/index.html
```

Swagger UIでの認証方法:
1. 右上の「Authorize」ボタンをクリック
2. `Bearer <your_token>` を入力
3. 「Authorize」をクリック

## プロジェクト構成

```
.
├── cmd/
│   ├── sales/          # メインアプリケーション
│   └── token/          # JWTトークン生成ツール
├── internal/
│   ├── auth/           # JWT認証ロジック
│   ├── config/         # 設定管理
│   ├── di/             # 依存性注入コンテナ
│   ├── domain/         # ドメインモデル (Domain, Target, Task, Log)
│   ├── external/       # 外部APIクライアント (ViewDNS)
│   ├── infrastructure/ # データベース・Redis接続
│   ├── interfaces/
│   │   ├── handler/    # HTTPハンドラー (API統合)
│   │   ├── middleware/ # ミドルウェア (JWT認証、slogログ)
│   │   └── repository/ # データアクセス層 (GORM, OpenAI)
│   ├── usecase/        # ビジネスロジック (Domain, Target, Task, Log, GPT, Fetch)
│   └── util/           # ユーティリティ
├── migrations/         # データベースマイグレーション (sql-migrate)
├── docs/              # Swagger生成ドキュメント
├── .air.toml          # Air設定ファイル
├── dbconfig.yml       # マイグレーション設定
├── Dockerfile         # 本番環境用Dockerファイル
└── docker-compose.*   # Docker Compose設定 (dev/prod)
```

## 開発

### テストの実行

```bash
go test ./...
```

### 特定のテストの実行

```bash
go test ./internal/interfaces/repository/...
```

### Swaggerドキュメントの更新

APIエンドポイントやドキュメントコメントを変更した後:

```bash
swag init -g cmd/sales/main.go
```

## ドメインステータスのフロー

ドメインは以下のステータスを遷移します:

1. `unknown` - 初期状態
2. `initialize` - 初期化済み
3. `check_view` - 閲覧可否チェック中
4. `check_japan` - 日本語サイトかチェック中
5. `crawl_comp_info` - 企業情報クローリング中（GPT-5-nanoによる業種判定を含む）
6. `done` - 完了

### 業種判定機能

`crawl_comp_info` ステータスのドメインに対して、OpenAI GPTを使用して自動的に業種を判定します。

APIを使用して業種判定を実行:
```bash
curl -X POST http://localhost:8080/api/domains/analyze
```

業種判定は日本標準産業分類に基づいて業種を自動選択します。

## エラーハンドリング

APIは構造化されたエラーレスポンスを返します:

```json
{
  "error": "error_code",
  "message": "人間が読めるエラーメッセージ"
}
```

主なエラーコード:
- `validation_error` (400) - 入力値が不正
- `unauthorized` (401) - トークンが無効または未設定
- `not_found` (404) - リソースが見つからない
- `database_error` (500) - データベース操作失敗

## ログ

全てのログはslogを使用してJSON形式で出力されます:

```json
{
  "time": "2025-10-01T12:00:00Z",
  "level": "INFO",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/api/domains",
  "status": 200,
  "latency": "5ms"
}
```

## セキュリティ

- **JWT認証**: トークンの有効期限は1年
- **パスワード**: SHA256ハッシュで保存
- **CORS**: デフォルトで有効化済み
- **注意**: 現在認証はコメントアウトされています。本番環境では `cmd/sales/main.go:68` の認証ミドルウェアを有効化してください

## トラブルシューティング

### トークン生成時にパニックが発生する

`.env` ファイルに `PASSWORD` と `JWT_SECRET` が正しく設定されているか確認してください。

### データベース接続エラー

1. MySQLサーバーが起動しているか確認
2. `.env` のデータベース設定が正しいか確認
3. データベースが存在するか確認: `CREATE DATABASE sales;`

### マイグレーションが失敗する

`dbconfig.yml` のデータソース設定が `.env` と一致しているか確認してください。

## 保守・運用

### 新しいAPIエンドポイントの追加

1. `internal/interfaces/handler/` にハンドラーメソッドを追加
2. Swaggerアノテーションを記述
3. `cmd/sales/main.go` でルートを登録
4. `swag init -g cmd/sales/main.go` でドキュメント更新

### データベーススキーマの変更

1. `migrations/` に新しいマイグレーションファイルを作成
2. `sql-migrate up` で適用
3. 必要に応じて `internal/domain/` のモデルを更新

### 環境変数の追加

1. `internal/config/config.go` の `Environment` 構造体に追加
2. `.env.sample` に例を追加
3. README.md を更新

## ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

## コントリビューション

1. リポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成