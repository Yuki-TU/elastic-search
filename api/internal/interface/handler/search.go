package handler

import (
	"net/http"
	"strconv"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/application/usecase"
	"github.com/Yuki-TU/elastic-search/api/pkg/utils"
)

// SearchHandler handles search-related HTTP requests
type SearchHandler struct {
	searchUseCase *usecase.SearchUseCase
}

// NewSearchHandler creates a new SearchHandler
func NewSearchHandler(searchUseCase *usecase.SearchUseCase) *SearchHandler {
	return &SearchHandler{
		searchUseCase: searchUseCase,
	}
}

// Search handles basic search requests
// GET /search?q={query}&index={index}&from={from}&size={size}
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// Set headers
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// Parse query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		rw.WriteBadRequestError("Query parameter 'q' is required")
		return
	}

	index := r.URL.Query().Get("index")
	from, _ := strconv.Atoi(r.URL.Query().Get("from"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	// Create search request
	req := &dto.SearchRequest{
		Query: query,
		Index: index,
		From:  from,
		Size:  size,
	}

	// Perform search
	result, err := h.searchUseCase.Search(ctx, req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// Return search results
	rw.WriteSearchResult(result)
}

// AdvancedSearch handles advanced search requests with filters and sorting
// POST /search
func (h *SearchHandler) AdvancedSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// Set headers
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// Parse request body
	var req dto.SearchRequest
	if err := utils.ParseRequestBody(r, &req); err != nil {
		rw.WriteError(err)
		return
	}

	// Perform advanced search
	result, err := h.searchUseCase.AdvancedSearch(ctx, &req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// Return search results
	rw.WriteSearchResult(result)
}

// OptionsHandler handles CORS preflight requests
func (h *SearchHandler) OptionsHandler(w http.ResponseWriter, r *http.Request) {
	utils.SetCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}
