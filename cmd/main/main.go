package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"medosTest/internal/app/endpoint"
	"medosTest/internal/pkg/dal"
	"medosTest/internal/pkg/refresh"
	"medosTest/internal/pkg/resthandler"
	"medosTest/pkg/jwt"
	"net/http"
)

func main() {

	db, err := dal.New("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	jwtG, err := jwt.NewGenerator("HS512", "topsecret")
	if err != nil {
		log.Fatal(err)
	}

	ref := refresh.NewHandler(sha256.New, "refsecret")

	endp := endpoint.New(&db, jwtG, ref)

	newtoken := resthandler.New(endp.GetTokens).SetMethods("POST")

	reftokens := resthandler.New(endp.RefreshTokens).SetMethods("PUT", "POST")

	mux := http.NewServeMux()

	mux.Handle("/auth/Tokens", newtoken)
	mux.Handle("/auth/Refresh", reftokens)

	err = http.ListenAndServe(":4000", mux)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(1)
}
