package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jeheskielSunloy77/libra-link/internal/application"
	applicationdto "github.com/jeheskielSunloy77/libra-link/internal/application/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	httpdto "github.com/jeheskielSunloy77/libra-link/internal/interface/http/dto"
	"github.com/jeheskielSunloy77/libra-link/internal/interface/http/response"
)

type ReadingProgressHandler struct {
	*ResourceHandler[domain.ReadingProgress, *applicationdto.StoreReadingProgressInput, *applicationdto.UpdateReadingProgressInput, *httpdto.StoreReadingProgressRequest, *httpdto.UpdateReadingProgressRequest]
	service application.ReadingProgressService
}

func NewReadingProgressHandler(h Handler, service application.ReadingProgressService) *ReadingProgressHandler {
	return &ReadingProgressHandler{
		ResourceHandler: NewResourceHandler[domain.ReadingProgress, *applicationdto.StoreReadingProgressInput, *applicationdto.UpdateReadingProgressInput, *httpdto.StoreReadingProgressRequest, *httpdto.UpdateReadingProgressRequest]("reading_progress", h, service),
		service:         service,
	}
}

func (h *ReadingProgressHandler) Store() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.StoreReadingProgressRequest) (*response.Response[domain.ReadingProgress], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		input := req.ToUsecase()
		input.UserID = userID

		entity, err := h.service.Store(c.UserContext(), input)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.ReadingProgress]{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Successfully created reading_progress!",
			Data:    entity,
		}
		return &resp, nil
	}, http.StatusCreated, &httpdto.StoreReadingProgressRequest{})
}

type BookmarkHandler struct {
	*ResourceHandler[domain.Bookmark, *applicationdto.StoreBookmarkInput, *applicationdto.UpdateBookmarkInput, *httpdto.StoreBookmarkRequest, *httpdto.UpdateBookmarkRequest]
	service application.BookmarkService
}

func NewBookmarkHandler(h Handler, service application.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{
		ResourceHandler: NewResourceHandler[domain.Bookmark, *applicationdto.StoreBookmarkInput, *applicationdto.UpdateBookmarkInput, *httpdto.StoreBookmarkRequest, *httpdto.UpdateBookmarkRequest]("bookmark", h, service),
		service:         service,
	}
}

func (h *BookmarkHandler) Store() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.StoreBookmarkRequest) (*response.Response[domain.Bookmark], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		input := req.ToUsecase()
		input.UserID = userID

		entity, err := h.service.Store(c.UserContext(), input)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.Bookmark]{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Successfully created bookmark!",
			Data:    entity,
		}
		return &resp, nil
	}, http.StatusCreated, &httpdto.StoreBookmarkRequest{})
}

type AnnotationHandler struct {
	*ResourceHandler[domain.Annotation, *applicationdto.StoreAnnotationInput, *applicationdto.UpdateAnnotationInput, *httpdto.StoreAnnotationRequest, *httpdto.UpdateAnnotationRequest]
	service application.AnnotationService
}

func NewAnnotationHandler(h Handler, service application.AnnotationService) *AnnotationHandler {
	return &AnnotationHandler{
		ResourceHandler: NewResourceHandler[domain.Annotation, *applicationdto.StoreAnnotationInput, *applicationdto.UpdateAnnotationInput, *httpdto.StoreAnnotationRequest, *httpdto.UpdateAnnotationRequest]("annotation", h, service),
		service:         service,
	}
}

func (h *AnnotationHandler) Store() fiber.Handler {
	return Handle(h.Handler, func(c *fiber.Ctx, req *httpdto.StoreAnnotationRequest) (*response.Response[domain.Annotation], error) {
		userID, err := parseAuthenticatedUserID(c)
		if err != nil {
			return nil, err
		}

		input := req.ToUsecase()
		input.UserID = userID

		entity, err := h.service.Store(c.UserContext(), input)
		if err != nil {
			return nil, err
		}

		resp := response.Response[domain.Annotation]{
			Status:  http.StatusCreated,
			Success: true,
			Message: "Successfully created annotation!",
			Data:    entity,
		}
		return &resp, nil
	}, http.StatusCreated, &httpdto.StoreAnnotationRequest{})
}
