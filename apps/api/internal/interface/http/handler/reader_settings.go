package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jeheskielSunloy77/libra-link/internal/application"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	httpdto "github.com/jeheskielSunloy77/libra-link/internal/interface/http/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/interface/http/response"
)

type ReaderSettingsHandler struct {
	Handler
	preferencesService application.UserPreferencesService
	readerStateService application.UserReaderStateService
}

func NewReaderSettingsHandler(h Handler, preferencesService application.UserPreferencesService, readerStateService application.UserReaderStateService) *ReaderSettingsHandler {
	return &ReaderSettingsHandler{
		Handler:            h,
		preferencesService: preferencesService,
		readerStateService: readerStateService,
	}
}

func (h *ReaderSettingsHandler) GetPreferences() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, _ *httpdto.Empty) (*response.Response[domain.UserPreferences], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		prefs, err := h.preferencesService.GetByUserID(c.UserContext(), userID)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.UserPreferences]{
			Status:  http.StatusOK,
			Success: true,
			Message: "Successfully fetched preferences!",
			Data:    prefs,
		}
		return &resp, nil
	}, http.StatusOK, &httpdto.Empty{})
}

func (h *ReaderSettingsHandler) PatchPreferences() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.UpdateUserPreferencesRequest) (*response.Response[domain.UserPreferences], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		updated, err := h.preferencesService.Patch(c.UserContext(), userID, req.ToUsecase())
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.UserPreferences]{
			Status:  http.StatusOK,
			Success: true,
			Message: "Successfully updated preferences!",
			Data:    updated,
		}
		return &resp, nil
	}, http.StatusOK, &httpdto.UpdateUserPreferencesRequest{})
}

func (h *ReaderSettingsHandler) GetReaderState() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, _ *httpdto.Empty) (*response.Response[domain.UserReaderState], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		state, err := h.readerStateService.GetByUserID(c.UserContext(), userID)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.UserReaderState]{
			Status:  http.StatusOK,
			Success: true,
			Message: "Successfully fetched reader state!",
			Data:    state,
		}
		return &resp, nil
	}, http.StatusOK, &httpdto.Empty{})
}

func (h *ReaderSettingsHandler) PatchReaderState() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.UpdateUserReaderStateRequest) (*response.Response[domain.UserReaderState], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		input, err := req.ToUsecase()
		if err != nil {
			return nil, err
		}

		updated, err := h.readerStateService.Patch(c.UserContext(), userID, input)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.UserReaderState]{
			Status:  http.StatusOK,
			Success: true,
			Message: "Successfully updated reader state!",
			Data:    updated,
		}
		return &resp, nil
	}, http.StatusOK, &httpdto.UpdateUserReaderStateRequest{})
}
