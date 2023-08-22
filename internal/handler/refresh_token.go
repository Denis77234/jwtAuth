package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"

	"medodsTest/internal/model"
	"medodsTest/internal/service"
)

func RefreshToken(tk *service.TokenManager, allowedMethods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowed := checkMethods(r, allowedMethods)
		if !allowed {
			err := fmt.Errorf("method %v not allowed, use:%v\n", r.Method, allowedMethods)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		accessFromUser := getAccessTokenFromHeader(r)
		if accessFromUser == "" {
			http.Error(w, "invalid access token", http.StatusBadRequest)
			return
		}

		var refreshFromUser model.RefreshToken

		err := json.NewDecoder(r.Body).Decode(&refreshFromUser)
		if err != nil {
			http.Error(w, "invalid refresh token", http.StatusBadRequest)
			return
		}

		if refreshFromUser.Token == "" {
			http.Error(w, "invalid refresh token", http.StatusBadRequest)
			return
		}

		access, refresh, err := tk.RefreshTokens(r.Context(), accessFromUser, refreshFromUser.Token)
		if err != nil {
			if oneOfErrors(err, mongo.ErrNoDocuments, service.ErrInvalidToken, service.ErrInvalidFormat) {
				http.Error(w, "invalid token", http.StatusBadRequest)
			} else if errors.Is(err, service.ErrExpiredToken) {
				http.Error(w, "expired token", http.StatusBadRequest)
			} else {
				log.Printf("refresh tokens: %v\n", err)
				http.Error(w, "something went wrong", http.StatusInternalServerError)
			}
			return
		}

		tokensJson, err := marshalTokens(access, refresh)
		if err != nil {
			log.Printf("marshal tokens error: %v\n", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(tokensJson)
		if err != nil {
			log.Printf("Writing response error: %v\n", err)
			return
		}
	}
}
