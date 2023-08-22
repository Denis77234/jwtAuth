package app

import (
	"context"
	"crypto/sha256"
	"log"
	"net/http"
	"time"

	"medodsTest/internal/handler"
	"medodsTest/internal/mongo"
	"medodsTest/internal/service"
	"medodsTest/pkg/jwt"
)

func Start(ctx context.Context) {
	cfg := newCfg()

	log.Println("connecting to db...")
	db, err := mongo.New(cfg.mongoUri)
	if err != nil {
		log.Fatalf("database initialisation error: %v\n", err)
	}

	jwtG, err := jwt.NewGenerator(jwt.HS512, cfg.accessKey)
	if err != nil {
		log.Fatalf("jwt generator initialisation error: %v\n", err)
	}

	tk, err := service.NewTokenManager(&db, jwtG, cfg.refreshKey, sha256.New)
	if err != nil {
		log.Fatalf("token manager initialisation error: %v\n", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/auth/Tokens", handler.GetToken(tk, http.MethodPost))
	mux.Handle("/auth/Refresh", handler.RefreshToken(tk, http.MethodPut))

	server := &http.Server{Addr: cfg.port, Handler: mux}

	go func() {
		log.Println("server started")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server initialisation error: %v\n", err)
		}
	}()

	<-ctx.Done()

	timedCtxServ, cancel1 := context.WithTimeout(context.Background(), time.Second*15)

	defer cancel1()

	err = server.Shutdown(timedCtxServ)
	if err != nil {
		log.Fatalf("server shutdown failed:%+v\n", err)
	}

	timedCtxDb, cancel2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel2()

	err = db.Close(timedCtxDb)
	if err != nil {
		log.Fatalf("database shutdown failed:%+v\n", err)
	}

	log.Println("app exited properly")
}
