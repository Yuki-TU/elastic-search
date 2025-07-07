package usecase

import (
	"context"

	"github.com/Yuki-TU/elastic-search/api/internal/application/dto"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/entity"
	"github.com/Yuki-TU/elastic-search/api/internal/domain/service"
	"github.com/Yuki-TU/elastic-search/api/pkg/errors"
)

// DocumentUseCase はドキュメント関連のビジネスロジックを処理する
type DocumentUseCase struct {
	documentService service.DocumentHandler
}

// NewDocumentUseCase は新しい DocumentUseCase を作成する
func NewDocumentUseCase(documentService service.DocumentHandler) *DocumentUseCase {
	return &DocumentUseCase{
		documentService: documentService,
	}
}

// CreateDocument は新しいドキュメントを作成する
func (uc *DocumentUseCase) CreateDocument(ctx context.Context, req *dto.CreateDocumentRequest) (*dto.DocumentDTO, error) {
	// リクエストを検証
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// ドメインサービスを通じてドキュメントを作成
	doc, err := uc.documentService.CreateDocument(ctx, req.Index, req.Source)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	return uc.entityToDTO(doc), nil
}

// CreateDocumentWithID は指定されたIDでドキュメントを作成する
func (uc *DocumentUseCase) CreateDocumentWithID(ctx context.Context, req *dto.CreateDocumentRequest) (*dto.DocumentDTO, error) {
	// リクエストを検証
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if req.ID == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "CreateDocumentWithIDではドキュメントIDを空にできません")
	}

	// ドメインサービスを通じてIDありでドキュメントを作成
	doc, err := uc.documentService.CreateDocumentWithID(ctx, req.Index, req.ID, req.Source)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	return uc.entityToDTO(doc), nil
}

// GetDocument はインデックスとIDでドキュメントを取得する
func (uc *DocumentUseCase) GetDocument(ctx context.Context, index, id string) (*dto.DocumentDTO, error) {
	// 入力を検証
	if index == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "インデックスは空にできません")
	}
	if id == "" {
		return nil, errors.NewAppError(errors.ErrCodeValidationFailed, "ドキュメントIDは空にできません")
	}

	// ドメインサービスを通じてドキュメントを取得
	doc, err := uc.documentService.GetDocument(ctx, index, id)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	return uc.entityToDTO(doc), nil
}

// UpdateDocument は既存のドキュメントを更新する
func (uc *DocumentUseCase) UpdateDocument(ctx context.Context, req *dto.UpdateDocumentRequest) (*dto.DocumentDTO, error) {
	// リクエストを検証
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// ドメインサービスを通じてドキュメントを更新
	doc, err := uc.documentService.UpdateDocument(ctx, req.Index, req.ID, req.Source)
	if err != nil {
		return nil, err
	}

	// DTOに変換
	return uc.entityToDTO(doc), nil
}

// DeleteDocument はドキュメントを削除する
func (uc *DocumentUseCase) DeleteDocument(ctx context.Context, req *dto.DeleteDocumentRequest) error {
	// リクエストを検証
	if req.Index == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "インデックスは空にできません")
	}
	if req.ID == "" {
		return errors.NewAppError(errors.ErrCodeValidationFailed, "ドキュメントIDは空にできません")
	}

	// ドメインサービスを通じてドキュメントを削除
	return uc.documentService.DeleteDocument(ctx, req.Index, req.ID)
}

// entityToDTO はエンティティをDTOに変換するヘルパーメソッド
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
