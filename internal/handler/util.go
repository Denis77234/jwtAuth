package handler

import (
	"errors"
	"net/http"
	"strings"
)

func getAccessTokenFromHeader(r *http.Request) string {
	rawToken := r.Header.Get("Authorization")
	accessToken := strings.TrimSpace(strings.Replace(rawToken, "Bearer", "", 1))

	return accessToken
}

func checkMethods(r *http.Request, allowedMethods []string) bool {
	for _, method := range allowedMethods {
		if r.Method == method {
			return true
		}
	}

	return false
}

func oneOfErrors(myerr error, targetsErr ...error) bool {

	for _, err := range targetsErr {
		if errors.Is(myerr, err) {
			return true
		}
	}

	return false

}
