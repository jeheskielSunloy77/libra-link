package application

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
	"github.com/jeheskielSunloy77/libra-link/internal/app/sqlerr"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"gorm.io/gorm"
)

type EbookService interface {
	ResourceService[domain.Ebook, *applicationdto.StoreEbookInput, *applicationdto.UpdateEbookInput]
	AttachMetadata(ctx context.Context, input *applicationdto.AttachGoogleMetadataInput) (*domain.EbookGoogleMetadata, error)
	DetachMetadata(ctx context.Context, ebookID uuid.UUID) error
}

type ebookService struct {
	ResourceService[domain.Ebook, *applicationdto.StoreEbookInput, *applicationdto.UpdateEbookInput]
	repo         port.EbookRepository
	metadataRepo port.EbookGoogleMetadataRepository
}

func NewEbookService(repo port.EbookRepository, metadataRepo port.EbookGoogleMetadataRepository) EbookService {
	return &ebookService{
		ResourceService: NewResourceService[domain.Ebook, *applicationdto.StoreEbookInput, *applicationdto.UpdateEbookInput]("ebook", repo),
		repo:            repo,
		metadataRepo:    metadataRepo,
	}
}

func (s *ebookService) AttachMetadata(ctx context.Context, input *applicationdto.AttachGoogleMetadataInput) (*domain.EbookGoogleMetadata, error) {
	if input == nil {
		return nil, errs.NewBadRequestError("metadata payload is required", true, nil, nil)
	}

	if _, err := s.repo.GetByID(ctx, input.EbookID, nil); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	metadata := &domain.EbookGoogleMetadata{
		EbookID:       input.EbookID,
		GoogleBooksID: input.GoogleBooksID,
		ISBN10:        input.ISBN10,
		ISBN13:        input.ISBN13,
		Publisher:     input.Publisher,
		PublishedDate: input.PublishedDate,
		PageCount:     input.PageCount,
		Categories:    input.Categories,
		ThumbnailURL:  input.ThumbnailURL,
		InfoLink:      input.InfoLink,
		RawPayload:    input.RawPayload,
	}

	if err := s.metadataRepo.Upsert(ctx, metadata); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	stored, err := s.metadataRepo.GetByEbookID(ctx, input.EbookID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return stored, nil
}

func (s *ebookService) DetachMetadata(ctx context.Context, ebookID uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, ebookID, nil); err != nil {
		return sqlerr.HandleError(err)
	}

	if err := s.metadataRepo.SoftDeleteByEbookID(ctx, ebookID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errs.NewNotFoundError("metadata not found", true)
		}
		return sqlerr.HandleError(err)
	}

	return nil
}
