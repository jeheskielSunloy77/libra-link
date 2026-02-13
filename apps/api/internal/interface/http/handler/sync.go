package handler

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
	"github.com/jeheskielSunloy77/libra-link/internal/application"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	httpdto "github.com/jeheskielSunloy77/libra-link/internal/interface/http/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/interface/http/response"
	httputils "github.com/jeheskielSunloy77/libra-link/internal/interface/http/utils"
)

type SyncHandler struct {
	Handler
	service application.SyncService
}

func NewSyncHandler(h Handler, service application.SyncService) *SyncHandler {
	return &SyncHandler{Handler: h, service: service}
}

func (h *SyncHandler) StoreEvent() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.StoreSyncEventRequest) (*response.Response[domain.SyncEvent], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		input := req.ToUsecase()
		input.UserID = userID

		event, err := h.service.StoreEvent(c.UserContext(), input)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.SyncEvent]{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Sync event stored successfully!",
			Data:    event,
		}
		return &resp, nil
	}, http.StatusCreated, &httpdto.StoreSyncEventRequest{})
}

func (h *SyncHandler) ListEvents() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, _ *httpdto.Empty) (response.PaginatedResponse[domain.SyncEvent], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return response.PaginatedResponse[domain.SyncEvent]{}, err
		}

		limit := httputils.ParseQueryInt(c.Query("limit"), 200, 100)
		var since *time.Time
		if rawSince := c.Query("since"); rawSince != "" {
			parsed, parseErr := time.Parse(time.RFC3339, rawSince)
			if parseErr != nil {
				return response.PaginatedResponse[domain.SyncEvent]{}, errs.NewBadRequestError("invalid since value; expected RFC3339 datetime", true, nil, nil)
			}
			utc := parsed.UTC()
			since = &utc
		}

		events, err := h.service.ListEvents(c.UserContext(), userID, since, limit)
		if err != nil {
			return response.PaginatedResponse[domain.SyncEvent]{}, err
		}

		resp := response.NewPaginatedResponse("Successfully fetched sync events!", events, int64(len(events)), limit, 0)
		return resp, nil
	}, http.StatusOK, &httpdto.Empty{})
}
