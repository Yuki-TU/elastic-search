package dto

import (
	"time"
)

// DocumentDTO はレスポンス内のドキュメントを表す
type DocumentDTO struct {
	ID       string         `json:"id"`
	Index    string         `json:"index"`
	Source   map[string]any `json:"source"`
	Version  int64          `json:"version"`
	Created  time.Time      `json:"created"`
	Modified time.Time      `json:"modified"`
}

// SearchResponse は検索レスポンスを表す
type SearchResponse struct {
	Query    SearchQueryDTO `json:"query"`
	Results  []HitDTO       `json:"results"`
	Total    int64          `json:"total"`
	MaxScore float64        `json:"max_score,omitempty"`
	Took     int64          `json:"took"`
	TimedOut bool           `json:"timed_out,omitempty"`
}

// SearchQueryDTO はレスポンス内の検索クエリを表す
type SearchQueryDTO struct {
	Query   string            `json:"query"`
	Index   string            `json:"index,omitempty"`
	Filters map[string]string `json:"filters,omitempty"`
	From    int               `json:"from"`
	Size    int               `json:"size"`
	Sort    []SortFieldDTO    `json:"sort,omitempty"`
}

// HitDTO はレスポンス内の検索ヒットを表す
type HitDTO struct {
	Index  string         `json:"index"`
	ID     string         `json:"id"`
	Score  float64        `json:"score"`
	Source map[string]any `json:"source"`
}

// ErrorResponse はエラーレスポンスを表す
type ErrorResponse struct {
	Error ErrorDTO `json:"error"`
}

// ErrorDTO はエラー詳細を表す
type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HealthResponse はヘルスチェックレスポンスを表す
type HealthResponse struct {
	Status  string                 `json:"status"`
	Service string                 `json:"service"`
	Version string                 `json:"version"`
	Checks  map[string]interface{} `json:"checks"`
}

// NewErrorResponse は新しいエラーレスポンスを作成する
func NewErrorResponse(code, message, details string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewSearchResponse は新しい検索レスポンスを作成する
func NewSearchResponse(query SearchQueryDTO, results []HitDTO, total int64, maxScore float64, took int64, timedOut bool) *SearchResponse {
	response := &SearchResponse{
		Query:    query,
		Results:  results,
		Total:    total,
		Took:     took,
		TimedOut: timedOut,
	}

	// 結果がある場合のみmax_scoreを含める
	if len(results) > 0 {
		response.MaxScore = maxScore
	}

	return response
}

// NewHealthResponse は新しいヘルスレスポンスを作成する
func NewHealthResponse(status, service, version string, checks map[string]interface{}) *HealthResponse {
	return &HealthResponse{
		Status:  status,
		Service: service,
		Version: version,
		Checks:  checks,
	}
}
