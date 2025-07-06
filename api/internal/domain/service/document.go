package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Yuki-TU/elastic-search/api/internal/domain/entity"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/repository"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
)

// DocumentService provides business logic for document operations
type DocumentService struct {
	repo repository.ElasticsearchRepository
}

// NewDocumentService creates a new DocumentService
func NewDocumentService(repo repository.ElasticsearchRepository) *DocumentService {
	return &DocumentService{
		repo: repo,
	}
}

// CreateDocument creates a new document
func (s *DocumentService) CreateDocument(ctx context.Context, index string, source map[string]any) (*entity.Document, error) {
	// Validate input
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}

	if len(source) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document source cannot be empty")
	}

	// Create document entity
	doc := entity.NewDocument(index, source)

	// Apply business rules
	if err := s.applyBusinessRules(doc); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to create document")
	}

	return doc, nil
}

// GetDocument retrieves a document by ID
func (s *DocumentService) GetDocument(ctx context.Context, index, id string) (*entity.Document, error) {
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}

	if id == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty")
	}

	doc, err := s.repo.GetDocument(ctx, index, id)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentNotFound, "Document not found")
	}

	return doc, nil
}

// UpdateDocument updates an existing document
func (s *DocumentService) UpdateDocument(ctx context.Context, index, id string, source map[string]any) (*entity.Document, error) {
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}

	if id == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty")
	}

	if len(source) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document source cannot be empty")
	}

	// Get existing document
	doc, err := s.repo.GetDocument(ctx, index, id)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentNotFound, "Document not found")
	}

	// Update document
	doc.UpdateSource(source)

	// Apply business rules
	if err := s.applyBusinessRules(doc); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentUpdateFailed, "Failed to update document")
	}

	return doc, nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(ctx context.Context, index, id string) error {
	if index == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}

	if id == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty")
	}

	// Check if document exists
	_, err := s.repo.GetDocument(ctx, index, id)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentNotFound, "Document not found")
	}

	// Delete document
	if err := s.repo.DeleteDocument(ctx, index, id); err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentDeleteFailed, "Failed to delete document")
	}

	return nil
}

// BulkIndexDocuments creates multiple documents in a single operation
func (s *DocumentService) BulkIndexDocuments(ctx context.Context, docs []*entity.Document) error {
	if len(docs) == 0 {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "No documents provided for bulk indexing")
	}

	// Validate all documents
	for i, doc := range docs {
		if err := s.validateDocument(doc); err != nil {
			return errors.NewAppError(errors.ErrCodeValidationFailed, fmt.Sprintf("Document %d validation failed: %v", i, err))
		}

		// Apply business rules
		if err := s.applyBusinessRules(doc); err != nil {
			return errors.NewAppError(errors.ErrCodeValidationFailed, fmt.Sprintf("Document %d business rule validation failed: %v", i, err))
		}
	}

	// Perform bulk indexing
	if err := s.repo.BulkIndex(ctx, docs); err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to bulk index documents")
	}

	return nil
}

// CreateDocumentWithID creates a document with a specific ID
func (s *DocumentService) CreateDocumentWithID(ctx context.Context, index, id string, source map[string]any) (*entity.Document, error) {
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}

	if id == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty")
	}

	if len(source) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document source cannot be empty")
	}

	// Check if document already exists
	_, err := s.repo.GetDocument(ctx, index, id)
	if err == nil {
		return nil, errors.NewDocumentExistsError(index, id)
	}

	// Create document entity
	doc := entity.NewDocument(index, source)
	doc.SetID(id)

	// Apply business rules
	if err := s.applyBusinessRules(doc); err != nil {
		return nil, err
	}

	// Save to repository
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to create document")
	}

	return doc, nil
}

// applyBusinessRules applies business rules to a document
func (s *DocumentService) applyBusinessRules(doc *entity.Document) error {
	// Add timestamp fields if not present
	if _, exists := doc.GetField("created_at"); !exists {
		doc.SetField("created_at", time.Now().Format(time.RFC3339))
	}

	if _, exists := doc.GetField("updated_at"); !exists {
		doc.SetField("updated_at", time.Now().Format(time.RFC3339))
	} else {
		// Update the timestamp
		doc.SetField("updated_at", time.Now().Format(time.RFC3339))
	}

	// Validate required fields (example business rule)
	if err := s.validateRequiredFields(doc); err != nil {
		return err
	}

	// Apply data transformations
	if err := s.applyDataTransformations(doc); err != nil {
		return err
	}

	return nil
}

// validateDocument validates a document
func (s *DocumentService) validateDocument(doc *entity.Document) error {
	if doc == nil {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Document cannot be nil")
	}

	if doc.Index == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Document index cannot be empty")
	}

	if len(doc.Source) == 0 {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Document source cannot be empty")
	}

	return nil
}

// validateRequiredFields validates required fields in a document
func (s *DocumentService) validateRequiredFields(doc *entity.Document) error {
	// Example: Check if certain fields are required based on the index
	switch doc.Index {
	case "users":
		if _, exists := doc.GetField("email"); !exists {
			return errors.NewAppError(errors.ErrCodeValidationFailed, "Email field is required for users index")
		}
		if _, exists := doc.GetField("name"); !exists {
			return errors.NewAppError(errors.ErrCodeValidationFailed, "Name field is required for users index")
		}
	case "products":
		if _, exists := doc.GetField("name"); !exists {
			return errors.NewAppError(errors.ErrCodeValidationFailed, "Name field is required for products index")
		}
		if _, exists := doc.GetField("price"); !exists {
			return errors.NewAppError(errors.ErrCodeValidationFailed, "Price field is required for products index")
		}
	}

	return nil
}

// applyDataTransformations applies data transformations to a document
func (s *DocumentService) applyDataTransformations(doc *entity.Document) error {
	// Example: Normalize email addresses
	if email, exists := doc.GetField("email"); exists {
		if emailStr, ok := email.(string); ok {
			doc.SetField("email", normalizeEmail(emailStr))
		}
	}

	// Example: Add computed fields
	if firstName, exists := doc.GetField("first_name"); exists {
		if lastName, exists := doc.GetField("last_name"); exists {
			if firstNameStr, ok := firstName.(string); ok {
				if lastNameStr, ok := lastName.(string); ok {
					doc.SetField("full_name", firstNameStr+" "+lastNameStr)
				}
			}
		}
	}

	return nil
}

// normalizeEmail normalizes an email address
func normalizeEmail(email string) string {
	// Simple email normalization - convert to lowercase
	// In a real application, you might want more sophisticated normalization
	return strings.ToLower(email)
}
