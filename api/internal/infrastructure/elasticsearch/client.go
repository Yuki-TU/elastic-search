package elasticsearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Yuki-TU/elastic-search/api/config"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
	"github.com/elastic/go-elasticsearch/v9"
)

// Client wraps the Elasticsearch client with additional functionality
type Client struct {
	es     *elasticsearch.Client
	config *config.Config
}

// ClientConfig represents the configuration for the Elasticsearch client
type ClientConfig struct {
	URLs                   []string
	Username               string
	Password               string
	APIKey                 string
	CertificateFingerprint string
	CloudID                string
	MaxRetries             int
	RetryOnStatus          []int
	RetryOnTimeout         bool
	EnableMetrics          bool
	EnableDebugLogger      bool
	EnableCompression      bool
	MaxIdleConnsPerHost    int
	ResponseHeaderTimeout  time.Duration
	RequestTimeout         time.Duration
	SniffOnStart           bool
	SniffInterval          time.Duration
	HealthcheckInterval    time.Duration
	DiscoverNodesOnStart   bool
	DiscoverNodesInterval  time.Duration
	EnableRetryOnTimeout   bool
	DisableRetry           bool
	UseResponseCheckOnly   bool
	CompressRequestBody    bool
}

// NewClient creates a new Elasticsearch client
func NewClient(conf *config.Config) (*Client, error) {
	// Create Elasticsearch configuration
	esConfig := elasticsearch.Config{
		Addresses: []string{conf.ElasticsearchURL},

		// Transport configuration
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 10 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},

		// Retry configuration
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			return time.Duration(i) * 100 * time.Millisecond
		},
		MaxRetries: 3,

		// Discover nodes configuration
		DiscoverNodesOnStart:  false,
		DiscoverNodesInterval: 60 * time.Second,

		// Health check configuration
		EnableMetrics:     true,
		EnableDebugLogger: conf.Environment == "development",
	}

	// Create Elasticsearch client
	es, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, errors.NewElasticsearchConnectionError(err)
	}

	client := &Client{
		es:     es,
		config: conf,
	}

	// Test the connection
	if err := client.ping(); err != nil {
		return nil, errors.NewElasticsearchConnectionError(err)
	}

	return client, nil
}

// NewClientWithConfig creates a new Elasticsearch client with custom configuration
func NewClientWithConfig(clientConfig *ClientConfig) (*Client, error) {
	// Create Elasticsearch configuration
	esConfig := elasticsearch.Config{
		Addresses: clientConfig.URLs,
		Username:  clientConfig.Username,
		Password:  clientConfig.Password,
		APIKey:    clientConfig.APIKey,
		CloudID:   clientConfig.CloudID,

		// Transport configuration
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   clientConfig.MaxIdleConnsPerHost,
			ResponseHeaderTimeout: clientConfig.ResponseHeaderTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},

		// Retry configuration
		RetryOnStatus: clientConfig.RetryOnStatus,
		RetryBackoff: func(i int) time.Duration {
			return time.Duration(i) * 100 * time.Millisecond
		},
		MaxRetries: clientConfig.MaxRetries,

		// Discover nodes configuration
		DiscoverNodesOnStart:  clientConfig.DiscoverNodesOnStart,
		DiscoverNodesInterval: clientConfig.DiscoverNodesInterval,

		// Health check configuration
		EnableMetrics:     clientConfig.EnableMetrics,
		EnableDebugLogger: clientConfig.EnableDebugLogger,
	}

	// Create Elasticsearch client
	es, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, errors.NewElasticsearchConnectionError(err)
	}

	client := &Client{
		es: es,
	}

	// Test the connection
	if err := client.ping(); err != nil {
		return nil, errors.NewElasticsearchConnectionError(err)
	}

	return client, nil
}

// GetClient returns the underlying Elasticsearch client
func (c *Client) GetClient() *elasticsearch.Client {
	return c.es
}

// Ping tests the connection to Elasticsearch
func (c *Client) Ping() error {
	return c.ping()
}

// ping performs a ping to Elasticsearch
func (c *Client) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.es.Ping(
		c.es.Ping.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping failed with status: %s", res.Status())
	}

	return nil
}

