package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/infrastructure/elasticsearch"
	"github.com/Yuki-TU/elastic-search/api/pkg/utils"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	esClient *elasticsearch.Client
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(esClient *elasticsearch.Client) *HealthHandler {
	return &HealthHandler{
		esClient: esClient,
	}
}

// HealthCheck handles basic health check requests
// GET /health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// Set headers
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// Check ElasticSearch connection
	esHealth := h.checkElasticsearchHealth(ctx)

	// Overall health status
	overallStatus := "healthy"
	if isHealthy, ok := esHealth["is_healthy"].(bool); !ok || !isHealthy {
		overallStatus = "unhealthy"
	}

	// Create health response using DTO
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

// OptionsHandler handles CORS preflight requests
func (h *HealthHandler) OptionsHandler(w http.ResponseWriter, r *http.Request) {
	utils.SetCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}

// checkElasticsearchHealth checks the health of the ElasticSearch cluster
func (h *HealthHandler) checkElasticsearchHealth(ctx context.Context) map[string]any {
	// Create a context with timeout for the health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Perform health check
	info, err := h.esClient.Info(healthCtx)
	if err != nil {
		return map[string]any{
			"is_healthy": false,
			"error":      err.Error(),
			"status":     "unavailable",
		}
	}

	// Extract information from the response
	healthInfo := map[string]any{
		"is_healthy":    true,
		"status":        "available",
		"response_time": "< 5s",
	}

	// Extract cluster name
	if clusterName, ok := info["cluster_name"].(string); ok {
		healthInfo["cluster_name"] = clusterName
	}

	// Extract version information
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
