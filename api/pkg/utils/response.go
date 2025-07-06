package utils

import (
	"encoding/json"
	"net/http"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
)

// ResponseWriter provides utilities for writing HTTP responses
type ResponseWriter struct {
	writer http.ResponseWriter
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{writer: w}
}

// WriteJSON writes a JSON response
func (rw *ResponseWriter) WriteJSON(statusCode int, data any) error {
	rw.writer.Header().Set("Content-Type", "application/json")
	rw.writer.WriteHeader(statusCode)
	return json.NewEncoder(rw.writer).Encode(data)
}

// WriteSuccess writes data directly without wrapper
func (rw *ResponseWriter) WriteSuccess(data any, message string) error {
	return rw.WriteJSON(http.StatusOK, data)
}

// WriteError writes an error response
func (rw *ResponseWriter) WriteError(err error) error {
	if appErr := errors.GetAppError(err); appErr != nil {
		errorResponse := dto.NewErrorResponse(
			string(appErr.Code),
			appErr.Message,
			appErr.Details,
		)
		return rw.WriteJSON(appErr.HTTPStatus, errorResponse)
	}

	// Handle generic errors
	errorResponse := dto.NewErrorResponse(
		"INTERNAL_ERROR",
		"An internal error occurred",
		err.Error(),
	)
	return rw.WriteJSON(http.StatusInternalServerError, errorResponse)
}

// WriteValidationError writes a validation error response
func (rw *ResponseWriter) WriteValidationError(field, message string) error {
	appErr := errors.NewValidationError(field, message)
	return rw.WriteError(appErr)
}

// WriteNotFoundError writes a not found error response
func (rw *ResponseWriter) WriteNotFoundError(resource string) error {
	appErr := errors.NewAppError(errors.ErrCodeDocumentNotFound, "Resource not found: "+resource)
	return rw.WriteError(appErr)
}

// WriteBadRequestError writes a bad request error response
func (rw *ResponseWriter) WriteBadRequestError(message string) error {
	appErr := errors.NewAppError(errors.ErrCodeInvalidRequest, message)
	return rw.WriteError(appErr)
}

// WriteInternalError writes an internal server error response
func (rw *ResponseWriter) WriteInternalError(message string, cause error) error {
	appErr := errors.NewInternalError(message, cause)
	return rw.WriteError(appErr)
}

// WriteDocument writes a document response directly
func (rw *ResponseWriter) WriteDocument(document *dto.DocumentDTO, message string) error {
	return rw.WriteJSON(http.StatusOK, document)
}

// WriteSearchResult writes a search result response
func (rw *ResponseWriter) WriteSearchResult(result *dto.SearchResponse) error {
	return rw.WriteJSON(http.StatusOK, result)
}

// WriteCreated writes a created response with the data directly
func (rw *ResponseWriter) WriteCreated(data any, message string) error {
	return rw.WriteJSON(http.StatusCreated, data)
}

// WriteNoContent writes a no content response
func (rw *ResponseWriter) WriteNoContent() error {
	rw.writer.WriteHeader(http.StatusNoContent)
	return nil
}

// Helper functions for common response patterns

// WriteJSONError writes a JSON error response with custom status code
func WriteJSONError(w http.ResponseWriter, statusCode int, code, message, details string) error {
	errorResponse := dto.NewErrorResponse(code, message, details)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(errorResponse)
}

// WriteJSONSuccess writes data directly without wrapper
func WriteJSONSuccess(w http.ResponseWriter, statusCode int, data any, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// SetCORSHeaders sets CORS headers
func SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

// SetSecurityHeaders sets security headers
func SetSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
}

// ParseRequestBody parses JSON request body
func ParseRequestBody(r *http.Request, v any) error {
	if r.Body == nil {
		return errors.NewAppError(errors.ErrCodeInvalidRequest, "Request body is empty")
	}

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.NewAppError(errors.ErrCodeInvalidRequest, "Invalid JSON format: "+err.Error())
	}

	return nil
}
