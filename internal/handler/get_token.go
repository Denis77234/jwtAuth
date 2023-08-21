package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"medodsTest/internal/model"
	"medodsTest/internal/service"
)

func GetToken(tm *service.TokenManager, allowedMethods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowed := checkMethods(r, allowedMethods)
		if !allowed {
			err := fmt.Errorf("method %v not allowed, use:%v\n", r.Method, allowedMethods)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var GUID model.Guid

		err := json.NewDecoder(r.Body).Decode(&GUID)
		if err != nil {
			http.Error(w, "invalid guid", http.StatusBadRequest)
			return
		}

		if GUID.Guid == "" {
			http.Error(w, "invalid guid", http.StatusBadRequest)
			return
		}

		tokensJson, err := tm.GetTokens(r.Context(), GUID.Guid)
		if err != nil {
			log.Printf("GetTokens error: %v\n", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, err = w.Write(tokensJson)
		if err != nil {
			log.Printf("Writing response error: %v\n", err)
			return
		}
	}
}
