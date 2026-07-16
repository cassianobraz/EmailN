package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type myHandler struct{}

func (m myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("myHandler"))
}

func main() {
	r := chi.NewRouter()

	m := myHandler{}
	r.Handle("/handler", m)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		product := r.URL.Query().Get("product")
		if product != "" {
			w.Write([]byte(product))
		} else {
			w.Write([]byte("teste"))
		}
	})

	http.ListenAndServe(":3000", r)
}
