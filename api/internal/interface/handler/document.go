package handler

import (
	"net/http"
	"strings"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/application/usecase"
	"github.com/Yuki-TU/elastic-search/api/pkg/utils"
)

// DocumentHandler はドキュメント関連のHTTPリクエストを処理する
type DocumentHandler struct {
	documentUseCase *usecase.DocumentUseCase
}

// NewDocumentHandler は新しい DocumentHandler を作成する
func NewDocumentHandler(documentUseCase *usecase.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{
		documentUseCase: documentUseCase,
	}
}

// CreateDocument はドキュメント作成リクエストを処理する
// POST /documents
func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// ヘッダーを設定
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// リクエストボディを解析
	var req dto.CreateDocumentRequest
	if err := utils.ParseRequestBody(r, &req); err != nil {
		rw.WriteError(err)
		return
	}

	// ドキュメントを作成
	result, err := h.documentUseCase.CreateDocument(ctx, &req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// 成功レスポンスを返す
	rw.WriteCreated(result, "Document created successfully")
}

// GetDocument はドキュメント取得リクエストを処理する
// GET /documents/{index}/{id}
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// ヘッダーを設定
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// パスパラメータを抽出
	index := h.getPathParam(r, "index")
	id := h.getPathParam(r, "id")

	if index == "" || id == "" {
		rw.WriteBadRequestError("Index and ID are required")
		return
	}

	// ドキュメントを取得
	result, err := h.documentUseCase.GetDocument(ctx, index, id)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// 成功レスポンスを返す
	rw.WriteDocument(result, "Document retrieved successfully")
}

// UpdateDocument はドキュメント更新/作成リクエストを処理する
// PUT /documents/{index}/{id}
func (h *DocumentHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// ヘッダーを設定
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// パスパラメータを抽出
	index := h.getPathParam(r, "index")
	id := h.getPathParam(r, "id")

	if index == "" || id == "" {
		rw.WriteBadRequestError("Index and ID are required")
		return
	}

	// リクエストボディを解析
	var req dto.UpdateDocumentRequest
	if err := utils.ParseRequestBody(r, &req); err != nil {
		rw.WriteError(err)
		return
	}

	// パスからインデックスとIDを設定
	req.Index = index
	req.ID = id

	// ドキュメントを更新
	result, err := h.documentUseCase.UpdateDocument(ctx, &req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// 成功レスポンスを返す
	rw.WriteDocument(result, "Document updated successfully")
}

// DeleteDocument はドキュメント削除リクエストを処理する
// DELETE /documents/{index}/{id}
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// ヘッダーを設定
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// パスパラメータを抽出
	index := h.getPathParam(r, "index")
	id := h.getPathParam(r, "id")

	if index == "" || id == "" {
		rw.WriteBadRequestError("Index and ID are required")
		return
	}

	// 削除リクエストを作成
	req := &dto.DeleteDocumentRequest{
		Index: index,
		ID:    id,
	}

	// ドキュメントを削除
	err := h.documentUseCase.DeleteDocument(ctx, req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// 成功レスポンスを返す
	rw.WriteNoContent()
}

// OptionsHandler はCORSプリフライトリクエストを処理する
func (h *DocumentHandler) OptionsHandler(w http.ResponseWriter, r *http.Request) {
	utils.SetCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}

// getPathParam はリクエストからパスパラメータを抽出する
func (h *DocumentHandler) getPathParam(r *http.Request, param string) string {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

	switch param {
	case "index":
		if len(pathParts) >= 2 {
			return pathParts[1]
		}
	case "id":
		if len(pathParts) >= 3 {
			return pathParts[2]
		}
	}

	return ""
}
