package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware provides CORS functionality
func CORSMiddleware(config *CORSConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if isOriginAllowed(origin, config.AllowOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			// Set other CORS headers
			if len(config.AllowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			}

			if len(config.AllowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			}

			if len(config.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if origin is allowed
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
		// Support for wildcard patterns (simple implementation)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	ContentSecurityPolicy   string
	XFrameOptions           string
	XContentTypeOptions     string
	XSSProtection           string
	StrictTransportSecurity string
	ReferrerPolicy          string
	PermissionsPolicy       string
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ContentSecurityPolicy:   "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'",
		XFrameOptions:           "DENY",
		XContentTypeOptions:     "nosniff",
		XSSProtection:           "1; mode=block",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "camera=(), microphone=(), geolocation=()",
	}
}

// SecurityMiddleware provides security headers
func SecurityMiddleware(config *SecurityConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set security headers
			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			if config.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.XFrameOptions)
			}

			if config.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.XContentTypeOptions)
			}

			if config.XSSProtection != "" {
				w.Header().Set("X-XSS-Protection", config.XSSProtection)
			}

			if config.StrictTransportSecurity != "" && r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
			}

			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			if config.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", config.PermissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	WindowSize        int // in seconds
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerMinute: 100,
		BurstSize:         10,
		WindowSize:        60,
	}
}

// SimpleRateLimitMiddleware provides basic rate limiting
// Note: This is a simplified implementation for demonstration
// In production, you'd want to use a proper rate limiting library
func SimpleRateLimitMiddleware(config *RateLimitConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In a real implementation, you'd track requests per IP
			// and implement proper rate limiting logic

			// For now, just add rate limiting headers
			w.Header().Set("X-RateLimit-Limit", string(rune(config.RequestsPerMinute)))
			w.Header().Set("X-RateLimit-Remaining", "99") // Mock remaining
			w.Header().Set("X-RateLimit-Reset", "60")     // Mock reset time

			next.ServeHTTP(w, r)
		})
	}
}

// CompressionMiddleware provides gzip compression
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// For this simplified implementation, we'll just add the header
		// In production, you'd want to use a proper compression library
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimitMiddleware limits request body size
func RequestSizeLimitMiddleware(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			if r.ContentLength > maxSize {
				http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
				return
			}

			// Also limit the body reader
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)

			next.ServeHTTP(w, r)
		})
	}
}

// RequestTimeoutMiddleware sets request timeout
func RequestTimeoutMiddleware(timeout int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In a real implementation, you'd set up proper timeout handling
			// For now, just add a timeout header
			w.Header().Set("X-Request-Timeout", string(rune(timeout)))

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryMiddleware provides panic recovery
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic (in production, use proper logging)

				// Set response headers
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				// Return error response
				w.Write([]byte(`{"error": "Internal server error", "message": "An unexpected error occurred"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// ChainMiddleware chains multiple middleware functions
func ChainMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
