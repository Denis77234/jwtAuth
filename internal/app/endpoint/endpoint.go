package endpoint

import (
	"context"
	"encoding/json"
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

func New(db refreshDB, jwtG jwt.Generator, refH refresh.Handler) *Endpoint {
	e := &Endpoint{jwtG: jwtG, db: db, refH: refH}
	return e
}

func (e *Endpoint) GetTokens(w http.ResponseWriter, r *http.Request) {
	accFromCookie, refFromCookie, _ := e.tokensFromCookie(r)
	if accFromCookie != "" || refFromCookie != "" {
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

	e.setTokens(GUID.GUID, "add", w)

	w.WriteHeader(http.StatusOK)
}

func (e Endpoint) RefreshTokens(w http.ResponseWriter, r *http.Request) {

	accFromCookie, refFromCookie, err := e.tokensFromCookie(r)
	if err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	if ok := e.refH.Validate(refFromCookie, accFromCookie); !ok {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	_, accPayload, err := jwt.ParseToStruct(accFromCookie)
	if err != nil {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	guid := accPayload.Sub
	if guid == "" {
		http.Error(w, "invalid request format", http.StatusBadRequest)
		return
	}

	refFromDB, err := e.db.Find(context.TODO(), guid)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}
		log.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	err = bcrypt2.CompareHashAndPassword(refFromDB.Refresh, []byte(refFromCookie))
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

	e.setTokens(guid, "update", w)

	w.WriteHeader(http.StatusOK)
}
