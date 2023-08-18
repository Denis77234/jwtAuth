package app

import (
	"crypto/sha256"
	"log"
	"net/http"

	handler2 "medosTest/internal/handler"
	"medosTest/internal/mongo"
	"medosTest/internal/refresh"
	"medosTest/internal/service"
	"medosTest/pkg/jwt"
)

type Myapp struct {
	tk  *service.TokenManager
	mux *http.ServeMux
}

func New() *Myapp {
	m := &Myapp{}

	cfg := newCfg()

	db, err := mongo.New(cfg.mongoUri)
	if err != nil {
		log.Fatalf("database initialisation error: %v\n", err)
	}

	jwtG, err := jwt.NewGenerator("HS512", cfg.accessKey)
	if err != nil {
		log.Fatalf("jwt generator initialisation error: %v\n", err)
	}

	ref := refresh.NewHandler(sha256.New, cfg.refreshKey)

	m.tk = service.New(&db, jwtG, ref)

	m.mux = http.NewServeMux()

	m.mux.Handle("/auth/Tokens", handler2.GetToken(m.tk, "POST"))
	m.mux.Handle("/auth/Refresh", handler2.RefreshToken(m.tk, "PUT"))

	return m
}

func (m *Myapp) Start() {
	err := http.ListenAndServe(":4000", m.mux)
	if err != nil {
		log.Fatalf("server initialisation error: %v\n", err)
	}
}
