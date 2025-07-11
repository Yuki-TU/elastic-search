package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Yuki-TU/elastic-search/api/internal/domain/entity"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/repository"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
)

// Searcher は検索サービスのインターフェース
type Searcher interface {
	Search(ctx context.Context, queryStr string, index string, from, size int) (*entity.SearchResult, error)
	AdvancedSearch(ctx context.Context, queryStr string, index string, filters map[string]string, sortFields []entity.SortField, from, size int) (*entity.SearchResult, error)
	MultiSearch(ctx context.Context, queries []entity.SearchQuery) ([]*entity.SearchResult, error)
	SuggestSearch(ctx context.Context, queryStr string, index string, field string, size int) (*entity.SearchResult, error)
	FacetedSearch(ctx context.Context, queryStr string, index string, facetFields []string, from, size int) (*entity.SearchResult, error)
}

// SearchService は検索操作のビジネスロジックを提供する
type SearchService struct {
	repo repository.ElasticsearchRepository
}

// NewSearchService は新しいSearchServiceを作成する
func NewSearchService(repo repository.ElasticsearchRepository) *SearchService {
	return &SearchService{
		repo: repo,
	}
}

// Search は検索操作を実行する
func (s *SearchService) Search(ctx context.Context, queryStr string, index string, from, size int) (*entity.SearchResult, error) {
	// 入力を検証
	if queryStr == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Search query cannot be empty")
	}

	if size < 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Size must be non-negative")
	}

	if from < 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "From must be non-negative")
	}

	// デフォルト値を適用
	if size == 0 {
		size = 10
	}

	// 検索クエリを作成
	query := entity.NewSearchQuery(queryStr)
	query.SetIndex(index)
	query.SetPagination(from, size)

	// クエリにビジネスルールを適用
	if err := s.applySearchBusinessRules(query); err != nil {
		return nil, err
	}

	// 検索を実行
	result, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Search operation failed")
	}

	// 結果を後処理
	if err := s.postProcessSearchResults(result); err != nil {
		return nil, err
	}

	return result, nil
}

// AdvancedSearch はフィルターとソートを含む高度な検索を実行する
func (s *SearchService) AdvancedSearch(ctx context.Context, queryStr string, index string, filters map[string]string, sortFields []entity.SortField, from, size int) (*entity.SearchResult, error) {
	// 入力を検証
	if queryStr == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Search query cannot be empty")
	}

	if size < 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Size must be non-negative")
	}

	if from < 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "From must be non-negative")
	}

	// デフォルト値を適用
	if size == 0 {
		size = 10
	}

	// 検索クエリを作成
	query := entity.NewSearchQuery(queryStr)
	query.SetIndex(index)
	query.SetPagination(from, size)

	// フィルターを追加
	for field, value := range filters {
		if field != "" && value != "" {
			query.AddFilter(field, value)
		}
	}

	// ソートを追加
	for _, sortField := range sortFields {
		if sortField.Field != "" && (sortField.Order == "asc" || sortField.Order == "desc") {
			query.AddSort(sortField.Field, sortField.Order)
		}
	}

	// クエリにビジネスルールを適用
	if err := s.applySearchBusinessRules(query); err != nil {
		return nil, err
	}

	// 検索を実行
	result, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Advanced search operation failed")
	}

	// 結果を後処理
	if err := s.postProcessSearchResults(result); err != nil {
		return nil, err
	}

	return result, nil
}

// MultiSearch は一度のリクエストで複数の検索操作を実行する
func (s *SearchService) MultiSearch(ctx context.Context, queries []entity.SearchQuery) ([]*entity.SearchResult, error) {
	if len(queries) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "No search queries provided")
	}

	// 全てのクエリを検証
	for i, query := range queries {
		if err := s.validateSearchQuery(&query); err != nil {
			return nil, errors.NewAppError(errors.ErrCodeValidationFailed, fmt.Sprintf("Query %d validation failed: %v", i, err))
		}

		// 各クエリにビジネスルールを適用
		if err := s.applySearchBusinessRules(&query); err != nil {
			return nil, errors.NewAppError(errors.ErrCodeValidationFailed, fmt.Sprintf("Query %d business rule validation failed: %v", i, err))
		}
	}

	// クエリポインターに変換
	queryPointers := make([]*entity.SearchQuery, len(queries))
	for i := range queries {
		queryPointers[i] = &queries[i]
	}

	// マルチ検索を実行
	results, err := s.repo.MultiSearch(ctx, queryPointers)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Multi-search operation failed")
	}

	// 全ての結果を後処理
	for _, result := range results {
		if err := s.postProcessSearchResults(result); err != nil {
			return nil, err
		}
	}

	return results, nil
}

