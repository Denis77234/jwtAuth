package service

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"medodsTest/internal/model"
	"medodsTest/pkg/jwt"
)

const (
	AccessTokenLifetime  = 24 * 10 * time.Hour
	RefreshTokenLifetime = 24 * 30 * time.Hour
)

func (tm *TokenManager) makeRefreshToken(guid, refresh string, iat int64) (model.Token, error) {
	refExpirationTime := time.Now().Add(RefreshTokenLifetime)

	bcryptedToken, err := bcrypt.GenerateFromPassword([]byte(refresh), 5)
	if err != nil {
		return model.Token{}, err
	}

	refToken := model.Token{GUID: guid, Refresh: bcryptedToken, ExpTime: refExpirationTime.Unix(), Iat: iat}

	return refToken, nil
}

func (tm *TokenManager) guidIatFromToken(access string) (string, int64, error) {
	_, accPayload, err := jwt.ParseToStruct(access)
	if err != nil {
		return "", 0, ErrInvalidFormat
	}

	guid := accPayload.Sub

	if guid == "" {
		return "", 0, ErrInvalidFormat
	}

	iat := accPayload.Iat

	if iat == 0 {
		return "", 0, ErrInvalidFormat
	}

	return guid, iat, nil
}

func (tm *TokenManager) validateRefresh(refreshFromCookie string, refreshFromDb model.Token) error {
	err := bcrypt.CompareHashAndPassword(refreshFromDb.Refresh, []byte(refreshFromCookie))
	if err != nil {
		return ErrInvalidToken
	}

	if refreshFromDb.ExpTime < time.Now().Unix() {
		return ErrExpiredToken
	}

	return nil
}
