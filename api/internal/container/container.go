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

// Container holds all dependencies
type Container struct {
	// Configuration
	Config *config.Config

	// Infrastructure
	ElasticsearchClient *elasticsearch.Client
	ElasticsearchRepo   repository.ElasticsearchRepository
	Logger              *log.Logger

	// Domain Services
	DocumentService *service.DocumentService
	SearchService   *service.SearchService

	// Use Cases
	DocumentUseCase *usecase.DocumentUseCase
	SearchUseCase   *usecase.SearchUseCase

	// Handlers
	DocumentHandler *handler.DocumentHandler
	SearchHandler   *handler.SearchHandler
	HealthHandler   *handler.HealthHandler

	// Middleware
	LoggingMiddleware *middleware.LoggingMiddleware
}

// NewContainer creates a new container with all dependencies
func NewContainer() (*Container, error) {
	container := &Container{}

	// Initialize configuration
	container.Config = config.NewConfig()

	// Initialize logger
	container.Logger = log.New(os.Stdout, "[ElasticSearch-API] ", log.LstdFlags|log.Lshortfile)

	// Initialize infrastructure
	if err := container.initInfrastructure(); err != nil {
		return nil, err
	}

	// Initialize domain services
	container.initDomainServices()

	// Initialize use cases
	container.initUseCases()

	// Initialize handlers
	container.initHandlers()

	// Initialize middleware
	container.initMiddleware()

	return container, nil
}

// initInfrastructure initializes infrastructure components
func (c *Container) initInfrastructure() error {
	var err error

	// Initialize Elasticsearch client
	c.ElasticsearchClient, err = elasticsearch.NewClient(c.Config)
	if err != nil {
		return err
	}

	// Initialize Elasticsearch repository
	c.ElasticsearchRepo = elasticsearch.NewRepository(c.ElasticsearchClient)

	return nil
}

// initDomainServices initializes domain services
func (c *Container) initDomainServices() {
	// Initialize document service
	c.DocumentService = service.NewDocumentService(c.ElasticsearchRepo)

	// Initialize search service
	c.SearchService = service.NewSearchService(c.ElasticsearchRepo)
}

// initUseCases initializes use cases
func (c *Container) initUseCases() {
	// Initialize document use case
	c.DocumentUseCase = usecase.NewDocumentUseCase(c.DocumentService)

	// Initialize search use case
	c.SearchUseCase = usecase.NewSearchUseCase(c.SearchService)
}

// initHandlers initializes handlers
func (c *Container) initHandlers() {
	// Initialize document handler
	c.DocumentHandler = handler.NewDocumentHandler(c.DocumentUseCase)

	// Initialize search handler
	c.SearchHandler = handler.NewSearchHandler(c.SearchUseCase)

	// Initialize health handler
	c.HealthHandler = handler.NewHealthHandler(c.ElasticsearchClient)
}

// initMiddleware initializes middleware
func (c *Container) initMiddleware() {
	// Initialize logging middleware
	c.LoggingMiddleware = middleware.NewLoggingMiddleware(c.Logger)
}

// Cleanup performs cleanup operations
func (c *Container) Cleanup() error {
	if c.ElasticsearchClient != nil {
		return c.ElasticsearchClient.Close()
	}
	return nil
}

// GetConfig returns the configuration
func (c *Container) GetConfig() *config.Config {
	return c.Config
}

// GetLogger returns the logger
func (c *Container) GetLogger() *log.Logger {
	return c.Logger
}

// GetElasticsearchClient returns the Elasticsearch client
func (c *Container) GetElasticsearchClient() *elasticsearch.Client {
	return c.ElasticsearchClient
}

// GetElasticsearchRepo returns the Elasticsearch repository
func (c *Container) GetElasticsearchRepo() repository.ElasticsearchRepository {
	return c.ElasticsearchRepo
}

// GetDocumentService returns the document service
func (c *Container) GetDocumentService() *service.DocumentService {
	return c.DocumentService
}

// GetSearchService returns the search service
func (c *Container) GetSearchService() *service.SearchService {
	return c.SearchService
}

