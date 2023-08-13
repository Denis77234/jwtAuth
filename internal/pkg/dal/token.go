package dal

import "time"

type Token struct {
	GUID     string
	Refresh  string
	InspTime int64
}

const insptime = int64(time.Second * 60 * 60 * 24 * 30)

func newToken(guid, refresh string) Token {
	inspTime := time.Now().Unix() + insptime

	t := Token{GUID: guid, Refresh: refresh, InspTime: inspTime}

	return t
}
