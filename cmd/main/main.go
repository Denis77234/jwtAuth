package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type header struct {
	alg string
	typ string
}

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

	jwt := makeHeader()

	_, err := fmt.Fprintf(w, "header is %v", jwt)
	if err != nil {
		log.Println(err)
	}
}

func makeHeader() string {
	hdr := header{alg: "sha512", typ: "jwt"}

	bytes, err := json.Marshal(hdr)
	if err != nil {
		log.Println(err)
	}
	b64Hdr := base64.StdEncoding.EncodeToString(bytes)

	return b64Hdr
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/getToken", getToken)

	err := http.ListenAndServe(":4000", mux)
	if err != nil {
		log.Fatal(err)
	}
}
