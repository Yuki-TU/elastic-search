package handler

import (
	"net/http"
	"strings"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/application/usecase"
	"github.com/Yuki-TU/elastic-search/api/pkg/utils"
)

// DocumentHandler handles document-related HTTP requests
type DocumentHandler struct {
	documentUseCase *usecase.DocumentUseCase
}

// NewDocumentHandler creates a new DocumentHandler
func NewDocumentHandler(documentUseCase *usecase.DocumentUseCase) *DocumentHandler {
	return &DocumentHandler{
		documentUseCase: documentUseCase,
	}
}

// CreateDocument handles document creation requests
// POST /documents
func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// Set headers
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// Parse request body
	var req dto.CreateDocumentRequest
	if err := utils.ParseRequestBody(r, &req); err != nil {
		rw.WriteError(err)
		return
	}

	// Create document
	result, err := h.documentUseCase.CreateDocument(ctx, &req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// Return success response
	rw.WriteCreated(result, "Document created successfully")
}

// GetDocument handles document retrieval requests
// GET /documents/{index}/{id}
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// Set headers
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// Extract path parameters
	index := h.getPathParam(r, "index")
	id := h.getPathParam(r, "id")

	if index == "" || id == "" {
		rw.WriteBadRequestError("Index and ID are required")
		return
	}

	// Get document
	result, err := h.documentUseCase.GetDocument(ctx, index, id)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// Return success response
	rw.WriteDocument(result, "Document retrieved successfully")
}

// UpdateDocument handles document update/create requests
// PUT /documents/{index}/{id}
func (h *DocumentHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// Set headers
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// Extract path parameters
	index := h.getPathParam(r, "index")
	id := h.getPathParam(r, "id")

	if index == "" || id == "" {
		rw.WriteBadRequestError("Index and ID are required")
		return
	}

	// Parse request body
	var req dto.UpdateDocumentRequest
	if err := utils.ParseRequestBody(r, &req); err != nil {
		rw.WriteError(err)
		return
	}

	// Set index and ID from path
	req.Index = index
	req.ID = id

	// Update document
	result, err := h.documentUseCase.UpdateDocument(ctx, &req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// Return success response
	rw.WriteDocument(result, "Document updated successfully")
}

// DeleteDocument handles document deletion requests
// DELETE /documents/{index}/{id}
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rw := utils.NewResponseWriter(w)

	// Set headers
	utils.SetCORSHeaders(w)
	utils.SetSecurityHeaders(w)

	// Extract path parameters
	index := h.getPathParam(r, "index")
	id := h.getPathParam(r, "id")

	if index == "" || id == "" {
		rw.WriteBadRequestError("Index and ID are required")
		return
	}

	// Create delete request
	req := &dto.DeleteDocumentRequest{
		Index: index,
		ID:    id,
	}

	// Delete document
	err := h.documentUseCase.DeleteDocument(ctx, req)
	if err != nil {
		rw.WriteError(err)
		return
	}

	// Return success response
	rw.WriteNoContent()
}

// OptionsHandler handles CORS preflight requests
func (h *DocumentHandler) OptionsHandler(w http.ResponseWriter, r *http.Request) {
	utils.SetCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}

// getPathParam extracts a path parameter from the request
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
