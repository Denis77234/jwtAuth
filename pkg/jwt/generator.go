package jwt

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"strings"
)

const (
	HS512 = "HS512"
	HS256 = "HS256"
	HS1   = "HS1"
)

var algomap = map[string]func() hash.Hash{
	HS512: sha512.New,
	HS256: sha256.New,
	HS1:   sha1.New,
}

type Generator struct {
	algorithm string //name of algorithm for encoding
	alg       func() hash.Hash
	key       string
}

func NewGenerator(algorithm string, key string) (Generator, error) {
	algorithm = strings.ToUpper(algorithm)

	alg, ok := algomap[algorithm]
	if !ok {
		return Generator{}, fmt.Errorf("invalid algorithm name: %v\n", algorithm)
	}

	jg := Generator{algorithm: algorithm, alg: alg, key: key}

	return jg, nil
}

func (g *Generator) makeSignature(h string, p string) string {
	hashFunc := hmac.New(g.alg, []byte(g.key))

	str := fmt.Sprintf("%s.%s", h, p)

	hashFunc.Write([]byte(str))

	b64Sig := base64.RawURLEncoding.EncodeToString(hashFunc.Sum(nil))

	return b64Sig

}

func (g Generator) Generate(p Payload) (string, error) {
	h := header{Alg: g.algorithm, Typ: "jwt"}

	hdr, err := h.base64()
	if err != nil {
		return "", fmt.Errorf("b64 header encoding error: %w", err)
	}

	payload, err := p.base64()
	if err != nil {
		return "", fmt.Errorf("b64 payload encoding error: %w", err)
	}

	sign := g.makeSignature(hdr, payload)

	jwt := fmt.Sprintf("%s.%s.%s", hdr, payload, sign)

	return jwt, nil
}
