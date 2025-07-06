# 🔍 Elasticsearch + Go デモアプリケーション

このプロジェクトは、**Elasticsearch**と**Go**を組み合わせた検索機能のデモンストレーションです。  
Dockerを使用してローカル環境で簡単に実行でき、RESTful APIを通じてElasticsearchの強力な検索機能を体験できます。

## ✨ 主な機能

- 📄 **ドキュメントの登録**: JSON形式でデータをElasticsearchに保存
- 🔎 **全文検索**: 高速で柔軟な検索クエリをサポート  
- 🚀 **RESTful API**: 使いやすいHTTPエンドポイント（7つのコアエンドポイント）
- 🐳 **Docker対応**: ワンコマンドで環境構築
- 🏗️ **Clean Architecture**: 保守性の高いアーキテクチャ設計

## 🏗️ アーキテクチャ

このプロジェクトはClean Architectureパターンを採用し、レイヤーごとに責務を分離しています：

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP/REST API                           │
│                      (11 endpoints)                        │
├─────────────────────────────────────────────────────────────┤
│                    Interface Layer                         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Document       │  │  Search         │  │  Health         │  │
│  │  Handler        │  │  Handler        │  │  Handler        │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                   Application Layer                        │
│  ┌─────────────────┐  ┌─────────────────┐                    │
│  │  Document       │  │  Search         │                    │
│  │  UseCase        │  │  UseCase        │                    │
│  └─────────────────┘  └─────────────────┘                    │
├─────────────────────────────────────────────────────────────┤
│                     Domain Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  Document       │  │  Search         │  │  Entities       │  │
│  │  Service        │  │  Service        │  │  (Document,     │  │
│  │                 │  │                 │  │   SearchResult) │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                 Infrastructure Layer                       │
│  ┌─────────────────┐  ┌─────────────────┐                    │
│  │  Elasticsearch  │  │  HTTP           │                    │
│  │  Repository     │  │  Middleware     │                    │
│  └─────────────────┘  └─────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

### 🎯 設計の特徴

- **エンドポイント数の最適化**: 28個から11個に削減（61%削減）
- **コアな機能に集中**: ドキュメント操作、検索、ヘルスチェックに特化
- **依存関係の逆転**: 上位レイヤーが下位レイヤーに依存しない設計
- **テスタビリティ**: 各レイヤーが独立してテスト可能

## 🛠️ 必要な環境

