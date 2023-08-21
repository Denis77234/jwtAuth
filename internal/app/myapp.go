package app

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"medodsTest/internal/handler"
	"medodsTest/internal/mongo"
	"medodsTest/internal/service"
	"medodsTest/pkg/jwt"
)

const algorithmName = "HS512"

func Start(ctx context.Context) {
	cfg := newCfg()

	fmt.Println("connecting to db...")
	db, err := mongo.New(cfg.mongoUri)
	if err != nil {
		log.Fatalf("database initialisation error: %v\n", err)
	}

	jwtG, err := jwt.NewGenerator(algorithmName, cfg.accessKey)
	if err != nil {
		log.Fatalf("jwt generator initialisation error: %v\n", err)
	}

	tk := service.NewTokenManager(&db, jwtG, cfg.refreshKey, sha256.New)

	mux := http.NewServeMux()

	mux.Handle("/auth/Tokens", handler.GetToken(tk, "POST"))
	mux.Handle("/auth/Refresh", handler.RefreshToken(tk, "PUT"))

	server := &http.Server{Addr: cfg.port, Handler: mux}

	go func() {
		fmt.Println("server started")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server initialisation error: %v\n", err)
		}
	}()

	<-ctx.Done()

	stop(server, &db)

	log.Println("app exited properly")
}

func stop(serv *http.Server, db io.Closer) {

	timedCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	defer cancel()

	err := serv.Shutdown(timedCtx)
	if err != nil {
		log.Fatalf("server shutdown failed:%+v\n", err)
	}

	err = db.Close()
	if err != nil {
		log.Fatalf("database shutdown failed:%+v\n", err)
	}

}
