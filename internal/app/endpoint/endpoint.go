package endpoint

import (
	"context"
	"fmt"
	"log"
	"medosTest/internal/pkg/dal"
	"medosTest/internal/pkg/refresh"
	"medosTest/pkg/jwt"
	"net/http"
	"time"
)

type refreshDB interface {
	AddRefresh(ctx context.Context, refresh, GUID string) error
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

func (e *Endpoint) tokens(id string) (access string, refresh string) {
	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: time.Now().Unix()}

	access = e.jwtG.Generate(payload)

	refresh = e.refH.Generate(access)

	return access, refresh
}

func (e *Endpoint) GetToken(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(401)
		_, err := w.Write([]byte("Invalid id"))
		if err != nil {
			log.Println(err)
		}
		return
	}

	acc, ref := e.tokens(id)

	_, err := fmt.Fprintf(w, "ACCESS: %v\n REFRESH: %v\n", acc, ref)
	if err != nil {
		log.Println(err)
	}
}