- [Docker](https://www.docker.com/) 20.10以上
- [Docker Compose](https://docs.docker.com/compose/) 2.0以上

## 🚀 クイックスタート

### 1. リポジトリをクローン

```bash
git clone <repository-url>
cd elastic-search
```

### 2. 環境を起動

```bash
docker-compose up -d
```

初回起動時は、Dockerイメージのダウンロードとビルドが行われます（数分かかる場合があります）。

### 3. サービスの確認

起動完了後、以下のURLでサービスにアクセスできます：

- 🔍 **Elasticsearch**: http://localhost:9200
- 🌐 **Goアプリケーション**: http://localhost:8080

## 📋 API リファレンス

### 🏥 ヘルスチェック

```bash
GET /health
```

Elasticsearchの接続状態とクラスター情報を確認します。

**例:**

```bash
curl http://localhost:8080/health
```

### 📝 ドキュメント操作

#### ドキュメントの作成

```bash
POST /documents
```

新しいドキュメントをElasticsearchに追加します。

**リクエスト例:**

```bash
curl -X POST http://localhost:8080/documents \
  -H "Content-Type: application/json" \
  -d '{
    "index": "articles",
    "source": {
      "title": "Elasticsearchの活用法", 
      "content": "Elasticsearchは高性能な検索エンジンとして多くの企業で活用されています。"
    }
  }'
```

#### ドキュメントの取得

```bash
GET /documents/{index}/{id}
```

指定したIDのドキュメントを取得します。

**例:**

```bash
curl "http://localhost:8080/documents/articles/abc123"
```

#### ドキュメントの更新

```bash
PUT /documents/{index}/{id}
```

指定したIDのドキュメントを更新します。

**例:**

```bash
curl -X PUT http://localhost:8080/documents/articles/abc123 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Elasticsearchの活用法（更新版）", 
    "content": "Elasticsearchは高性能な検索エンジンとして多くの企業で活用されています。"
  }'
```

#### ドキュメントの削除

```bash
DELETE /documents/{index}/{id}
```

指定したIDのドキュメントを削除します。

**例:**

```bash
curl -X DELETE "http://localhost:8080/documents/articles/abc123"
```

### 🔍 検索

#### 基本検索

```bash
GET /search?q={検索語}&index={インデックス名}
```

指定したキーワードでドキュメントを検索します。

**検索例:**

```bash
# 単一キーワード検索
curl "http://localhost:8080/search?q=Elasticsearch&index=articles"

# 複数キーワード検索
curl "http://localhost:8080/search?q=検索+エンジン&index=articles"
```

#### 高度な検索

```bash
POST /search
```

より複雑な検索クエリを実行します。

**例:**

```bash
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "Elasticsearch",
    "index": "articles",
    "from": 0,
    "size": 10,
    "sort": [{"field": "title", "order": "asc"}]
  }'
```

## 💡 使用例

### サンプルデータの登録と検索
```bash
# 1. サンプルドキュメントを登録
curl -X POST http://localhost:8080/documents \
  -H "Content-Type: application/json" \
  -d '{
    "index": "articles",
    "source": {
      "title": "機械学習入門", 
      "content": "機械学習は人工知能の一分野で、データから学習するアルゴリズムです。"
    }
  }'

curl -X POST http://localhost:8080/documents \
  -H "Content-Type: application/json" \
  -d '{
    "index": "articles",
    "source": {
      "title": "Docker活用術", 
      "content": "Dockerはコンテナ技術を使ってアプリケーションを効率的にデプロイできます。"
    }
  }'

# 2. 検索を実行
curl "http://localhost:8080/search?q=機械学習&index=articles"
curl "http://localhost:8080/search?q=Docker&index=articles"
```

## 🛑 サービスの停止

### 通常の停止
```bash
docker-compose down
```

### データを含めて完全削除

```bash
docker-compose down -v
```

⚠️ **注意**: `-v`オプションを使用すると、Elasticsearchに保存されたすべてのデータが削除されます。

## 🔧 トラブルシューティング

### よくある問題と解決方法

#### ポートが既に使用されている

```bash
# 使用中のポートを確認
lsof -i :8080
lsof -i :9200

# 必要に応じてプロセスを終了
kill -9 <PID>
```

#### Elasticsearchが起動しない

```bash
# コンテナのログを確認
docker-compose logs elasticsearch

# メモリ不足の場合、Docker Desktopのメモリ設定を増やしてください（推奨: 4GB以上）
```

## 📁 プロジェクト構成

```
elastic-search/
├── api/                          # Go APIサーバー
│   ├── cmd/                      # アプリケーションエントリーポイント
│   │   └── server/
│   │       └── main.go           # メインアプリケーション
│   ├── config/                   # 設定管理
│   │   └── config.go
│   ├── internal/                 # アプリケーション内部パッケージ
│   │   ├── application/          # アプリケーション層
│   │   │   ├── dto/              # データ転送オブジェクト
│   │   │   │   ├── request.go
│   │   │   │   └── response.go
│   │   │   └── usecase/          # ビジネスロジック
│   │   │       ├── document.go
│   │   │       └── search.go
│   │   ├── container/            # 依存関係注入
│   │   │   └── container.go
│   │   ├── domain/               # ドメイン層
│   │   │   ├── entity/           # エンティティ
│   │   │   │   ├── document.go
│   │   │   │   └── search.go
│   │   │   ├── repository/       # リポジトリインターフェース
│   │   │   │   └── elasticsearch.go
│   │   │   └── service/          # ドメインサービス
│   │   │       ├── document.go
│   │   │       └── search.go
│   │   ├── infrastructure/       # インフラストラクチャ層
│   │   │   ├── elasticsearch/    # Elasticsearch実装
│   │   │   │   ├── client.go
│   │   │   │   └── repository.go
│   │   │   └── http/             # HTTP設定
│   │   └── interface/            # インターフェース層
│   │       ├── handler/          # HTTPハンドラー
│   │       │   ├── document.go
│   │       │   ├── health.go
│   │       │   └── search.go
│   │       └── middleware/       # ミドルウェア
│   │           ├── cors.go
│   │           └── logging.go
│   ├── pkg/                      # 共通パッケージ
│   │   ├── errors/               # エラーハンドリング
│   │   │   └── errors.go
│   │   └── utils/                # ユーティリティ
│   │       └── response.go
│   ├── go.mod                    # Go依存関係
│   └── go.sum
├── docker/                       # Dockerファイル
│   ├── elasticsearch/
│   │   └── Dockerfile
│   └── golang/
│       └── Dockerfile
├── db/                           # データベース関連
│   ├── mapping.json              # Elasticsearchマッピング
│   ├── test_users.ndjson         # テストデータ
│   └── seed_data.sh              # データシード
├── compose.yaml                  # Docker Compose設定
├── Makefile                      # ビルド・実行用コマンド
└── README.md                     # このファイル
```

## 🎯 エンドポイント一覧

このAPIは**11のコアエンドポイント**を提供しています：

| メソッド | パス                      | 説明             |
| -------- | ------------------------- | ---------------- |
| GET      | `/health`                 | ヘルスチェック   |
| POST     | `/documents`              | ドキュメント作成 |
| GET      | `/documents/{index}/{id}` | ドキュメント取得 |
| PUT      | `/documents/{index}/{id}` | ドキュメント更新 |
| DELETE   | `/documents/{index}/{id}` | ドキュメント削除 |
| GET      | `/search`                 | 基本検索         |
| POST     | `/search`                 | 高度な検索       |
| OPTIONS  | `/documents`              | CORS対応         |
| OPTIONS  | `/documents/{index}/{id}` | CORS対応         |
| OPTIONS  | `/search`                 | CORS対応         |
| OPTIONS  | `/health`                 | CORS対応         |

## 🤝 コントリビューション

プルリクエストやIssueの報告を歓迎します！

## 📄 ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。
