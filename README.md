# 🔍 Elasticsearch + Go デモアプリケーション

このプロジェクトは、**Elasticsearch**と**Go**を組み合わせた検索機能のデモンストレーションです。  
Dockerを使用してローカル環境で簡単に実行でき、RESTful APIを通じてElasticsearchの強力な検索機能を体験できます。

## ✨ 主な機能

- 📄 **ドキュメントの登録**: JSON形式でデータをElasticsearchに保存
- 🔎 **全文検索**: 高速で柔軟な検索クエリをサポート  
- 🚀 **RESTful API**: 使いやすいHTTPエンドポイント
- 🐳 **Docker対応**: ワンコマンドで環境構築

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
GET /
```

Elasticsearchの接続状態とクラスター情報を確認します。

**例:**

```bash
curl http://localhost:8080/
```

### 📝 ドキュメント登録

```bash
POST /index
```

新しいドキュメントをElasticsearchに追加します。

**リクエスト例:**

```bash
curl -X POST http://localhost:8080/index \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Elasticsearchの活用法", 
    "content": "Elasticsearchは高性能な検索エンジンとして多くの企業で活用されています。"
  }'
```

**レスポンス例:**
```json
{
  "message": "Document indexed successfully",
  "id": "abc123"
}
```

### 🔍 検索

```bash
GET /search?q={検索語}
```

指定したキーワードでドキュメントを検索します。

**検索例:**

```bash
# 単一キーワード検索
curl "http://localhost:8080/search?q=Elasticsearch"

# 複数キーワード検索
curl "http://localhost:8080/search?q=検索+エンジン"
```

## 💡 使用例

### サンプルデータの登録と検索
```bash
# 1. サンプルドキュメントを登録
curl -X POST http://localhost:8080/index \
  -H "Content-Type: application/json" \
  -d '{"title": "機械学習入門", "content": "機械学習は人工知能の一分野で、データから学習するアルゴリズムです。"}'

curl -X POST http://localhost:8080/index \
  -H "Content-Type: application/json" \
  -d '{"title": "Docker活用術", "content": "Dockerはコンテナ技術を使ってアプリケーションを効率的にデプロイできます。"}'

# 2. 検索を実行
curl "http://localhost:8080/search?q=機械学習"
curl "http://localhost:8080/search?q=Docker"
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
├── api/                  # Go APIサーバー
│   ├── config/           # 設定ファイル
│   ├── main.go           # メインアプリケーション
│   ├── go.mod            # Go依存関係
│   └── go.sum
├── docker/               # Dockerファイル
│   ├── elasticsearch/
│   └── golang/
├── db/                   # データベース関連
│   ├── mapping.json      # Elasticsearchマッピング
│   └── test_users.ndjson # テストデータ
├── compose.yaml          # Docker Compose設定
└── README.md          　 # このファイル
```

## 🤝 コントリビューション

プルリクエストやIssueの報告を歓迎します！

## 📄 ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は[LICENSE](LICENSE)ファイルをご覧ください。
