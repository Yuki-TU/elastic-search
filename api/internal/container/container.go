package container

import (
	"log"
	"os"

	"github.com/Yuki-TU/elastic-search/api/config"
	"github.com/Yuki-TU/elastic-search/api/internal/application/usecase"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/repository"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/service"
	"github.com/Yuki-TU/elastic-search/api/internal/infrastructure/elasticsearch"
	"github.com/Yuki-TU/elastic-search/api/internal/interface/handler"
	"github.com/Yuki-TU/elastic-search/api/internal/interface/middleware"
)

// Container は全ての依存関係を保持する
type Container struct {
	// 設定
	Config *config.Config

	// インフラストラクチャ
	ElasticsearchClient *elasticsearch.Client
	ElasticsearchRepo   repository.ElasticsearchRepository
	Logger              *log.Logger

	// ドメインサービス
	DocumentService *service.DocumentService
	SearchService   *service.SearchService

	// ユースケース
	DocumentUseCase *usecase.DocumentUseCase
	SearchUseCase   *usecase.SearchUseCase

	// ハンドラー
	DocumentHandler *handler.DocumentHandler
	SearchHandler   *handler.SearchHandler
	HealthHandler   *handler.HealthHandler

	// ミドルウェア
	LoggingMiddleware *middleware.LoggingMiddleware
}

// NewContainer は全ての依存関係を持つ新しいコンテナを作成する
func NewContainer() (*Container, error) {
	container := &Container{}

	// 設定を初期化
	container.Config = config.NewConfig()

	// ロガーを初期化
	container.Logger = log.New(os.Stdout, "[ElasticSearch-API] ", log.LstdFlags|log.Lshortfile)

	// インフラストラクチャを初期化
	if err := container.initInfrastructure(); err != nil {
		return nil, err
	}

	// ドメインサービスを初期化
	container.initDomainServices()

	// ユースケースを初期化
	container.initUseCases()

	// ハンドラーを初期化
	container.initHandlers()

	// ミドルウェアを初期化
	container.initMiddleware()

	return container, nil
}

// initInfrastructure はインフラストラクチャコンポーネントを初期化する
func (c *Container) initInfrastructure() error {
	var err error

	// Elasticsearchクライアントを初期化
	c.ElasticsearchClient, err = elasticsearch.NewClient(c.Config)
	if err != nil {
		return err
	}

	// Elasticsearchリポジトリを初期化
	c.ElasticsearchRepo = elasticsearch.NewRepository(c.ElasticsearchClient)

	return nil
}

// initDomainServices はドメインサービスを初期化する
func (c *Container) initDomainServices() {
	// ドキュメントサービスを初期化
	c.DocumentService = service.NewDocumentService(c.ElasticsearchRepo)

	// 検索サービスを初期化
	c.SearchService = service.NewSearchService(c.ElasticsearchRepo)
}

// initUseCases はユースケースを初期化する
func (c *Container) initUseCases() {
	// ドキュメントユースケースを初期化
	c.DocumentUseCase = usecase.NewDocumentUseCase(c.DocumentService)

	// 検索ユースケースを初期化
	c.SearchUseCase = usecase.NewSearchUseCase(c.SearchService)
}

// initHandlers はハンドラーを初期化する
func (c *Container) initHandlers() {
	// ドキュメントハンドラーを初期化
	c.DocumentHandler = handler.NewDocumentHandler(c.DocumentUseCase)

	// 検索ハンドラーを初期化
	c.SearchHandler = handler.NewSearchHandler(c.SearchUseCase)

	// ヘルスハンドラーを初期化
	c.HealthHandler = handler.NewHealthHandler(c.ElasticsearchClient)
}

// initMiddleware はミドルウェアを初期化する
func (c *Container) initMiddleware() {
	// ログミドルウェアを初期化
	c.LoggingMiddleware = middleware.NewLoggingMiddleware(c.Logger)
}

// Cleanup はクリーンアップ操作を実行する
func (c *Container) Cleanup() error {
	if c.ElasticsearchClient != nil {
		return c.ElasticsearchClient.Close()
	}
	return nil
}

// GetConfig は設定を返す
func (c *Container) GetConfig() *config.Config {
	return c.Config
}

// GetLogger はロガーを返す
func (c *Container) GetLogger() *log.Logger {
	return c.Logger
}

// GetElasticsearchClient はElasticsearchクライアントを返す
func (c *Container) GetElasticsearchClient() *elasticsearch.Client {
	return c.ElasticsearchClient
}

// GetElasticsearchRepo はElasticsearchリポジトリを返す
func (c *Container) GetElasticsearchRepo() repository.ElasticsearchRepository {
	return c.ElasticsearchRepo
}

// GetDocumentService はドキュメントサービスを返す
func (c *Container) GetDocumentService() *service.DocumentService {
	return c.DocumentService
}

// GetSearchService は検索サービスを返す
func (c *Container) GetSearchService() *service.SearchService {
	return c.SearchService
}

// GetDocumentUseCase はドキュメントユースケースを返す
func (c *Container) GetDocumentUseCase() *usecase.DocumentUseCase {
	return c.DocumentUseCase
}

// GetSearchUseCase は検索ユースケースを返す
func (c *Container) GetSearchUseCase() *usecase.SearchUseCase {
	return c.SearchUseCase
}

// GetDocumentHandler はドキュメントハンドラーを返す
func (c *Container) GetDocumentHandler() *handler.DocumentHandler {
	return c.DocumentHandler
}

// GetSearchHandler は検索ハンドラーを返す
func (c *Container) GetSearchHandler() *handler.SearchHandler {
	return c.SearchHandler
}

// GetHealthHandler はヘルスハンドラーを返す
func (c *Container) GetHealthHandler() *handler.HealthHandler {
	return c.HealthHandler
}

// GetLoggingMiddleware はログミドルウェアを返す
func (c *Container) GetLoggingMiddleware() *middleware.LoggingMiddleware {
	return c.LoggingMiddleware
}

// インターフェースの実装確認
var (
	_ ContainerInterface = (*Container)(nil)
)

// ContainerInterface はコンテナインターフェースを定義する
type ContainerInterface interface {
	GetConfig() *config.Config
	GetLogger() *log.Logger
	GetElasticsearchClient() *elasticsearch.Client
	GetElasticsearchRepo() repository.ElasticsearchRepository
	GetDocumentService() *service.DocumentService
	GetSearchService() *service.SearchService
	GetDocumentUseCase() *usecase.DocumentUseCase
	GetSearchUseCase() *usecase.SearchUseCase
	GetDocumentHandler() *handler.DocumentHandler
	GetSearchHandler() *handler.SearchHandler
	GetHealthHandler() *handler.HealthHandler
	GetLoggingMiddleware() *middleware.LoggingMiddleware
	Cleanup() error
}
