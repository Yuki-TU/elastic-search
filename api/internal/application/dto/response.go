package dto

import (
	"time"
)

// DocumentDTO represents a document in response
type DocumentDTO struct {
	ID       string         `json:"id"`
	Index    string         `json:"index"`
	Source   map[string]any `json:"source"`
	Version  int64          `json:"version"`
	Created  time.Time      `json:"created"`
	Modified time.Time      `json:"modified"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Query    SearchQueryDTO `json:"query"`
	Results  []HitDTO       `json:"results"`
	Total    int64          `json:"total"`
	MaxScore float64        `json:"max_score,omitempty"`
	Took     int64          `json:"took"`
	TimedOut bool           `json:"timed_out,omitempty"`
}

// SearchQueryDTO represents a search query in response
type SearchQueryDTO struct {
	Query   string            `json:"query"`
	Index   string            `json:"index,omitempty"`
	Filters map[string]string `json:"filters,omitempty"`
	From    int               `json:"from"`
	Size    int               `json:"size"`
	Sort    []SortFieldDTO    `json:"sort,omitempty"`
}

// HitDTO represents a search hit in response
type HitDTO struct {
	Index  string         `json:"index"`
	ID     string         `json:"id"`
	Score  float64        `json:"score"`
	Source map[string]any `json:"source"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDTO `json:"error"`
}

// ErrorDTO represents error details
type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string                 `json:"status"`
	Service string                 `json:"service"`
	Version string                 `json:"version"`
	Checks  map[string]interface{} `json:"checks"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code, message, details string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDTO{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// NewSearchResponse creates a new search response
func NewSearchResponse(query SearchQueryDTO, results []HitDTO, total int64, maxScore float64, took int64, timedOut bool) *SearchResponse {
	response := &SearchResponse{
		Query:    query,
		Results:  results,
		Total:    total,
		Took:     took,
		TimedOut: timedOut,
	}

	// Only include max_score if there are results
	if len(results) > 0 {
		response.MaxScore = maxScore
	}

	return response
}

// NewHealthResponse creates a new health response
func NewHealthResponse(status, service, version string, checks map[string]interface{}) *HealthResponse {
	return &HealthResponse{
		Status:  status,
		Service: service,
		Version: version,
		Checks:  checks,
	}
}
