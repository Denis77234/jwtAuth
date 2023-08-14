package restDecorator

import (
	"log"
	"net/http"
	"strings"
)

type RestDecorator struct {
	hf      http.HandlerFunc
	methods []string
}

func New(hf http.HandlerFunc) *RestDecorator {
	return &RestDecorator{hf: hf}
}

func (d *RestDecorator) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for _, method := range d.methods {
		if r.Method == method {
			d.hf(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	_, err := w.Write([]byte("Wrong method"))
	if err != nil {
		log.Println(err)
	}

}

func (d *RestDecorator) SetMethods(methods ...string) *RestDecorator {
	for _, m := range methods {
		d.methods = append(d.methods, strings.ToUpper(m))
	}
	return d
}
