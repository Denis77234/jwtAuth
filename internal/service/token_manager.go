package service

import (
	"context"
	"fmt"
	"time"

	"medosTest/internal/models"
	"medosTest/internal/refresh"
	"medosTest/pkg/jwt"
)

type refreshStorage interface {
	Add(ctx context.Context, token models.Token) error
	Find(ctx context.Context, guid string) (models.Token, error)
	Delete(ctx context.Context, guid string) error
	Update(ctx context.Context, guid string, upd models.Token) error
}

type TokenManager struct {
	db   refreshStorage
	jwtG jwt.Generator
	refH refresh.Handler
}

func New(db refreshStorage, jwtG jwt.Generator, refH refresh.Handler) *TokenManager {
	e := &TokenManager{jwtG: jwtG, db: db, refH: refH}
	return e
}

func (t *TokenManager) GetTokens(guid string) (string, string, error) {
	err := t.deleteTokenIfInDb(guid)
	if err != nil {
		return "", "", fmt.Errorf("can't delete setted tokens: %w\n", err)
	}

	accExpirationTime := time.Now().Add(time.Hour * AccExp)
	newAcc, newRef, err := t.tokens(guid, accExpirationTime)
	if err != nil {
		return "", "", fmt.Errorf("token generation: %w\n", err)
	}

	refreshToken, err := t.makeRefreshToken(guid, newRef)
	if err != nil {
		return "", "", fmt.Errorf("make refresh token error: %w\n", err)
	}

	err = t.db.Add(context.TODO(), refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("database access error: %w\n", err)
	}

	return newAcc, newRef, nil
}

func (t *TokenManager) RefreshTokens(access, refresh string) (string, string, error) {
	if ok := t.refH.Validate(refresh, access); !ok {
		return "", "", ValidationErr
	}

	guid, err := t.guidFromToken(access)
	if err != nil {
		return "", "", err
	}

	refFromDB, err := t.db.Find(context.TODO(), guid)
	if err != nil {
		return "", "", fmt.Errorf("cant't find refresh: %w\n", err)
	}

	err = t.validateRefresh(refresh, refFromDB)
	if err != nil {
		if err == ExpiredToken {
			err = t.db.Delete(context.TODO(), guid)
			if err != nil {
				return "", "", fmt.Errorf("can't delete expired refresh: %w\n", err)
			}
		}
		return "", "", err
	}

	accExpirationTime := time.Now().Add(time.Hour * AccExp)
	newAcc, newRef, err := t.tokens(guid, accExpirationTime)
	if err != nil {
		return "", "", fmt.Errorf("can't generate tokens: %w\n", err)
	}

	refreshToken, err := t.makeRefreshToken(guid, newRef)
	if err != nil {
		return "", "", fmt.Errorf("can't make refresh token: %w\n", err)
	}

	err = t.db.Update(context.TODO(), guid, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("can't update token: %w\n", err)
	}

	return newAcc, newRef, nil
}
