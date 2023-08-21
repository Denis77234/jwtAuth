package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Payload struct {
	Iss string `json:"iss,omitempty"`
	Sub string `json:"sub,omitempty"`
	Aud string `json:"aud,omitempty"`
	Exp int64  `json:"exp,omitempty"`
	Nbf string `json:"nbf,omitempty"`
	Jti string `json:"jti,omitempty"`
	Iat int64  `json:"iat,omitempty"`
}

func (b Payload) base64() (string, error) {
	bytes, err := json.Marshal(b)
	if err != nil {
		return "", fmt.Errorf("marshal error: %w\n", err)
	}
	b64Body := base64.RawURLEncoding.EncodeToString(bytes)

	return b64Body, nil
}
