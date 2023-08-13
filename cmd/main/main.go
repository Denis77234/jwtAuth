package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"medosTest/internal/pkg/dal"
	"medosTest/internal/pkg/refresh"
	"medosTest/pkg/jwt"
	"net/http"
	"time"
)

func getToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(405)
		_, err := w.Write([]byte("Wrong method"))
		if err != nil {
			log.Println(err)
		}
		return
	}

	id := r.URL.Query().Get("id")

	if id == "" {
		w.WriteHeader(405)
		_, err := w.Write([]byte("Invalid id"))
		if err != nil {
			log.Println(err)
		}
		return
	}

	jwtGen, err := jwt.NewGenerator("HS512", "topSecret")
	if err != nil {
		w.WriteHeader(500)
		_, err := w.Write([]byte("Something went wrong"))
		log.Println(err)
		return
	}

	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: time.Now().Unix()}
	jwToken := jwtGen.Generate(payload)

	refHan := refresh.NewHandler(sha256.New, "refreshKey")

	refreshToken := refHan.Generate(jwToken)

	val := refHan.Validate(refreshToken, jwToken)

	mongoDB, err := dal.New("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	ref, err := mongoDB.FindRefresh(context.TODO(), id)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(ref)
	_, err = fmt.Fprintf(w, "ACCESS: %v\n REFRESH: %v\n VAL:%v", jwToken, refreshToken, val)
	if err != nil {
		log.Println(err)
	}
}

//func validJWT(jwt string) bool {
//	parse := strings.Split(jwt, ".")
//	check := makeSignature(parse[0], parse[1])
//	return check == parse[2]
//}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/getToken", getToken)

	err := http.ListenAndServe(":4000", mux)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(1)
}
