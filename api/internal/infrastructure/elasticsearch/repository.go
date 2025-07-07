package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/Yuki-TU/elastic-search/api/internal/domain/entity"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/repository"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
)

// Repository はElasticsearchRepositoryインターフェースを実装する
type Repository struct {
	client *Client
}

// NewRepository は新しいElasticsearchリポジトリを作成する
func NewRepository(client *Client) repository.ElasticsearchRepository {
	return &Repository{
		client: client,
	}
}

// CreateDocument はElasticsearchに新しいドキュメントを作成する
func (r *Repository) CreateDocument(ctx context.Context, doc *entity.Document) error {
	// ドキュメントをJSONに変換
	body, err := json.Marshal(doc.Source)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to marshal document")
	}

	// ドキュメントを作成
	res, err := r.client.es.Index(
		doc.Index,
		bytes.NewReader(body),
		r.client.es.Index.WithContext(ctx),
		r.client.es.Index.WithDocumentID(doc.ID),
		r.client.es.Index.WithRefresh("true"),
	)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to index document")
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewAppError(errors.ErrCodeDocumentCreateFailed, fmt.Sprintf("Document indexing failed with status: %s", res.Status()))
	}

	// レスポンスを解析してドキュメントIDを取得
	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to parse index response")
	}

	// レスポンスからドキュメントIDを設定
	if id, ok := result["_id"].(string); ok {
		doc.SetID(id)
	}

	return nil
}

// GetDocument はIDでドキュメントを取得する
func (r *Repository) GetDocument(ctx context.Context, index, id string) (*entity.Document, error) {
	res, err := r.client.es.Get(
		index,
		id,
		r.client.es.Get.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentNotFound, "Failed to get document")
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, errors.NewDocumentNotFoundError(index, id)
		}
		return nil, errors.NewAppError(errors.ErrCodeDocumentNotFound, fmt.Sprintf("Document retrieval failed with status: %s", res.Status()))
	}

	// レスポンスを解析
	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentNotFound, "Failed to parse get response")
	}

	// ドキュメントデータを抽出
	source, ok := result["_source"].(map[string]any)
	if !ok {
		return nil, errors.NewAppError(errors.ErrCodeDocumentNotFound, "Invalid document format")
	}

	// ドキュメントエンティティを作成
	doc := entity.NewDocument(index, source)
	doc.SetID(id)

	// バージョンが利用可能な場合は設定
	if version, ok := result["_version"].(float64); ok {
		doc.Version = int64(version)
	}

	return doc, nil
}

// UpdateDocument は既存のドキュメントを更新する
func (r *Repository) UpdateDocument(ctx context.Context, doc *entity.Document) error {
	// ドキュメントをJSONに変換
	body, err := json.Marshal(doc.Source)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentUpdateFailed, "Failed to marshal document")
	}

	// ドキュメントを更新
	res, err := r.client.es.Index(
		doc.Index,
		bytes.NewReader(body),
		r.client.es.Index.WithContext(ctx),
		r.client.es.Index.WithDocumentID(doc.ID),
		r.client.es.Index.WithRefresh("true"),
	)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentUpdateFailed, "Failed to update document")
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewAppError(errors.ErrCodeDocumentUpdateFailed, fmt.Sprintf("Document update failed with status: %s", res.Status()))
	}

	// レスポンスを解析してバージョンを取得
	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentUpdateFailed, "Failed to parse update response")
	}

	// ドキュメントバージョンを更新
	if version, ok := result["_version"].(float64); ok {
		doc.Version = int64(version)
	}

	return nil
}

// DeleteDocument はIDでドキュメントを削除する
func (r *Repository) DeleteDocument(ctx context.Context, index, id string) error {
	res, err := r.client.es.Delete(
		index,
		id,
		r.client.es.Delete.WithContext(ctx),
		r.client.es.Delete.WithRefresh("true"),
	)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentDeleteFailed, "Failed to delete document")
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return errors.NewDocumentNotFoundError(index, id)
		}
		return errors.NewAppError(errors.ErrCodeDocumentDeleteFailed, fmt.Sprintf("Document deletion failed with status: %s", res.Status()))
	}

	return nil
}

