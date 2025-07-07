package entity

// SearchQuery は検索クエリ構造を表す
type SearchQuery struct {
	Query   string            `json:"query"`
	Index   string            `json:"index,omitempty"`
	Filters map[string]string `json:"filters,omitempty"`
	From    int               `json:"from"`
	Size    int               `json:"size"`
	Sort    []SortField       `json:"sort,omitempty"`
}

// SortField はソートフィールドを表す
type SortField struct {
	Field string `json:"field"`
	Order string `json:"order"` // "asc" または "desc"
}

// SearchResult は検索操作の結果を表す
type SearchResult struct {
	Query    SearchQuery `json:"query"`
	Hits     []Hit       `json:"hits"`
	Total    int64       `json:"total"`
	MaxScore float64     `json:"max_score"`
	Took     int64       `json:"took"`
	TimedOut bool        `json:"timed_out"`
}

// Hit は単一の検索結果を表す
type Hit struct {
	Index  string         `json:"_index"`
	ID     string         `json:"_id"`
	Score  float64        `json:"_score"`
	Source map[string]any `json:"_source"`
}

// NewSearchQuery は新しい SearchQuery インスタンスを作成する
func NewSearchQuery(query string) *SearchQuery {
	return &SearchQuery{
		Query:   query,
		Filters: make(map[string]string),
		From:    0,
		Size:    10,
		Sort:    []SortField{},
	}
}

// SetIndex は検索対象のインデックスを設定する
func (sq *SearchQuery) SetIndex(index string) {
	sq.Index = index
}

// AddFilter は検索クエリにフィルターを追加する
func (sq *SearchQuery) AddFilter(field, value string) {
	sq.Filters[field] = value
}

// SetPagination はページネーションパラメータを設定する
func (sq *SearchQuery) SetPagination(from, size int) {
	sq.From = from
	sq.Size = size
}

// AddSort は検索クエリにソートフィールドを追加する
func (sq *SearchQuery) AddSort(field, order string) {
	sq.Sort = append(sq.Sort, SortField{
		Field: field,
		Order: order,
	})
}

// NewSearchResult は新しい SearchResult インスタンスを作成する
func NewSearchResult(query SearchQuery) *SearchResult {
	return &SearchResult{
		Query: query,
		Hits:  []Hit{},
		Total: 0,
		Took:  0,
	}
}

// AddHit は検索結果にヒットを追加する
func (sr *SearchResult) AddHit(hit Hit) {
	sr.Hits = append(sr.Hits, hit)
}

// HasResults は検索結果があるかどうかを返す
func (sr *SearchResult) HasResults() bool {
	return len(sr.Hits) > 0
}

// GetTotalPages は総ページ数を返す
func (sr *SearchResult) GetTotalPages() int64 {
	if sr.Query.Size == 0 {
		return 0
	}
	return (sr.Total + int64(sr.Query.Size) - 1) / int64(sr.Query.Size)
}

// GetCurrentPage は現在のページ番号を返す
func (sr *SearchResult) GetCurrentPage() int64 {
	if sr.Query.Size == 0 {
		return 0
	}
	return int64(sr.Query.From)/int64(sr.Query.Size) + 1
}
