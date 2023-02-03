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
	GetManifests(c context.Context, userId string) ([]ManifestModel, error)
}

type UserModel struct {
	gorm.Model
	ID        string    `json:"id" gorm:"primarykey"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	LastLogin time.Time `json:"lastLogin"`
	Nonce     string    `json:"nonce"`
	TierID    uint      `json:"tierId" gorm:"default:0"`
	Tier      TierModel `json:"tier"`
}

type TierModel struct {
	gorm.Model
	Name string
}

type ManifestModel struct {
	gorm.Model
	Address string    `json:"address" gorm:"uniqueindex"`
	ChainId int64     `json:"chainId"`
	User    UserModel `json:"user"`
	UserID  string    `json:"user_id"`
}
