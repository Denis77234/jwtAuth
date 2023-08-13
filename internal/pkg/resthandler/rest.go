package resthandler

import (
	"log"
	"net/http"
	"strings"
)

type RestHandler struct {
	hf      http.HandlerFunc
	methods []string
}

func New(hf http.HandlerFunc) *RestHandler {
	return &RestHandler{hf: hf}
}

func (h *RestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for _, method := range h.methods {
		if r.Method != method {
			w.WriteHeader(405)
			_, err := w.Write([]byte("Wrong method"))
			if err != nil {
				log.Println(err)
			}
			return
		}
	}

	h.hf(w, r)
}

func (r *RestHandler) SetMethods(methods ...string) *RestHandler {
	for _, m := range methods {
		r.methods = append(r.methods, strings.ToUpper(m))
	}
	return r
}
