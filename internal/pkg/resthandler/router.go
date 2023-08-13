package resthandler

import (
	"log"
	"net/http"
)

type RestHandler struct {
	hf http.HandlerFunc
}

func (h RestHandler) setMethod(method string) {
	if h.hf == nil {
		panic("unable to assign handler to nil function")
	}

	NewHf := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(405)
			_, err := w.Write([]byte("Wrong method"))
			if err != nil {
				log.Println(err)

			}
			h.hf(w, r)
		}
	}

	h.hf = NewHf
}

func (h RestHandler) GET() {
	h.setMethod("GET")
}

func (h RestHandler) POST() {
	h.setMethod("POST")
}
