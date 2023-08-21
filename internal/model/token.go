package model

type Token struct {
	GUID    string
	Refresh []byte
	ExpTime int64
}
