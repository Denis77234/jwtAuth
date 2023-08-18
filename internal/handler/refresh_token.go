package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"

	service2 "medosTest/internal/service"
)

func RefreshToken(tk *service2.TokenManager, allowedMethods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := checkMethods(r, allowedMethods)
		if err != nil {
			strErr := fmt.Sprint(err)
			http.Error(w, strErr, http.StatusBadRequest)
			return
		}

		accFromCookie, refFromCookie, err := tokensFromCookie(r)
		if err != nil {
			http.Error(w, "invalid request format", http.StatusBadRequest)
			return
		}

		newAcc, newRef, err := tk.RefreshTokens(accFromCookie, refFromCookie)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) || errors.Is(err, service2.ValidationErr) || errors.Is(err, service2.ErrInvalidFormat) || errors.Is(err, service2.InvalidToken) {
				http.Error(w, "invalid token", http.StatusBadRequest)
				return
			} else if errors.Is(err, service2.ExpiredToken) {
				http.Error(w, "expired token", http.StatusBadRequest)
				return
			} else {
				log.Printf("refresh tokens: %v\n", err)
				http.Error(w, "something went wrong", http.StatusInternalServerError)
				return
			}
		}

		setCookie(newAcc, newRef, w)

		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
	}
}
