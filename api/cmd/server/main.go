package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Yuki-TU/elastic-search/api/internal/container"
	"github.com/Yuki-TU/elastic-search/api/internal/interface/middleware"
)

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	container  *container.Container
}

// NewServer creates a new server instance
func NewServer() (*Server, error) {
	// Initialize dependency injection container
	cont, err := container.NewContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	// Create HTTP server
	server := &Server{
		container: cont,
	}

	// Setup routes and middleware
	server.setupServer()

	return server, nil
}

// setupServer configures the HTTP server with routes and middleware
func (s *Server) setupServer() {
	// Create main router
	mux := http.NewServeMux()

	// Setup routes
	s.setupRoutes(mux)

	// Setup middleware chain
	handler := s.setupMiddleware(mux)

	// Configure HTTP server
	port := s.container.GetConfig().Port
	if port == "" {
		port = "8080" // Default port
	}

	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// setupRoutes configures all application routes
func (s *Server) setupRoutes(mux *http.ServeMux) {
	// Get handlers from container
	documentHandler := s.container.GetDocumentHandler()
	searchHandler := s.container.GetSearchHandler()
	healthHandler := s.container.GetHealthHandler()

	// Document routes
	mux.HandleFunc("POST /documents", documentHandler.CreateDocument)
	mux.HandleFunc("GET /documents/{index}/{id}", documentHandler.GetDocument)
	mux.HandleFunc("PUT /documents/{index}/{id}", documentHandler.UpdateDocument)
	mux.HandleFunc("DELETE /documents/{index}/{id}", documentHandler.DeleteDocument)
	mux.HandleFunc("OPTIONS /documents", documentHandler.OptionsHandler)
	mux.HandleFunc("OPTIONS /documents/{index}/{id}", documentHandler.OptionsHandler)

	// Search routes
	mux.HandleFunc("GET /search", searchHandler.Search)
	mux.HandleFunc("POST /search", searchHandler.AdvancedSearch)
	mux.HandleFunc("OPTIONS /search", searchHandler.OptionsHandler)

	// Health routes
	mux.HandleFunc("GET /health", healthHandler.HealthCheck)
	mux.HandleFunc("OPTIONS /health", healthHandler.OptionsHandler)
}

// setupMiddleware configures the middleware chain
func (s *Server) setupMiddleware(handler http.Handler) http.Handler {
	logger := s.container.GetLogger()

	// Create middleware chain
	middlewares := []func(http.Handler) http.Handler{
		// Recovery middleware (should be first)
		middleware.RecoveryMiddleware,

		// CORS middleware
		middleware.CORSMiddleware(middleware.DefaultCORSConfig()),

		// Security middleware
		middleware.SecurityMiddleware(middleware.DefaultSecurityConfig()),

		// Request size limit (10MB)
		middleware.RequestSizeLimitMiddleware(10 * 1024 * 1024),

		// Rate limiting
		middleware.SimpleRateLimitMiddleware(middleware.DefaultRateLimitConfig()),

		// Request timeout (30 seconds)
		middleware.RequestTimeoutMiddleware(30),

		// Logging middleware (should be after recovery but before business logic)
		middleware.StructuredLogMiddleware(logger),

		// Error logging middleware
		middleware.ErrorLogMiddleware(logger),

		// Compression middleware
		middleware.CompressionMiddleware,
	}

	// Apply middleware chain
	return middleware.ChainMiddleware(middlewares...)(handler)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	logger := s.container.GetLogger()

	// Log server startup information
	config := s.container.GetConfig()
	logger.Printf("Starting server on port %s", s.httpServer.Addr)
	logger.Printf("Environment: %s", config.Environment)
	logger.Printf("Elasticsearch URL: %s", config.ElasticsearchURL)

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	logger := s.container.GetLogger()
	logger.Println("Shutting down server...")

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	// Cleanup container resources
	if err := s.container.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup container: %w", err)
	}

	logger.Println("Server stopped successfully")
	return nil
}

// main is the application entry point
func main() {
	// Create server
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup graceful shutdown
	go func() {
		// Create a channel to listen for interrupt signals
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Wait for signal
		<-c
		log.Println("Received shutdown signal")

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Stop server
		if err := server.Stop(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	// Start server
	log.Println("Starting Elasticsearch API server...")
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
