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

// DocumentHandler はドキュメントサービスのインターフェース
type DocumentHandler interface {
	CreateDocument(ctx context.Context, index string, source map[string]any) (*entity.Document, error)
	GetDocument(ctx context.Context, index, id string) (*entity.Document, error)
	UpdateDocument(ctx context.Context, index, id string, source map[string]any) (*entity.Document, error)
	DeleteDocument(ctx context.Context, index, id string) error
	BulkIndexDocuments(ctx context.Context, docs []*entity.Document) error
	CreateDocumentWithID(ctx context.Context, index, id string, source map[string]any) (*entity.Document, error)
}

// DocumentService はドキュメント操作のビジネスロジックを提供する
type DocumentService struct {
	repo repository.ElasticsearchRepository
}

// NewDocumentService は新しいDocumentServiceを作成する
func NewDocumentService(repo repository.ElasticsearchRepository) *DocumentService {
	return &DocumentService{
		repo: repo,
	}
}

// CreateDocument は新しいドキュメントを作成する
func (s *DocumentService) CreateDocument(ctx context.Context, index string, source map[string]any) (*entity.Document, error) {
	// 入力を検証
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}

	if len(source) == 0 {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "Document source cannot be empty")
	}

	// ドキュメントエンティティを作成
	doc := entity.NewDocument(index, source)

	// ビジネスルールを適用
	if err := s.applyBusinessRules(doc); err != nil {
		return nil, err
	}

	// リポジトリに保存
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to create document")
	}

	return doc, nil
}

// GetDocument はIDでドキュメントを取得する
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

// UpdateDocument は既存のドキュメントを更新する
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

	// 既存のドキュメントを取得
	doc, err := s.repo.GetDocument(ctx, index, id)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentNotFound, "Document not found")
	}

	// ドキュメントを更新
	doc.UpdateSource(source)

	// ビジネスルールを適用
	if err := s.applyBusinessRules(doc); err != nil {
		return nil, err
	}

	// リポジトリに保存
	if err := s.repo.UpdateDocument(ctx, doc); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentUpdateFailed, "Failed to update document")
	}

	return doc, nil
}

// DeleteDocument はドキュメントを削除する
func (s *DocumentService) DeleteDocument(ctx context.Context, index, id string) error {
	if index == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Index cannot be empty")
	}

	if id == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "Document ID cannot be empty")
	}

	// ドキュメントの存在確認
	_, err := s.repo.GetDocument(ctx, index, id)
	if err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentNotFound, "Document not found")
	}

	// ドキュメントを削除
	if err := s.repo.DeleteDocument(ctx, index, id); err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentDeleteFailed, "Failed to delete document")
	}

	return nil
}

// BulkIndexDocuments は複数のドキュメントを一度に作成する
func (s *DocumentService) BulkIndexDocuments(ctx context.Context, docs []*entity.Document) error {
	if len(docs) == 0 {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "No documents provided for bulk indexing")
	}

	// 全てのドキュメントを検証
	for i, doc := range docs {
		if err := s.validateDocument(doc); err != nil {
			return errors.NewAppError(errors.ErrCodeValidationFailed, fmt.Sprintf("Document %d validation failed: %v", i, err))
		}

		// ビジネスルールを適用
		if err := s.applyBusinessRules(doc); err != nil {
			return errors.NewAppError(errors.ErrCodeValidationFailed, fmt.Sprintf("Document %d business rule validation failed: %v", i, err))
		}
	}

	// バルクインデックスを実行
	if err := s.repo.BulkIndex(ctx, docs); err != nil {
		return errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to bulk index documents")
	}

	return nil
}

// CreateDocumentWithID は指定されたIDでドキュメントを作成する
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

	// ドキュメントが既に存在するかを確認
	_, err := s.repo.GetDocument(ctx, index, id)
	if err == nil {
		return nil, errors.NewDocumentExistsError(index, id)
	}

	// ドキュメントエンティティを作成
	doc := entity.NewDocument(index, source)
	doc.SetID(id)

	// ビジネスルールを適用
	if err := s.applyBusinessRules(doc); err != nil {
		return nil, err
	}

	// リポジトリに保存
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, errors.WrapError(err, errors.ErrCodeDocumentCreateFailed, "Failed to create document")
	}

	return doc, nil
}

// applyBusinessRules はドキュメントにビジネスルールを適用する
func (s *DocumentService) applyBusinessRules(doc *entity.Document) error {
	// タイムスタンプフィールドが存在しない場合は追加
	if _, exists := doc.GetField("created_at"); !exists {
		doc.SetField("created_at", time.Now().Format(time.RFC3339))
	}

	if _, exists := doc.GetField("updated_at"); !exists {
		doc.SetField("updated_at", time.Now().Format(time.RFC3339))
	} else {
		// タイムスタンプを更新
		doc.SetField("updated_at", time.Now().Format(time.RFC3339))
	}

	// 必須フィールドを検証（ビジネスルールの例）
	if err := s.validateRequiredFields(doc); err != nil {
		return err
	}

	// データ変換を適用
	if err := s.applyDataTransformations(doc); err != nil {
		return err
	}

	return nil
}

// validateDocument はドキュメントを検証する
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

// validateRequiredFields はドキュメントの必須フィールドを検証する
func (s *DocumentService) validateRequiredFields(doc *entity.Document) error {
	// 例: インデックスに基づいて特定のフィールドが必須かを確認
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

// applyDataTransformations はドキュメントにデータ変換を適用する
func (s *DocumentService) applyDataTransformations(doc *entity.Document) error {
	// 例: メールアドレスの正規化
	if email, exists := doc.GetField("email"); exists {
		if emailStr, ok := email.(string); ok {
			doc.SetField("email", normalizeEmail(emailStr))
		}
	}

	// 例: 算出フィールドの追加
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

// normalizeEmail はメールアドレスを正規化する
func normalizeEmail(email string) string {
	// シンプルなメール正規化 - 小文字に変換
	// 実際のアプリケーションでは、より高度な正規化が必要かもしれません
	return strings.ToLower(email)
}

// インターフェースの実装確認
var _ DocumentHandler = (*DocumentService)(nil)