// Search は検索操作を実行する
func (r *Repository) Search(ctx context.Context, query *entity.SearchQuery) (*entity.SearchResult, error) {
	// 検索クエリを構築
	searchQuery := r.buildSearchQuery(query)

	// クエリをJSONに変換
	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Failed to marshal search query")
	}

	// 検索を実行
	res, err := r.client.es.Search(
		r.client.es.Search.WithContext(ctx),
		r.client.es.Search.WithIndex(query.Index),
		r.client.es.Search.WithBody(bytes.NewReader(body)),
		r.client.es.Search.WithFrom(query.From),
		r.client.es.Search.WithSize(query.Size),
	)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Failed to perform search")
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.NewAppError(errors.ErrCodeSearchFailed, fmt.Sprintf("Search failed with status: %s", res.Status()))
	}

	// レスポンスを解析
	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Failed to parse search response")
	}

	// 検索結果を構築
	searchResult := r.buildSearchResult(query, result)

	return searchResult, nil
}

// MultiSearch は複数の検索操作を実行する
func (r *Repository) MultiSearch(ctx context.Context, queries []*entity.SearchQuery) ([]*entity.SearchResult, error) {
	// マルチ検索ボディを構築
	var body bytes.Buffer
	for _, query := range queries {
		// ヘッダー
		header := map[string]any{
			"index": query.Index,
		}
		headerJSON, _ := json.Marshal(header)
		body.Write(headerJSON)
		body.WriteByte('\n')

		// クエリ
		searchQuery := r.buildSearchQuery(query)
		queryJSON, _ := json.Marshal(searchQuery)
		body.Write(queryJSON)
		body.WriteByte('\n')
	}

	// マルチ検索を実行
	res, err := r.client.es.Msearch(
		&body,
		r.client.es.Msearch.WithContext(ctx),
	)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Failed to perform multi-search")
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.NewAppError(errors.ErrCodeSearchFailed, fmt.Sprintf("Multi-search failed with status: %s", res.Status()))
	}

	// レスポンスを解析
	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeSearchFailed, "Failed to parse multi-search response")
	}

	// 検索結果を構築
	var results []*entity.SearchResult
	if responses, ok := result["responses"].([]any); ok {
		for i, response := range responses {
			if responseMap, ok := response.(map[string]any); ok {
				searchResult := r.buildSearchResult(queries[i], responseMap)
				results = append(results, searchResult)
			}
		}
	}

	return results, nil
}

// CreateIndex は新しいインデックスを作成する
func (r *Repository) CreateIndex(ctx context.Context, index string, mapping map[string]any) error {
	// マッピングをJSONに変換
	body, err := json.Marshal(mapping)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeIndexCreateFailed, "Failed to marshal index mapping")
	}

	// インデックスを作成
	res, err := r.client.es.Indices.Create(
		index,
		r.client.es.Indices.Create.WithContext(ctx),
		r.client.es.Indices.Create.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeIndexCreateFailed, "Failed to create index")
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewAppError(errors.ErrCodeIndexCreateFailed, fmt.Sprintf("Index creation failed with status: %s", res.Status()))
	}

	return nil
}

// DeleteIndex はインデックスを削除する
func (r *Repository) DeleteIndex(ctx context.Context, index string) error {
	res, err := r.client.es.Indices.Delete(
		[]string{index},
		r.client.es.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeIndexDeleteFailed, "Failed to delete index")
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return errors.NewIndexNotFoundError(index)
		}
		return errors.NewAppError(errors.ErrCodeIndexDeleteFailed, fmt.Sprintf("Index deletion failed with status: %s", res.Status()))
	}

	return nil
}

// IndexExists はインデックスが存在するかを確認する
func (r *Repository) IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := r.client.es.Indices.Exists(
		[]string{index},
		r.client.es.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, errors.WrapError(err, errors.ErrCodeIndexNotFound, "Failed to check index existence")
	}
	defer res.Body.Close()

	return res.StatusCode == 200, nil
}

// BulkIndex はドキュメントのバルクインデックスを実行する
func (r *Repository) BulkIndex(ctx context.Context, documents []*entity.Document) error {
	// バルクボディを構築
	var body bytes.Buffer
	for _, doc := range documents {
		// アクションとメタデータ
		action := map[string]any{
			"index": map[string]any{
				"_index": doc.Index,
				"_id":    doc.ID,
			},
		}
		actionJSON, _ := json.Marshal(action)
		body.Write(actionJSON)
		body.WriteByte('\n')

		// ドキュメントソース
		sourceJSON, _ := json.Marshal(doc.Source)
		body.Write(sourceJSON)
		body.WriteByte('\n')
	}

	// バルク操作を実行
	res, err := r.client.es.Bulk(
		&body,
		r.client.es.Bulk.WithContext(ctx),
		r.client.es.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to perform bulk indexing")
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewAppError(errors.ErrCodeDocumentCreateFailed, fmt.Sprintf("Bulk indexing failed with status: %s", res.Status()))
	}

	return nil
}

