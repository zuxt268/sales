# Sales API

ドメイン管理APIサーバー（JWT認証付き）

## 機能

- ドメイン管理（CRUD操作）
- JWT認証によるBearer token認証
- ViewDNS APIを使用した逆引きIP検索
- slogによる構造化ログ
- Swagger/OpenAPI ドキュメント
- Air によるホットリロード（開発環境）
- sql-migrate によるデータベースマイグレーション

## 技術スタック

- **フレームワーク**: Echo v4
- **データベース**: MySQL + GORM
- **認証**: JWT (golang-jwt/jwt)
- **APIドキュメント**: Swagger (swaggo)
- **ログ**: slog (構造化JSONログ)
- **テスト**: testcontainers-go
- **マイグレーション**: sql-migrate

## 必要な環境

- Go 1.25+
- MySQL 5.7+
- Docker (testcontainers用)

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

# サーバー設定
ADDRESS=:8091

# ViewDNS API設定
VIEW_DNS_API_URL=https://api.viewdns.info
API_KEY=your_viewdns_api_key

# 認証設定
# パスワードのSHA256ハッシュを生成: echo -n "your_password" | openssl sha256
PASSWORD=your_sha256_password_hash

# JWTシークレットキーを生成: openssl rand -base64 48
JWT_SECRET=your_jwt_secret_key
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

全てのエンドポイントはJWT認証が必要です。

### ドメイン管理

- `GET /api/domains` - ドメイン一覧取得（ページネーション・フィルタリング対応）
- `PUT /api/domain` - ドメイン情報更新
- `DELETE /api/domain` - ドメイン削除
- `POST /api/fetch` - ViewDNS逆引きIPからドメイン情報取得

### ドキュメント

- `GET /swagger/index.html` - Swagger UI

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
│   ├── di/             # 依存性注入
│   ├── domain/         # ドメインモデル・ビジネスロジック
│   ├── external/       # 外部APIクライアント
│   ├── infrastructure/ # データベース・インフラ
│   ├── interfaces/
│   │   ├── handler/    # HTTPハンドラー
│   │   ├── middleware/ # カスタムミドルウェア（JWT認証、ログ）
│   │   └── repository/ # データアクセス層
│   └── usecase/        # ユースケース実装
├── migrations/         # データベースマイグレーション
├── docs/              # Swaggerドキュメント
└── .air.toml          # Air設定ファイル
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
4. `crawl_comp_info` - 企業情報クローリング中
5. `phone` - 電話番号処理中
6. `done` - 完了

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

- JWTトークンの有効期限は1年
- パスワードはSHA256でハッシュ化
- 全APIエンドポイントはBearer token認証が必要
- CORSはデフォルトで有効

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