package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
	"github.com/jeheskielSunloy77/libra-link/internal/interface/http/middleware"
)

func parseAuthenticatedUserID(c *fiber.Ctx) (uuid.UUID, error) {
	raw := middleware.GetUserID(c)
	if raw == "" {
		return uuid.Nil, errs.NewUnauthorizedError("Unauthorized", false)
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errs.NewUnauthorizedError("Unauthorized", false)
	}

	return id, nil
}
