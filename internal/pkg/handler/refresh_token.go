package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"

	"medosTest/internal/pkg/service"
)

func RefreshToken(tk *service.TokenManager, allowedMethods ...string) http.HandlerFunc {
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
			if errors.As(err, &mongo.ErrNoDocuments) || errors.As(err, &service.ValidationErr) || errors.As(err, &service.ErrInvalidFormat) || errors.As(err, &service.InvalidToken) {
				http.Error(w, "invalid token", http.StatusBadRequest)
				return
			} else if errors.As(err, &service.ExpiredToken) {
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
