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
	ErrCantValidate  = errors.New("validation error")
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("expired token")
)

type refreshStorage interface {
	Add(ctx context.Context, token model.Token) error
	Find(ctx context.Context, guid string) (model.Token, error)
	Delete(ctx context.Context, guid string) error
	Update(ctx context.Context, guid string, upd model.Token) error
}

type TokenManager struct {
	db               refreshStorage
	jwtGenerator     jwt.Generator
	refreshAlgorithm func() hash.Hash
	refreshSecretKet string
}

func NewTokenManager(db refreshStorage, jwtG jwt.Generator, refKey string, refAlg func() hash.Hash) *TokenManager {
	tm := &TokenManager{jwtGenerator: jwtG, db: db, refreshSecretKet: refKey, refreshAlgorithm: refAlg}
	return tm
}

func (tm *TokenManager) GetTokens(ctx context.Context, guid string) ([]byte, error) {
	err := tm.db.Delete(ctx, guid)
	if err != nil {
		return nil, fmt.Errorf("can't delete setted tokens: %w\n", err)
	}

	accExpirationTime := time.Now().Add(AccessTokenLifetime)
	newAcc, newRef, err := tm.generateTokens(guid, accExpirationTime)
	if err != nil {
		return nil, fmt.Errorf("token generation: %w\n", err)
	}

	refreshToken, err := tm.makeRefreshToken(guid, newRef)
	if err != nil {
		return nil, fmt.Errorf("make refresh token error: %w\n", err)
	}

	tokensJson, err := tm.marshalTokens(newAcc, newRef)
	if err != nil {
		return nil, fmt.Errorf("generateTokens marshalling error: %w\n", err)
	}

	err = tm.db.Add(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("database access error: %w\n", err)
	}

	return tokensJson, nil
}

func (tm *TokenManager) RefreshTokens(ctx context.Context, access, refresh string) ([]byte, error) {
	if ok := tm.validatePair(refresh, access); !ok {
		return nil, ErrCantValidate
	}

	guid, err := tm.guidFromToken(access)
	if err != nil {
		return nil, err
	}

	refFromDB, err := tm.db.Find(ctx, guid)
	if err != nil {
		return nil, fmt.Errorf("cant't find refresh: %w\n", err)
	}

	err = tm.validateRefresh(refresh, refFromDB)
	if err != nil {
		if err == ErrExpiredToken {
			err = tm.db.Delete(ctx, guid)
			if err != nil {
				return nil, fmt.Errorf("can't delete expired refresh: %w\n", err)
			}
		}
		return nil, fmt.Errorf("refresh validation err: %w\n", err)
	}

	accExpirationTime := time.Now().Add(AccessTokenLifetime)
	newAcc, newRef, err := tm.generateTokens(guid, accExpirationTime)
	if err != nil {
		return nil, fmt.Errorf("can't generateRefresh generateTokens: %w\n", err)
	}

	refreshToken, err := tm.makeRefreshToken(guid, newRef)
	if err != nil {
		return nil, fmt.Errorf("can't make refresh token: %w\n", err)
	}

	tokensJson, err := tm.marshalTokens(newAcc, newRef)
	if err != nil {
		return nil, fmt.Errorf("can't marshall tokens: %w\n", err)
	}

	err = tm.db.Update(ctx, guid, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("can't update token: %w\n", err)
	}

	return tokensJson, nil
}

func (tm *TokenManager) generateRefresh(accessToken string) string {
	hashFunc := hmac.New(tm.refreshAlgorithm, []byte(tm.refreshSecretKet))

	hashFunc.Write([]byte(accessToken))

	return base64.RawURLEncoding.EncodeToString(hashFunc.Sum(nil))
}

func (tm *TokenManager) validatePair(refresh, access string) bool {
	check := tm.generateRefresh(access)

	val := check == refresh

	return val
}

func (tm *TokenManager) generateTokens(id string, expTime time.Time) (access string, refresh string, err error) {
	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: time.Now().Unix(), Exp: expTime.Unix()}

	access, err = tm.jwtGenerator.Generate(payload)
	if err != nil {
		return "", "", fmt.Errorf("jwt generator error: %w\n", err)

	}

	refresh = tm.generateRefresh(access)

	return access, refresh, nil
}
