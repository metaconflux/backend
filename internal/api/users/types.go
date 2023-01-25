package users

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/metaconflux/backend/internal/api/users/jwtmaker"
	"github.com/metaconflux/backend/internal/api/users/repository"
	"github.com/metaconflux/backend/internal/cache"
	"github.com/metaconflux/backend/internal/resolver"
)

type UserApi struct {
	cache      cache.ICache
	resolver   resolver.IResolver
	jwtmaker   *jwtmaker.JWTMaker
	repository repository.UserRepository
}

func NewUserAPI(maker *jwtmaker.JWTMaker, cache cache.ICache, resolver resolver.IResolver, repository repository.UserRepository) UserApi {
	return UserApi{
		cache:      cache,
		resolver:   resolver,
		jwtmaker:   maker,
		repository: repository,
	}
}

type SignInBody struct {
	Address   string        `json:"address"`
	Nonce     string        `json:"nonce"`
	Lifetime  time.Duration `json:"lifetime"`
	Signature string        `json:"signature"`
	Text      string        `json:"text"`
}

type User struct {
	Nonce   string         `json:"nonce"`
	Address common.Address `json:"address"`
}

type AuthTextResult struct {
	Text       string     `json:"text"`
	SignInData SignInBody `json:"data"`
}

type SignInResult struct {
	Token string `json:"token"`
}
