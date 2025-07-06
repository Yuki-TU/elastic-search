package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents error codes
type ErrorCode string

const (
	// Document related errors
	ErrCodeDocumentNotFound     ErrorCode = "DOCUMENT_NOT_FOUND"
	ErrCodeDocumentExists       ErrorCode = "DOCUMENT_EXISTS"
	ErrCodeInvalidDocument      ErrorCode = "INVALID_DOCUMENT"
	ErrCodeDocumentCreateFailed ErrorCode = "DOCUMENT_CREATE_FAILED"
	ErrCodeDocumentUpdateFailed ErrorCode = "DOCUMENT_UPDATE_FAILED"
	ErrCodeDocumentDeleteFailed ErrorCode = "DOCUMENT_DELETE_FAILED"

	// Search related errors
	ErrCodeSearchFailed  ErrorCode = "SEARCH_FAILED"
	ErrCodeInvalidQuery  ErrorCode = "INVALID_QUERY"
	ErrCodeSearchTimeout ErrorCode = "SEARCH_TIMEOUT"

	// Index related errors
	ErrCodeIndexNotFound     ErrorCode = "INDEX_NOT_FOUND"
	ErrCodeIndexExists       ErrorCode = "INDEX_EXISTS"
	ErrCodeIndexCreateFailed ErrorCode = "INDEX_CREATE_FAILED"
	ErrCodeIndexDeleteFailed ErrorCode = "INDEX_DELETE_FAILED"
	ErrCodeInvalidMapping    ErrorCode = "INVALID_MAPPING"

	// Validation errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrCodeMissingParameter ErrorCode = "MISSING_PARAMETER"
	ErrCodeInvalidParameter ErrorCode = "INVALID_PARAMETER"

	// Infrastructure errors
	ErrCodeElasticsearchDown ErrorCode = "ELASTICSEARCH_DOWN"
	ErrCodeConnectionFailed  ErrorCode = "CONNECTION_FAILED"
	ErrCodeTimeout           ErrorCode = "TIMEOUT"
	ErrCodeInternalError     ErrorCode = "INTERNAL_ERROR"

	// Authentication/Authorization errors
	ErrCodeUnauthorized         ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden            ErrorCode = "FORBIDDEN"
	ErrCodeAuthenticationFailed ErrorCode = "AUTHENTICATION_FAILED"
)

// AppError represents a custom application error
type AppError struct {
	Code       ErrorCode      `json:"code"`
	Message    string         `json:"message"`
	Details    string         `json:"details,omitempty"`
	Cause      error          `json:"-"`
	Timestamp  time.Time      `json:"timestamp"`
	Context    map[string]any `json:"context,omitempty"`
	HTTPStatus int            `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Timestamp:  time.Now(),
		Context:    make(map[string]any),
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// NewAppErrorWithCause creates a new application error with a cause
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

// NewAppErrorWithDetails creates a new application error with details
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

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value any) *AppError {
	e.Context[key] = value
	return e
}

// WithHTTPStatus sets the HTTP status code
func (e *AppError) WithHTTPStatus(status int) *AppError {
	e.HTTPStatus = status
	return e
}

// getHTTPStatusForCode returns the appropriate HTTP status code for an error code
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

// Common error constructors
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

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from error
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// WrapError wraps a generic error into an AppError
func WrapError(err error, code ErrorCode, message string) *AppError {
	return NewAppErrorWithCause(code, message, err)
}
