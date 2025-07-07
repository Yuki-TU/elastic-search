package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode はエラーコードを表す
type ErrorCode string

const (
	// ドキュメント関連のエラー
	ErrCodeDocumentNotFound     ErrorCode = "DOCUMENT_NOT_FOUND"
	ErrCodeDocumentExists       ErrorCode = "DOCUMENT_EXISTS"
	ErrCodeInvalidDocument      ErrorCode = "INVALID_DOCUMENT"
	ErrCodeDocumentCreateFailed ErrorCode = "DOCUMENT_CREATE_FAILED"
	ErrCodeDocumentUpdateFailed ErrorCode = "DOCUMENT_UPDATE_FAILED"
	ErrCodeDocumentDeleteFailed ErrorCode = "DOCUMENT_DELETE_FAILED"

	// 検索関連のエラー
	ErrCodeSearchFailed  ErrorCode = "SEARCH_FAILED"
	ErrCodeInvalidQuery  ErrorCode = "INVALID_QUERY"
	ErrCodeSearchTimeout ErrorCode = "SEARCH_TIMEOUT"

	// インデックス関連のエラー
	ErrCodeIndexNotFound     ErrorCode = "INDEX_NOT_FOUND"
	ErrCodeIndexExists       ErrorCode = "INDEX_EXISTS"
	ErrCodeIndexCreateFailed ErrorCode = "INDEX_CREATE_FAILED"
	ErrCodeIndexDeleteFailed ErrorCode = "INDEX_DELETE_FAILED"
	ErrCodeInvalidMapping    ErrorCode = "INVALID_MAPPING"

	// バリデーションエラー
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrCodeMissingParameter ErrorCode = "MISSING_PARAMETER"
	ErrCodeInvalidParameter ErrorCode = "INVALID_PARAMETER"

	// インフラストラクチャエラー
	ErrCodeElasticsearchDown ErrorCode = "ELASTICSEARCH_DOWN"
	ErrCodeConnectionFailed  ErrorCode = "CONNECTION_FAILED"
	ErrCodeTimeout           ErrorCode = "TIMEOUT"
	ErrCodeInternalError     ErrorCode = "INTERNAL_ERROR"

	// 認証/認可エラー
	ErrCodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden            ErrorCode = "FORBIDDEN"
	ErrCodeAuthenticationFailed ErrorCode = "AUTHENTICATION_FAILED"
)

// AppError はカスタムアプリケーションエラーを表す
type AppError struct {
	Code       ErrorCode      `json:"code"`
	Message    string         `json:"message"`
	Details    string         `json:"details,omitempty"`
	Cause      error          `json:"-"`
	Timestamp  time.Time      `json:"timestamp"`
	Context    map[string]any `json:"context,omitempty"`
	HTTPStatus int            `json:"-"`
}

// Error は error インターフェースを実装する
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap は基になるエラーを返す
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError は新しいアプリケーションエラーを作成する
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Timestamp:  time.Now(),
		Context:    make(map[string]any),
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// NewAppErrorWithCause は原因となるエラーを含む新しいアプリケーションエラーを作成する
func NewAppErrorWithCause(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Cause:      cause,
		Timestamp:  time.Now(),
		Context:    make(map[string]any),
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// NewAppErrorWithDetails は詳細情報を含む新しいアプリケーションエラーを作成する
func NewAppErrorWithDetails(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		Timestamp:  time.Now(),
		Context:    make(map[string]any),
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// WithContext はエラーにコンテキストを追加する
func (e *AppError) WithContext(key string, value any) *AppError {
	e.Context[key] = value
	return e
}

// WithHTTPStatus は HTTP ステータスコードを設定する
func (e *AppError) WithHTTPStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}

// getHTTPStatusForCode はエラーコードに対応する適切な HTTP ステータスコードを返す
func getHTTPStatusForCode(code ErrorCode) int {
	switch code {
	case ErrCodeDocumentNotFound, ErrCodeIndexNotFound:
		return http.StatusNotFound
	case ErrCodeDocumentExists, ErrCodeIndexExists:
		return http.StatusConflict
	case ErrCodeValidationFailed, ErrCodeInvalidRequest, ErrCodeMissingParameter,
		ErrCodeInvalidParameter, ErrCodeInvalidQuery, ErrCodeInvalidDocument, ErrCodeInvalidMapping:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeAuthenticationFailed:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeTimeout, ErrCodeSearchTimeout:
		return http.StatusRequestTimeout
	case ErrCodeElasticsearchDown, ErrCodeConnectionFailed:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// 共通エラーコンストラクタ
func NewDocumentNotFoundError(index, id string) *AppError {
	return NewAppError(ErrCodeDocumentNotFound, fmt.Sprintf("Document not found: %s/%s", index, id))
}

func NewDocumentExistsError(index, id string) *AppError {
	return NewAppError(ErrCodeDocumentExists, fmt.Sprintf("Document already exists: %s/%s", index, id))
}

func NewIndexNotFoundError(index string) *AppError {
	return NewAppError(ErrCodeIndexNotFound, fmt.Sprintf("Index not found: %s", index))
}

func NewIndexExistsError(index string) *AppError {
	return NewAppError(ErrCodeIndexExists, fmt.Sprintf("Index already exists: %s", index))
}

func NewValidationError(field, message string) *AppError {
	return NewAppError(ErrCodeValidationFailed, fmt.Sprintf("Validation failed for field '%s': %s", field, message))
}

func NewSearchError(query string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeSearchFailed, fmt.Sprintf("Search failed for query: %s", query), cause)
}

func NewElasticsearchConnectionError(cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeConnectionFailed, "Failed to connect to Elasticsearch", cause)
}

func NewTimeoutError(operation string) *AppError {
	return NewAppError(ErrCodeTimeout, fmt.Sprintf("Operation timed out: %s", operation))
}

func NewInternalError(message string, cause error) *AppError {
	return NewAppErrorWithCause(ErrCodeInternalError, message, cause)
}

// IsAppError はエラーが AppError かどうかをチェックする
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError はエラーから AppError を抽出する
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// WrapError は一般的なエラーを AppError にラップする
func WrapError(err error, code ErrorCode, message string) *AppError {
	return NewAppErrorWithCause(code, message, err)
}
