.DEFAULT_GOAL := help

.PHONY: up
up: ## 起動
	docker compose up -d

.PHONY: down
down: ## 停止
	docker compose down

.PHONY: build
build: ## ビルド
	docker compose build

.PHONY: log
log: ## ログを表示
	docker compose logs -f app

.PHONY: in
in: ## コンテナに入る
	docker compose exec app sh

.PHONY: seed
seed: ## データを投入
	./db/seed_data.sh

.PHONY: help
help: ## オプションを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
