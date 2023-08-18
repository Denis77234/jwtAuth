package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func (h header) base64() (string, error) {
	bytes, err := json.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("b64 header marshall error: %w\n", err)
	}

	b64Hdr := base64.RawURLEncoding.EncodeToString(bytes)

	return b64Hdr, nil
}
