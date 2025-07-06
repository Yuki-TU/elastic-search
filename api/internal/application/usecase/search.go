package usecase

import (
	"context"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/entity"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/service"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
)

// SearchUseCase handles search-related operations
type SearchUseCase struct {
	searchService *service.SearchService
}

// NewSearchUseCase creates a new SearchUseCase
func NewSearchUseCase(searchService *service.SearchService) *SearchUseCase {
	return &SearchUseCase{
		searchService: searchService,
	}
}

// Search performs a basic search operation
func (uc *SearchUseCase) Search(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Set defaults
	req.SetDefaults()

	// Perform search through domain service
	result, err := uc.searchService.Search(ctx, req.Query, req.Index, req.From, req.Size)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	return uc.entityToDTO(result), nil
}

// AdvancedSearch performs an advanced search with filters and sorting
func (uc *SearchUseCase) AdvancedSearch(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Set defaults
	req.SetDefaults()

	// Convert sort fields to entity type
	sortFields := make([]entity.SortField, len(req.Sort))
	for i, sort := range req.Sort {
		sortFields[i] = entity.SortField{
			Field: sort.Field,
			Order: sort.Order,
		}
	}

	// Perform advanced search through domain service
	result, err := uc.searchService.AdvancedSearch(ctx, req.Query, req.Index, req.Filters, sortFields, req.From, req.Size)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	return uc.entityToDTO(result), nil
}

// MultiSearch performs multiple search operations
func (uc *SearchUseCase) MultiSearch(ctx context.Context, requests []*dto.SearchRequest) ([]*dto.SearchResponse, error) {
	// Validate requests
	if len(requests) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "No search requests provided")
	}

	// Convert DTOs to entities
	queries := make([]entity.SearchQuery, len(requests))
	for i, req := range requests {
		if err := req.Validate(); err != nil {
			return nil, err
		}
		req.SetDefaults()

		query := entity.SearchQuery{
			Query:   req.Query,
			Index:   req.Index,
			Filters: req.Filters,
			From:    req.From,
			Size:    req.Size,
		}

		// Convert sort fields
		for _, sort := range req.Sort {
			query.Sort = append(query.Sort, entity.SortField{
				Field: sort.Field,
				Order: sort.Order,
			})
		}

		queries[i] = query
	}

	// Perform multi-search through domain service
	results, err := uc.searchService.MultiSearch(ctx, queries)
	if err != nil {
		return nil, err
	}

	// Convert results to DTOs
	responses := make([]*dto.SearchResponse, len(results))
	for i, result := range results {
		responses[i] = uc.entityToDTO(result)
	}

	return responses, nil
}

// SuggestSearch performs a suggest/autocomplete search
func (uc *SearchUseCase) SuggestSearch(ctx context.Context, query, index, field string, size int) (*dto.SearchResponse, error) {
	// Validate input
	if query == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Search query cannot be empty")
	}
	if field == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Field for suggestion cannot be empty")
	}

	// Set default size
	if size <= 0 {
		size = 5
	}

	// Perform suggest search through domain service
	result, err := uc.searchService.SuggestSearch(ctx, query, index, field, size)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	return uc.entityToDTO(result), nil
}

// FacetedSearch performs a faceted search with aggregations
func (uc *SearchUseCase) FacetedSearch(ctx context.Context, req *dto.SearchRequest, facetFields []string) (*dto.SearchResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if len(facetFields) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Facet fields cannot be empty")
	}

	// Set defaults
	req.SetDefaults()

	// Perform faceted search through domain service
	result, err := uc.searchService.FacetedSearch(ctx, req.Query, req.Index, facetFields, req.From, req.Size)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	return uc.entityToDTO(result), nil
}

// SearchByField performs a search within a specific field
func (uc *SearchUseCase) SearchByField(ctx context.Context, field, value, index string, from, size int) (*dto.SearchResponse, error) {
	// Validate input
	if field == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Field cannot be empty")
	}
	if value == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Value cannot be empty")
	}

	// Set defaults
	if size <= 0 {
		size = 10
	}
	if from < 0 {
		from = 0
	}

	// Create search request with field filter
	req := &dto.SearchRequest{
		Query:   value,
		Index:   index,
		Filters: map[string]string{field: value},
		From:    from,
		Size:    size,
	}

	// Perform search
	return uc.Search(ctx, req)
}

// SearchSimilar finds documents similar to a given document
func (uc *SearchUseCase) SearchSimilar(ctx context.Context, index, id string, fields []string, size int) (*dto.SearchResponse, error) {
	// Validate input
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}
	if id == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty")
	}

	// Set defaults
	if size <= 0 {
		size = 10
	}

	// This would typically use Elasticsearch's "more like this" functionality
	// For now, we'll return an empty result as this requires specific implementation
	result := &entity.SearchResult{
		Query: entity.SearchQuery{
			Query: "similar:" + id,
			Index: index,
			From:  0,
			Size:  size,
		},
		Hits:     []entity.Hit{},
		Total:    0,
		MaxScore: 0.0,
		Took:     0,
		TimedOut: false,
	}

	return uc.entityToDTO(result), nil
}

// GetSearchStatistics returns search statistics and analytics
func (uc *SearchUseCase) GetSearchStatistics(ctx context.Context, index string) (map[string]any, error) {
	// Validate input
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}

	// This would typically gather statistics from Elasticsearch
	// For now, we'll return mock statistics
	stats := map[string]any{
		"index":            index,
		"total_documents":  0,
		"search_count":     0,
		"average_response": "0ms",
		"popular_queries":  []string{},
		"recent_searches":  []string{},
		"failed_searches":  0,
		"success_rate":     "100%",
	}

	return stats, nil
}

// ValidateSearchQuery validates a search query without executing it
func (uc *SearchUseCase) ValidateSearchQuery(ctx context.Context, req *dto.SearchRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// Additional query validation can be added here
	// For example, checking for malicious queries, syntax validation, etc.

	return nil
}

// Helper method to convert entity to DTO
func (uc *SearchUseCase) entityToDTO(result *entity.SearchResult) *dto.SearchResponse {
	hits := make([]dto.HitDTO, len(result.Hits))
	for i, hit := range result.Hits {
		hits[i] = dto.HitDTO{
			Index:  hit.Index,
			ID:     hit.ID,
			Score:  hit.Score,
			Source: hit.Source,
		}
	}

	// Convert query
	queryDTO := dto.SearchQueryDTO{
		Query:   result.Query.Query,
		Index:   result.Query.Index,
		Filters: result.Query.Filters,
		From:    result.Query.From,
		Size:    result.Query.Size,
	}

	// Convert sort fields
	for _, sort := range result.Query.Sort {
		queryDTO.Sort = append(queryDTO.Sort, dto.SortFieldDTO{
			Field: sort.Field,
			Order: sort.Order,
		})
	}

	return &dto.SearchResponse{
		Query:    queryDTO,
		Results:  hits,
		Total:    result.Total,
		MaxScore: result.MaxScore,
		Took:     result.Took,
		TimedOut: result.TimedOut,
	}
}
