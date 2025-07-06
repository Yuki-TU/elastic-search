package usecase

import (
	"context"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/entity"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/service"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
)

// DocumentUseCase handles document-related business logic
type DocumentUseCase struct {
	documentService *service.DocumentService
}

// NewDocumentUseCase creates a new DocumentUseCase
func NewDocumentUseCase(documentService *service.DocumentService) *DocumentUseCase {
	return &DocumentUseCase{
		documentService: documentService,
	}
}

// CreateDocument creates a new document
func (uc *DocumentUseCase) CreateDocument(ctx context.Context, req *dto.CreateDocumentRequest) (*dto.DocumentDTO, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Create document through domain service
	doc, err := uc.documentService.CreateDocument(ctx, req.Index, req.Source)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	return uc.entityToDTO(doc), nil
}

// CreateDocumentWithID creates a document with a specific ID
func (uc *DocumentUseCase) CreateDocumentWithID(ctx context.Context, req *dto.CreateDocumentRequest) (*dto.DocumentDTO, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if req.ID == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty for CreateDocumentWithID")
	}

	// Create document with ID through domain service
	doc, err := uc.documentService.CreateDocumentWithID(ctx, req.Index, req.ID, req.Source)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	return uc.entityToDTO(doc), nil
}

// GetDocument retrieves a document by index and ID
func (uc *DocumentUseCase) GetDocument(ctx context.Context, index, id string) (*dto.DocumentDTO, error) {
	// Validate input
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}
	if id == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty")
	}

	// Get document through domain service
	doc, err := uc.documentService.GetDocument(ctx, index, id)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	return uc.entityToDTO(doc), nil
}

// UpdateDocument updates an existing document
func (uc *DocumentUseCase) UpdateDocument(ctx context.Context, req *dto.UpdateDocumentRequest) (*dto.DocumentDTO, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Update document through domain service
	doc, err := uc.documentService.UpdateDocument(ctx, req.Index, req.ID, req.Source)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	return uc.entityToDTO(doc), nil
}

// DeleteDocument deletes a document
func (uc *DocumentUseCase) DeleteDocument(ctx context.Context, req *dto.DeleteDocumentRequest) error {
	// Validate request
	if req.Index == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}
	if req.ID == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty")
	}

	// Delete document through domain service
	return uc.documentService.DeleteDocument(ctx, req.Index, req.ID)
}

// Helper method to convert entity to DTO
func (uc *DocumentUseCase) entityToDTO(doc *entity.Document) *dto.DocumentDTO {
	return &dto.DocumentDTO{
		ID:       doc.ID,
		Index:    doc.Index,
		Source:   doc.Source,
		Version:  doc.Version,
		Created:  doc.Created,
		Modified: doc.Modified,
	}
}
