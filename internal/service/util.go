package service

import (
	"encoding/json"
	"time"

	"golang.org/x/crypto/bcrypt"

	"medodsTest/internal/model"
	"medodsTest/pkg/jwt"
)

const (
	AccessTokenLifetime  = 24 * 10 * time.Hour
	RefreshTokenLifetime = 24 * 30 * time.Hour
)

func (tm *TokenManager) makeRefreshToken(guid, refresh string) (model.Token, error) {
	refExpirationTime := time.Now().Add(RefreshTokenLifetime)

	bcryptedToken, err := bcrypt.GenerateFromPassword([]byte(refresh), 5)
	if err != nil {
		return model.Token{}, err
	}

	refToken := model.Token{GUID: guid, Refresh: bcryptedToken, ExpTime: refExpirationTime.Unix()}

	return refToken, nil
}

func (tm *TokenManager) guidFromToken(access string) (string, error) {
	_, accPayload, err := jwt.ParseToStruct(access)
	if err != nil {
		return "", ErrInvalidFormat
	}

	guid := accPayload.Sub

	if guid == "" {
		return "", ErrInvalidFormat
	}

	return guid, nil
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

func (tm TokenManager) marshalTokens(acc, ref string) ([]byte, error) {
	pair := model.TokenPair{Access: acc, Refresh: ref}

	bytes, err := json.Marshal(pair)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
