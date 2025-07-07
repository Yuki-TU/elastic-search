package handler

import (
	"net/http"
	"strconv"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/application/usecase"
	"github.com/Yuki-TU/elastic-search/api/pkg/utils"
)

// SearchHandler は検索関連のHTTPリクエストを処理する
type SearchHandler struct {
	searchUseCase *usecase.SearchUseCase
}

// NewSearchHandler は新しい SearchHandler を作成する
func NewSearchHandler(searchUseCase *usecase.SearchUseCase) *SearchHandler {
	return &SearchHandler{
		searchUseCase: searchUseCase,
	}
}

// Search は基本的な検索リクエストを処理する
// GET /search?q={query}&index={index}&from={from}&size={size}
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// ヘッダーを設定
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// クエリパラメータを解析
	query := r.URL.Query().Get("q")
	if query == "" {
		rw.WriteBadRequestError("Query parameter 'q' is required")
		return
	}

	index := r.URL.Query().Get("index")
	from, _ := strconv.Atoi(r.URL.Query().Get("from"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))

	// 検索リクエストを作成
	req := &dto.SearchRequest{
		Query: query,
		Index: index,
		From:  from,
		Size:  size,
	}

	// 検索を実行
	result, err := h.searchUseCase.Search(ctx, req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// 検索結果を返す
	rw.WriteSearchResult(result)
}

// AdvancedSearch はフィルターとソートを含む高度な検索リクエストを処理する
// POST /search
func (h *SearchHandler) AdvancedSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// ヘッダーを設定
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// リクエストボディを解析
	var req dto.SearchRequest
	if err := utils.ParseRequestBody(r, &req); err != nil {
		rw.WriteError(err)
		return
	}

	// 高度な検索を実行
	result, err := h.searchUseCase.AdvancedSearch(ctx, &req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// 検索結果を返す
	rw.WriteSearchResult(result)
}

// OptionsHandler はCORSプリフライトリクエストを処理する
func (h *SearchHandler) OptionsHandler(w http.ResponseWriter, r *http.Request) {
	utils.SetCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}
