package handler

import (
	"fmt"
	"log"
	"net/http"

	"medosTest/internal/service"
)

func GetToken(tk *service.TokenManager, allowedMethods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := checkMethods(r, allowedMethods)
		if err != nil {
			strErr := fmt.Sprint(err)
			http.Error(w, strErr, http.StatusBadRequest)
			return
		}

		_, _, err = tokensFromCookie(r)
		if err != http.ErrNoCookie {
			http.Error(w, "tokens already set", http.StatusBadRequest)
			return
		}

		guid, err := getGUID(r)
		if err != nil {
			http.Error(w, "invalid guid", http.StatusBadRequest)
			return
		}

		access, refresh, err := tk.GetTokens(guid)
		if err != nil {
			log.Printf("GetTokens error: %v\n", err)

			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		setCookie(access, refresh, w)

		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
	}
}
