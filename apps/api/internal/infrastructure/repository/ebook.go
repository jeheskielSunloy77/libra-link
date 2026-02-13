package repository

import (
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/config"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/cache"
	"gorm.io/gorm"
)

type EbookRepository = port.EbookRepository

type ebookRepository struct {
	ResourceRepository[domain.Ebook]
}

func NewEbookRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) EbookRepository {
	return &ebookRepository{
		ResourceRepository: NewResourceRepository[domain.Ebook](cfg, db, cacheClient),
	}
}
