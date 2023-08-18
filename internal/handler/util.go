package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"medosTest/internal/models"
	"medosTest/internal/service"
)

var errMissCookies = errors.New("miss cookies")

func tokensFromCookie(r *http.Request) (string, string, error) {
	acc, err1 := r.Cookie("Access")

	ref, err2 := r.Cookie("Refresh")

	if err1 != nil && err2 != nil {
		return "", "", errMissCookies
	}

	if err1 != nil || err2 != nil {
		return "", "", http.ErrNoCookie
	}

	return acc.Value, ref.Value, nil
}

func getGUID(r *http.Request) (string, error) {
	var GUID models.GUID

	err := json.NewDecoder(r.Body).Decode(&GUID)
	if err != nil {
		return "", err
	}

	if GUID.GUID == "" {
		return "", err
	}
	return GUID.GUID, nil
}

func setCookie(access, refresh string, w http.ResponseWriter) {
	cookieExpTime := time.Now().Add(time.Hour * service.RefreshExp)

	http.SetCookie(w, &http.Cookie{
		Name:     "Access",
		Value:    access,
		Expires:  cookieExpTime,
		HttpOnly: true,
		Path:     "/auth",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "Refresh",
		Value:    refresh,
		Expires:  cookieExpTime,
		HttpOnly: true,
		Path:     "/auth",
	})
}

func checkMethods(r *http.Request, allowedMethods []string) error {
	for _, method := range allowedMethods {
		if r.Method == method {
			return nil
		}
	}

	return fmt.Errorf("method %v not allowed, use:%v\n", r.Method, allowedMethods)
}
