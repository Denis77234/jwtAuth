package jwt

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

var algomap = map[string]func() hash.Hash{
	"HS512": sha512.New,
	"HS256": sha256.New,
	"HS1":   sha1.New,
}
