package repository

import (
	"context"

	"github.com/Yuki-TU/elastic-search/api/internal/domain/entity"
)

// ElasticsearchRepository defines the interface for Elasticsearch operations
type ElasticsearchRepository interface {
	// Document operations
	CreateDocument(ctx context.Context, doc *entity.Document) error
	GetDocument(ctx context.Context, index, id string) (*entity.Document, error)
	UpdateDocument(ctx context.Context, doc *entity.Document) error
	DeleteDocument(ctx context.Context, index, id string) error

	// Search operations
	Search(ctx context.Context, query *entity.SearchQuery) (*entity.SearchResult, error)
	MultiSearch(ctx context.Context, queries []*entity.SearchQuery) ([]*entity.SearchResult, error)

	// Index operations
	CreateIndex(ctx context.Context, index string, mapping map[string]any) error
	DeleteIndex(ctx context.Context, index string) error
	IndexExists(ctx context.Context, index string) (bool, error)

	// Bulk operations
	BulkIndex(ctx context.Context, documents []*entity.Document) error
	BulkDelete(ctx context.Context, indices []string, ids []string) error

	// Health and info
	Health(ctx context.Context) error
	Info(ctx context.Context) (map[string]any, error)
}

// SearchOptions provides additional options for search operations
type SearchOptions struct {
	Timeout           string
	Preference        string
	Routing           string
	ExpandWildcards   []string
	AllowNoIndices    bool
	IgnoreUnavailable bool
}

// BulkItem represents a single item in a bulk operation
type BulkItem struct {
	Index  string         `json:"_index"`
	ID     string         `json:"_id"`
	Source map[string]any `json:"_source"`
	Action string         `json:"action"` // "index", "create", "update", "delete"
}

// BulkResponse represents the response from a bulk operation
type BulkResponse struct {
	Took   int64            `json:"took"`
	Errors bool             `json:"errors"`
	Items  []map[string]any `json:"items"`
}
