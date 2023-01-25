package sqlite

import (
	"errors"
	"time"

	"github.com/metaconflux/backend/internal/resolver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Resolver struct {
	resolver.IResolver
	db *gorm.DB
}

type ResolverModel struct {
	gorm.Model
	Key      string `gorm:"uniqueindex"`
	Value    string
	Lifetime time.Time
}

func NewResolver(db *gorm.DB) (resolver.IResolver, error) {
	err := db.AutoMigrate(&ResolverModel{})
	if err != nil {
		return nil, err
	}
	return &Resolver{
		db: db,
	}, nil
}

func (r *Resolver) Get(key string) (string, error) {
	var model ResolverModel
	result := r.db.Find(&model, "key = ?", key)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", resolver.ErrNotFound
		}

		return "", result.Error
	}

	if model.Key == "" {
		return "", resolver.ErrNotFound
	}

	if model.Lifetime.Unix() > 0 && model.Lifetime.Before(time.Now()) {
		return model.Value, resolver.ErrLifetime
	}

	return model.Value, nil
}
func (r *Resolver) Set(key string, val string, timeout int64) error {

	var t time.Time

	if timeout > 0 {
		t = time.Now().Add(time.Duration(timeout) * time.Minute)
	}

	model := ResolverModel{
		Key:      key,
		Value:    val,
		Lifetime: t,
	}

	result := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "lifetime", "updated_at"}),
	}).Create(&model)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
