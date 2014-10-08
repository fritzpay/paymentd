package payment

import (
	"github.com/gorilla/mux"
	"net/http"
)

type API struct {
	router *mux.Router
}

func NewAPI(r *mux.Router) *API {
	a := &API{
		router: r,
	}
	r.HandleFunc("/payment", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("payment"))
	})
	return a
}
