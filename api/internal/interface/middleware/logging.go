package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// RequestIDKey is the key for request ID in context
type RequestIDKey struct{}

// LoggingMiddleware provides request logging functionality
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// LogRequest logs HTTP requests with timing information
func (m *LoggingMiddleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID
		requestID := generateRequestID()

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
		r = r.WithContext(ctx)

		// Start timer
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Log request start
		m.logger.Printf("[%s] %s %s - Request started", requestID, r.Method, r.URL.Path)

		// Process request
		next.ServeHTTP(ww, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request completion
		m.logger.Printf("[%s] %s %s - %d - %v",
			requestID, r.Method, r.URL.Path, ww.statusCode, duration)
	})
}

// LogRequestWithBody logs HTTP requests including body information
func (m *LoggingMiddleware) LogRequestWithBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID
		requestID := generateRequestID()

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
		r = r.WithContext(ctx)

		// Start timer
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Log request details
		m.logger.Printf("[%s] %s %s - %s - Content-Length: %d - Request started",
			requestID, r.Method, r.URL.Path, r.RemoteAddr, r.ContentLength)

		// Process request
		next.ServeHTTP(ww, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request completion with more details
		m.logger.Printf("[%s] %s %s - %s - %d - %v - User-Agent: %s",
			requestID, r.Method, r.URL.Path, r.RemoteAddr, ww.statusCode, duration, r.UserAgent())
	})
}

// LogError logs errors with context
func (m *LoggingMiddleware) LogError(requestID string, err error, message string) {
	if requestID == "" {
		requestID = "unknown"
	}
	m.logger.Printf("[%s] ERROR: %s - %v", requestID, message, err)
}

// LogInfo logs informational messages
func (m *LoggingMiddleware) LogInfo(requestID string, message string) {
	if requestID == "" {
		requestID = "unknown"
	}
	m.logger.Printf("[%s] INFO: %s", requestID, message)
}

// LogWarning logs warning messages
func (m *LoggingMiddleware) LogWarning(requestID string, message string) {
	if requestID == "" {
		requestID = "unknown"
	}
	m.logger.Printf("[%s] WARNING: %s", requestID, message)
}

// LogDebug logs debug messages
func (m *LoggingMiddleware) LogDebug(requestID string, message string) {
	if requestID == "" {
		requestID = "unknown"
	}
	m.logger.Printf("[%s] DEBUG: %s", requestID, message)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// AccessLogMiddleware provides access log functionality
func AccessLogMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate request ID
			requestID := generateRequestID()

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
			r = r.WithContext(ctx)

			// Start timer
			start := time.Now()

			// Wrap response writer to capture status code
			ww := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)

			// Log in common log format
			logger.Printf("%s - - [%s] \"%s %s %s\" %d %d %v",
				r.RemoteAddr,
				start.Format("02/Jan/2006:15:04:05 -0700"),
				r.Method,
				r.URL.Path,
				r.Proto,
				ww.statusCode,
				0, // Response size (not captured in this simple implementation)
				duration,
			)
		})
	}
}

// ErrorLogMiddleware provides error logging functionality
func ErrorLogMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap response writer to capture status code
			ww := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(ww, r)

			// Log errors (4xx and 5xx status codes)
			if ww.statusCode >= 400 {
				requestID := GetRequestID(r.Context())
				logger.Printf("[%s] ERROR: %s %s - Status: %d - Remote: %s - User-Agent: %s",
					requestID,
					r.Method,
					r.URL.Path,
					ww.statusCode,
					r.RemoteAddr,
					r.UserAgent(),
				)
			}
		})
	}
}

// StructuredLogMiddleware provides structured logging
func StructuredLogMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate request ID
			requestID := generateRequestID()

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
			r = r.WithContext(ctx)

			// Start timer
			start := time.Now()

			// Wrap response writer to capture status code
			ww := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(ww, r)

			// Calculate duration
			duration := time.Since(start)

			// Structured log entry
			logger.Printf(`{"request_id":"%s","method":"%s","path":"%s","status":%d,"duration_ms":%d,"remote_addr":"%s","user_agent":"%s","timestamp":"%s"}`,
				requestID,
				r.Method,
				r.URL.Path,
				ww.statusCode,
				duration.Milliseconds(),
				r.RemoteAddr,
				r.UserAgent(),
				start.Format(time.RFC3339),
			)
		})
	}
}
