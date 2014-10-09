package payment

import (
	"github.com/gorilla/mux"
	"net/http"
)

// API represents the payment API in the version 1.x
type API struct {
	router *mux.Router
}

// NewAPI creates a new payment API
func NewAPI(r *mux.Router) *API {
	a := &API{
		router: r,
	}
	r.HandleFunc("/payment", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("payment"))
	})
	return a
}
