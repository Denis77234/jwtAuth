package jwt

import (
	"crypto/hmac"
	"errors"
	"hash"
	"strings"

	"encoding/base64"
	"fmt"
)

type Generator struct {
	algorithm string //name of algorithm for encoding
	alg       func() hash.Hash
	key       string
}

func NewGenerator(algorithm string, key string) (Generator, error) {

	algorithm = strings.ToUpper(algorithm)

	alg, ok := algomap[algorithm]
	if !ok {
		return Generator{}, errors.New("invalid algorithm name")
	}

	jg := Generator{algorithm: algorithm, alg: alg, key: key}

	return jg, nil
}

func (g *Generator) makeSignature(h string, p string) string {

	hashFunc := hmac.New(g.alg, []byte(g.key))

	str := fmt.Sprintf("%v.%v", h, p)

	hashFunc.Write([]byte(str))

	b64Sig := base64.RawURLEncoding.EncodeToString(hashFunc.Sum(nil))

	return b64Sig

}

func (g Generator) Generate(p Payload) string {

	h := header{Alg: g.algorithm, Typ: "jwt"}

	hdr := h.base64()

	payload := p.base64()

	sign := g.makeSignature(hdr, payload)

	jwt := fmt.Sprintf("%v.%v.%v", hdr, payload, sign)

	return jwt
}