// BulkDelete はドキュメントのバルク削除を実行する
func (r *Repository) BulkDelete(ctx context.Context, indices []string, ids []string) error {
	if len(indices) != len(ids) {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Indices and IDs arrays must have the same length")
	}

	// バルクボディを構築
	var body bytes.Buffer
	for i, index := range indices {
		// アクションとメタデータ
		action := map[string]any{
			"delete": map[string]any{
				"_index": index,
				"_id":    ids[i],
			},
		}
		actionJSON, _ := json.Marshal(action)
		body.Write(actionJSON)
		body.WriteByte('\n')
	}

	// バルク操作を実行
	res, err := r.client.es.Bulk(
		&body,
		r.client.es.Bulk.WithContext(ctx),
		r.client.es.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentDeleteFailed, "Failed to perform bulk deletion")
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewAppError(errors.ErrCodeDocumentDeleteFailed, fmt.Sprintf("Bulk deletion failed with status: %s", res.Status()))
	}

	return nil
}

// Health はElasticsearchクラスターの健康状態を返す
func (r *Repository) Health(ctx context.Context) error {
	healthy, err := r.client.IsHealthy(ctx)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeElasticsearchDown, "Failed to check cluster health")
	}

	if !healthy {
		return errors.NewAppError(errors.ErrCodeElasticsearchDown, "Elasticsearch cluster is not healthy")
	}

	return nil
}

// Info はElasticsearchクラスターの情報を返す
func (r *Repository) Info(ctx context.Context) (map[string]any, error) {
	info, err := r.client.Info(ctx)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeElasticsearchDown, "Failed to get cluster info")
	}

	return info, nil
}

// ヘルパーメソッド

// buildSearchQuery はSearchQueryエンティティからElasticsearchクエリを構築する
func (r *Repository) buildSearchQuery(query *entity.SearchQuery) map[string]any {
	esQuery := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query.Query,
				"fields": []string{"*"},
			},
		},
		"from": query.From,
		"size": query.Size,
	}

	// フィルターを追加
	if len(query.Filters) > 0 {
		filters := make([]map[string]any, 0, len(query.Filters))
		for field, value := range query.Filters {
			if field == "_facets" {
				// ファセット集約を処理
				continue
			}
			filters = append(filters, map[string]any{
				"term": map[string]any{
					field: value,
				},
			})
		}

		if len(filters) > 0 {
			esQuery["query"] = map[string]any{
				"bool": map[string]any{
					"must":   esQuery["query"],
					"filter": filters,
				},
			}
		}
	}

	// ソートを追加
	if len(query.Sort) > 0 {
		sort := make([]map[string]any, 0, len(query.Sort))
		for _, sortField := range query.Sort {
			sort = append(sort, map[string]any{
				sortField.Field: map[string]any{
					"order": sortField.Order,
				},
			})
		}
		esQuery["sort"] = sort
	}

	return esQuery
}

// buildSearchResult はElasticsearchレスポンスからSearchResultエンティティを構築する
func (r *Repository) buildSearchResult(query *entity.SearchQuery, result map[string]any) *entity.SearchResult {
	searchResult := entity.NewSearchResult(*query)

	// ヒットを抽出
	if hits, ok := result["hits"].(map[string]any); ok {
		// 総ヒット数
		if total, ok := hits["total"].(map[string]any); ok {
			if value, ok := total["value"].(float64); ok {
				searchResult.Total = int64(value)
			}
		}

		// 最大スコア
		if maxScore, ok := hits["max_score"].(float64); ok {
			searchResult.MaxScore = maxScore
		}

		// 個別のヒット
		if hitsList, ok := hits["hits"].([]any); ok {
			for _, hit := range hitsList {
				if hitMap, ok := hit.(map[string]any); ok {
					entityHit := entity.Hit{
						Index:  getString(hitMap, "_index"),
						ID:     getString(hitMap, "_id"),
						Score:  getFloat64(hitMap, "_score"),
						Source: getMap(hitMap, "_source"),
					}
					searchResult.AddHit(entityHit)
				}
			}
		}
	}

	// タイミング情報を抽出
	if took, ok := result["took"].(float64); ok {
		searchResult.Took = int64(took)
	}

	if timedOut, ok := result["timed_out"].(bool); ok {
		searchResult.TimedOut = timedOut
	}

	return searchResult
}

// 型変換用のヘルパー関数
func getString(m map[string]any, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64(m map[string]any, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0.0
}

func getMap(m map[string]any, key string) map[string]any {
	if val, ok := m[key].(map[string]any); ok {
		return val
	}
	return nil
}
