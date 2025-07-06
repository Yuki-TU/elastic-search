package dto

import (
	"time"
)

// CreateDocumentRequest represents a request to create a document
type CreateDocumentRequest struct {
	Index  string         `json:"index" binding:"required"`
	ID     string         `json:"id,omitempty"`
	Source map[string]any `json:"source" binding:"required"`
}

// UpdateDocumentRequest represents a request to update a document
type UpdateDocumentRequest struct {
	Index  string         `json:"index" binding:"required"`
	ID     string         `json:"id" binding:"required"`
	Source map[string]any `json:"source" binding:"required"`
}

// DeleteDocumentRequest represents a request to delete a document
type DeleteDocumentRequest struct {
	Index string `json:"index" binding:"required"`
	ID    string `json:"id" binding:"required"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query   string            `json:"query" binding:"required"`
	Index   string            `json:"index,omitempty"`
	Filters map[string]string `json:"filters,omitempty"`
	From    int               `json:"from,omitempty"`
	Size    int               `json:"size,omitempty"`
	Sort    []SortFieldDTO    `json:"sort,omitempty"`
}

// SortFieldDTO represents a sort field in request
type SortFieldDTO struct {
	Field string `json:"field" binding:"required"`
	Order string `json:"order" binding:"required"` // "asc" or "desc"
}

// BulkIndexRequest represents a bulk index request
type BulkIndexRequest struct {
	Documents []BulkDocumentRequest `json:"documents" binding:"required"`
}

// BulkDocumentRequest represents a single document in bulk request
type BulkDocumentRequest struct {
	Index  string         `json:"index" binding:"required"`
	ID     string         `json:"id,omitempty"`
	Source map[string]any `json:"source" binding:"required"`
}

// CreateIndexRequest represents a request to create an index
type CreateIndexRequest struct {
	Index   string         `json:"index" binding:"required"`
	Mapping map[string]any `json:"mapping,omitempty"`
}

// Validate validates the CreateDocumentRequest
func (req *CreateDocumentRequest) Validate() error {
	if req.Index == "" {
		return ErrIndexRequired
	}
	if len(req.Source) == 0 {
		return ErrSourceRequired
	}
	return nil
}

// Validate validates the UpdateDocumentRequest
func (req *UpdateDocumentRequest) Validate() error {
	if req.Index == "" {
		return ErrIndexRequired
	}
	if req.ID == "" {
		return ErrIDRequired
	}
	if len(req.Source) == 0 {
		return ErrSourceRequired
	}
	return nil
}

// Validate validates the SearchRequest
func (req *SearchRequest) Validate() error {
	if req.Query == "" {
		return ErrQueryRequired
	}
	if req.Size < 0 {
		return ErrInvalidSize
	}
	if req.From < 0 {
		return ErrInvalidFrom
	}
	for _, sort := range req.Sort {
		if sort.Field == "" {
			return ErrSortFieldRequired
		}
		if sort.Order != "asc" && sort.Order != "desc" {
			return ErrInvalidSortOrder
		}
	}
	return nil
}

// SetDefaults sets default values for SearchRequest
func (req *SearchRequest) SetDefaults() {
	if req.Size == 0 {
		req.Size = 10
	}
	if req.From == 0 {
		req.From = 0
	}
}

// Custom errors for validation
var (
	ErrIndexRequired     = NewValidationError("index is required")
	ErrIDRequired        = NewValidationError("id is required")
	ErrSourceRequired    = NewValidationError("source is required")
	ErrQueryRequired     = NewValidationError("query is required")
	ErrInvalidSize       = NewValidationError("size must be non-negative")
	ErrInvalidFrom       = NewValidationError("from must be non-negative")
	ErrSortFieldRequired = NewValidationError("sort field is required")
	ErrInvalidSortOrder  = NewValidationError("sort order must be 'asc' or 'desc'")
)

// ValidationError represents a validation error
type ValidationError struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
		Time:    time.Now(),
	}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}
