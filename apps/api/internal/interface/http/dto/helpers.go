package dto

import (
	"github.com/google/uuid"
	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
)

func parseUUID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errs.NewBadRequestError("invalid UUID value", true, []errs.FieldError{{Field: "id", Error: "must be a valid uuid"}}, nil)
	}
	return id, nil
}
