package jwt

import (
	"encoding/base64"
	"encoding/json"
	"log"
)

type header struct {
	Alg string `json:"alg"` //value will be set automatically
	Typ string `json:"typ"`
}

func (h header) base64() string {

	bytes, err := json.Marshal(h)
	if err != nil {
		log.Println(err)
	}

	b64Hdr := base64.RawURLEncoding.EncodeToString(bytes)

	return b64Hdr
}
