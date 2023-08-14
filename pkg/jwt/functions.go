package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

func parse(jwt string) (header, payload, signature string) {
	slice := strings.Split(jwt, ".")
	header = slice[0]
	payload = slice[1]
	signature = slice[2]
	return header, payload, signature
}

func Signature(jwt string) (signature string) {
	_, _, signature = parse(jwt)
	return signature
}

func ParseToStruct(jwt string) (head header, payload Payload, err error) {
	headStr, payloadStr, _ := parse(jwt)

	byteH, err := base64.RawURLEncoding.DecodeString(headStr)
	if err != nil {
		return head, payload, err
	}

	byteP, err := base64.RawURLEncoding.DecodeString(payloadStr)
	if err != nil {
		return head, payload, err
	}

	err = json.Unmarshal(byteH, &head)
	if err != nil {
		return head, payload, err
	}
	err = json.Unmarshal(byteP, &payload)
	if err != nil {
		return head, payload, err
	}

	return head, payload, nil
}
