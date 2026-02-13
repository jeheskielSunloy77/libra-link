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

type EbookHandler struct {
	*ResourceHandler[domain.Ebook, *applicationdto.StoreEbookInput, *applicationdto.UpdateEbookInput, *httpdto.StoreEbookRequest, *httpdto.UpdateEbookRequest]
	service application.EbookService
}

func NewEbookHandler(h Handler, service application.EbookService) *EbookHandler {
	return &EbookHandler{
		ResourceHandler: NewResourceHandler[domain.Ebook, *applicationdto.StoreEbookInput, *applicationdto.UpdateEbookInput, *httpdto.StoreEbookRequest, *httpdto.UpdateEbookRequest]("ebook", h, service),
		service:         service,
	}
}

func (h *EbookHandler) Store() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.StoreEbookRequest) (*response.Response[domain.Ebook], error) {
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

		resp := response.Response[domain.Ebook]{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Successfully created ebook!",
			Data:    entity,
		}
		return &resp, nil
	}, http.StatusCreated, &httpdto.StoreEbookRequest{})
}

func (h *EbookHandler) AttachMetadata() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.AttachGoogleMetadataRequest) (*response.Response[domain.EbookGoogleMetadata], error) {
		ebookID, err := httputils.ParseUUIDParam(c.Params("id"))
		if err != nil {
			return nil, err
		}

		input := req.ToUsecase()
		input.EbookID = ebookID

		metadata, err := h.service.AttachMetadata(c.UserContext(), input)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.EbookGoogleMetadata]{
			Status:  http.StatusOK,
			Success: true,
			Message: "Successfully attached metadata!",
			Data:    metadata,
		}
		return &resp, nil
	}, http.StatusOK, &httpdto.AttachGoogleMetadataRequest{})
}

func (h *EbookHandler) DetachMetadata() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, _ *httpdto.Empty) (*response.Response[any], error) {
		ebookID, err := httputils.ParseUUIDParam(c.Params("id"))
		if err != nil {
			return nil, err
		}

		if err := h.service.DetachMetadata(c.UserContext(), ebookID); err != nil {
			return nil, err
		}

		resp := response.Response[any]{
			Status:  http.StatusOK,
			Success: true,
			Message: "Successfully detached metadata!",
		}
		return &resp, nil
	}, http.StatusOK, &httpdto.Empty{})
}
