package endpoint

import (
	"context"
	"fmt"
	bcrypt2 "golang.org/x/crypto/bcrypt"
	"log"
	"medosTest/internal/pkg/dal"
	"medosTest/internal/pkg/refresh"
	"medosTest/pkg/jwt"
	"net/http"
	"time"
)

type refreshDB interface {
	AddRefresh(ctx context.Context, token dal.Token) error
	FindRefresh(ctx context.Context, guid string) (dal.Token, error)
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

func (e *Endpoint) tokens(id string, expTime time.Time) (access string, refresh string) {
	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: time.Now().Unix(), Exp: expTime.Unix()}

	access = e.jwtG.Generate(payload)

	refresh = e.refH.Generate(access)

	return access, refresh
}

func (e *Endpoint) GetToken(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write([]byte("Invalid id"))
		if err != nil {
			log.Println(err)
		}
		return
	}

	accExpirationTime := time.Now().Add(time.Hour * 24 * 3)
	refExpirationTime := time.Now().Add(time.Hour * 24 * 30)

	acc, ref := e.tokens(id, accExpirationTime)

	bcrypt, err := bcrypt2.GenerateFromPassword([]byte(ref), 5)
	if err != nil {
		log.Println(err)
	}

	refToken := dal.NewToken(id, bcrypt, refExpirationTime.Unix())

	err = e.db.AddRefresh(context.TODO(), refToken)
	if err != nil {
		log.Println(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Access",
		Value:    acc,
		Expires:  accExpirationTime,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Refresh",
		Value:    ref,
		Expires:  refExpirationTime,
		HttpOnly: true,
	})

	_, err = fmt.Fprintf(w, "ACCESS: %v\n REFRESH: %v\n", acc, ref)
	if err != nil {
		log.Println(err)
	}
}
