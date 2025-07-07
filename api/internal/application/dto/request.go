package dto

import (
	"time"
)

// CreateDocumentRequest はドキュメント作成リクエストを表す
type CreateDocumentRequest struct {
	Index  string         `json:"index" binding:"required"`
	ID     string         `json:"id,omitempty"`
	Source map[string]any `json:"source" binding:"required"`
}

// UpdateDocumentRequest はドキュメント更新リクエストを表す
type UpdateDocumentRequest struct {
	Index  string         `json:"index" binding:"required"`
	ID     string         `json:"id" binding:"required"`
	Source map[string]any `json:"source" binding:"required"`
}

// DeleteDocumentRequest はドキュメント削除リクエストを表す
type DeleteDocumentRequest struct {
	Index string `json:"index" binding:"required"`
	ID    string `json:"id" binding:"required"`
}

// SearchRequest は検索リクエストを表す
type SearchRequest struct {
	Query   string            `json:"query" binding:"required"`
	Index   string            `json:"index,omitempty"`
	Filters map[string]string `json:"filters,omitempty"`
	From    int               `json:"from,omitempty"`
	Size    int               `json:"size,omitempty"`
	Sort    []SortFieldDTO    `json:"sort,omitempty"`
}

// SortFieldDTO はリクエスト内のソートフィールドを表す
type SortFieldDTO struct {
	Field string `json:"field" binding:"required"`
	Order string `json:"order" binding:"required"` // "asc" または "desc"
}

// BulkIndexRequest はバルクインデックスリクエストを表す
type BulkIndexRequest struct {
	Documents []BulkDocumentRequest `json:"documents" binding:"required"`
}

// BulkDocumentRequest はバルクリクエスト内の単一ドキュメントを表す
type BulkDocumentRequest struct {
	Index  string         `json:"index" binding:"required"`
	ID     string         `json:"id,omitempty"`
	Source map[string]any `json:"source" binding:"required"`
}

// CreateIndexRequest はインデックス作成リクエストを表す
type CreateIndexRequest struct {
	Index   string         `json:"index" binding:"required"`
	Mapping map[string]any `json:"mapping,omitempty"`
}

// Validate は CreateDocumentRequest を検証する
func (req *CreateDocumentRequest) Validate() error {
	if req.Index == "" {
		return ErrIndexRequired
	}
	if len(req.Source) == 0 {
		return ErrSourceRequired
	}
	return nil
}

// Validate は UpdateDocumentRequest を検証する
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

// Validate は SearchRequest を検証する
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

// SetDefaults は SearchRequest のデフォルト値を設定する
func (req *SearchRequest) SetDefaults() {
	if req.Size == 0 {
		req.Size = 10
	}
	if req.From == 0 {
		req.From = 0
	}
}

// バリデーション用のカスタムエラー
var (
	ErrIndexRequired     = NewValidationError("インデックスは必須です")
	ErrIDRequired        = NewValidationError("IDは必須です")
	ErrSourceRequired    = NewValidationError("ソースは必須です")
	ErrQueryRequired     = NewValidationError("クエリは必須です")
	ErrInvalidSize       = NewValidationError("サイズは非負の値である必要があります")
	ErrInvalidFrom       = NewValidationError("fromは非負の値である必要があります")
	ErrSortFieldRequired = NewValidationError("ソートフィールドは必須です")
	ErrInvalidSortOrder  = NewValidationError("ソート順序は 'asc' または 'desc' である必要があります")
)

// ValidationError はバリデーションエラーを表す
type ValidationError struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// NewValidationError は新しいバリデーションエラーを作成する
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
		Time:    time.Now(),
	}
}

// Error は error インターフェースを実装する
func (e *ValidationError) Error() string {
	return e.Message
}