// SuggestSearch はサジェスト/オートコンプリート検索を実行する
func (s *SearchService) SuggestSearch(ctx context.Context, queryStr string, index string, field string, size int) (*entity.SearchResult, error) {
	if queryStr == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Search query cannot be empty")
	}

	if field == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Field for suggestion cannot be empty")
	}

	if size <= 0 {
		size = 5 // デフォルトサジェストサイズ
	}

	// サジェスト用のプレフィックスクエリを作成
	suggestQuery := fmt.Sprintf("%s*", queryStr)

	// 検索クエリを作成
	query := entity.NewSearchQuery(suggestQuery)
	query.SetIndex(index)
	query.SetPagination(0, size)

	// ビジネスルールを適用
	if err := s.applySearchBusinessRules(query); err != nil {
		return nil, err
	}

	// 検索を実行
	result, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Suggest search operation failed")
	}

	// 結果を後処理
	if err := s.postProcessSearchResults(result); err != nil {
		return nil, err
	}

	return result, nil
}

// FacetedSearch は集約を含むファセット検索を実行する
func (s *SearchService) FacetedSearch(ctx context.Context, queryStr string, index string, facetFields []string, from, size int) (*entity.SearchResult, error) {
	if queryStr == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Search query cannot be empty")
	}

	if len(facetFields) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Facet fields cannot be empty")
	}

	// Apply default values
	if size == 0 {
		size = 10
	}

	// Create search query
	query := entity.NewSearchQuery(queryStr)
	query.SetIndex(index)
	query.SetPagination(from, size)

	// Add facet information to query (this would be handled by the repository implementation)
	// For now, we'll store facet fields in the query filters as a special marker
	query.AddFilter("_facets", strings.Join(facetFields, ","))

	// Apply business rules
	if err := s.applySearchBusinessRules(query); err != nil {
		return nil, err
	}

	// Perform search
	result, err := s.repo.Search(ctx, query)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Faceted search operation failed")
	}

	// Post-process results
	if err := s.postProcessSearchResults(result); err != nil {
		return nil, err
	}

	return result, nil
}

// applySearchBusinessRules applies business rules to search queries
func (s *SearchService) applySearchBusinessRules(query *entity.SearchQuery) error {
	// Sanitize query string
	query.Query = s.sanitizeQuery(query.Query)

	// Apply maximum result size limit
	if query.Size > 1000 {
		query.Size = 1000
	}

	// Apply maximum offset limit
	if query.From > 10000 {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "From offset cannot exceed 10000")
	}

	// Add default sorting if none specified
	if len(query.Sort) == 0 {
		query.AddSort("_score", "desc")
	}

	// Validate sort fields
	for _, sortField := range query.Sort {
		if !s.isValidSortField(sortField.Field) {
			return errors.NewAppError(errors.ErrCodeValidationFailed, fmt.Sprintf("Invalid sort field: %s", sortField.Field))
		}
	}

	return nil
}

// postProcessSearchResults post-processes search results
func (s *SearchService) postProcessSearchResults(result *entity.SearchResult) error {
	if result == nil {
		return nil
	}

	// Apply business rules to results
	for i := range result.Hits {
		hit := &result.Hits[i]

		// Remove sensitive fields from results
		if hit.Source != nil {
			s.removeSensitiveFields(hit.Source)
		}

		// Add computed fields
		if err := s.addComputedFields(hit); err != nil {
			return err
		}
	}

	return nil
}

// validateSearchQuery validates a search query
func (s *SearchService) validateSearchQuery(query *entity.SearchQuery) error {
	if query.Query == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Search query cannot be empty")
	}

	if query.Size < 0 {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Size must be non-negative")
	}

	if query.From < 0 {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "From must be non-negative")
	}

	return nil
}

// sanitizeQuery sanitizes a search query string
func (s *SearchService) sanitizeQuery(query string) string {
	// Remove potentially dangerous characters
	query = strings.ReplaceAll(query, "<", "")
	query = strings.ReplaceAll(query, ">", "")
	query = strings.ReplaceAll(query, "\"", "\\\"")

	// Trim whitespace
	query = strings.TrimSpace(query)

	return query
}

// isValidSortField checks if a field is valid for sorting
func (s *SearchService) isValidSortField(field string) bool {
	// Define allowed sort fields
	allowedFields := map[string]bool{
		"_score":     true,
		"_id":        true,
		"created_at": true,
		"updated_at": true,
		"name":       true,
		"title":      true,
		"date":       true,
		"price":      true,
		"rating":     true,
	}

	return allowedFields[field]
}

// removeSensitiveFields removes sensitive fields from search results
func (s *SearchService) removeSensitiveFields(source map[string]any) {
	sensitiveFields := []string{
		"password",
		"password_hash",
		"secret",
		"token",
		"api_key",
		"private_key",
		"ssn",
		"credit_card",
	}

	for _, field := range sensitiveFields {
		delete(source, field)
	}
}

// addComputedFields adds computed fields to search results
func (s *SearchService) addComputedFields(hit *entity.Hit) error {
	if hit.Source == nil {
		return nil
	}

	// Add a computed field showing the match score category
	if hit.Score >= 0.8 {
		hit.Source["_match_quality"] = "high"
	} else if hit.Score >= 0.5 {
		hit.Source["_match_quality"] = "medium"
	} else {
		hit.Source["_match_quality"] = "low"
	}

	// Add index metadata
	hit.Source["_source_index"] = hit.Index

	return nil
}

// インターフェースの実装確認
var _ Searcher = (*SearchService)(nil)