// GetDocumentUseCase returns the document use case
func (c *Container) GetDocumentUseCase() *usecase.DocumentUseCase {
	return c.DocumentUseCase
}

// GetSearchUseCase returns the search use case
func (c *Container) GetSearchUseCase() *usecase.SearchUseCase {
	return c.SearchUseCase
}

// GetDocumentHandler returns the document handler
func (c *Container) GetDocumentHandler() *handler.DocumentHandler {
	return c.DocumentHandler
}

// GetSearchHandler returns the search handler
func (c *Container) GetSearchHandler() *handler.SearchHandler {
	return c.SearchHandler
}

// GetHealthHandler returns the health handler
func (c *Container) GetHealthHandler() *handler.HealthHandler {
	return c.HealthHandler
}

// GetLoggingMiddleware returns the logging middleware
func (c *Container) GetLoggingMiddleware() *middleware.LoggingMiddleware {
	return c.LoggingMiddleware
}

// Interface compliance checks
var (
	_ ContainerInterface = (*Container)(nil)
)

// ContainerInterface defines the container interface
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

// MockContainer is a mock container for testing
type MockContainer struct {
	MockConfig              *config.Config
	MockLogger              *log.Logger
	MockElasticsearchClient *elasticsearch.Client
	MockElasticsearchRepo   repository.ElasticsearchRepository
	MockDocumentService     *service.DocumentService
	MockSearchService       *service.SearchService
	MockDocumentUseCase     *usecase.DocumentUseCase
	MockSearchUseCase       *usecase.SearchUseCase
	MockDocumentHandler     *handler.DocumentHandler
	MockSearchHandler       *handler.SearchHandler
	MockHealthHandler       *handler.HealthHandler
	MockLoggingMiddleware   *middleware.LoggingMiddleware
}

// NewMockContainer creates a new mock container
func NewMockContainer() *MockContainer {
	return &MockContainer{}
}

// GetConfig returns mock config
func (m *MockContainer) GetConfig() *config.Config {
	return m.MockConfig
}

// GetLogger returns mock logger
func (m *MockContainer) GetLogger() *log.Logger {
	return m.MockLogger
}

// GetElasticsearchClient returns mock Elasticsearch client
func (m *MockContainer) GetElasticsearchClient() *elasticsearch.Client {
	return m.MockElasticsearchClient
}

// GetElasticsearchRepo returns mock Elasticsearch repository
func (m *MockContainer) GetElasticsearchRepo() repository.ElasticsearchRepository {
	return m.MockElasticsearchRepo
}

// GetDocumentService returns mock document service
func (m *MockContainer) GetDocumentService() *service.DocumentService {
	return m.MockDocumentService
}

// GetSearchService returns mock search service
func (m *MockContainer) GetSearchService() *service.SearchService {
	return m.MockSearchService
}

// GetDocumentUseCase returns mock document use case
func (m *MockContainer) GetDocumentUseCase() *usecase.DocumentUseCase {
	return m.MockDocumentUseCase
}

// GetSearchUseCase returns mock search use case
func (m *MockContainer) GetSearchUseCase() *usecase.SearchUseCase {
	return m.MockSearchUseCase
}

// GetDocumentHandler returns mock document handler
func (m *MockContainer) GetDocumentHandler() *handler.DocumentHandler {
	return m.MockDocumentHandler
}

// GetSearchHandler returns mock search handler
func (m *MockContainer) GetSearchHandler() *handler.SearchHandler {
	return m.MockSearchHandler
}

// GetHealthHandler returns mock health handler
func (m *MockContainer) GetHealthHandler() *handler.HealthHandler {
	return m.MockHealthHandler
}

// GetLoggingMiddleware returns mock logging middleware
func (m *MockContainer) GetLoggingMiddleware() *middleware.LoggingMiddleware {
	return m.MockLoggingMiddleware
}

// Cleanup performs mock cleanup
func (m *MockContainer) Cleanup() error {
	return nil
}

// Interface compliance check
var (
	_ ContainerInterface = (*MockContainer)(nil)
)