// Info returns information about the Elasticsearch cluster
func (c *Client) Info(ctx context.Context) (map[string]any, error) {
	res, err := c.es.Info(
		c.es.Info.WithContext(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("cluster info request failed with status: %s", res.Status())
	}

	var info map[string]any
	if err := c.parseResponse(res.Body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse cluster info response: %w", err)
	}

	return info, nil
}

// Health returns the health status of the Elasticsearch cluster
func (c *Client) Health(ctx context.Context) (map[string]any, error) {
	res, err := c.es.Cluster.Health(
		c.es.Cluster.Health.WithContext(ctx),
		c.es.Cluster.Health.WithLevel("cluster"),
		c.es.Cluster.Health.WithWaitForStatus("yellow"),
		c.es.Cluster.Health.WithTimeout(time.Second*30),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster health: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("cluster health request failed with status: %s", res.Status())
	}

	var health map[string]any
	if err := c.parseResponse(res.Body, &health); err != nil {
		return nil, fmt.Errorf("failed to parse cluster health response: %w", err)
	}

	return health, nil
}

// Stats returns cluster statistics
func (c *Client) Stats(ctx context.Context) (map[string]any, error) {
	res, err := c.es.Cluster.Stats(
		c.es.Cluster.Stats.WithContext(ctx),
		c.es.Cluster.Stats.WithNodeID("_local"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster stats: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("cluster stats request failed with status: %s", res.Status())
	}

	var stats map[string]any
	if err := c.parseResponse(res.Body, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse cluster stats response: %w", err)
	}

	return stats, nil
}

// Close closes the Elasticsearch client
func (c *Client) Close() error {
	// The elasticsearch client doesn't have a close method in v8
	// Connection pooling is handled automatically
	return nil
}

// IsHealthy checks if the Elasticsearch cluster is healthy
func (c *Client) IsHealthy(ctx context.Context) (bool, error) {
	health, err := c.Health(ctx)
	if err != nil {
		return false, err
	}

	status, ok := health["status"].(string)
	if !ok {
		return false, fmt.Errorf("invalid health status format")
	}

	return status == "green" || status == "yellow", nil
}

// WaitForHealthy waits for the cluster to become healthy
func (c *Client) WaitForHealthy(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for cluster to become healthy: %w", ctx.Err())
		case <-ticker.C:
			healthy, err := c.IsHealthy(ctx)
			if err != nil {
				log.Printf("Error checking cluster health: %v", err)
				continue
			}

			if healthy {
				return nil
			}
		}
	}
}

// GetVersion returns the Elasticsearch version
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	info, err := c.Info(ctx)
	if err != nil {
		return "", err
	}

	version, ok := info["version"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid version format in cluster info")
	}

	versionNumber, ok := version["number"].(string)
	if !ok {
		return "", fmt.Errorf("invalid version number format")
	}

	return versionNumber, nil
}

// GetClusterName returns the Elasticsearch cluster name
func (c *Client) GetClusterName(ctx context.Context) (string, error) {
	info, err := c.Info(ctx)
	if err != nil {
		return "", err
	}

	clusterName, ok := info["cluster_name"].(string)
	if !ok {
		return "", fmt.Errorf("invalid cluster name format")
	}

	return clusterName, nil
}

// EnableSniffer enables the sniffer to discover nodes
func (c *Client) EnableSniffer(interval time.Duration) {
	// In elasticsearch v8, sniffing is configured at client creation time
	// This method is kept for compatibility but doesn't do anything
	log.Printf("Sniffer configuration should be set during client creation")
}

// SetLogger sets a custom logger for the client
func (c *Client) SetLogger(logger any) {
	// In elasticsearch v8, logger is configured at client creation time
	// This method is kept for compatibility but doesn't do anything
	log.Printf("Logger configuration should be set during client creation")
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() *config.Config {
	return c.config
}

// parseResponse is a helper function to parse JSON response
func (c *Client) parseResponse(body io.Reader, v any) error {
	return json.NewDecoder(body).Decode(v)
}

// DefaultClientConfig returns a default client configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		URLs:                  []string{"http://localhost:9200"},
		MaxRetries:            3,
		RetryOnStatus:         []int{502, 503, 504, 429},
		RetryOnTimeout:        true,
		EnableMetrics:         true,
		EnableDebugLogger:     false,
		EnableCompression:     false,
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: 10 * time.Second,
		RequestTimeout:        30 * time.Second,
		SniffOnStart:          false,
		SniffInterval:         60 * time.Second,
		HealthcheckInterval:   30 * time.Second,
		DiscoverNodesOnStart:  false,
		DiscoverNodesInterval: 60 * time.Second,
		EnableRetryOnTimeout:  true,
		DisableRetry:          false,
		UseResponseCheckOnly:  false,
		CompressRequestBody:   false,
	}
}
