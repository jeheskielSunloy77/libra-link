package repository

import (
	"github.com/jeheskielSunloy77/libra-link/internal/application/port"
	"github.com/jeheskielSunloy77/libra-link/internal/domain"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/config"
	"github.com/jeheskielSunloy77/libra-link/internal/infrastructure/lib/cache"
	"gorm.io/gorm"
)

type AnnotationRepository = port.AnnotationRepository

type annotationRepository struct {
	ResourceRepository[domain.Annotation]
}

func NewAnnotationRepository(cfg *config.Config, db *gorm.DB, cacheClient cache.Cache) AnnotationRepository {
	return &annotationRepository{
		ResourceRepository: NewResourceRepository[domain.Annotation](cfg, db, cacheClient),
	}
}
