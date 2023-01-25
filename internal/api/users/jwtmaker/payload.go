package jwtmaker

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type Payload struct {
	ID        uuid.UUID
	User      string
	IssuedAt  time.Time
	ExpiredAt time.Time
}

func NewPayload(user string, lifetime time.Duration) (*jwt.StandardClaims, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &jwt.StandardClaims{
		//		Audience:  "",
		ExpiresAt: time.Now().Add(lifetime).Unix(),
		Id:        tokenID.String(),
		IssuedAt:  time.Now().Unix(),
		//		Issuer:    "",
		NotBefore: time.Now().Unix(),
		Subject:   user,
	}, nil
}
