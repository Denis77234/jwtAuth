package app

import (
	"crypto/sha256"
	"log"
	"net/http"

	"medosTest/internal/pkg/handler"
	"medosTest/internal/pkg/mongo"
	"medosTest/internal/pkg/refresh"
	"medosTest/internal/pkg/service"
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

	m.mux.Handle("/auth/Tokens", handler.GetToken(m.tk, "POST"))
	m.mux.Handle("/auth/Refresh", handler.RefreshToken(m.tk, "PUT"))

	return m
}

func (m *Myapp) Start() {
	err := http.ListenAndServe(":4000", m.mux)
	if err != nil {
		log.Fatalf("server initialisation error: %v\n", err)
	}
}
