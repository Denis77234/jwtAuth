package endpoint

import (
	"medosTest/pkg/jwt"
	"net/http"
	"time"
)

const accExp = 24 * 10

const refreshExp = 24 * 30

func (e *Endpoint) tokens(id string, expTime time.Time) (access string, refresh string) {
	payload := jwt.Payload{Sub: id, Iss: "medodsTest", Iat: time.Now().Unix(), Exp: expTime.Unix()}

	access = e.jwtG.Generate(payload)

	refresh = e.refH.Generate(access)

	return access, refresh
}

func (e Endpoint) tokensFromCookie(r *http.Request) (string, string, error) {
	acc, err := r.Cookie("Access")
	if err != nil {
		return "", "", err
	}
	ref, err := r.Cookie("Refresh")
	if err != nil {
		return "", "", err
	}

	return acc.Value, ref.Value, nil
}
