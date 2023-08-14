package myapp

import (
	"crypto/sha256"
	"log"
	"medosTest/internal/app/endpoint"
	"medosTest/internal/pkg/dal"
	"medosTest/internal/pkg/refresh"
	"medosTest/internal/pkg/restDecorator"
	"medosTest/pkg/jwt"
	"net/http"
)

type Myapp struct {
	endpoint *endpoint.Endpoint
	mux      *http.ServeMux
}

func New() *Myapp {
	m := &Myapp{}

	db, err := dal.New("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	jwtG, err := jwt.NewGenerator("HS512", "topsecret")
	if err != nil {
		log.Fatal(err)
	}

	ref := refresh.NewHandler(sha256.New, "refsecret")

	m.endpoint = endpoint.New(&db, jwtG, ref)

	getTokens := restDecorator.New(m.endpoint.GetTokens).SetMethods("POST")

	refreshTokens := restDecorator.New(m.endpoint.RefreshTokens).SetMethods("PUT")

	m.mux = http.NewServeMux()

	m.mux.Handle("/auth/Tokens", getTokens)
	m.mux.Handle("/auth/Refresh", refreshTokens)

	return m
}

func (m *Myapp) Start() {
	err := http.ListenAndServe(":4000", m.mux)
	if err != nil {
		log.Fatal(err)
	}
}
