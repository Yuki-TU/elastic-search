package entity

// SearchQuery represents a search query structure
type SearchQuery struct {
	Query   string            `json:"query"`
	Index   string            `json:"index,omitempty"`
	Filters map[string]string `json:"filters,omitempty"`
	From    int               `json:"from"`
	Size    int               `json:"size"`
	Sort    []SortField       `json:"sort,omitempty"`
}

// SortField represents a field to sort by
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" or "desc"
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Query    SearchQuery `json:"query"`
	Hits     []Hit       `json:"hits"`
	Total    int64       `json:"total"`
	MaxScore float64     `json:"max_score"`
	Took     int64       `json:"took"`
	TimedOut bool        `json:"timed_out"`
}

// Hit represents a single search result
type Hit struct {
	Index  string         `json:"_index"`
	ID     string         `json:"_id"`
	Score  float64        `json:"_score"`
	Source map[string]any `json:"_source"`
}

// NewSearchQuery creates a new SearchQuery instance
func NewSearchQuery(query string) *SearchQuery {
	return &SearchQuery{
		Query:   query,
		Filters: make(map[string]string),
		From:    0,
		Size:    10,
		Sort:    []SortField{},
	}
}

// SetIndex sets the index to search in
func (sq *SearchQuery) SetIndex(index string) {
	sq.Index = index
}

// AddFilter adds a filter to the search query
func (sq *SearchQuery) AddFilter(field, value string) {
	sq.Filters[field] = value
}

// SetPagination sets the pagination parameters
func (sq *SearchQuery) SetPagination(from, size int) {
	sq.From = from
	sq.Size = size
}

// AddSort adds a sort field to the search query
func (sq *SearchQuery) AddSort(field, order string) {
	sq.Sort = append(sq.Sort, SortField{
		Field: field,
		Order: order,
	})
}

// NewSearchResult creates a new SearchResult instance
func NewSearchResult(query SearchQuery) *SearchResult {
	return &SearchResult{
		Query: query,
		Hits:  []Hit{},
		Total: 0,
		Took:  0,
	}
}

// AddHit adds a hit to the search result
func (sr *SearchResult) AddHit(hit Hit) {
	sr.Hits = append(sr.Hits, hit)
}

// HasResults returns true if there are search results
func (sr *SearchResult) HasResults() bool {
	return len(sr.Hits) > 0
}

// GetTotalPages returns the total number of pages
func (sr *SearchResult) GetTotalPages() int64 {
	if sr.Query.Size == 0 {
		return 0
	}
	return (sr.Total + int64(sr.Query.Size) - 1) / int64(sr.Query.Size)
}

// GetCurrentPage returns the current page number
func (sr *SearchResult) GetCurrentPage() int64 {
	if sr.Query.Size == 0 {
		return 0
	}
	return int64(sr.Query.From)/int64(sr.Query.Size) + 1
}
