package sqliterepo

import (
	"github.com/metaconflux/backend/internal/api/users/repository"
	"gorm.io/gorm"
)

type Sqlite struct {
	repository.UserRepository
	db *gorm.DB
}

func NewSqliteRepository(db *gorm.DB) (repository.UserRepository, error) {
	return &Sqlite{
		db: db,
	}, nil
}
