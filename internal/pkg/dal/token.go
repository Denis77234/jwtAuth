package dal

type Token struct {
	GUID    string
	Refresh []byte
	ExpTime int64
}

func NewToken(guid string, refresh []byte, expTime int64) Token {

	t := Token{GUID: guid, Refresh: refresh, ExpTime: expTime}

	return t
}
