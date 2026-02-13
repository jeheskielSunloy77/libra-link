package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jeheskielSunloy77/libra-link/internal/application"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	httpdto "github.com/jeheskielSunloy77/libra-link/internal/interface/http/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/interface/http/response"
	httputils "github.com/jeheskielSunloy77/libra-link/internal/interface/http/utils"
)

type ShareHandler struct {
	*ResourceHandler[domain.Share, *applicationdto.StoreShareInput, *applicationdto.UpdateShareInput, *httpdto.StoreShareRequest, *httpdto.UpdateShareRequest]
	service application.ShareService
}

func NewShareHandler(h Handler, service application.ShareService) *ShareHandler {
	return &ShareHandler{
		ResourceHandler: NewResourceHandler[domain.Share, *applicationdto.StoreShareInput, *applicationdto.UpdateShareInput, *httpdto.StoreShareRequest, *httpdto.UpdateShareRequest]("share", h, service),
		service:         service,
	}
}

func (h *ShareHandler) Store() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.StoreShareRequest) (*response.Response[domain.Share], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		input := req.ToUsecase()
		input.OwnerUserID = userID

		entity, err := h.service.Store(c.UserContext(), input)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.Share]{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Successfully created share!",
			Data:    entity,
		}
		return &resp, nil
	}, http.StatusCreated, &httpdto.StoreShareRequest{})
}

func (h *ShareHandler) Borrow() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.BorrowShareRequest) (*response.Response[domain.Borrow], error) {
		shareID, err := httputils.ParseUUIDParam(c.Params("id"))
		if err != nil {
			return nil, err
		}
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		borrow, err := h.service.Borrow(c.UserContext(), &applicationdto.BorrowShareInput{
			ShareID:             shareID,
			BorrowerUserID:      userID,
			LegalAcknowledgedAt: &req.LegalAcknowledged,
		})
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.Borrow]{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Borrow created successfully!",
			Data:    borrow,
		}
		return &resp, nil
	}, http.StatusCreated, &httpdto.BorrowShareRequest{})
}

func (h *ShareHandler) ReturnBorrow() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, _ *httpdto.Empty) (*response.Response[domain.Borrow], error) {
		borrowID, err := httputils.ParseUUIDParam(c.Params("id"))
		if err != nil {
			return nil, err
		}
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		borrow, err := h.service.ReturnBorrow(c.UserContext(), &applicationdto.ReturnBorrowInput{
			BorrowID:       borrowID,
			BorrowerUserID: userID,
		})
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.Borrow]{
			Status:  http.StatusOK,
			Success: true,
			Message: "Borrow returned successfully!",
			Data:    borrow,
		}
		return &resp, nil
	}, http.StatusOK, &httpdto.Empty{})
}

func (h *ShareHandler) UpsertReview() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.UpsertShareReviewRequest) (*response.Response[domain.ShareReview], error) {
		shareID, err := httputils.ParseUUIDParam(c.Params("id"))
		if err != nil {
			return nil, err
		}
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		review, err := h.service.UpsertReview(c.UserContext(), &applicationdto.UpsertShareReviewInput{
			ShareID:    shareID,
			UserID:     userID,
			Rating:     req.Rating,
			ReviewText: req.ReviewText,
		})
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.ShareReview]{
			Status:  http.StatusOK,
			Success: true,
			Message: "Review saved successfully!",
			Data:    review,
		}
		return &resp, nil
	}, http.StatusOK, &httpdto.UpsertShareReviewRequest{})
}

func (h *ShareHandler) CreateReport() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.CreateShareReportRequest) (*response.Response[domain.ShareReport], error) {
		shareID, err := httputils.ParseUUIDParam(c.Params("id"))
		if err != nil {
			return nil, err
		}
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		report, err := h.service.CreateReport(c.UserContext(), &applicationdto.CreateShareReportInput{
			ShareID:        shareID,
			ReporterUserID: userID,
			Reason:         req.Reason,
			Details:        req.Details,
		})
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.ShareReport]{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Report created successfully!",
			Data:    report,
		}
		return &resp, nil
	}, http.StatusCreated, &httpdto.CreateShareReportRequest{})
}
