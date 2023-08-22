package service

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"time"

	"medodsTest/internal/model"
	"medodsTest/pkg/jwt"
)

var (
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("expired token")
)

type refreshStorage interface {
	Add(ctx context.Context, token model.Token) error
	Find(ctx context.Context, guid string, iat int64) (model.Token, error)
	Delete(ctx context.Context, guid string, iat int64) error
	Update(ctx context.Context, guid string, iat int64, upd model.Token) error
}

type TokenManager struct {
	db               refreshStorage
	jwtGenerator     jwt.Generator
	refreshAlgorithm func() hash.Hash
	refreshSecretKey string
}

func NewTokenManager(db refreshStorage, jwtG jwt.Generator, refKey string, refAlg func() hash.Hash) (*TokenManager, error) {

	if refAlg == nil {
		return nil, errors.New("nil algorithm function")
	}

	tm := &TokenManager{jwtGenerator: jwtG, db: db, refreshSecretKey: refKey, refreshAlgorithm: refAlg}
	return tm, nil
}

func (tm *TokenManager) GetTokens(ctx context.Context, guid string) (string, string, error) {
	inspirationTime := time.Now().UnixMilli()

	accExpirationTime := time.Now().Add(AccessTokenLifetime)
	newAcc, newRef, err := tm.generateTokens(guid, accExpirationTime, inspirationTime)
	if err != nil {
		return "", "", fmt.Errorf("token generation: %w", err)
	}

	refreshToken, err := tm.makeRefreshToken(guid, newRef, inspirationTime)
	if err != nil {
		return "", "", fmt.Errorf("make refresh token error: %w", err)
	}

	err = tm.db.Add(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("database access error: %w", err)
	}

	return newAcc, newRef, nil
}

func (tm *TokenManager) RefreshTokens(ctx context.Context, access, refresh string) (string, string, error) {
	if ok := tm.validatePair(refresh, access); !ok {
		return "", "", ErrInvalidToken
	}

	guid, iat, err := tm.guidIatFromToken(access)
	if err != nil {
		return "", "", err
	}

	refFromDB, err := tm.db.Find(ctx, guid, iat)
	if err != nil {
		return "", "", fmt.Errorf("cant't find refresh: %w", err)
	}

	err = tm.validateRefresh(refresh, refFromDB)
	if err != nil {
		if err == ErrExpiredToken {
			err = tm.db.Delete(ctx, guid, iat)
			if err != nil {
				return "", "", fmt.Errorf("can't delete expired refresh: %w", err)
			}
		}
		return "", "", fmt.Errorf("refresh validation err: %w", err)
	}

	inspirationTime := time.Now().UnixMilli()

	accExpirationTime := time.Now().Add(AccessTokenLifetime)
	newAcc, newRef, err := tm.generateTokens(guid, accExpirationTime, inspirationTime)
	if err != nil {
		return "", "", fmt.Errorf("can't generate tokens: %w", err)
	}

	refreshToken, err := tm.makeRefreshToken(guid, newRef, inspirationTime)
	if err != nil {
		return "", "", fmt.Errorf("can't make refresh token: %w", err)
	}

	err = tm.db.Update(ctx, guid, iat, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("can't update token: %w", err)
	}

	return newAcc, newRef, nil
}

func (tm *TokenManager) generateRefresh(accessToken string) string {
	hashFunc := hmac.New(tm.refreshAlgorithm, []byte(tm.refreshSecretKey))

	hashFunc.Write([]byte(accessToken))

	return base64.RawURLEncoding.EncodeToString(hashFunc.Sum(nil))
}

func (tm *TokenManager) validatePair(refresh, access string) bool {
	check := tm.generateRefresh(access)

	val := check == refresh

	return val
}

func (tm *TokenManager) generateTokens(id string, expTime time.Time, iat int64) (access string, refresh string, err error) {
	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: iat, Exp: expTime.Unix()}

	access, err = tm.jwtGenerator.Generate(payload)
	if err != nil {
		return "", "", fmt.Errorf("jwt generator error: %w", err)

	}

	refresh = tm.generateRefresh(access)

	return access, refresh, nil
}
