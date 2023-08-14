package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	bcrypt2 "golang.org/x/crypto/bcrypt"
	"log"
	"medosTest/internal/pkg/models"
	"medosTest/internal/pkg/refresh"
	"medosTest/pkg/jwt"
	"net/http"
	"time"
)

type refreshDB interface {
	Add(ctx context.Context, token models.Token) error
	Find(ctx context.Context, guid string) (models.Token, error)
	Delete(ctx context.Context, guid string) error
	Update(ctx context.Context, guid string, upd models.Token) error
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

	accExpirationTime := time.Now().Add(time.Hour * accExp)
	refExpirationTime := time.Now().Add(time.Hour * refreshExp)

	acc, ref := e.tokens(GUID.GUID, accExpirationTime)

	bcrypt, err := bcrypt2.GenerateFromPassword([]byte(ref), 5)
	if err != nil {
		log.Println(err)
	}

	refToken := models.NewToken(GUID.GUID, bcrypt, refExpirationTime.Unix())

	err = e.db.Add(context.TODO(), refToken)
	if err != nil {
		log.Println(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Access",
		Value:    acc,
		Expires:  refExpirationTime,
		HttpOnly: true,
		Path:     "/auth",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Refresh",
		Value:    ref,
		Expires:  refExpirationTime,
		HttpOnly: true,
		Path:     "/auth",
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

	_, accPayload, err := e.jwtG.ParseToStruct(acc)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	guid := accPayload.Sub

	refFromDB, err := e.db.Find(context.TODO(), guid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	err = bcrypt2.CompareHashAndPassword(refFromDB.Refresh, []byte(ref))
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	if refFromDB.ExpTime < time.Now().Unix() {
		http.Error(w, "expired token", http.StatusBadRequest)

		e.db.Delete(context.TODO(), guid)

		//Not all browsers allow setting cookies with a 4XX code
		http.SetCookie(w, &http.Cookie{
			Name:     "Refresh",
			Value:    "",
			Expires:  time.Now(),
			HttpOnly: true,
			Path:     "/auth",
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "Access",
			Value:    "",
			Expires:  time.Now(),
			HttpOnly: true,
			Path:     "/auth",
		})
		return
	}

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

	err = e.db.Update(context.TODO(), guid, refToken)
	if err != nil {
		log.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
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
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "Acc:%v\n Ref:%v\n", acc, ref)
}
