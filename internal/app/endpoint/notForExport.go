package endpoint

import (
	"context"
	bcrypt2 "golang.org/x/crypto/bcrypt"
	"log"
	"medosTest/internal/pkg/models"
	"medosTest/pkg/jwt"
	"net/http"
	"time"
)

const accExp = 24 * 10

const refreshExp = 24 * 30

func (e *Endpoint) tokens(id string, expTime time.Time) (access string, refresh string) {
	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: time.Now().Unix(), Exp: expTime.Unix()}

	access = e.jwtG.Generate(payload)

	refresh = e.refH.Generate(access)

	return access, refresh
}

func (e *Endpoint) tokensFromCookie(r *http.Request) (string, string, error) {
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

func (e Endpoint) setTokens(guid string, w http.ResponseWriter, option string) {
	accExpirationTime := time.Now().Add(time.Hour * accExp)
	refExpirationTime := time.Now().Add(time.Hour * refreshExp)

	newAcc, newRef := e.tokens(guid, accExpirationTime)

	bcrypt, err := bcrypt2.GenerateFromPassword([]byte(newRef), 5)
	if err != nil {
		log.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	refToken := models.NewToken(guid, bcrypt, refExpirationTime.Unix())

	switch option {
	case "update":
		err = e.db.Update(context.TODO(), guid, refToken)
		if err != nil {
			log.Println(err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

	case "add":
		err = e.db.Add(context.TODO(), refToken)
		if err != nil {
			log.Println(err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Access",
		Value:    newAcc,
		Expires:  refExpirationTime,
		HttpOnly: true,
		Path:     "/auth",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Refresh",
		Value:    newRef,
		Expires:  refExpirationTime,
		HttpOnly: true,
		Path:     "/auth",
	})

	w.Header().Set("Cache-Control", "no-store")
}
