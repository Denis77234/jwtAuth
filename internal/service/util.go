package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	bcrypt2 "golang.org/x/crypto/bcrypt"

	"medosTest/internal/models"
	"medosTest/pkg/jwt"
)

const (
	AccExp     = 24 * 10
	RefreshExp = 24 * 30
)

var (
	ValidationErr    = errors.New("validation error")
	ErrInvalidFormat = errors.New("invalid format")
	InvalidToken     = errors.New("invalid token")
	ExpiredToken     = errors.New("expired token")
)

func (t *TokenManager) tokens(id string, expTime time.Time) (access string, refresh string, err error) {
	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: time.Now().Unix(), Exp: expTime.Unix()}

	access, err = t.jwtG.Generate(payload)
	if err != nil {
		return "", "", fmt.Errorf("jwt geneartor error: %v\n", err)

	}

	refresh = t.refH.Generate(access)

	return access, refresh, nil
}

func (t *TokenManager) makeRefreshToken(guid, refresh string) (models.Token, error) {
	refExpirationTime := time.Now().Add(time.Hour * RefreshExp)

	bcryptedToken, err := bcrypt2.GenerateFromPassword([]byte(refresh), 5)
	if err != nil {
		return models.Token{}, fmt.Errorf("bcryption error: %v\n", err)
	}

	refToken := models.Token{GUID: guid, Refresh: bcryptedToken, ExpTime: refExpirationTime.Unix()}

	return refToken, nil
}

func (t *TokenManager) deleteTokenIfInDb(guid string) error {
	tokenExistsInDb := true

	_, err := t.db.Find(context.TODO(), guid)
	if err != nil {
		if errors.Is(mongo.ErrNoDocuments, err) {
			tokenExistsInDb = false
		} else {
			return fmt.Errorf("database access error: %v\n", err)
		}
	}

	if tokenExistsInDb {
		err = t.db.Delete(context.TODO(), guid)
		if err != nil {
			return fmt.Errorf("database access error: %v\n", err)
		}
	}
	return nil
}

func (t *TokenManager) guidFromToken(access string) (string, error) {
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

func (t *TokenManager) validateRefresh(refreshFromCookie string, refreshFromDb models.Token) error {
	err := bcrypt2.CompareHashAndPassword(refreshFromDb.Refresh, []byte(refreshFromCookie))
	if err != nil {
		return InvalidToken
	}

	if refreshFromDb.ExpTime < time.Now().Unix() {
		return ExpiredToken
	}

	return nil
}
