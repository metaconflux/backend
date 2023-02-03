package sqliterepo

import (
	"context"
	"time"

	"github.com/metaconflux/backend/internal/api/users/repository"
)

func (r *Sqlite) Migrate() error {
	return r.db.AutoMigrate(&repository.TierModel{}, &repository.UserModel{}, &repository.ManifestModel{})
}

func (r *Sqlite) Create(c context.Context, user repository.UserModel) error {
	result := r.db.Create(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *Sqlite) Get(c context.Context, id string) (repository.UserModel, error) {
	var user repository.UserModel
	result := r.db.Find(&user, "id = ?", id)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}
func (r *Sqlite) GetByEmail(c context.Context, email string) (repository.UserModel, error) {
	var user repository.UserModel
	result := r.db.Find(&user, "email = ?", email)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}
func (r *Sqlite) GetByAddress(c context.Context, address string) (repository.UserModel, error) {
	var user repository.UserModel
	result := r.db.Find(&user, "address = ?", address)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func (r *Sqlite) NewLogin(c context.Context, id string, nonce string) error {
	result := r.db.Model(&repository.UserModel{}).Where("id = ?", id).Updates(&repository.UserModel{Nonce: nonce, LastLogin: time.Now()})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *Sqlite) CreateManifest(c context.Context, manifest repository.ManifestModel) error {
	result := r.db.Create(&manifest)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *Sqlite) GetManifests(c context.Context, userId string) ([]repository.ManifestModel, error) {
	var manifests []repository.ManifestModel
	result := r.db.Find(&manifests, "user_id = ?", userId)
	if result.Error != nil {
		return nil, result.Error
	}

	return manifests, nil
}
