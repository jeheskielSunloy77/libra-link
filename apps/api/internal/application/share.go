package application

import (
	"context"
	"errors"
	"time"

	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
	"github.com/jeheskielSunloy77/libra-link/internal/app/sqlerr"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"gorm.io/gorm"
)

type ShareService interface {
	ResourceService[domain.Share, *applicationdto.StoreShareInput, *applicationdto.UpdateShareInput]
	Borrow(ctx context.Context, input *applicationdto.BorrowShareInput) (*domain.Borrow, error)
	ReturnBorrow(ctx context.Context, input *applicationdto.ReturnBorrowInput) (*domain.Borrow, error)
	UpsertReview(ctx context.Context, input *applicationdto.UpsertShareReviewInput) (*domain.ShareReview, error)
	CreateReport(ctx context.Context, input *applicationdto.CreateShareReportInput) (*domain.ShareReport, error)
}

type shareService struct {
	ResourceService[domain.Share, *applicationdto.StoreShareInput, *applicationdto.UpdateShareInput]
	shareRepo  port.ShareRepository
	borrowRepo port.BorrowRepository
	reviewRepo port.ShareReviewRepository
	reportRepo port.ShareReportRepository
}

func NewShareService(shareRepo port.ShareRepository, borrowRepo port.BorrowRepository, reviewRepo port.ShareReviewRepository, reportRepo port.ShareReportRepository) ShareService {
	return &shareService{
		ResourceService: NewResourceService[domain.Share, *applicationdto.StoreShareInput, *applicationdto.UpdateShareInput]("share", shareRepo),
		shareRepo:       shareRepo,
		borrowRepo:      borrowRepo,
		reviewRepo:      reviewRepo,
		reportRepo:      reportRepo,
	}
}

func (s *shareService) Borrow(ctx context.Context, input *applicationdto.BorrowShareInput) (*domain.Borrow, error) {
	if input == nil {
		return nil, errs.NewBadRequestError("borrow payload is required", true, nil, nil)
	}

	share, err := s.shareRepo.GetByID(ctx, input.ShareID, nil)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}

	if share.Status != domain.ShareStatusActive {
		return nil, errs.NewBadRequestError("share is not available for borrow", true, nil, nil)
	}

	if share.OwnerUserID == input.BorrowerUserID {
		return nil, errs.NewBadRequestError("share owner cannot borrow own share", true, nil, nil)
	}

	activeCount, err := s.borrowRepo.CountActiveByShare(ctx, share.ID)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	if activeCount >= int64(share.MaxConcurrentBorrows) {
		return nil, errs.NewBadRequestError("share reached maximum concurrent borrows", true, nil, nil)
	}

	if _, err := s.borrowRepo.GetActiveByShareAndBorrower(ctx, share.ID, input.BorrowerUserID); err == nil {
		return nil, errs.NewBadRequestError("user already has an active borrow for this share", true, nil, nil)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sqlerr.HandleError(err)
	}

	now := time.Now().UTC()
	borrow := &domain.Borrow{
		ShareID:             share.ID,
		BorrowerUserID:      input.BorrowerUserID,
		StartedAt:           now,
		DueAt:               now.Add(time.Duration(share.BorrowDurationHours) * time.Hour),
		Status:              domain.BorrowStatusActive,
		LegalAcknowledgedAt: now,
	}

	if err := s.borrowRepo.Store(ctx, borrow); err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return borrow, nil
}

func (s *shareService) ReturnBorrow(ctx context.Context, input *applicationdto.ReturnBorrowInput) (*domain.Borrow, error) {
	if input == nil {
		return nil, errs.NewBadRequestError("return payload is required", true, nil, nil)
	}

	borrow, err := s.borrowRepo.GetByID(ctx, input.BorrowID, nil)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}

	if borrow.Status != domain.BorrowStatusActive {
		return nil, errs.NewBadRequestError("borrow is not active", true, nil, nil)
	}

	share, err := s.shareRepo.GetByID(ctx, borrow.ShareID, nil)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}

	isBorrower := borrow.BorrowerUserID == input.BorrowerUserID
	isOwner := share.OwnerUserID == input.BorrowerUserID
	if !isBorrower && !isOwner {
		return nil, errs.NewForbiddenError("not allowed to return this borrow", true)
	}

	now := time.Now().UTC()
	if _, err := s.borrowRepo.Update(ctx, *borrow, map[string]any{
		"status":      domain.BorrowStatusReturned,
		"returned_at": now,
		"updated_at":  now,
	}); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	updated, err := s.borrowRepo.GetByID(ctx, borrow.ID, nil)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}
	return updated, nil
}

func (s *shareService) UpsertReview(ctx context.Context, input *applicationdto.UpsertShareReviewInput) (*domain.ShareReview, error) {
	if input == nil {
		return nil, errs.NewBadRequestError("review payload is required", true, nil, nil)
	}

	if _, err := s.shareRepo.GetByID(ctx, input.ShareID, nil); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	existing, err := s.reviewRepo.GetByShareAndUser(ctx, input.ShareID, input.UserID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sqlerr.HandleError(err)
		}

		review := &domain.ShareReview{
			ShareID:    input.ShareID,
			UserID:     input.UserID,
			Rating:     input.Rating,
			ReviewText: input.ReviewText,
		}
		if err := s.reviewRepo.Store(ctx, review); err != nil {
			return nil, sqlerr.HandleError(err)
		}
		return review, nil
	}

	now := time.Now().UTC()
	if _, err := s.reviewRepo.Update(ctx, *existing, map[string]any{
		"rating":      input.Rating,
		"review_text": input.ReviewText,
		"updated_at":  now,
		"deleted_at":  nil,
	}); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	updated, err := s.reviewRepo.GetByID(ctx, existing.ID, nil)
	if err != nil {
		return nil, sqlerr.HandleError(err)
	}

	return updated, nil
}

func (s *shareService) CreateReport(ctx context.Context, input *applicationdto.CreateShareReportInput) (*domain.ShareReport, error) {
	if input == nil {
		return nil, errs.NewBadRequestError("report payload is required", true, nil, nil)
	}

	if _, err := s.shareRepo.GetByID(ctx, input.ShareID, nil); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	report := &domain.ShareReport{
		ShareID:        input.ShareID,
		ReporterUserID: input.ReporterUserID,
		Reason:         input.Reason,
		Details:        input.Details,
		Status:         domain.ReportStatusOpen,
	}
	if err := s.reportRepo.Store(ctx, report); err != nil {
		return nil, sqlerr.HandleError(err)
	}

	return report, nil
}
