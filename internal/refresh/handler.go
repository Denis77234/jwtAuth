package refresh

import (
	"crypto/hmac"
	"encoding/base64"
	"hash"
)

type Handler struct {
	key string
	alg func() hash.Hash
}

func NewHandler(alg func() hash.Hash, key string) Handler {
	h := Handler{key: key, alg: alg}

	return h
}

func (h Handler) Generate(accessToken string) string {
	hashFunc := hmac.New(h.alg, []byte(h.key))

	hashFunc.Write([]byte(accessToken))

	return base64.RawURLEncoding.EncodeToString(hashFunc.Sum(nil))
}

func (h Handler) Validate(refresh, access string) bool {
	check := h.Generate(access)

	val := check == refresh

	return val
}
