package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	Migrate() error
	Create(c context.Context, user UserModel) error
	Get(c context.Context, id string) (UserModel, error)
	GetByEmail(c context.Context, email string) (UserModel, error)
	GetByAddress(c context.Context, address string) (UserModel, error)
	NewLogin(c context.Context, id string, nonce string) error
	CreateManifest(c context.Context, manifest ManifestModel) error
	GetManifests(c context.Context, id string) ([]ManifestModel, error)
}

type UserModel struct {
	gorm.Model
	ID        string    `json:"id" gorm:"primarykey"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	LastLogin time.Time `json:"lastLogin"`
	Nonce     string    `json:"nonce"`
}

type ManifestModel struct {
	gorm.Model
	Address string    `json:"address"`
	User    UserModel `json:"user"`
	UserID  string
}
