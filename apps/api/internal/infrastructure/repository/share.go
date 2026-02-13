package repository

import (
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/config"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/cache"
	"gorm.io/gorm"
)

type ShareRepository = port.ShareRepository

type shareRepository struct {
	ResourceRepository[domain.Share]
}

func NewShareRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) ShareRepository {
	return &shareRepository{
		ResourceRepository: NewResourceRepository[domain.Share](cfg, db, cacheClient),
	}
}
