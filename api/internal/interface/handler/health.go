package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/infrastructure/elasticsearch"
	"github.com/Yuki-TU/elastic-search/api/pkg/utils"
)

// HealthHandler はヘルスチェックリクエストを処理する
type HealthHandler struct {
	esClient *elasticsearch.Client
}

// NewHealthHandler は新しい HealthHandler を作成する
func NewHealthHandler(esClient *elasticsearch.Client) *HealthHandler {
	return &HealthHandler{
		esClient: esClient,
	}
}

// HealthCheck は基本的なヘルスチェックリクエストを処理する
// GET /health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// ヘッダーを設定
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// ElasticSearch接続をチェック
	esHealth := h.checkElasticsearchHealth(ctx)

	// 全体的なヘルス状態
	overallStatus := "healthy"
	if isHealthy, ok := esHealth["is_healthy"].(bool); !ok || !isHealthy {
		overallStatus = "unhealthy"
	}

	// DTOを使用してヘルスレスポンスを作成
	healthResponse := dto.NewHealthResponse(
		overallStatus,
		"elasticsearch-api",
		"1.0.0",
		map[string]interface{}{
			"elasticsearch": esHealth,
		},
	)

	if overallStatus == "healthy" {
		rw.WriteJSON(http.StatusOK, healthResponse)
	} else {
		rw.WriteJSON(http.StatusServiceUnavailable, healthResponse)
	}
}

// OptionsHandler はCORSプリフライトリクエストを処理する
func (h *HealthHandler) OptionsHandler(w http.ResponseWriter, r *http.Request) {
	utils.SetCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}

// checkElasticsearchHealth はElasticSearchクラスターのヘルスをチェックする
func (h *HealthHandler) checkElasticsearchHealth(ctx context.Context) map[string]any {
	// ヘルスチェック用にタイムアウト付きのコンテキストを作成
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// ヘルスチェックを実行
	info, err := h.esClient.Info(healthCtx)
	if err != nil {
		return map[string]any{
			"is_healthy": false,
			"error":      err.Error(),
			"status":     "unavailable",
		}
	}

	// レスポンスから情報を抽出
	healthInfo := map[string]any{
		"is_healthy":    true,
		"status":        "available",
		"response_time": "< 5s",
	}

	// クラスター名を抽出
	if clusterName, ok := info["cluster_name"].(string); ok {
		healthInfo["cluster_name"] = clusterName
	}

	// バージョン情報を抽出
	if version, ok := info["version"].(map[string]any); ok {
		if versionNumber, ok := version["number"].(string); ok {
			healthInfo["version"] = versionNumber
		}
		if luceneVersion, ok := version["lucene_version"].(string); ok {
			healthInfo["lucene_version"] = luceneVersion
		}
	}

	return healthInfo
}
