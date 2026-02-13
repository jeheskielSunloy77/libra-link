package repository

import (
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/config"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/cache"
	"gorm.io/gorm"
)

type BookmarkRepository = port.BookmarkRepository

type bookmarkRepository struct {
	ResourceRepository[domain.Bookmark]
}

func NewBookmarkRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) BookmarkRepository {
	return &bookmarkRepository{
		ResourceRepository: NewResourceRepository[domain.Bookmark](cfg, db, cacheClient),
	}
}
