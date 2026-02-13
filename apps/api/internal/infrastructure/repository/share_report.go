package repository

import (
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/config"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/cache"
	"gorm.io/gorm"
)

type ShareReportRepository = port.ShareReportRepository

type shareReportRepository struct {
	ResourceRepository[domain.ShareReport]
}

func NewShareReportRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) ShareReportRepository {
	return &shareReportRepository{
		ResourceRepository: NewResourceRepository[domain.ShareReport](cfg, db, cacheClient),
	}
}
