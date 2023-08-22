package model

type Token struct {
	GUID    string
	Iat     int64
	Refresh []byte
	ExpTime int64
}
