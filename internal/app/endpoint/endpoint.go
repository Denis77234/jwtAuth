package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	bcrypt2 "golang.org/x/crypto/bcrypt"
	"log"
	"medosTest/internal/pkg/models"
	"medosTest/internal/pkg/refresh"
	"medosTest/pkg/jwt"
	"net/http"
	"time"
)

type refreshDB interface {
	AddRefresh(ctx context.Context, token models.Token) error
	FindRefresh(ctx context.Context, guid string) (models.Token, error)
}

type Endpoint struct {
	db   refreshDB
	jwtG jwt.Generator
	refH refresh.Handler
}

func New(db refreshDB, jwtG jwt.Generator, refH refresh.Handler) Endpoint {
	e := Endpoint{jwtG: jwtG, db: db, refH: refH}
	return e
}

func (e *Endpoint) GetTokens(w http.ResponseWriter, r *http.Request) {
	accC, refC, _ := e.tokensFromCookie(r)
	if accC != "" || refC != "" {
		http.Error(w, "tokens already set", http.StatusBadRequest)
		return
	}

	var GUID models.GUID

	err := json.NewDecoder(r.Body).Decode(&GUID)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if GUID.GUID == "" {
		http.Error(w, "invalid id", http.StatusUnauthorized)
		return
	}

	accExpirationTime := time.Now().Add(time.Hour * 24 * 3)
	refExpirationTime := time.Now().Add(time.Hour * 24 * 30)

	acc, ref := e.tokens(GUID.GUID, accExpirationTime)

	bcrypt, err := bcrypt2.GenerateFromPassword([]byte(ref), 5)
	if err != nil {
		log.Println(err)
	}

	refToken := models.NewToken(GUID.GUID, bcrypt, refExpirationTime.Unix())

	err = e.db.AddRefresh(context.TODO(), refToken)
	if err != nil {
		log.Println(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Access",
		Value:    acc,
		Expires:  refExpirationTime,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Refresh",
		Value:    ref,
		Expires:  refExpirationTime,
		HttpOnly: true,
	})

	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
}

func (e Endpoint) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	acc, ref, err := e.tokensFromCookie(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if ok := e.refH.Validate(ref, acc); !ok {
		log.Println(err)
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Acc:%v\n Ref:%v\n", acc, ref)
}

func (e *Endpoint) tokens(id string, expTime time.Time) (access string, refresh string) {
	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: time.Now().Unix(), Exp: expTime.Unix()}

	access = e.jwtG.Generate(payload)

	refresh = e.refH.Generate(access)

	return access, refresh
}

func (e Endpoint) tokensFromCookie(r *http.Request) (string, string, error) {
	acc, err := r.Cookie("Access")
	if err != nil {
		return "", "", err
	}
	ref, err := r.Cookie("Refresh")
	if err != nil {
		return "", "", err
	}

	return acc.Value, ref.Value, nil
}
