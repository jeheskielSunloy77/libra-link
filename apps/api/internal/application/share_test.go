package application

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type testBorrowRepo struct {
	*repository.MockResourceRepository[domain.Borrow]
}

func (r *testBorrowRepo) CountActiveByShare(ctx context.Context, shareID uuid.UUID) (int64, error) {
	items, _, err := r.GetMany(ctx, repository.GetManyOptions{})
	if err != nil {
		return 0, err
	}

	var count int64
	for i := range items {
		if items[i].ShareID == shareID && items[i].Status == domain.BorrowStatusActive {
			count++
		}
	}
	return count, nil
}

func (r *testBorrowRepo) GetActiveByShareAndBorrower(ctx context.Context, shareID uuid.UUID, borrowerID uuid.UUID) (*domain.Borrow, error) {
	items, _, err := r.GetMany(ctx, repository.GetManyOptions{})
	if err != nil {
		return nil, err
	}

	for i := range items {
		if items[i].ShareID == shareID && items[i].BorrowerUserID == borrowerID && items[i].Status == domain.BorrowStatusActive {
			borrow := items[i]
			return &borrow, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

type testShareReviewRepo struct {
	*repository.MockResourceRepository[domain.ShareReview]
}

func (r *testShareReviewRepo) GetByShareAndUser(ctx context.Context, shareID uuid.UUID, userID uuid.UUID) (*domain.ShareReview, error) {
	items, _, err := r.GetMany(ctx, repository.GetManyOptions{})
	if err != nil {
		return nil, err
	}

	for i := range items {
		if items[i].ShareID == shareID && items[i].UserID == userID {
			review := items[i]
			return &review, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func newShareServiceForTest() ShareService {
	shareRepo := repository.NewMockResourceRepository[domain.Share](false)
	borrowRepo := &testBorrowRepo{MockResourceRepository: repository.NewMockResourceRepository[domain.Borrow](false)}
	reviewRepo := &testShareReviewRepo{MockResourceRepository: repository.NewMockResourceRepository[domain.ShareReview](false)}
	reportRepo := repository.NewMockResourceRepository[domain.ShareReport](false)

	return NewShareService(shareRepo, borrowRepo, reviewRepo, reportRepo)
}

func TestShareServiceBorrow_RejectsOwnerBorrowingOwnShare(t *testing.T) {
	ctx := context.Background()
	service := newShareServiceForTest()

	shareOwnerID := uuid.New()
	share, err := service.Store(ctx, &applicationdto.StoreShareInput{
		EbookID:              uuid.New(),
		OwnerUserID:          shareOwnerID,
		BorrowDurationHours:  24,
		MaxConcurrentBorrows: 1,
	})
	require.NoError(t, err)

	_, err = service.Borrow(ctx, &applicationdto.BorrowShareInput{
		ShareID:        share.ID,
		BorrowerUserID: shareOwnerID,
	})
	require.Error(t, err)

	var httpErr *errs.ErrorResponse
	require.ErrorAs(t, err, &httpErr)
	require.Equal(t, http.StatusBadRequest, httpErr.Status)
}

func TestShareServiceUpsertReview_UpdatesExistingReview(t *testing.T) {
	ctx := context.Background()
	service := newShareServiceForTest()

	ownerID := uuid.New()
	share, err := service.Store(ctx, &applicationdto.StoreShareInput{
		EbookID:              uuid.New(),
		OwnerUserID:          ownerID,
		BorrowDurationHours:  24,
		MaxConcurrentBorrows: 1,
	})
	require.NoError(t, err)

	userID := uuid.New()
	first, err := service.UpsertReview(ctx, &applicationdto.UpsertShareReviewInput{
		ShareID: share.ID,
		UserID:  userID,
		Rating:  3,
	})
	require.NoError(t, err)

	text := "Updated review"
	second, err := service.UpsertReview(ctx, &applicationdto.UpsertShareReviewInput{
		ShareID:    share.ID,
		UserID:     userID,
		Rating:     5,
		ReviewText: &text,
	})
	require.NoError(t, err)
	require.Equal(t, first.ID, second.ID)
	require.Equal(t, int16(5), second.Rating)
	require.NotNil(t, second.ReviewText)
	require.Equal(t, text, *second.ReviewText)
}

func TestShareServiceBorrow_RejectsWhenShareAtCapacity(t *testing.T) {
	ctx := context.Background()
	service := newShareServiceForTest()

	share, err := service.Store(ctx, &applicationdto.StoreShareInput{
		EbookID:              uuid.New(),
		OwnerUserID:          uuid.New(),
		BorrowDurationHours:  24,
		MaxConcurrentBorrows: 1,
	})
	require.NoError(t, err)

	_, err = service.Borrow(ctx, &applicationdto.BorrowShareInput{
		ShareID:        share.ID,
		BorrowerUserID: uuid.New(),
	})
	require.NoError(t, err)

	_, err = service.Borrow(ctx, &applicationdto.BorrowShareInput{
		ShareID:        share.ID,
		BorrowerUserID: uuid.New(),
	})
	require.Error(t, err)

	var httpErr *errs.ErrorResponse
	require.True(t, errors.As(err, &httpErr))
	require.Equal(t, http.StatusBadRequest, httpErr.Status)
}
