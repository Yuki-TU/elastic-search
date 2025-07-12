package usecase

import (
	"context"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/entity"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/service"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
)

// SearchUseCaser は検索ユースケースのインターフェース
type SearchUseCaser interface {
	Search(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error)
	AdvancedSearch(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error)
	MultiSearch(ctx context.Context, requests []*dto.SearchRequest) ([]*dto.SearchResponse, error)
	SuggestSearch(ctx context.Context, query, index, field string, size int) (*dto.SearchResponse, error)
	FacetedSearch(ctx context.Context, req *dto.SearchRequest, facetFields []string) (*dto.SearchResponse, error)
	SearchByField(ctx context.Context, field, value, index string, from, size int) (*dto.SearchResponse, error)
	SearchSimilar(ctx context.Context, index, id string, fields []string, size int) (*dto.SearchResponse, error)
	GetSearchStatistics(ctx context.Context, index string) (map[string]any, error)
	ValidateSearchQuery(ctx context.Context, req *dto.SearchRequest) error
}

// SearchUseCase は検索関連の操作を処理する
type SearchUseCase struct {
	searchService service.Searcher
}

// NewSearchUseCase は新しい SearchUseCase を作成する
func NewSearchUseCase(searchService service.Searcher) *SearchUseCase {
	return &SearchUseCase{
		searchService: searchService,
	}
}

// Search は基本的な検索操作を実行する
func (uc *SearchUseCase) Search(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	// リクエストを検証
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// デフォルト値を設定
	req.SetDefaults()

	// ドメインサービスを通じて検索を実行
	result, err := uc.searchService.Search(ctx, req.Query, req.Index, req.From, req.Size)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	return uc.entityToDTO(result), nil
}

// AdvancedSearch はフィルターとソートを含む高度な検索を実行する
func (uc *SearchUseCase) AdvancedSearch(ctx context.Context, req *dto.SearchRequest) (*dto.SearchResponse, error) {
	// リクエストを検証
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// デフォルト値を設定
	req.SetDefaults()

	// ソートフィールドをエンティティ型に変換
	sortFields := make([]entity.SortField, len(req.Sort))
	for i, sort := range req.Sort {
		sortFields[i] = entity.SortField{
			Field: sort.Field,
			Order: sort.Order,
		}
	}

	// ドメインサービスを通じて高度な検索を実行
	result, err := uc.searchService.AdvancedSearch(ctx, req.Query, req.Index, req.Filters, sortFields, req.From, req.Size)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	return uc.entityToDTO(result), nil
}

// MultiSearch は複数の検索操作を実行する
func (uc *SearchUseCase) MultiSearch(ctx context.Context, requests []*dto.SearchRequest) ([]*dto.SearchResponse, error) {
	// リクエストを検証
	if len(requests) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "検索リクエストが提供されていません")
	}

	// DTOをエンティティに変換
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

		// ソートフィールドを変換
		for _, sort := range req.Sort {
			query.Sort = append(query.Sort, entity.SortField{
				Field: sort.Field,
				Order: sort.Order,
			})
		}

		queries[i] = query
	}

	// ドメインサービスを通じてマルチ検索を実行
	results, err := uc.searchService.MultiSearch(ctx, queries)
	if err != nil {
		return nil, err
	}

	// 結果をDTOに変換
	responses := make([]*dto.SearchResponse, len(results))
	for i, result := range results {
		responses[i] = uc.entityToDTO(result)
	}

	return responses, nil
}

// SuggestSearch はサジェスト/オートコンプリート検索を実行する
func (uc *SearchUseCase) SuggestSearch(ctx context.Context, query, index, field string, size int) (*dto.SearchResponse, error) {
	// 入力を検証
	if query == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "検索クエリは空にできません")
	}
	if field == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "サジェスト用のフィールドは空にできません")
	}

	// デフォルトサイズを設定
	if size <= 0 {
		size = 5
	}

	// ドメインサービスを通じてサジェスト検索を実行
	result, err := uc.searchService.SuggestSearch(ctx, query, index, field, size)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	return uc.entityToDTO(result), nil
}

// FacetedSearch は集約を含むファセット検索を実行する
func (uc *SearchUseCase) FacetedSearch(ctx context.Context, req *dto.SearchRequest, facetFields []string) (*dto.SearchResponse, error) {
	// リクエストを検証
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if len(facetFields) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "ファセットフィールドは空にできません")
	}

	// デフォルト値を設定
	req.SetDefaults()

	// ドメインサービスを通じてファセット検索を実行
	result, err := uc.searchService.FacetedSearch(ctx, req.Query, req.Index, facetFields, req.From, req.Size)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	return uc.entityToDTO(result), nil
}

// SearchByField は特定のフィールド内で検索を実行する
func (uc *SearchUseCase) SearchByField(ctx context.Context, field, value, index string, from, size int) (*dto.SearchResponse, error) {
	// 入力を検証
	if field == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "フィールドは空にできません")
	}
	if value == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "値は空にできません")
	}

	// デフォルト値を設定
	if size <= 0 {
		size = 10
	}
	if from < 0 {
		from = 0
	}

	// フィールドフィルター付きの検索リクエストを作成
	req := &dto.SearchRequest{
		Query:   value,
		Index:   index,
		Filters: map[string]string{field: value},
		From:    from,
		Size:    size,
	}

	// 検索を実行
	return uc.Search(ctx, req)
}

// SearchSimilar は指定されたドキュメントに類似したドキュメントを検索する
func (uc *SearchUseCase) SearchSimilar(ctx context.Context, index, id string, fields []string, size int) (*dto.SearchResponse, error) {
	// 入力を検証
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "インデックスは空にできません")
	}
	if id == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "ドキュメントIDは空にできません")
	}

	// デフォルト値を設定
	if size <= 0 {
		size = 10
	}

	// 通常はElasticsearchの「more like this」機能を使用する
	// 今は特定の実装が必要なため空の結果を返す
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

// GetSearchStatistics は検索統計と分析を返す
func (uc *SearchUseCase) GetSearchStatistics(ctx context.Context, index string) (map[string]any, error) {
	// 入力を検証
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "インデックスは空にできません")
	}

	// 通常はElasticsearchから統計を収集する
	// 今はモック統計を返す
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

// ValidateSearchQuery は検索クエリを実行せずに検証する
func (uc *SearchUseCase) ValidateSearchQuery(ctx context.Context, req *dto.SearchRequest) error {
	// リクエストを検証
	if err := req.Validate(); err != nil {
		return err
	}

	// 追加のクエリ検証をここに追加できる
	// 例：悪意のあるクエリのチェック、構文検証など

	return nil
}

// entityToDTO はエンティティをDTOに変換するヘルパーメソッド
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

	// クエリを変換
	queryDTO := dto.SearchQueryDTO{
		Query:   result.Query.Query,
		Index:   result.Query.Index,
		Filters: result.Query.Filters,
		From:    result.Query.From,
		Size:    result.Query.Size,
	}

	// ソートフィールドを変換
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
