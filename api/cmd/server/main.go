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

// Server は HTTP サーバーを表す
type Server struct {
	httpServer *http.Server
	container  *container.Container
}

// NewServer は新しいサーバーインスタンスを作成する
func NewServer() (*Server, error) {
	// 依存関係注入コンテナを初期化
	cont, err := container.NewContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize container: %w", err)
	}

	// HTTP サーバーを作成
	server := &Server{
		container: cont,
	}

	// ルートとミドルウェアを設定
	server.setupServer()

	return server, nil
}

// setupServer は HTTP サーバーにルートとミドルウェアを設定する
func (s *Server) setupServer() {
	// メインルーターを作成
	mux := http.NewServeMux()

	// ルートを設定
	s.setupRoutes(mux)

	// ミドルウェアチェーンを設定
	handler := s.setupMiddleware(mux)

	// HTTP サーバーを設定
	port := s.container.GetConfig().Port
	if port == "" {
		port = "8080" // デフォルトポート
	}

	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// setupRoutes は全てのアプリケーションルートを設定する
func (s *Server) setupRoutes(mux *http.ServeMux) {
	// コンテナからハンドラーを取得
	documentHandler := s.container.GetDocumentHandler()
	searchHandler := s.container.GetSearchHandler()
	healthHandler := s.container.GetHealthHandler()

	// ドキュメントルート
	mux.HandleFunc("POST /documents", documentHandler.CreateDocument)
	mux.HandleFunc("GET /documents/{index}/{id}", documentHandler.GetDocument)
	mux.HandleFunc("PUT /documents/{index}/{id}", documentHandler.UpdateDocument)
	mux.HandleFunc("DELETE /documents/{index}/{id}", documentHandler.DeleteDocument)
	mux.HandleFunc("OPTIONS /documents", documentHandler.OptionsHandler)
	mux.HandleFunc("OPTIONS /documents/{index}/{id}", documentHandler.OptionsHandler)

	// 検索ルート
	mux.HandleFunc("GET /search", searchHandler.Search)
	mux.HandleFunc("POST /search", searchHandler.AdvancedSearch)
	mux.HandleFunc("OPTIONS /search", searchHandler.OptionsHandler)

	// ヘルスルート
	mux.HandleFunc("GET /health", healthHandler.HealthCheck)
	mux.HandleFunc("OPTIONS /health", healthHandler.OptionsHandler)
}

// setupMiddleware はミドルウェアチェーンを設定する
func (s *Server) setupMiddleware(handler http.Handler) http.Handler {
	logger := s.container.GetLogger()

	// ミドルウェアチェーンを作成
	middlewares := []func(http.Handler) http.Handler{
		// リカバリーミドルウェア（最初に配置）
		middleware.RecoveryMiddleware,

		// CORS ミドルウェア
		middleware.CORSMiddleware(middleware.DefaultCORSConfig()),

		// セキュリティミドルウェア
		middleware.SecurityMiddleware(middleware.DefaultSecurityConfig()),

		// リクエストサイズ制限（10MB）
		middleware.RequestSizeLimitMiddleware(10 * 1024 * 1024),

		// レート制限
		middleware.SimpleRateLimitMiddleware(middleware.DefaultRateLimitConfig()),

		// リクエストタイムアウト（30秒）
		middleware.RequestTimeoutMiddleware(30),

		// ログミドルウェア（リカバリー後、ビジネスロジック前に配置）
		middleware.StructuredLogMiddleware(logger),

		// エラーログミドルウェア
		middleware.ErrorLogMiddleware(logger),

		// 圧縮ミドルウェア
		middleware.CompressionMiddleware,
	}

	// ミドルウェアチェーンを適用
	return middleware.ChainMiddleware(middlewares...)(handler)
}

// Start は HTTP サーバーを開始する
func (s *Server) Start() error {
	logger := s.container.GetLogger()

	// サーバー起動情報をログ出力
	config := s.container.GetConfig()
	logger.Printf("Starting server on port %s", s.httpServer.Addr)
	logger.Printf("Environment: %s", config.Environment)
	logger.Printf("Elasticsearch URL: %s", config.ElasticsearchURL)

	// サーバーを開始
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop は HTTP サーバーを優雅に停止する
func (s *Server) Stop(ctx context.Context) error {
	logger := s.container.GetLogger()
	logger.Println("Shutting down server...")

	// HTTP サーバーをシャットダウン
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	// コンテナリソースをクリーンアップ
	if err := s.container.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup container: %w", err)
	}

	logger.Println("Server stopped successfully")
	return nil
}

// main はアプリケーションのエントリーポイント
func main() {
	// サーバーを作成
	server, err := NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 優雅なシャットダウンを設定
	go func() {
		// 割り込みシグナルを待機するチャンネルを作成
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// シグナルを待機
		<-c
		log.Println("Received shutdown signal")

		// タイムアウト付きのシャットダウンコンテキストを作成
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// サーバーを停止
		if err := server.Stop(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	// サーバーを開始
	log.Println("Starting Elasticsearch API server...")
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
